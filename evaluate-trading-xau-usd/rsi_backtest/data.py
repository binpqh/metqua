from __future__ import annotations

from pathlib import Path

import pandas as pd


TIMEFRAME_TO_FREQ = {
    "1m": "1min",
    "5m": "5min",
    "10m": "10min",
    "15m": "15min",
    "30m": "30min",
    "1h": "1h",
    "3h": "3h",
    "5h": "5h",
    "1d": "1d",
}


def load_candles(csv_path: str | Path) -> pd.DataFrame:
    frame = pd.read_csv(csv_path)
    required = {"timestamp", "open", "high", "low", "close"}
    missing = required - set(frame.columns)
    if missing:
        raise ValueError(f"Missing required columns: {sorted(missing)}")

    frame["timestamp"] = pd.to_datetime(frame["timestamp"], utc=True)
    frame = frame.sort_values("timestamp").drop_duplicates(subset=["timestamp"]).reset_index(drop=True)

    if "volume" not in frame.columns:
        frame["volume"] = 0.0

    numeric_cols = ["open", "high", "low", "close", "volume"]
    frame[numeric_cols] = frame[numeric_cols].astype(float)
    return frame


def resample_candles(frame: pd.DataFrame, timeframe: str) -> pd.DataFrame:
    if timeframe not in TIMEFRAME_TO_FREQ:
        raise ValueError(f"Unsupported timeframe: {timeframe}")

    freq = TIMEFRAME_TO_FREQ[timeframe]
    temp = frame.copy().set_index("timestamp")
    result = (
        temp.resample(freq, label="right", closed="right")
        .agg(
            {
                "open": "first",
                "high": "max",
                "low": "min",
                "close": "last",
                "volume": "sum",
            }
        )
        .dropna(subset=["open", "high", "low", "close"])
        .reset_index()
    )
    return result