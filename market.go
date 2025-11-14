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

// TechnicalIndicators는 한 시점의 기술적 분석 지표들을 담는 구조체입니다.
type TechnicalIndicators struct {
	BBState  BBWState
	PlusDI   float64
	MinusDI  float64
	VWZScore float64
	ZScore   float64
	EmaShort float64
	EmaLong  float64
	ADX      float64
}

// StrategyDataContext는 백테스팅 전략에 필요한 모든 데이터 시리즈를 보관하는 컨테이너입니다.
type StrategyDataContext struct {
	Candles    CandleSticks
	EmaShort   []float64
	EmaLong    []float64
	ZScores    []float64
	VwzScores  []float64
	PlusDI     []float64
	MinusDI    []float64
	AdxSeries  []float64
	BbwzScores []float64
	Bbw        []float64
	Dx         []float64
}
