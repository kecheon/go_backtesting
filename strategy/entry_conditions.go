package strategy

import (
	"go-backtesting/config"
)

// EntryCondition defines the signature for a function that checks for a trading signal.
type EntryCondition func(indicators TechnicalIndicators, config *config.Config) (bool, bool)

// DetermineEntrySignal determines the entry signal based on the indicators.
func DetermineEntrySignal(indicators TechnicalIndicators, config *config.Config, longCondition EntryCondition, shortCondition EntryCondition) (string, bool, bool) {
	adx := last(indicators.ADX)
	if adx > config.ADXThreshold &&
		adx < config.AdxUpperThreshold {
		// indicators.ADX[0] < indicators.ADX[1] &&
		// indicators.ADX[1] < indicators.ADX[2] {
		if entry, stop := longCondition(indicators, config); entry || stop {
			return "long", entry, stop
		}
		if entry, stop := shortCondition(indicators, config); entry || stop {
			return "short", entry, stop
		}
	}
	return "", false, false
}

// DefaultLongCondition provides the default logic for a long entry signal.
func DefaultLongCondition(indicators TechnicalIndicators, config *config.Config) (bool, bool) {
	entry := last(indicators.EmaShort) > last(indicators.EmaLong) && last(indicators.ZScore) < 0.0
	return entry, false
}

// DefaultShortCondition provides the default logic for a short entry signal.
func DefaultShortCondition(indicators TechnicalIndicators, config *config.Config) (bool, bool) {
	entry := last(indicators.EmaShort) < last(indicators.EmaLong) && last(indicators.ZScore) > 0.0
	return entry, false
}

// MACDLongCondition checks for a bullish MACD crossover.
func MACDLongCondition(indicators TechnicalIndicators, config *config.Config) (bool, bool) {
	// A bullish crossover occurs when the MACD histogram crosses from negative to positive.
	if len(indicators.MACDHistogram) < 2 {
		return false, false
	}
	prev := indicators.MACDHistogram[len(indicators.MACDHistogram)-2]
	curr := indicators.MACDHistogram[len(indicators.MACDHistogram)-1]
	entry := prev < 0 && curr > 0
	return entry, false
}

// MACDShortCondition checks for a bearish MACD crossover.
func MACDShortCondition(indicators TechnicalIndicators, config *config.Config) (bool, bool) {
	// A bearish crossover occurs when the MACD histogram crosses from positive to negative.
	if len(indicators.MACDHistogram) < 2 {
		return false, false
	}
	prev := indicators.MACDHistogram[len(indicators.MACDHistogram)-2]
	curr := indicators.MACDHistogram[len(indicators.MACDHistogram)-1]
	entry := prev > 0 && curr < 0
	return entry, false
}

// last safely returns the last element of a slice, or 0 if the slice is empty.
func last(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	return data[len(data)-1]
}
