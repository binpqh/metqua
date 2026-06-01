from __future__ import annotations

from pathlib import Path
import csv

from rsi_backtest.strategy import Scenario
from rsi_backtest.engine import run_batch


def run_grid(data_path: str, out_csv: str):
    buy_thresholds = ["<5", "<10", "<12", "<15", "<18", "<20"]
    sell_thresholds = [">70", ">75", ">80", ">85", ">90"]

    rows = []
    for t in buy_thresholds:
        s = Scenario(name=f"buy_{t}_sl15_tp15", rsi_condition=t, command="BUY", stop_loss_abs=15.0, take_profit_abs=15.0)
        trades, summary = run_batch(data_path=data_path, timeframe="5m", scenarios=[s], output_dir=Path("results_grid")/s.name)
        rec = summary.iloc[0].to_dict()
        rec.update({"side": "BUY", "rsi": t})
        rows.append(rec)

    for t in sell_thresholds:
        s = Scenario(name=f"sell_{t}_sl15_tp15", rsi_condition=t, command="SELL", stop_loss_abs=15.0, take_profit_abs=15.0)
        trades, summary = run_batch(data_path=data_path, timeframe="5m", scenarios=[s], output_dir=Path("results_grid")/s.name)
        rec = summary.iloc[0].to_dict()
        rec.update({"side": "SELL", "rsi": t})
        rows.append(rec)

    # write aggregate CSV
    keys = ["side", "rsi", "trades", "wins", "losses", "win_rate", "total_pnl_abs", "profit_factor"]
    with open(out_csv, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=keys)
        writer.writeheader()
        for r in rows:
            writer.writerow({k: r.get(k, None) for k in keys})

    return out_csv


if __name__ == "__main__":
    data = "data/fetched_5m_2026-04-16_2026-06-01.csv"
    out = "results_grid/grid_summary.csv"
    Path("results_grid").mkdir(exist_ok=True)
    path = run_grid(data, out)
    print("Grid search completed. Summary:")
    import pandas as pd

    df = pd.read_csv(path)
    df_sorted = df[df['trades']>=10].sort_values(['win_rate','trades'], ascending=[False, False])
    print(df_sorted.to_string(index=False))
