package main

// createTechnicalIndicators는 주어진 인덱스 i에 대한 모든 기술 지표를 수집하여
// TechnicalIndicators 구조체를 생성하고 반환하는 StrategyDataContext의 메소드입니다.
func (s *StrategyDataContext) createTechnicalIndicators(i int) TechnicalIndicators {
	return TechnicalIndicators{
		BBState:  DetectBBWState(s.Candles[:i+1], 20, 2.0, 0),
		PlusDI:   s.PlusDI[i],
		MinusDI:  s.MinusDI[i],
		VWZScore: s.VwzScores[i],
		ZScore:   s.ZScores[i],
		EmaShort: s.EmaShort[i],
		EmaLong:  s.EmaLong[i],
		ADX:      s.AdxSeries[i],
	}
}

// determineEntrySignal은 주어진 기술 지표 세트를 기반으로 진입 신호를 결정합니다.
// 이 함수는 TechnicalIndicators 구조체와 ADX 임계값을 입력으로 받습니다.
// 진입 조건이 충족되면 진입 방향("long" 또는 "short")과 true를 반환하고,
// 그렇지 않으면 빈 문자열과 false를 반환합니다.
func determineEntrySignal(indicators TechnicalIndicators, adxThreshold float64) (string, bool) {
	// 롱 포지션 진입 조건:
	// 1. 볼린저 밴드 폭이 강세 확장 상태 (ExpandingBullish)
	// 2. +DI가 -DI보다 큼 (상승 추세 우위)
	// 3. VWZ-Score가 1.0 미만
	// 4. Z-Score가 1.0 미만
	// 5. 단기 EMA가 장기 EMA보다 큼 (상승 추세)
	longCondition := indicators.BBState.Status == ExpandingBullish &&
		indicators.PlusDI > indicators.MinusDI &&
		indicators.VWZScore < 1.0 &&
		indicators.ZScore < 1.0 &&
		indicators.EmaShort > indicators.EmaLong

	// 숏 포지션 진입 조건:
	// 1. 볼린저 밴드 폭이 약세 확장 상태 (ExpandingBearish)
	// 2. -DI가 +DI보다 큼 (하락 추세 우위)
	// 3. VWZ-Score가 -1.0 초과
	// 4. Z-Score가 -1.0 초과
	// 5. 단기 EMA가 장기 EMA보다 작음 (하락 추세)
	shortCondition := indicators.BBState.Status == ExpandingBearish &&
		indicators.MinusDI > indicators.PlusDI &&
		indicators.VWZScore > -1.0 &&
		indicators.ZScore > -1.0 &&
		indicators.EmaShort < indicators.EmaLong

	// 최종 진입 결정:
	// ADX가 임계값보다 크고, 롱 또는 숏 조건 중 하나를 만족해야 함
	if indicators.ADX > adxThreshold && (longCondition || shortCondition) {
		if longCondition {
			return "long", true
		}
		return "short", true
	}

	// 진입 조건이 충족되지 않으면 신호 없음
	return "", false
}
