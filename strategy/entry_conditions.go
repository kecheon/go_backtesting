package strategy

import "go-backtesting/config"

// EntryCondition defines the signature for a function that checks for a trading signal.
type EntryCondition func(indicators TechnicalIndicators) bool

// DetermineEntrySignal determines the entry signal based on the indicators.
func DetermineEntrySignal(indicators TechnicalIndicators, config *config.Config, longCondition EntryCondition, shortCondition EntryCondition) (string, bool) {
	if indicators.ADX > config.ADXThreshold && indicators.ADX < config.AdxUpperThreshold {
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
	return indicators.EmaShort > indicators.EmaLong && indicators.ZScore < 0.0
}

// DefaultShortCondition provides the default logic for a short entry signal.
func DefaultShortCondition(indicators TechnicalIndicators) bool {
	return indicators.EmaShort < indicators.EmaLong && indicators.ZScore > 0.0
}

// MACDLongCondition checks for a bullish MACD crossover.
func MACDLongCondition(indicators TechnicalIndicators) bool {
	// A bullish crossover occurs when the MACD histogram crosses from negative to positive.
	return indicators.PrevMACDHistogram < 0 && indicators.MACDHistogram > 0
}

// MACDShortCondition checks for a bearish MACD crossover.
func MACDShortCondition(indicators TechnicalIndicators) bool {
	// A bearish crossover occurs when the MACD histogram crosses from positive to negative.
	return indicators.PrevMACDHistogram > 0 && indicators.MACDHistogram < 0
}
