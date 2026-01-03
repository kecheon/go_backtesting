# POC Strategy Backtester

This project implements a Volume Profile + Candle Pattern strategy backtester in Go.

## Prerequisites

- Go 1.25+ installed

## Quick Start

1.  **Download Data**
    Fetch the last 30 days of 1H BTC-USDT-SWAP data from OKX:
    ```bash
    make download
    ```
    This creates `btc_1h.csv`.

2.  **Run Backtest**
    Execute the strategy backtest:
    ```bash
    make backtest
    ```
    This will:
    - Load the CSV data.
    - Run the simulation.
    - Print statistics to the console.
    - Generate `trades.csv` (log) and `chart.html` (visualization).

3.  **View Results**
    Open `chart.html` in your web browser to view the interactive chart with trade markers.

## Build Binaries

To build standalone executables:
```bash
make build
```
The binaries will be placed in the `bin/` directory:
- `bin/downloader`
- `bin/backtester`
