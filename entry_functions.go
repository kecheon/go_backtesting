package main

// determineEntrySignal은 주어진 데이터 포인트를 기반으로 진입 신호를 결정합니다.
// 이 함수는 캔들 데이터, BBW 상태, 각종 지표 및 임계값을 입력으로 받습니다.
// 진입 조건이 충족되면 진입 방향("long" 또는 "short")과 true를 반환하고,
// 그렇지 않으면 빈 문자열과 false를 반환합니다.
func determineEntrySignal(
	bbState BBWState,
	plusDI float64,
	minusDI float64,
	vwzScore float64,
	zscore float64,
	emaShort float64,
	emaLong float64,
	adx float64,
	adxThreshold float64,
) (string, bool) {
	// 롱 포지션 진입 조건:
	// 1. 볼린저 밴드 폭이 강세 확장 상태 (ExpandingBullish)
	// 2. +DI가 -DI보다 큼 (상승 추세 우위)
	// 3. VWZ-Score가 1.0 미만
	// 4. Z-Score가 1.0 미만
	// 5. 단기 EMA가 장기 EMA보다 큼 (상승 추세)
	longCondition := bbState.Status == ExpandingBullish && plusDI > minusDI && vwzScore < 1.0 && zscore < 1.0 && emaShort > emaLong

	// 숏 포지션 진입 조건:
	// 1. 볼린저 밴드 폭이 약세 확장 상태 (ExpandingBearish)
	// 2. -DI가 +DI보다 큼 (하락 추세 우위)
	// 3. VWZ-Score가 -1.0 초과
	// 4. Z-Score가 -1.0 초과
	// 5. 단기 EMA가 장기 EMA보다 작음 (하락 추세)
	shortCondition := bbState.Status == ExpandingBearish && minusDI > plusDI && vwzScore > -1.0 && zscore > -1.0 && emaShort < emaLong

	// 최종 진입 결정:
	// ADX가 임계값보다 크고, 롱 또는 숏 조건 중 하나를 만족해야 함
	if adx > adxThreshold && (longCondition || shortCondition) {
		if longCondition {
			return "long", true
		}
		return "short", true
	}

	// 진입 조건이 충족되지 않으면 신호 없음
	return "", false
}
