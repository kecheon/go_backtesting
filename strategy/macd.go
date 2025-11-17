package strategy

import "github.com/markcheno/go-talib"

// CalculateMACD computes the MACD indicator values.
// It returns the raw, unpadded results from the talib library.
func CalculateMACD(closePrices []float64, fastPeriod, slowPeriod, signalPeriod int) (macd, signal, histogram []float64) {
	// Let the core talib function do the heavy lifting.
	// Note: The returned slices will be shorter than the input due to the initial periods
	// required for calculation. The calling code is responsible for alignment if needed.
	macd, signal, histogram = talib.Macd(closePrices, fastPeriod, slowPeriod, signalPeriod)
	return
}
