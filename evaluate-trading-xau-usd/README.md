# XAUUSD RSI Backtest (MVP)

Ứng dụng CLI backtest use case RSI cho XAUUSD với các input:
- `RSI condition`: `<20`, `>70`, `15-20`, `10-15`
- `command`: `BUY` hoặc `SELL`
- `stop_loss_pct`, `take_profit_pct`
- timeframe: `1m`, `5m`, `10m`, `15m`, `30m`, `1h`, `3h`, `5h`, `1d`

## Cấu trúc chính
- `main.py`: CLI entrypoint
- `rsi_backtest/data.py`: load + resample dữ liệu nến
- `rsi_backtest/indicators.py`: RSI Wilder
- `rsi_backtest/engine.py`: engine backtest, intrabar SL/TP
- `scenarios.sample.json`: mẫu các use case
- `data/sample_xauusd_1m.csv`: dữ liệu mẫu để chạy nhanh
- `tests/test_backtest.py`: unit tests

## Cài đặt
```bash
cd /home/pqhung1/metqua/evaluate-trading-xau-usd
pip install -r requirements.txt
```

## Chạy backtest
```bash
cd /home/pqhung1/metqua/evaluate-trading-xau-usd
python main.py \
  --data data/sample_xauusd_1m.csv \
  --scenario-file scenarios.sample.json \
  --timeframe 1m \
  --output-dir results \
  --tie-breaker sl_first
```

Kết quả nằm trong thư mục `results/`:
- `trades.csv`
- `summary.csv`
- `summary.json`

## Chạy tests
```bash
cd /home/pqhung1/metqua/evaluate-trading-xau-usd
pytest -q
```

## Gợi ý data source
- Free-first: Dukascopy (primary), Yahoo Finance (fallback)
- Có thể nâng cấp paid API sau qua adapter provider (OANDA/Twelve Data)