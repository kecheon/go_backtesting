package strategy

import (
	"fmt"
	"go-backtesting/config"
	"go-backtesting/market"

	"github.com/markcheno/go-talib"
)

// initializeStrategyDataContext initializes the strategy data context.
func InitializeStrategyDataContext(config *config.Config) (*StrategyDataContext, error) {
	// 1. Read Candles from CSV
	candles, err := market.ReadCandlesFromCSV(config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read candle data: %w", err)
	}

	if len(candles) == 0 {
		return &StrategyDataContext{Candles: candles}, nil // Return empty context if no candles
	}

	// 2. Prepare data for TALib
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}

	// 3. Calculate all indicator series
	emaShort := talib.Ema(closes, config.EmaPeriod)
	emaLong := talib.Ema(closes, config.EmaPeriod*10)
	zScores := ZScores(candles, config.VWZPeriod)
	vwzScores := VWZScores(candles, config.VWZPeriod, config.VWZScore.MinStdDev)
	bbw, _, _, _ := BBW(candles, 20, 2.0)
	bbwzScores := NormalizeBBW(bbw, 50)

	adxSeries := talib.Adx(highs, lows, closes, config.ADXPeriod)
	plusDI := talib.PlusDI(highs, lows, closes, config.ADXPeriod)
	minusDI := talib.MinusDI(highs, lows, closes, config.ADXPeriod)
	dx := talib.Dx(highs, lows, closes, config.ADXPeriod)

	// 4. Create and return the context
	return &StrategyDataContext{
		Candles:    candles,
		EmaShort:   emaShort,
		EmaLong:    emaLong,
		ZScores:    zScores,
		VwzScores:  vwzScores,
		PlusDI:     plusDI,
		MinusDI:    minusDI,
		AdxSeries:  adxSeries,
		BbwzScores: bbwzScores,
		Bbw:        bbw,
		DX:         dx,
	}, nil
}
