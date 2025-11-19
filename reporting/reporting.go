package reporting

import (
	"fmt"
	"go-backtesting/market"
	"go-backtesting/strategy"
	"math"
	"os"
	"strings"
	"text/template"
	"time"
)

// PrintTradeAnalysis prints a detailed analysis of the trades.
func PrintTradeAnalysis(result strategy.BacktestResult, strategyData *strategy.StrategyDataContext) {
	fmt.Printf("\n--- Trade Entry Analysis ---\n")
	fmt.Println("Signals that resulted in a trade:")
	fmt.Println("----------------------------------------------------------------------------------------------------------------------------------")
	fmt.Printf("%-5s %-5s %-20s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n", "Idx", "Type", "Timestamp", "ZScore", "VWZScore", "BBW", "ADX", "Volume", "PlusDI", "MinusDI", "DX", "PnL", "PnL(%)")
	fmt.Println("----------------------------------------------------------------------------------------------------------------------------------")

	for i, trade := range result.Trades {
		indicators := trade.EntryIndicators
		zStr := "NaN"
		if !math.IsNaN(indicators.VWZScore[len(indicators.VWZScore)-1]) {
			zStr = fmt.Sprintf("%.4f", indicators.ZScore)
		}

		vwzStr := "NaN"
		if !math.IsNaN(indicators.ZScore[len(indicators.ZScore)-1]) {
			vwzStr = fmt.Sprintf("%.4f", indicators.VWZScore)
		}

		entryIndex := -1
		for j, c := range strategyData.Candles {
			if c.Time.Equal(trade.EntryTime) {
				entryIndex = j
				break
			}
		}

		bbwStr := "NaN"
		if entryIndex != -1 && entryIndex < len(strategyData.Bbw) && !math.IsNaN(strategyData.Bbw[entryIndex]) {
			bbwStr = fmt.Sprintf("%.4f", strategyData.Bbw[entryIndex])
		}
		dxStr := "NaN"
		if entryIndex != -1 && entryIndex < len(strategyData.DX) && !math.IsNaN(strategyData.DX[entryIndex]) {
			dxStr = fmt.Sprintf("%.2f", strategyData.DX[entryIndex])
		}

		volStr := "NaN"
		if entryIndex != -1 {
			volStr = fmt.Sprintf("%.2f", strategyData.Candles[entryIndex].Vol)
		}

		pnlPctStr := fmt.Sprintf("%.2f%%", trade.PnlPercentage)

		fmt.Printf("%-5d %-5s %-20s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10s %-10.2f %-10s\n",
			i,
			trade.Direction,
			trade.EntryTime.Format("01-02 15:04"),
			zStr,
			vwzStr,
			bbwStr,
			fmt.Sprintf("%.2f", indicators.ADX),
			volStr,
			fmt.Sprintf("%.2f", indicators.PlusDI),
			fmt.Sprintf("%.2f", indicators.MinusDI),
			dxStr,
			trade.Pnl,
			pnlPctStr,
		)
	}
	fmt.Println("----------------------------------------------------------------------------------------------------------------------------------")
}

// ChartData contains the data for the HTML chart.
type ChartData struct {
	CandleData   string
	ZData        string
	VWZData      string
	EntrySignals string
	VolumeData   string
}

// GenerateHTMLChart generates an HTML chart of the backtest results.
func GenerateHTMLChart(candles market.CandleSticks, zScores []float64, vwzScores []float64, entrySignals []strategy.EntrySignal) {
	var candleData []string
	var zData []string
	var vwzData []string

	for i, c := range candles {
		ms := c.Time.UnixNano() / int64(time.Millisecond)
		candlePoint := fmt.Sprintf("{x: %d, o: %.4f, h: %.4f, l: %.4f, c: %.4f}", ms, c.Open, c.High, c.Low, c.Close)
		candleData = append(candleData, candlePoint)

		if math.IsNaN(zScores[i]) {
			zData = append(zData, fmt.Sprintf("{x: %d, y: null}", ms))
		} else {
			zData = append(zData, fmt.Sprintf("{x: %d, y: %.4f}", ms, zScores[i]))
		}
		if math.IsNaN(vwzScores[i]) {
			vwzData = append(vwzData, fmt.Sprintf("{x: %d, y: null}", ms))
		} else {
			vwzData = append(vwzData, fmt.Sprintf("{x: %d, y: %.4f}", ms, vwzScores[i]))
		}
	}

	candleDataJS := "[" + strings.Join(candleData, ",") + "]"
	zDataJS := "[" + strings.Join(zData, ",") + "]"
	vwzDataJS := "[" + strings.Join(vwzData, ",") + "]"

	var entrySignalData []string
	for _, s := range entrySignals {
		ms := s.Time.UnixNano() / int64(time.Millisecond)
		signalPoint := fmt.Sprintf("{x: %d, y: %.4f, direction: '%s'}", ms, s.Price, s.Direction)
		entrySignalData = append(entrySignalData, signalPoint)
	}
	entrySignalsJS := "[" + strings.Join(entrySignalData, ",") + "]"

	var volumeData []string
	for _, c := range candles {
		ms := c.Time.UnixNano() / int64(time.Millisecond)
		volumePoint := fmt.Sprintf("{x: %d, y: %.4f}", ms, c.Vol)
		volumeData = append(volumeData, volumePoint)
	}
	volumeDataJS := "[" + strings.Join(volumeData, ",") + "]"

	tmpl, err := template.ParseFiles("chart.html.template")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	data := ChartData{
		CandleData:   candleDataJS,
		ZData:        zDataJS,
		VWZData:      vwzDataJS,
		EntrySignals: entrySignalsJS,
		VolumeData:   volumeDataJS,
	}

	file, err := os.Create("chart.html")
	if err != nil {
		fmt.Println("Error creating chart.html:", err)
		return
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}
	fmt.Println("Generated chart.html with zoom & sync (scroll or drag to zoom)")
}

// PrintDetailedTradeRecords prints the detailed trade records.
func PrintDetailedTradeRecords(result strategy.BacktestResult) {
	fmt.Printf("\n--- Detailed Trade Records ---\n")
	fmt.Println("-----------------------------------------------------------------------------------------------------------------------------------------")
	fmt.Printf("%-5s %-5s %-20s %-15s %-20s %-15s %-10s %-10s %-10s\n",
		"Idx", "Type", "Entry Time", "Entry Price", "Exit Time", "Exit Price", "Pnl", "Pnl(%)", "Status")
	fmt.Println("-----------------------------------------------------------------------------------------------------------------------------------------")

	for i, trade := range result.Trades {
		status := "Loss"
		if trade.Pnl > 0 {
			status = "Win"
		}

		fmt.Printf("%-5d %-5s %-20s %-15.2f %-20s %-15.2f %-10.2f %-9.2f%% %-10s\n",
			i,
			trade.Direction,
			trade.EntryTime.Format("01-02 15:04:05"),
			trade.EntryPrice,
			trade.ExitTime.Format("01-02 15:04:05"),
			trade.ExitPrice,
			trade.Pnl,
			trade.PnlPercentage,
			status,
		)
	}
	fmt.Println("-----------------------------------------------------------------------------------------------------------------------------------------")
}

// PrintBacktestSummary prints the backtest summary.
func PrintBacktestSummary(result strategy.BacktestResult) {
	fmt.Printf("\n--- Backtest Summary ---\n")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("Total Trades: %d\n", result.TotalTrades)
	fmt.Printf("Win Rate: %.2f%%\n", result.WinRate)
	fmt.Printf("Wins: %d\n", result.WinCount)
	fmt.Printf("Losses: %d\n", result.LossCount)
	fmt.Printf("Total PnL: %.2f\n", result.TotalPnl) // PnL is in price points, not currency
	fmt.Println("-----------------------------------------------------------------")
}
