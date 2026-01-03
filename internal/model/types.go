package model

type Candle struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

type Trade struct {
	ID         int
	EntryTime  int64
	EntryPrice float64
	Side       string // "Long" or "Short"
	ExitTime   int64
	ExitPrice  float64
	Size       float64
	PnL        float64
	Reason     string // "Target", "StopLoss", "Pattern"
	Active     bool
	// For analysis
	EntryPattern string
}
