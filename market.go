package main

type DMIParams struct {
	Hedge       bool    `yaml:"hedge"`
	Period      int     `yaml:"period"`
	Threshold   float64 `yaml:"threshold"`
	DXThreshold float64 `yaml:"dxthreshold"`
	UpperBound  float64 `yaml:"upperbound"` // 최초 진입시 ADX값이 이것 보다 크면 끝물이므로 진입 거절
}

type ZScoreFilter struct {
	Enabled        bool    `yaml:"enabled"`
	Period         int     `yaml:"period"`
	LowerThreshold float64 `yaml:"lowerthreshold"`
	UpperThreshold float64 `yaml:"upperthreshold"`
}

type VWZScoreFilter struct {
	Enabled        bool    `yaml:"enabled"`
	Period         int     `yaml:"period"`
	LowerThreshold float64 `yaml:"lowerthreshold"`
	UpperThreshold float64 `yaml:"upperthreshold"`
}

type EMAZScoreFilter struct {
	Enabled        bool    `yaml:"enabled"`
	Period         int     `yaml:"period"`
	LowerThreshold float64 `yaml:"lowerthreshold"`
	UpperThreshold float64 `yaml:"upperthreshold"`
}

type BBStateFilter struct {
	Enabled    bool    `yaml:"enabled"`
	Period     int     `yaml:"period"`
	Multiplier float64 `yaml:"multiplier"`
	Threshold  float64 `yaml:"threshold"` // BBW average * Threshold. for example, when Threshold == 1.1 then bbw > bbwAverage * 1.1 would be the condition to decide BB trend
}

type VolatilityFilter struct {
	Enabled   bool    `yaml:"enabled"`
	Threshold float64 `yaml:"threshold"`
	Candles   int     `yaml:"candles"`
}

type MarketState string

const (
	BreakUp          MarketState = "BreakUp"
	BreakDown        MarketState = "BreakDown"
	Sideways         MarketState = "Sideways"
	Squeeze          MarketState = "Squeeze"
	ExpandingBullish MarketState = "ExpandingBullish"
	ExpandingBearish MarketState = "ExpandingBearish"
	Neutral          MarketState = "Neutral"
	Unknown          MarketState = "Unknown"
)

// Get method
func (ms MarketState) Get() string {
	return string(ms)
}
func IsTrending(state MarketState) bool {
	return state == BreakUp || state == BreakDown ||
		state == ExpandingBullish || state == ExpandingBearish
}
func IsBroken(state MarketState) bool {
	return state == BreakUp || state == BreakDown
}

type MarketTrend struct {
	Enabled      bool    `yaml:"enabled"`
	CandleSize   string  `yaml:"candlesize"`
	Threshold    float64 `yaml:"threshold"`
	EmaShort     int     `yaml:"emashort"`
	EmaLong      int     `yaml:"emalong"`
	BBPeriod     int     `yaml:"bbperiod"`     // BB term default 20
	BBMultiplier float64 `yaml:"bbmultiplier"` // 시그마 계수 default 1.5
	BBWThreshold float64 `yaml:"bbwthreshold"` // BBW 평균에 대한 multiplier default 1.1 (평균보다 10% 이상 크면 확장 작으면 수축)
}
