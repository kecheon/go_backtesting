package strategy

type PatternType string

const (
	PatternIncreasing PatternType = "increasing" // 상승형
	PatternDecreasing PatternType = "decreasing" // 하락형
	PatternMixed      PatternType = "mixed"      // 혼합형 (불규칙)
)

// AnalyzePattern inspects the last 3 values and returns the pattern type.
// Sign-aware: positive series => normal increasing (a < b < c).
//
//	negative series => "more negative" (a > b > c) is treated as Increasing (strengthening short).
func AnalyzePattern(values []float64) PatternType {
	if len(values) != 3 {
		return PatternMixed
	}
	a, b, c := values[0], values[1], values[2]

	// all non-negative -> standard increasing/decreasing
	if a >= 0 && b >= 0 && c >= 0 {
		switch {
		case a < b && b < c:
			return PatternIncreasing
		case a > b && b > c:
			return PatternDecreasing
		default:
			return PatternMixed
		}
	}

	// all non-positive -> for shorts: becoming more negative (a > b > c) means "strengthening short"
	if a <= 0 && b <= 0 && c <= 0 {
		switch {
		// a > b > c : e.g. -1, -2, -3 (더 음수) => short-strengthening => treat as Increasing (signal gets stronger)
		case a > b && b > c:
			return PatternIncreasing
		// a < b < c : e.g. -3, -2, -1 (음수지만 값이 커짐) => short-weakening => treat as Decreasing
		case a < b && b < c:
			return PatternDecreasing
		default:
			return PatternMixed
		}
	}

	// mixed signs -> ambiguous pattern (could add heuristics later)
	return PatternMixed
}

// DetectVolatilityExplosion detects sudden volatility expansion in BBW.
// Example: 0.01 → 0.03 → 0.05 (폭발적 증가)
func DetectVolatilityExplosion(values []float64, thresholdRatio float64) bool {
	if len(values) != 3 {
		return false
	}

	a, b, c := values[0], values[1], values[2]

	// 변동성 폭발 조건: 직전 대비 두 번 연속 threshold 이상 상승
	return (b > a*thresholdRatio) && (c > b*thresholdRatio)
}

// DetectSpike detects a large spike in value.
// zscore 또는 vwzscore 급등 감지용
func DetectSpike(values []float64, spikeDiff float64) bool {
	if len(values) != 3 {
		return false
	}

	a, b, c := values[0], values[1], values[2]

	// 급격한 값 증가 (예: 1.0 → 3.0 → 2.9)
	return (b-a >= spikeDiff) || (c-b >= spikeDiff)
}

// IsTrendStrengthening checks if ADX is strengthening.
func IsTrendStrengthening(adx []float64) bool {
	return AnalyzePattern(adx) == PatternIncreasing
}

type DirectionState struct {
	PlusDI  PatternType
	MinusDI PatternType
	DX      PatternType
}

func AnalyzeDirection(plusDI, minusDI, dx []float64) DirectionState {
	return DirectionState{
		PlusDI:  AnalyzePattern(plusDI),
		MinusDI: AnalyzePattern(minusDI),
		DX:      AnalyzePattern(dx),
	}
}

// IsHighQualitySignal evaluates whether the indicator state is suitable for entry.
type SignalQuality struct {
	Side            string
	ZScore          bool
	VWZScore        bool
	ADXGood         bool
	DIAligned       bool
	BandwidthStable bool
	NoSpike         bool
	OverallGood     bool
}

func EvaluateSignal(
	zscore, vwz, bbw, adx, plusDI, minusDI, dx []float64,
) string {

	// --- 1) 개별 패턴 분석 ---
	zPattern := AnalyzePattern(zscore)
	vPattern := AnalyzePattern(vwz)
	bPattern := AnalyzePattern(bbw)
	adxPattern := AnalyzePattern(adx)

	dir := AnalyzeDirection(plusDI, minusDI, dx)

	// --- 3) 부가 필터 ---
	volExplosion := DetectVolatilityExplosion(bbw, 1.5)
	zSpike := DetectSpike(zscore, 1.2)
	vSpike := DetectSpike(vwz, 1.0)

	longCondition := zPattern == PatternIncreasing &&
		vPattern == PatternIncreasing &&
		adxPattern == PatternIncreasing &&
		dir.PlusDI == PatternIncreasing && dir.MinusDI == PatternDecreasing &&
		bPattern == PatternIncreasing && !volExplosion &&
		!zSpike && !vSpike
	shortCondition := zPattern == PatternIncreasing &&
		vPattern == PatternIncreasing &&
		adxPattern == PatternIncreasing &&
		dir.PlusDI == PatternDecreasing && dir.MinusDI == PatternIncreasing &&
		bPattern == PatternIncreasing && !volExplosion &&
		!zSpike && !vSpike

	if longCondition {
		return "long"
	} else if shortCondition {
		return "short"
	}
	return "neutral"
}

type Pattern struct {
	Trend      string // "up", "down", "flat"
	Momentum   string // "increasing", "decreasing", "stable"
	Volatility string // "expanding", "contracting", "stable"
	Reversal   bool
	Breakout   bool
}

func AnalyzePatternZ(values []float64) Pattern {
	n := len(values)
	if n < 3 {
		return Pattern{}
	}

	a, b, c := values[n-3], values[n-2], values[n-1]

	p := Pattern{}

	// --- Trend ---
	switch {
	case c > b && b > a:
		p.Trend = "up"
	case c < b && b < a:
		p.Trend = "down"
	default:
		p.Trend = "flat"
	}

	// --- Momentum ---
	d1 := b - a
	d2 := c - b

	switch {
	case d2 > d1 && d2 > 0:
		p.Momentum = "increasing"
	case d2 < d1 && d2 < 0:
		p.Momentum = "decreasing"
	default:
		p.Momentum = "stable"
	}

	// --- Volatility (absolute diff) ---
	v1 := abs(d1)
	v2 := abs(d2)

	switch {
	case v2 > v1:
		p.Volatility = "expanding"
	case v2 < v1:
		p.Volatility = "contracting"
	default:
		p.Volatility = "stable"
	}

	// --- Reversal detection ---
	if (b > a && c < b) || (b < a && c > b) {
		p.Reversal = true
	}

	// --- Breakout detection ---
	if p.Trend == "up" && p.Momentum == "increasing" {
		p.Breakout = true
	}

	return p
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

type CombinedPattern struct {
	Side     string // "long", "short", "none"
	Strength string // "strong", "medium", "weak"
	Valid    bool
}

func CombinePatterns(z Pattern, v Pattern) CombinedPattern {
	cp := CombinedPattern{Side: "none", Valid: false}

	// --- Long logic ---
	if z.Breakout && v.Momentum == "increasing" {
		cp.Side = "long"
		cp.Valid = true
		if v.Volatility == "expanding" {
			cp.Strength = "strong"
		} else {
			cp.Strength = "medium"
		}
		return cp
	}

	// --- Short logic ---
	if z.Reversal && v.Momentum == "decreasing" {
		cp.Side = "short"
		cp.Valid = true
		if v.Volatility == "expanding" {
			cp.Strength = "strong"
		} else {
			cp.Strength = "medium"
		}
		return cp
	}

	return cp
}

func ShouldEnter(cp CombinedPattern) bool {
	return cp.Valid && cp.Strength != "medium"
}
