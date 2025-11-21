package strategy

// DefaultLongCondition provides the default logic for a long entry signal.
func BBWLongCondition(indicators TechnicalIndicators) bool {
	return last(indicators.BbwzScore) > 1.0 &&
		indicators.BBState.Status == ExpandingBullish &&
		last(indicators.EmaShort) > last(indicators.EmaLong) &&
		// indicators.EmaShort[0] < indicators.EmaShort[1] &&
		indicators.EmaShort[1] < indicators.EmaShort[2] &&
		// indicators.ZScore[1] < indicators.ZScore[2] &&
		indicators.VWZScore[1] < indicators.VWZScore[2] &&
		// indicators.BbwzScore[1] < indicators.BbwzScore[2] &&
		// indicators.PlusDI[1] < indicators.PlusDI[2] &&
		last(indicators.PlusDI) > last(indicators.MinusDI)
}

// DefaultShortCondition provides the default logic for a short entry signal.
func BBWShortCondition(indicators TechnicalIndicators) bool {
	return last(indicators.BbwzScore) < -1.0 &&
		indicators.BBState.Status == ExpandingBearish &&
		last(indicators.EmaShort) < last(indicators.EmaLong) &&
		// indicators.EmaShort[0] > indicators.EmaShort[1] &&
		indicators.EmaShort[1] > indicators.EmaShort[2] &&
		// indicators.ZScore[1] > indicators.ZScore[2] &&
		indicators.VWZScore[1] > indicators.VWZScore[2] &&
		// indicators.BbwzScore[1] > indicators.BbwzScore[2] &&
		// indicators.MinusDI[1] < indicators.MinusDI[2] &&
		last(indicators.PlusDI) < last(indicators.MinusDI)
}

func CombinedLongCondition(indicators TechnicalIndicators) bool {
	side := EvaluateSignal(indicators.ZScore, indicators.VWZScore,
		indicators.BbwzScore, indicators.ADX, indicators.PlusDI, indicators.MinusDI, indicators.DX)
	return side == "long"
}
func CombinedShortCondition(indicators TechnicalIndicators) bool {
	side := EvaluateSignal(indicators.ZScore, indicators.VWZScore,
		indicators.BbwzScore, indicators.ADX, indicators.PlusDI, indicators.MinusDI, indicators.DX)
	return side == "short"
}

func InverseLongCondition(indicators TechnicalIndicators) bool {
	return (last(indicators.BbwzScore) > 2.5 || last(indicators.BbwzScore) < -2.5) &&
		// last(indicators.VWZScore) < -1.5 &&
		// indicators.BBState.Status == Squeeze &&
		last(indicators.EmaShort) > last(indicators.EmaLong)
}
func InverseShortCondition(indicators TechnicalIndicators) bool {
	return (last(indicators.BbwzScore) > 2.5 || last(indicators.BbwzScore) < -2.5) &&
		// last(indicators.VWZScore) > 1.5 &&
		// indicators.BBState.Status == Squeeze &&
		last(indicators.EmaShort) < last(indicators.EmaLong)
}
