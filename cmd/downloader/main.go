package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

// Candle represents OHLCV data
type Candle struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

const (
	baseURL   = "https://www.okx.com"
	instId    = "BTC-USDT-SWAP"
	bar       = "1H"
	limit     = "100" // Max 100 per request usually for public, sometimes 300
	totalBars = 720   // 30 days * 24 hours
)

func main() {
	fmt.Println("Starting OKX Data Download...")

	var allCandles []Candle
	var after string // cursor for pagination (timestamp of last candle)

	// We need 720 bars. With limit 100, we need ~8 requests.
	for len(allCandles) < totalBars {
		url := fmt.Sprintf("%s/api/v5/market/candles?instId=%s&bar=%s&limit=%s", baseURL, instId, bar, limit)
		if after != "" {
			url += fmt.Sprintf("&after=%s", after)
		}

		// fmt.Println("Fetching:", url)
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var result struct {
			Code string     `json:"code"`
			Msg  string     `json:"msg"`
			Data [][]string `json:"data"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			panic(err)
		}

		if result.Code != "0" {
			panic(fmt.Sprintf("API Error: %s - %s", result.Code, result.Msg))
		}

		if len(result.Data) == 0 {
			break
		}

		for _, row := range result.Data {
			ts, _ := strconv.ParseInt(row[0], 10, 64)
			o, _ := strconv.ParseFloat(row[1], 64)
			h, _ := strconv.ParseFloat(row[2], 64)
			l, _ := strconv.ParseFloat(row[3], 64)
			c, _ := strconv.ParseFloat(row[4], 64)
			v, _ := strconv.ParseFloat(row[5], 64) // Usually contract volume
			// OKX SWAP volume is usually in contracts, sometimes volCcy is needed.
			// But for POC, any consistent volume metric works.

			allCandles = append(allCandles, Candle{
				Timestamp: ts,
				Open:      o,
				High:      h,
				Low:       l,
				Close:     c,
				Volume:    v,
			})
		}

		// Update cursor: use the timestamp of the last received candle
		lastTs := result.Data[len(result.Data)-1][0]
		after = lastTs

		// Sleep to be nice
		time.Sleep(200 * time.Millisecond)
	}

	// Sort candles by timestamp ascending (OKX returns descending)
	sort.Slice(allCandles, func(i, j int) bool {
		return allCandles[i].Timestamp < allCandles[j].Timestamp
	})

	// Trim to requested size if needed, though more is fine
	if len(allCandles) > totalBars {
		allCandles = allCandles[len(allCandles)-totalBars:]
	}

	// Save to CSV
	file, err := os.Create("btc_1h.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"timestamp", "open", "high", "low", "close", "volume"})
	for _, c := range allCandles {
		writer.Write([]string{
			fmt.Sprintf("%d", c.Timestamp),
			fmt.Sprintf("%f", c.Open),
			fmt.Sprintf("%f", c.High),
			fmt.Sprintf("%f", c.Low),
			fmt.Sprintf("%f", c.Close),
			fmt.Sprintf("%f", c.Volume),
		})
	}

	fmt.Printf("Successfully saved %d candles to btc_1h.csv\n", len(allCandles))
}
