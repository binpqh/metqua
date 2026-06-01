from __future__ import annotations

import pandas as pd

from rsi_backtest.data import resample_candles
from rsi_backtest.engine import BacktestEngine
from rsi_backtest.strategy import Scenario, condition_match


def test_condition_parser():
    assert condition_match(19.9, "<20")
    assert condition_match(70.1, ">70")
    assert condition_match(17.0, "15-20")
    assert not condition_match(21.0, "15-20")


def test_resample_to_5m():
    frame = pd.DataFrame(
        {
            "timestamp": pd.date_range("2026-01-01", periods=10, freq="1min", tz="UTC"),
            "open": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
            "high": [1.1, 2.1, 3.1, 4.1, 5.1, 6.1, 7.1, 8.1, 9.1, 10.1],
            "low": [0.9, 1.9, 2.9, 3.9, 4.9, 5.9, 6.9, 7.9, 8.9, 9.9],
            "close": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
            "volume": [1] * 10,
        }
    )
    result = resample_candles(frame, "5m")
    assert len(result) >= 2


def test_intrabar_sl_first_exit():
    candles = pd.DataFrame(
        {
            "timestamp": pd.date_range("2026-01-01", periods=20, freq="1min", tz="UTC"),
            "open": [100.0] * 20,
            "high": [100.5] * 20,
            "low": [99.5] * 20,
            "close": [100.0] * 20,
            "volume": [100] * 20,
        }
    )
    scenario = Scenario(
        name="always_buy",
        rsi_condition="15-70",
        command="BUY",
        stop_loss_pct=0.002,
        take_profit_pct=0.002,
    )

    engine = BacktestEngine(tie_breaker="sl_first", rsi_period=2)
    trades = engine.run(candles, scenario)

    assert len(trades) > 0
    assert trades[0].exit_reason.startswith("stop_loss")