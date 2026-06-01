from __future__ import annotations

import argparse
import json
from pathlib import Path

from rsi_backtest.engine import run_batch
from rsi_backtest.strategy import Scenario


def load_scenarios(path: str | Path) -> list[Scenario]:
    with open(path, "r", encoding="utf-8") as handle:
        payload = json.load(handle)

    scenarios: list[Scenario] = []
    for item in payload.get("scenarios", []):
        scenarios.append(
            Scenario(
                name=item["name"],
                rsi_condition=item["rsi_condition"],
                command=item["command"],
                stop_loss_pct=float(item["stop_loss_pct"]),
                take_profit_pct=float(item["take_profit_pct"]),
            )
        )
    return scenarios


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="RSI XAUUSD backtesting CLI")
    parser.add_argument("--data", required=True, help="Path to source candle CSV")
    parser.add_argument("--scenario-file", required=True, help="Path to scenarios JSON")
    parser.add_argument("--timeframe", default="1m", help="1m,5m,10m,15m,30m,1h,3h,5h,1d")
    parser.add_argument("--output-dir", default="results", help="Output directory for reports")
    parser.add_argument("--tie-breaker", default="sl_first", choices=["sl_first", "tp_first", "nearest"])
    parser.add_argument("--rsi-period", default=14, type=int)
    parser.add_argument("--fetch-yahoo", action="store_true", help="Fetch recent data from Yahoo Finance and save to --data path")
    parser.add_argument("--yahoo-symbol", default="GC=F", help="Yahoo ticker symbol for gold (default GC=F)")
    parser.add_argument("--yahoo-interval", default="1m", help="Interval used for yahoo fetch (1m,5m,15m etc)")
    parser.add_argument("--yahoo-period", default="7d", help="Period for yahoo fetch (7d,30d,1y etc)")
    parser.add_argument("--yahoo-start", default=None, help="Start date for yahoo fetch (YYYY-MM-DD)")
    parser.add_argument("--yahoo-end", default=None, help="End date for yahoo fetch (YYYY-MM-DD)")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    if args.fetch_yahoo:
        from rsi_backtest.fetchers import fetch_yahoo_gc, save_to_csv

        print(f"Fetching {args.yahoo_symbol} {args.yahoo_interval} {args.yahoo_period} from Yahoo...")
        if args.yahoo_start and args.yahoo_end:
            from rsi_backtest.fetchers import fetch_yahoo_range

            print(f"Using date range {args.yahoo_start} to {args.yahoo_end}")
            df = fetch_yahoo_range(symbol=args.yahoo_symbol, interval=args.yahoo_interval, start=args.yahoo_start, end=args.yahoo_end)
        else:
            df = fetch_yahoo_gc(symbol=args.yahoo_symbol, interval=args.yahoo_interval, period=args.yahoo_period)
        save_to_csv(df, args.data)
        print(f"Saved fetched data to {args.data}")

    scenarios = load_scenarios(args.scenario_file)
    trades, summary = run_batch(
        data_path=args.data,
        timeframe=args.timeframe,
        scenarios=scenarios,
        output_dir=args.output_dir,
        tie_breaker=args.tie_breaker,
        rsi_period=args.rsi_period,
    )

    print(f"Completed backtest: {len(summary)} scenarios, {len(trades)} trades")
    print(f"Reports written to: {args.output_dir}")


if __name__ == "__main__":
    main()