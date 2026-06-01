from __future__ import annotations

import argparse
from pathlib import Path

import pandas as pd

from rsi_backtest.fetchers import fetch_yahoo_range, save_to_csv
from rsi_backtest.strategy import Scenario
from rsi_backtest.engine import run_batch


def parse_args():
    p = argparse.ArgumentParser()
    p.add_argument("--symbol", default="GC=F")
    p.add_argument("--interval", default="5m")
    p.add_argument("--start", required=True)
    p.add_argument("--end", required=True)
    p.add_argument("--out-dir", default="results_experiments")
    return p.parse_args()


def main():
    args = parse_args()
    data_path = Path("data") / f"fetched_{args.interval}_{args.start}_{args.end}.csv"
    print(f"Fetching {args.symbol} {args.interval} {args.start}->{args.end}...")
    df = fetch_yahoo_range(symbol=args.symbol, interval=args.interval, start=args.start, end=args.end)
    save_to_csv(df, data_path)
    print(f"Saved {len(df)} candles to {data_path}")

    sl_gaps = [10.0, 15.0, 20.0]
    tp_gaps = [15.0, 20.0, 25.0]

    scenarios = []
    for sl in sl_gaps:
        for tp in tp_gaps:
            name = f"sell_rsi_lt20_sl{int(sl)}_tp{int(tp)}"
            scenarios.append(
                Scenario(
                    name=name,
                    rsi_condition="<20",
                    command="SELL",
                    stop_loss_abs=sl,
                    take_profit_abs=tp,
                )
            )

    out = Path(args.out_dir)
    out.mkdir(parents=True, exist_ok=True)

    trades_df, summary_df = run_batch(
        data_path=data_path,
        timeframe=args.interval,
        scenarios=scenarios,
        output_dir=out,
    )

    print(summary_df.to_string(index=False))


if __name__ == "__main__":
    main()
