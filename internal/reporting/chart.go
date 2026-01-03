package reporting

import (
	"os"
	"go-backtesting/internal/model"
	"html/template"
)

type Stats struct {
	TotalTrades   int
	WinRate       float64
	ProfitFactor  float64
	MaxDrawdown   float64
	TotalPnL      float64
}

func CalculateStats(trades []*model.Trade) Stats {
	if len(trades) == 0 {
		return Stats{}
	}

	wins := 0
	grossProfit := 0.0
	grossLoss := 0.0
	totalPnL := 0.0

	peak := 0.0
	drawdown := 0.0
	runningPnL := 0.0

	for _, t := range trades {
		if !t.Active {
			totalPnL += t.PnL
			runningPnL += t.PnL

			if t.PnL > 0 {
				wins++
				grossProfit += t.PnL
			} else {
				grossLoss += -t.PnL // positive value
			}

			// MDD
			if runningPnL > peak {
				peak = runningPnL
			}
			dd := peak - runningPnL
			if dd > drawdown {
				drawdown = dd
			}
		}
	}

	winRate := 0.0
	if len(trades) > 0 {
		winRate = (float64(wins) / float64(len(trades))) * 100.0
	}

	pf := 0.0
	if grossLoss > 0 {
		pf = grossProfit / grossLoss
	} else if grossProfit > 0 {
		pf = 999.0 // Infinite
	}

	return Stats{
		TotalTrades:  len(trades),
		WinRate:      winRate,
		ProfitFactor: pf,
		MaxDrawdown:  drawdown,
		TotalPnL:     totalPnL,
	}
}

// GenerateHTMLChart creates a simple HTML file with TradingView Lightweight Charts
func GenerateHTMLChart(candles []model.Candle, trades []*model.Trade, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl := `<!DOCTYPE html>
<html>
<head>
	<title>Backtest Results</title>
	<script src="https://unpkg.com/lightweight-charts/dist/lightweight-charts.standalone.production.js"></script>
	<style>
		body { padding: 0; margin: 0; }
		#chart { position: absolute; width: 100%; height: 100%; }
	</style>
</head>
<body>
	<div id="chart"></div>
	<script>
		const chart = LightweightCharts.createChart(document.getElementById('chart'), {
			width: window.innerWidth,
			height: window.innerHeight,
			layout: { backgroundColor: '#ffffff', textColor: '#333' },
			grid: { vertLines: { color: '#eee' }, horzLines: { color: '#eee' } },
		});

		const candleSeries = chart.addCandlestickSeries();
		const data = [
			{{range .Candles}}
			{ time: {{.Timestamp}}/1000, open: {{.Open}}, high: {{.High}}, low: {{.Low}}, close: {{.Close}} },
			{{end}}
		];
		candleSeries.setData(data);

		const markers = [
			{{range .Trades}}
			{ time: {{.EntryTime}}/1000, position: '{{if eq .Side "Long"}}belowBar{{else}}aboveBar{{end}}', color: '{{if eq .Side "Long"}}#2196F3{{else}}#E91E63{{end}}', shape: '{{if eq .Side "Long"}}arrowUp{{else}}arrowDown{{end}}', text: '{{.Side}} {{.EntryPattern}}' },
			{ time: {{.ExitTime}}/1000, position: '{{if eq .Side "Long"}}aboveBar{{else}}belowBar{{end}}', color: '{{if gt .PnL 0.0}}#4CAF50{{else}}#F44336{{end}}', shape: 'circle', text: 'Exit {{.Reason}}' },
			{{end}}
		];
		// Sort markers by time
		markers.sort((a, b) => a.time - b.time);
		candleSeries.setMarkers(markers);
	</script>
</body>
</html>`

	t := template.Must(template.New("chart").Parse(tmpl))
	return t.Execute(f, struct {
		Candles []model.Candle
		Trades  []*model.Trade
	}{
		Candles: candles,
		Trades:  trades,
	})
}
