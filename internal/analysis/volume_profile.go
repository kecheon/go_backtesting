package analysis

import (
	"math"
	"go-backtesting/internal/model"
)

type VolumeProfile struct {
	POC float64
	VAH float64
	VAL float64
}

// CalculateVolumeProfile computes the VP for the given slice of candles.
// binSizePct: e.g. 0.05 for 0.05%
// valueAreaPct: e.g. 0.70 for 70%
func CalculateVolumeProfile(candles []model.Candle, binSizePct float64, valueAreaPct float64) VolumeProfile {
	if len(candles) == 0 {
		return VolumeProfile{}
	}

	// 1. Determine Min/Max Price and Average Price to set bin size
	minPrice := candles[0].Low
	maxPrice := candles[0].High
	sumPrice := 0.0
	totalVolume := 0.0

	for _, c := range candles {
		if c.Low < minPrice {
			minPrice = c.Low
		}
		if c.High > maxPrice {
			maxPrice = c.High
		}
		sumPrice += c.Close
		totalVolume += c.Volume
	}
	avgPrice := sumPrice / float64(len(candles))

	// Bin Height based on average price
	binHeight := avgPrice * (binSizePct / 100.0)
	if binHeight == 0 {
		binHeight = 1.0 // Fallback
	}

	// 2. Create Bins
	// Map bin index to volume
	bins := make(map[int]float64)

	// Populate bins (using Typical Price)
	for _, c := range candles {
		typicalPrice := (c.High + c.Low + c.Close) / 3.0
		binIdx := int(math.Floor(typicalPrice / binHeight))
		bins[binIdx] += c.Volume
	}

	// 3. Find POC (Max Volume Bin)
	maxBinVol := -1.0
	pocBinIdx := 0

	for idx, vol := range bins {
		if vol > maxBinVol {
			maxBinVol = vol
			pocBinIdx = idx
		}
	}

	pocPrice := (float64(pocBinIdx) * binHeight) + (binHeight / 2.0)

	// 4. Calculate Value Area (VAH, VAL)
	targetVol := totalVolume * valueAreaPct
	currentVol := maxBinVol

	upIdx := pocBinIdx + 1
	downIdx := pocBinIdx - 1

	highestIdx := pocBinIdx
	lowestIdx := pocBinIdx

	for currentVol < targetVol {
		upVol := bins[upIdx]
		downVol := bins[downIdx]

		if upVol == 0 && downVol == 0 {
			break
		}

		if upVol >= downVol {
			currentVol += upVol
			if upIdx > highestIdx { highestIdx = upIdx }
			upIdx++
		} else {
			currentVol += downVol
			if downIdx < lowestIdx { lowestIdx = downIdx }
			downIdx--
		}
	}

	valPrice := (float64(lowestIdx) * binHeight)
	vahPrice := (float64(highestIdx) * binHeight) + binHeight

	return VolumeProfile{
		POC: pocPrice,
		VAH: vahPrice,
		VAL: valPrice,
	}
}
