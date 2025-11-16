package market

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

// Candle represents a single candlestick.
type Candle struct {
	Time  time.Time
	Open  float64
	High  float64
	Low   float64
	Close float64
	Vol   float64
}

// CandleSticks is a slice of Candles.
type CandleSticks []Candle

// ReadCandlesFromCSV reads a CSV file and returns a slice of CandleSticks.
func ReadCandlesFromCSV(filePath string) (CandleSticks, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	_, err = reader.Read() // Skip header
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("error: file is empty or contains only a header")
		}
		return nil, fmt.Errorf("error reading header: %w", err)
	}

	var candles CandleSticks
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading record: %w", err)
		}

		// CSV format: YYYY-MM-DD HH:MM:SS, open, high, low, close, volume
		t, err := time.Parse("2006-01-02 15:04:05", record[0])
		if err != nil {
			log.Printf("Error parsing timestamp, skipping record: %v", err)
			continue
		}

		open, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Printf("Error parsing open price, skipping record: %v", err)
			continue
		}
		high, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Printf("Error parsing high price, skipping record: %v", err)
			continue
		}
		low, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			log.Printf("Error parsing low price, skipping record: %v", err)
			continue
		}
		close, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			log.Printf("Error parsing close price, skipping record: %v", err)
			continue
		}
		vol, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			log.Printf("Error parsing volume, skipping record: %v", err)
			continue
		}

		candles = append(candles, Candle{
			Time:  t,
			Open:  open,
			High:  high,
			Low:   low,
			Close: close,
			Vol:   vol,
		})
	}
	return candles, nil
}
