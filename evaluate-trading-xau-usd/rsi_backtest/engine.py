from __future__ import annotations

from dataclasses import asdict, dataclass
from pathlib import Path

import pandas as pd

from .data import load_candles, resample_candles
from .indicators import rsi_wilder
from .strategy import Scenario, condition_match


@dataclass
class Trade:
    scenario: str
    side: str
    entry_time: pd.Timestamp
    exit_time: pd.Timestamp
    entry_price: float
    exit_price: float
    stop_loss_price: float
    take_profit_price: float
    exit_reason: str
    pnl_abs: float
    pnl_pct: float
    entry_rsi: float | None = None
    exit_rsi: float | None = None


class BacktestEngine:
    def __init__(self, tie_breaker: str = "sl_first", rsi_period: int = 14):
        if tie_breaker not in {"sl_first", "tp_first", "nearest"}:
            raise ValueError("tie_breaker must be one of sl_first, tp_first, nearest")
        self.tie_breaker = tie_breaker
        self.rsi_period = rsi_period

    def run(self, candles: pd.DataFrame, scenario: Scenario) -> list[Trade]:
        frame = candles.copy()
        frame["rsi"] = rsi_wilder(frame["close"], self.rsi_period)

        side = scenario.command.upper()
        if side not in {"BUY", "SELL"}:
            raise ValueError("command must be BUY or SELL")

        trades: list[Trade] = []
        open_position: dict | None = None

        for index in range(1, len(frame)):
            row = frame.iloc[index]

            if open_position is not None:
                exit_price, reason = self._check_exit(
                    side=open_position["side"],
                    low=row["low"],
                    high=row["high"],
                    stop_price=open_position["stop_loss_price"],
                    take_price=open_position["take_profit_price"],
                    entry_price=open_position["entry_price"],
                )
                if exit_price is not None:
                    pnl_abs = (exit_price - open_position["entry_price"]) if side == "BUY" else (open_position["entry_price"] - exit_price)
                    pnl_pct = pnl_abs / open_position["entry_price"]
                    trades.append(
                        Trade(
                            scenario=scenario.name,
                            side=side,
                            entry_time=open_position["entry_time"],
                            exit_time=row["timestamp"],
                            entry_price=open_position["entry_price"],
                            exit_price=exit_price,
                            stop_loss_price=open_position["stop_loss_price"],
                            take_profit_price=open_position["take_profit_price"],
                            exit_reason=reason,
                            pnl_abs=float(pnl_abs),
                            pnl_pct=float(pnl_pct),
                            entry_rsi=open_position.get("entry_rsi"),
                            exit_rsi=float(row.get("rsi", None)),
                        )
                    )
                    open_position = None

            if open_position is None and condition_match(float(row["rsi"]), scenario.rsi_condition):
                entry_price = float(row["close"])
                entry_rsi = float(row["rsi"])
                # support absolute SL/TP if provided, else use percent
                if scenario.stop_loss_abs is not None and scenario.take_profit_abs is not None:
                    if side == "BUY":
                        stop_loss_price = entry_price - scenario.stop_loss_abs
                        take_profit_price = entry_price + scenario.take_profit_abs
                    else:
                        stop_loss_price = entry_price + scenario.stop_loss_abs
                        take_profit_price = entry_price - scenario.take_profit_abs
                else:
                    if scenario.stop_loss_pct is None or scenario.take_profit_pct is None:
                        raise ValueError("Scenario must provide either abs or pct SL/TP")
                    if side == "BUY":
                        stop_loss_price = entry_price * (1 - scenario.stop_loss_pct)
                        take_profit_price = entry_price * (1 + scenario.take_profit_pct)
                    else:
                        stop_loss_price = entry_price * (1 + scenario.stop_loss_pct)
                        take_profit_price = entry_price * (1 - scenario.take_profit_pct)

                open_position = {
                    "side": side,
                    "entry_time": row["timestamp"],
                    "entry_price": entry_price,
                    "stop_loss_price": float(stop_loss_price),
                    "take_profit_price": float(take_profit_price),
                    "entry_rsi": entry_rsi,
                }

        return trades

    def _check_exit(
        self,
        side: str,
        low: float,
        high: float,
        stop_price: float,
        take_price: float,
        entry_price: float,
    ) -> tuple[float | None, str]:
        if side == "BUY":
            hit_sl = low <= stop_price
            hit_tp = high >= take_price
        else:
            hit_sl = high >= stop_price
            hit_tp = low <= take_price

        if not hit_sl and not hit_tp:
            return None, ""
        if hit_sl and not hit_tp:
            return stop_price, "stop_loss"
        if hit_tp and not hit_sl:
            return take_price, "take_profit"

        if self.tie_breaker == "sl_first":
            return stop_price, "stop_loss_same_bar"
        if self.tie_breaker == "tp_first":
            return take_price, "take_profit_same_bar"

        distance_sl = abs(entry_price - stop_price)
        distance_tp = abs(take_price - entry_price)
        if distance_sl <= distance_tp:
            return stop_price, "stop_loss_same_bar_nearest"
        return take_price, "take_profit_same_bar_nearest"


def summarize_trades(trades: list[Trade]) -> dict:
    if not trades:
        return {
            "trades": 0,
            "wins": 0,
            "losses": 0,
            "win_rate": 0.0,
            "total_pnl_abs": 0.0,
            "total_pnl_pct": 0.0,
            "profit_factor": 0.0,
            "avg_pnl_pct": 0.0,
        }

    profits = [trade.pnl_abs for trade in trades if trade.pnl_abs > 0]
    losses = [abs(trade.pnl_abs) for trade in trades if trade.pnl_abs < 0]
    total_pnl_abs = sum(trade.pnl_abs for trade in trades)
    total_pnl_pct = sum(trade.pnl_pct for trade in trades)
    wins = len(profits)
    losses_count = len(losses)

    return {
        "trades": len(trades),
        "wins": wins,
        "losses": losses_count,
        "win_rate": wins / len(trades),
        "total_pnl_abs": total_pnl_abs,
        "total_pnl_pct": total_pnl_pct,
        "profit_factor": (sum(profits) / sum(losses)) if losses else float("inf"),
        "avg_pnl_pct": total_pnl_pct / len(trades),
    }


def run_batch(
    data_path: str | Path,
    timeframe: str,
    scenarios: list[Scenario],
    output_dir: str | Path,
    tie_breaker: str = "sl_first",
    rsi_period: int = 14,
) -> tuple[pd.DataFrame, pd.DataFrame]:
    candles = load_candles(data_path)
    if timeframe != "1m":
        candles = resample_candles(candles, timeframe)

    engine = BacktestEngine(tie_breaker=tie_breaker, rsi_period=rsi_period)
    output = Path(output_dir)
    output.mkdir(parents=True, exist_ok=True)

    all_trades: list[Trade] = []
    summaries: list[dict] = []

    for scenario in scenarios:
        trades = engine.run(candles, scenario)
        all_trades.extend(trades)
        summary = summarize_trades(trades)
        summary.update({"scenario": scenario.name, "timeframe": timeframe})
        summaries.append(summary)

    trades_df = pd.DataFrame([asdict(trade) for trade in all_trades])
    summary_df = pd.DataFrame(summaries)

    trades_path = output / "trades.csv"
    summary_path = output / "summary.csv"
    summary_json_path = output / "summary.json"

    if not trades_df.empty:
        trades_df.to_csv(trades_path, index=False)
    else:
        pd.DataFrame(columns=[field for field in Trade.__dataclass_fields__]).to_csv(trades_path, index=False)

    summary_df.to_csv(summary_path, index=False)
    summary_df.to_json(summary_json_path, orient="records", indent=2)

    return trades_df, summary_df