package strategy

import (
	"go-backtesting/config"
	"log"
	"math"
)

// DefaultLongCondition provides the default logic for a long entry signal.
func BBWLongCondition(indicators TechnicalIndicators) (bool, bool) {
	entry := indicators.EmaShort[1] < indicators.EmaShort[2] &&
		last(indicators.EmaShort) > last(indicators.EmaLong) &&
		// indicators.BBState.Status == ExpandingBullish &&
		// indicators.ADX[0] < 25.0 &&
		// last(indicators.BbwzScore) > 1.0 &&
		// indicators.EmaShort[0] < indicators.EmaShort[1] &&
		// indicators.ZScore[1] < indicators.ZScore[2] &&
		indicators.VWZScore[1] < indicators.VWZScore[2] &&
		// indicators.VWZScore[1]*indicators.VWZScore[2] > 0 &&
		// indicators.VWZScore[2] > 1.5 &&
		// indicators.VWZScore[0] < indicators.VWZScore[1] &&
		// indicators.BbwzScore[1] < indicators.BbwzScore[2] &&
		// indicators.PlusDI[1] < indicators.PlusDI[2] &&
		// indicators.BoxFilter[2] < 1.0 &&
		// indicators.BoxFilter[2] > 1.0 && //indicators.ADX[2] < 25.0) &&
		// indicators.BoxFilter[1]*indicators.BoxFilter[2] > 0 &&
		indicators.BoxFilter[2] > 2.0 &&
		last(indicators.PlusDI) > last(indicators.MinusDI)
	return entry, false
}

// DefaultShortCondition provides the default logic for a short entry signal.
func BBWShortCondition(indicators TechnicalIndicators) (bool, bool) {
	entry := indicators.EmaShort[1] > indicators.EmaShort[2] &&
		last(indicators.EmaShort) < last(indicators.EmaLong) &&
		// indicators.BBState.Status == ExpandingBearish &&
		// indicators.ADX[0] < 25.0 &&
		// last(indicators.BbwzScore) < -1.0 &&
		// indicators.EmaShort[0] > indicators.EmaShort[1] &&
		// indicators.ZScore[1] > indicators.ZScore[2] &&
		indicators.VWZScore[1] > indicators.VWZScore[2] &&
		// indicators.VWZScore[1]*indicators.VWZScore[2] > 0 &&
		// indicators.VWZScore[2] < -1.5 &&
		// indicators.VWZScore[0] > indicators.VWZScore[1] &&
		// indicators.BbwzScore[1] > indicators.BbwzScore[2] &&
		// indicators.MinusDI[1] < indicators.MinusDI[2] &&
		// indicators.BoxFilter[2] > 1.0 && // indicators.ADX[2] < 25.0) &&
		// indicators.BoxFilter[1]*indicators.BoxFilter[2] > 0 &&
		indicators.BoxFilter[2] > 2.0 &&
		last(indicators.PlusDI) < last(indicators.MinusDI)
	return entry, false
}

func CombinedLongCondition(indicators TechnicalIndicators) (bool, bool) {
	side := EvaluateSignal(indicators.ZScore, indicators.VWZScore,
		indicators.BbwzScore, indicators.ADX, indicators.PlusDI, indicators.MinusDI, indicators.DX)
	return side == "long", false
}
func CombinedShortCondition(indicators TechnicalIndicators) (bool, bool) {
	side := EvaluateSignal(indicators.ZScore, indicators.VWZScore,
		indicators.BbwzScore, indicators.ADX, indicators.PlusDI, indicators.MinusDI, indicators.DX)
	return side == "short", false
}

func InverseLongCondition(indicators TechnicalIndicators) (bool, bool) {
	entry := (last(indicators.BbwzScore) > 2.5 || last(indicators.BbwzScore) < -2.5) &&
		// last(indicators.VWZScore) < -1.5 &&
		// indicators.BBState.Status == Squeeze &&
		last(indicators.EmaShort) > last(indicators.EmaLong)
	return entry, false
}
func InverseShortCondition(indicators TechnicalIndicators) (bool, bool) {
	entry := (last(indicators.BbwzScore) > 2.5 || last(indicators.BbwzScore) < -2.5) &&
		// last(indicators.VWZScore) > 1.5 &&
		// indicators.BBState.Status == Squeeze &&
		last(indicators.EmaShort) < last(indicators.EmaLong)
	return entry, false
}

func DMILongCondition(indicators TechnicalIndicators) (bool, bool) {
	// --- 1. Load Configuration ---
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	// if indicators.BbwzScore[2] > 1.0 {
	// 	return false
	// }
	// if indicators.BoxFilter[2] > cfg.BoxFilter.Threshold {
	// 	return false
	// }
	// 기존 ADX 조건
	condition1 := indicators.ADX[2] > cfg.ADXThreshold
	// ADX 증가 조건
	condition2 := indicators.ADX[2] > indicators.ADX[1]

	if !condition1 || !condition2 {
		// fmt.Printf("[%s 진입 거절] ADX가 작거나 증가하지 않음\n", symbol)
		return false, false
	}

	// DX 필터 추가
	dx := math.Abs(indicators.PlusDI[2] - indicators.MinusDI[2])
	if dx < cfg.ADXThreshold {
		// fmt.Printf("[%s 진입 거절] DX %.2f < %.2f (방향성 약함)\n", symbol, dx, dmiParam.DXThreshold)
		return false, false
	}

	// DI 방향 및 DI 증가 조건
	var condition3, diIncrease bool
	condition3 = indicators.PlusDI[2] > indicators.MinusDI[2]
	diIncrease = indicators.PlusDI[2] > indicators.PlusDI[1]

	if !condition3 {
		// fmt.Printf("[%s 진입 거절] 추세 방향 반대 side: %s PlusDI: %.2f, MinusDI: %.2f\n",
		return false, false
	}

	if !diIncrease {
		// fmt.Printf("[%s 진입 거절] 진입 방향 DI가 증가하지 않음\n", symbol)
		return false, false
	}
	return true, false
}

func DMIShortCondition(indicators TechnicalIndicators) (bool, bool) {

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	// if indicators.BbwzScore[2] > 1.0 {
	// 	return false
	// }
	// if indicators.BoxFilter[2] > cfg.BoxFilter.Threshold {
	// 	return false
	// }
	// 기존 ADX 조건
	condition1 := indicators.ADX[2] > cfg.ADXThreshold
	// ADX 증가 조건
	condition2 := indicators.ADX[2] > indicators.ADX[1]

	if !condition1 || !condition2 {
		// fmt.Printf("[%s 진입 거절] ADX가 작거나 증가하지 않음\n", symbol)
		return false, false
	}

	// DX 필터 추가
	dx := math.Abs(indicators.MinusDI[2] - indicators.PlusDI[2])
	if dx < cfg.ADXThreshold {
		// fmt.Printf("[%s 진입 거절] DX %.2f < %.2f (방향성 약함)\n", symbol, dx, dmiParam.DXThreshold)
		return false, false
	}

	// DI 방향 및 DI 증가 조건
	var condition3, diIncrease bool
	condition3 = indicators.PlusDI[2] < indicators.MinusDI[2]
	diIncrease = indicators.MinusDI[2] > indicators.MinusDI[1]

	if !condition3 {
		// fmt.Printf("[%s 진입 거절] 추세 방향 반대 side: %s PlusDI: %.2f, MinusDI: %.2f\n",
		return false, false
	}

	if !diIncrease {
		// fmt.Printf("[%s 진입 거절] 진입 방향 DI가 증가하지 않음\n", symbol)
		return false, false
	}
	return true, false
}
