package strategy

import "go-backtesting/config"

// EntryCondition defines the signature for a function that checks for a trading signal.
type EntryCondition func(indicators TechnicalIndicators) bool

// DetermineEntrySignal determines the entry signal based on the indicators.
func DetermineEntrySignal(indicators TechnicalIndicators, config *config.Config, longCondition EntryCondition, shortCondition EntryCondition) (string, bool) {
	adx := last(indicators.ADX)
	if adx > config.ADXThreshold && adx < config.AdxUpperThreshold {
		if longCondition(indicators) {
			return "long", true
		}
		if shortCondition(indicators) {
			return "short", true
		}
	}
	return "", false
}

// DefaultLongCondition provides the default logic for a long entry signal.
func DefaultLongCondition(indicators TechnicalIndicators) bool {
	return last(indicators.EmaShort) > last(indicators.EmaLong) && last(indicators.ZScore) < 0.0
}

// DefaultShortCondition provides the default logic for a short entry signal.
func DefaultShortCondition(indicators TechnicalIndicators) bool {
	return last(indicators.EmaShort) < last(indicators.EmaLong) && last(indicators.ZScore) > 0.0
}

// MACDLongCondition checks for a bullish MACD crossover.
func MACDLongCondition(indicators TechnicalIndicators) bool {
	// A bullish crossover occurs when the MACD histogram crosses from negative to positive.
	if len(indicators.MACDHistogram) < 2 {
		return false
	}
	prev := indicators.MACDHistogram[len(indicators.MACDHistogram)-2]
	curr := indicators.MACDHistogram[len(indicators.MACDHistogram)-1]
	return prev < 0 && curr > 0
}

// MACDShortCondition checks for a bearish MACD crossover.
func MACDShortCondition(indicators TechnicalIndicators) bool {
	// A bearish crossover occurs when the MACD histogram crosses from positive to negative.
	if len(indicators.MACDHistogram) < 2 {
		return false
	}
	prev := indicators.MACDHistogram[len(indicators.MACDHistogram)-2]
	curr := indicators.MACDHistogram[len(indicators.MACDHistogram)-1]
	return prev > 0 && curr < 0
}

// last safely returns the last element of a slice, or 0 if the slice is empty.
func last(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	return data[len(data)-1]
}
