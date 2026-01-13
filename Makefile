# Makefile for POC Strategy

.PHONY: all deps build download backtest clean

all: deps build

deps:
	go mod tidy

build: deps
	mkdir -p bin
	go build -o bin/downloader cmd/downloader/main.go
	go build -o bin/backtester cmd/backtester/main.go

download:
	go run cmd/downloader/main.go

backtest:
	go run cmd/backtester/main.go

clean:
	rm -rf bin
	rm -f btc_1h.csv
	rm -f trades.csv
	rm -f chart.html
