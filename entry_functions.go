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
		DX:       s.DX[i],
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
	longCondition := indicators.EmaShort > indicators.EmaLong &&
		// indicators.PlusDI > indicators.MinusDI &&
		// indicators.VWZScore > 1.5 &&
		// indicators.VWZScore < 2.5 &&
		// indicators.ZScore < 1.50 &&
		indicators.BBState.Status == ExpandingBullish &&
		indicators.BBState.BBW > 0.013

	// 숏 포지션 진입 조건:
	// 1. 볼린저 밴드 폭이 약세 확장 상태 (ExpandingBearish)
	// 2. -DI가 +DI보다 큼 (하락 추세 우위)
	// 3. VWZ-Score가 -1.0 초과
	// 4. Z-Score가 -1.0 초과
	// 5. 단기 EMA가 장기 EMA보다 작음 (하락 추세)
	shortCondition := indicators.EmaShort < indicators.EmaLong &&
		// indicators.MinusDI > indicators.PlusDI &&
		// indicators.VWZScore < -1.5 &&
		// indicators.VWZScore > -2.5 &&
		// indicators.ZScore > -1.5 &&
		indicators.BBState.Status == ExpandingBearish &&
		indicators.BBState.BBW > 0.013

	// 최종 진입 결정:
	// ADX가 임계값보다 크고, 롱 또는 숏 조건 중 하나를 만족해야 함
	if indicators.ADX > adxThreshold &&
		indicators.ADX < 50 &&
		(longCondition || shortCondition) {
		if longCondition {
			return "long", true
		}
		return "short", true
	}

	// 진입 조건이 충족되지 않으면 신호 없음
	return "", false
}

/**
func canEnterVFilter(d *DragonBot, exchangeName, symbol string, side domain.SideType, dmiParam config.DMIParams, initialEntry bool) bool {

	candles := d.MarketDataServices[exchangeName].GetCandleSticks(symbol)
	if len(candles) < 300 {
		return false
	}
	plusDI, minusDI, adxValue, _, _, err := DMIGivenCandles(candles, exchangeName, symbol, dmiParam.Period)
	if err != nil {
		fmt.Println(err)
	}
	n := len(adxValue) - 1

	// 기존 ADX 조건
	condition1 := adxValue[n] > dmiParam.Threshold
	if initialEntry && dmiParam.UpperBound > 0 {
		condition1 = adxValue[n] < dmiParam.UpperBound
	}

	// ADX 증가 조건
	condition2 := adxValue[n] > adxValue[n-1]

	if !condition1 || !condition2 {
		// fmt.Printf("[%s 진입 거절] ADX가 작거나 증가하지 않음\n", symbol)
		return false
	}

	// DX 필터 추가
	dx := math.Abs(plusDI[n] - minusDI[n])
	if dx < dmiParam.DXThreshold {
		// fmt.Printf("[%s 진입 거절] DX %.2f < %.2f (방향성 약함)\n", symbol, dx, dmiParam.DXThreshold)
		return false
	}

	// DI 방향 및 DI 증가 조건
	var condition3, diIncrease bool
	switch side {
	case domain.LongSide{}:
		condition3 = plusDI[n] > minusDI[n]
		diIncrease = plusDI[n] > plusDI[n-1]
	case domain.ShortSide{}:
		condition3 = plusDI[n] < minusDI[n]
		diIncrease = minusDI[n] > minusDI[n-1]
	}

	if !condition3 {
		// fmt.Printf("[%s 진입 거절] 추세 방향 반대 side: %s PlusDI: %.2f, MinusDI: %.2f\n",
		// symbol, side.String(), plusDI[n], minusDI[n])
		return false
	}

	if !diIncrease {
		// fmt.Printf("[%s 진입 거절] 진입 방향 DI가 증가하지 않음\n", symbol)
		return false
	}

	cfg := config.GetInstance()
	boxFilter := cfg.Get("BoxFilter").(config.BoxFilter)
	zScoreFilter := cfg.Get("ZScoreFilter").(config.ZScoreFilter)
	vWZScoreFilter := cfg.Get("VWZScoreFilter").(config.VWZScoreFilter)
	emaZScoreFilter := cfg.Get("EMAZScoreFilter").(config.EMAZScoreFilter)
	bbwFilter := cfg.Get("BBStateFilter").(config.BBStateFilter)

	if boxFilter.Enabled {
		// 2. 전체 캔들 데이터 길이만큼 슬라이스를 생성합니다.
		highs := make([]float64, len(candles))
		lows := make([]float64, len(candles))

		// 3. 데이터를 잘라내지 말고, 전체 캔들을 순회하며 슬라이스를 채웁니다.

		for i, c := range candles {
			highs[i] = c.High
			lows[i] = c.Low
		}
		period := boxFilter.Period
		highestHighs := talib.Max(highs, period)
		lowestLows := talib.Min(lows, period)

		minRangePct := boxFilter.Rate
		isRanging := ((highestHighs[len(highestHighs)-1] - lowestLows[len(lowestLows)-1]) / lowestLows[len(lowestLows)-1]) < minRangePct

		if isRanging {
			return false
		}
	}
	if zScoreFilter.Enabled {
		zScores := ZScores(candles, zScoreFilter.Period)
		lastIndex := len(zScores) - 1
		var condition bool
		if initialEntry {
			condition = math.Abs(zScores[lastIndex]) > zScoreFilter.LowerThreshold &&
				math.Abs(zScores[lastIndex]) < zScoreFilter.UpperThreshold
		} else {
			condition = math.Abs(zScores[lastIndex]) > zScoreFilter.LowerThreshold
		}
		if !condition { // ranging
			return false
		}
	}

	if vWZScoreFilter.Enabled {
		zScores := VWZScores(candles, vWZScoreFilter.Period)
		lastIndex := len(zScores) - 1
		var condition bool
		if initialEntry {
			condition = math.Abs(zScores[lastIndex]) > vWZScoreFilter.LowerThreshold &&
				math.Abs(zScores[lastIndex]) < vWZScoreFilter.UpperThreshold
		} else {
			condition = math.Abs(zScores[lastIndex]) > vWZScoreFilter.LowerThreshold
		}
		if !condition { // ranging
			return false
		}
	}

	if emaZScoreFilter.Enabled {
		zScores := ZScoresWithEMA(candles, emaZScoreFilter.Period, adxValue[n], 10, 50)
		lastIndex := len(zScores) - 1
		var condition bool
		if initialEntry {
			condition = math.Abs(zScores[lastIndex]) > emaZScoreFilter.LowerThreshold &&
				math.Abs(zScores[lastIndex]) < emaZScoreFilter.UpperThreshold
		} else {
			condition = math.Abs(zScores[lastIndex]) > emaZScoreFilter.LowerThreshold
		}
		if !condition { // ranging
			return false
		}
	}

	if bbwFilter.Enabled {
		bbw, _, _, _ := BBW(candles, bbwFilter.Period, bbwFilter.Multiplier)
		if bbw[len(bbw)-1] < bbwFilter.Threshold {
			msg := fmt.Sprintf("[%s] 진입거절 by BBW Filter BBW: %.4f", symbol, bbw[len(bbw)-1])
			log.Logger <- log.Log{Msg: msg, Level: logrus.DebugLevel}
			return false
		}
	}

	return true
}

**/
