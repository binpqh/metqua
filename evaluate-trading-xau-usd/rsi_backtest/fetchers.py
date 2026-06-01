from __future__ import annotations

from pathlib import Path
import logging

import pandas as pd

logger = logging.getLogger(__name__)


def fetch_yahoo_gc(symbol: str = "GC=F", interval: str = "1m", period: str = "7d") -> pd.DataFrame:
    try:
        import yfinance as yf
    except Exception as exc:  # pragma: no cover - optional dependency
        raise RuntimeError("yfinance is required for Yahoo fetcher") from exc

    df = yf.download(tickers=symbol, period=period, interval=interval, progress=False)
    if df.empty:
        raise RuntimeError("Yahoo returned empty dataframe")

    # Standardize columns
    df = df.reset_index()

    # find candidate column names robustly
    cols = {c.lower(): c for c in df.columns}

    def pick(*names):
        for n in names:
            if n.lower() in cols:
                return cols[n.lower()]
        return None

    ts_col = pick("Datetime", "Date", "timestamp") or df.columns[0]
    open_col = pick("Open", "open", "o")
    high_col = pick("High", "high", "h")
    low_col = pick("Low", "low", "l")
    close_col = pick("Close", "close", "c")
    vol_col = pick("Volume", "volume", "v")

    if not all([open_col, high_col, low_col, close_col]):
        raise RuntimeError(f"Required OHLC columns not found in Yahoo df; columns={list(df.columns)}")

    df["timestamp"] = pd.to_datetime(df[ts_col], utc=True)
    out = df[["timestamp", open_col, high_col, low_col, close_col, vol_col if vol_col else close_col]].copy()
    out.columns = ["timestamp", "open", "high", "low", "close", "volume"]
    out = out.dropna()
    return out


def fetch_yahoo_range(symbol: str, interval: str, start: str, end: str) -> pd.DataFrame:
    try:
        import yfinance as yf
    except Exception as exc:  # pragma: no cover - optional dependency
        raise RuntimeError("yfinance is required for Yahoo fetcher") from exc

    df = yf.download(tickers=symbol, start=start, end=end, interval=interval, progress=False)
    if df.empty:
        raise RuntimeError("Yahoo returned empty dataframe for given range")

    # flatten MultiIndex columns when ticker included
    if isinstance(df.columns, pd.MultiIndex):
        fields = {"Open", "High", "Low", "Close", "Volume"}
        new_cols = []
        for c in df.columns:
            if isinstance(c, tuple):
                pick = None
                for part in c:
                    if isinstance(part, str) and part in fields:
                        pick = part
                        break
                new_cols.append(pick if pick is not None else c[-1])
            else:
                new_cols.append(c)
        df.columns = new_cols

    df = df.reset_index()

    # find candidate column names robustly
    cols = {c.lower(): c for c in df.columns}

    def pick(*names):
        for n in names:
            if n.lower() in cols:
                return cols[n.lower()]
        return None

    ts_col = pick("Datetime", "Date", "timestamp") or df.columns[0]
    open_col = pick("Open", "open", "o")
    high_col = pick("High", "high", "h")
    low_col = pick("Low", "low", "l")
    close_col = pick("Close", "close", "c")
    vol_col = pick("Volume", "volume", "v")

    if not all([open_col, high_col, low_col, close_col]):
        raise RuntimeError(f"Required OHLC columns not found in Yahoo df; columns={list(df.columns)}")

    df["timestamp"] = pd.to_datetime(df[ts_col], utc=True)
    out = df[["timestamp", open_col, high_col, low_col, close_col, vol_col if vol_col else close_col]].copy()
    out.columns = ["timestamp", "open", "high", "low", "close", "volume"]
    out = out.dropna()
    return out


def save_to_csv(df: pd.DataFrame, path: str | Path) -> None:
    p = Path(path)
    p.parent.mkdir(parents=True, exist_ok=True)
    df.to_csv(p, index=False)


__all__ = ["fetch_yahoo_gc", "save_to_csv"]
