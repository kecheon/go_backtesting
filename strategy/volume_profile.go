package strategy

import (
	"go-backtesting/config"
	"go-backtesting/market"
	"math"
	"sort"
)

type VolumeProfile struct {
	POC       float64
	UpperPOCs []float64
	LowerPOCs []float64
}

// CalculateVolumeProfile calculates the Volume Profile for a given set of candles.
// It uses a rolling window approach, so `candles` should be the window (e.g., last 240 candles).
func CalculateVolumeProfile(candles []market.Candle, currentPrice float64, cfg config.VolumeProfileConfig) VolumeProfile {
	if len(candles) == 0 {
		return VolumeProfile{}
	}

	// 1. Determine Range
	minPrice := candles[0].Low
	maxPrice := candles[0].High

	for _, c := range candles {
		if c.Low < minPrice {
			minPrice = c.Low
		}
		if c.High > maxPrice {
			maxPrice = c.High
		}
	}

	// 2. Define Bin Size
	binSize := currentPrice * (cfg.BinSizePct / 100.0)
	if binSize <= 0 {
		binSize = 1.0 // Fallback to avoid division by zero
	}

	// 3. Create Bins and Aggregate Volume
	// Using a map for sparse distribution, though slice might be faster for dense data.
	// Given price range can be large, map is safer to avoid huge allocations if there are gaps.
	// Key is the bin index (floor(price / binSize))
	volumeBins := make(map[int]float64)

	for _, c := range candles {
		// Distribute volume across bins covered by the candle
		// Simple approach: midpoint or weighted?
		// Spec says: "accumulate each candle's volume to the corresponding price bin"
		// If a candle spans multiple bins, volume should ideally be distributed.
		// For simplicity and standard VP implementation, we often add to the bin of the Close price,
		// or distribute equally among bins touched by High-Low.
		// "Cluster" usually implies precise levels. Let's use the bin corresponding to the Close price as the primary anchor,
		// or better, distribute volume uniformly across the High-Low range.
		//
		// Decision: Uniform distribution across bins touched by the candle.

		lowBin := int(math.Floor(c.Low / binSize))
		highBin := int(math.Floor(c.High / binSize))

		numBins := highBin - lowBin + 1
		volPerBin := c.Vol / float64(numBins)

		for b := lowBin; b <= highBin; b++ {
			volumeBins[b] += volPerBin
		}
	}

	// 4. Find Global POC (Bin with Max Volume)
	var maxVol float64
	var pocBin int
	// We also need sorted bins to find local maxima
	var bins []int
	for b, vol := range volumeBins {
		bins = append(bins, b)
		if vol > maxVol {
			maxVol = vol
			pocBin = b
		}
	}
	sort.Ints(bins)

	pocPrice := (float64(pocBin) * binSize) + (binSize / 2.0)

	// 5. Find Local Maxima (UpperPOCs and LowerPOCs)
	// Definition of Local Maxima: A peak in the volume profile.
	// We need to smoothen or check neighbors to avoid noise.
	// Valid peak: Vol[i] > Vol[i-1] && Vol[i] > Vol[i+1] (simplified)

	// Convert map to a sorted slice of volumes for easier peak detection
	// We need to handle gaps (bins with 0 volume) if we iterate linearly.
	// Let's iterate from minBin to maxBin.
	if len(bins) == 0 {
		return VolumeProfile{POC: pocPrice}
	}
	minBin := bins[0]
	maxBin := bins[len(bins)-1]

	var upperPOCs []float64
	var lowerPOCs []float64

	// Helper to get vol safely
	getVol := func(b int) float64 {
		return volumeBins[b]
	}

	// Min distance in price terms
	minDistPrice := currentPrice * (cfg.MinPOCDistance / 100.0)

	// Identify candidate peaks
	var candidatePeaks []float64

	// Add Global POC as the first confirmed peak
	candidatePeaks = append(candidatePeaks, pocPrice)

	for b := minBin + 1; b < maxBin; b++ {
		v := getVol(b)
		if v == 0 {
			continue
		}

		// check neighbors (radius 1)
		// To be a peak, it should be higher than immediate non-zero neighbors or surrounding range?
		// Simple peak detection:
		if v > getVol(b-1) && v > getVol(b+1) {
			price := (float64(b) * binSize) + (binSize / 2.0)

			// Check prominence/significance?
			// The prompt says "Local Maxima...".
			// We will filter by distance later.
			if price != pocPrice { // avoid duplicate if global POC is detected here
				candidatePeaks = append(candidatePeaks, price)
			}
		}
	}

	// Filter peaks by distance.
	// If two peaks are too close, keep the one with higher volume?
	// Or simply, we just collect them and the strategy checks proximity.
	// The prompt says: "Local Maxima... UpperPOC and LowerPOC lists".
	// "MinPOCDistance" suggests we should filter.
	// Let's sort candidates by Volume descending to prioritize major peaks.

	type Peak struct {
		Price float64
		Vol   float64
	}
	var peaks []Peak
	for _, p := range candidatePeaks {
		// retrieve vol
		bin := int(math.Floor(p / binSize))
		peaks = append(peaks, Peak{Price: p, Vol: volumeBins[bin]})
	}

	// Sort by Volume Descending
	sort.Slice(peaks, func(i, j int) bool {
		return peaks[i].Vol > peaks[j].Vol
	})

	var finalPeaks []float64

	for _, p := range peaks {
		keep := true
		for _, existing := range finalPeaks {
			if math.Abs(p.Price - existing) < minDistPrice {
				keep = false
				break
			}
		}
		if keep {
			finalPeaks = append(finalPeaks, p.Price)
		}
	}

	// Classify into Upper/Lower relative to CURRENT price (or POC?)
	// "UpperPOC" usually means above the main POC or above current price?
	// Prompt: "Long Entry: current price enters LowerPOC range... Action: Long". This implies LowerPOC is below current price (Support).
	// Prompt: "Long Exit: price reaches UpperPOC". This implies UpperPOC is above current price (Resistance).
	// So Upper/Lower is relative to the CURRENT PRICE at the moment of calculation?
	// But calculation happens *before* strategy check.
	// Usually "Upper Value Area" vs "Lower Value Area".
	// Let's classify relative to the Global POC for structure, OR just return all peaks and let strategy filter.
	// Re-reading spec: "Find Local Maxima... manage UpperPOC and LowerPOC lists".
	// And "Long Entry: Condition 1: Current Price enters LowerPOC proximity".
	// If Upper/Lower are relative to global POC, then a LowerPOC could be above current price? Unlikely context.
	// Standard VP: Upper/Lower usually refers to Value Area High/Low or peaks above/below the main mode.
	// Given "Long Entry... LowerPOC", it strongly suggests Support levels.
	// I will split `finalPeaks` into those above Global POC and those below Global POC?
	// Or relative to Current Price?
	// If I use Current Price, the lists change dynamically even if the profile doesn't.
	// Let's use relative to Global POC for stability of the profile object, but actually the strategy uses them as support/resistance.
	// Let's stick to: UpperPOCs = Peaks > CurrentPrice, LowerPOCs = Peaks < CurrentPrice.
	// Wait, `CalculateVolumeProfile` takes `currentPrice`. So I can use that.

	for _, p := range finalPeaks {
		if p > currentPrice {
			upperPOCs = append(upperPOCs, p)
		} else if p < currentPrice {
			lowerPOCs = append(lowerPOCs, p)
		} else {
			// exact match, maybe add to both or ignore?
			// treat as lower (support) for safety?
			lowerPOCs = append(lowerPOCs, p)
		}
	}

	// Sort Upper ascending (closest to price first)
	sort.Float64s(upperPOCs)
	// Sort Lower descending (closest to price first)
	sort.Sort(sort.Reverse(sort.Float64Slice(lowerPOCs)))

	return VolumeProfile{
		POC:       pocPrice,
		UpperPOCs: upperPOCs,
		LowerPOCs: lowerPOCs,
	}
}
