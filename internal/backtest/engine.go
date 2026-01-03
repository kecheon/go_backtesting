package backtest

import (
	"fmt"
	"math"
	"go-backtesting/internal/analysis"
	"go-backtesting/internal/model"
)

type Config struct {
	LookbackPeriod int
	BinSizePct     float64
	ValueAreaPct   float64
	POCProximity   float64 // % distance to consider "near"
}

type Engine struct {
	Candles []model.Candle
	Config  Config

	Trades       []*model.Trade
	ActiveLongs  []*model.Trade
	ActiveShorts []*model.Trade

	TradeIDCounter int
}

func NewEngine(candles []model.Candle, cfg Config) *Engine {
	return &Engine{
		Candles: candles,
		Config:  cfg,
	}
}

func (e *Engine) Run() {
	// Need at least LookbackPeriod candles to start
	if len(e.Candles) < e.Config.LookbackPeriod {
		fmt.Println("Not enough data for lookback period")
		return
	}

	for i := e.Config.LookbackPeriod; i < len(e.Candles); i++ {
		// 1. Prepare Data
		// Window: [i-Lookback : i] (excludes current candle i? No, usually profile is built on PAST data to trade CURRENT).
		// So window is [i - Lookback : i]. The current candle is e.Candles[i].
		window := e.Candles[i-e.Config.LookbackPeriod : i]
		currentCandle := e.Candles[i]

		// Previous candle for pattern check
		// patterns.CheckPatterns uses slice, expects last one to be "current".
		// We pass window + current? Or just window?
		// "Signal on Close". We analyze the just-closed candle (index i) to trade on the Open of i+1?
		// Spec: "Action: 다음 캔들 시가에 롱 포지션 진입." (Enter on Next Open)
		// So we analyze candle `i`.

		// VP is built on PAST data. Does it include candle `i`?
		// Typically indicators are calculated on closed candles.
		// If we are at index `i`, it acts as the "Just Closed" candle.
		// VP should probably be built on window ending at `i`.

		vp := analysis.CalculateVolumeProfile(window, e.Config.BinSizePct, e.Config.ValueAreaPct)

		// Check Patterns on the current candle `i` (and its predecessor for Engulfing)
		// We need at least 2 candles for patterns.
		patternWindow := []model.Candle{e.Candles[i-1], e.Candles[i]}
		isBull, isBear, patternName := analysis.CheckPatterns(patternWindow)

		// 2. Manage Exits (Logic on Open of i? Or Close of i?)
		// We process exits based on price action of candle `i`?
		// Spec: "Target UpperPOC 도달 시..." (When price hits UpperPOC).
		// We can check if `i` High/Low hit the targets.

		e.checkExits(currentCandle, vp, isBull, isBear)

		// 3. Check Entries
		// If entry signal found, we enter on Open of `i+1` (next loop iteration? or simulate now?)
		// We usually record the trade as entering at `currentCandle.Close` (approximation) or `nextCandle.Open`.
		// Since we iterate `i`, `i` is the latest known candle. We can't know `i+1` Open yet unless we look ahead.
		// Standard backtest: Signal at `i` Close -> Trade execution at `i+1` Open.
		// But in this loop, we only see `i`.
		// Correct approach: Set "Pending Order" for next tick?
		// Or simpler: We look ahead to `i+1` if it exists.

		if i+1 < len(e.Candles) {
			nextOpen := e.Candles[i+1].Open
			e.checkEntries(currentCandle, nextOpen, e.Candles[i+1].Timestamp, vp, isBull, isBear, patternName)
		}
	}
}

func (e *Engine) checkEntries(curr model.Candle, entryPrice float64, entryTime int64, vp analysis.VolumeProfile, isBull, isBear bool, patternName string) {
	// Proximity Threshold
	prox := e.Config.POCProximity / 100.0 // e.g. 0.002

	// Long Entry: Near VAL + Bullish Pattern
	// Condition 1: Current Price near LowerPOC (VAL)
	// "Price" usually means Close, or Low/High range? Spec says "현재 가격이... 진입".
	// Let's use Close.
	// Near VAL: abs(Close - VAL) / VAL <= prox ??
	// Or price is within [VAL * (1-prox), VAL * (1+prox)]?
	// Or specifically "Entered the proximity range".
	// Let's use: Abs(curr.Close - vp.VAL) <= (vp.VAL * prox)

	distToVal := math.Abs(curr.Close - vp.VAL)
	isNearVal := distToVal <= (vp.VAL * prox)

	if isNearVal && isBull {
		// Doji check is inside CheckPatterns (returns false if Doji)
		// Enter Long
		t := &model.Trade{
			ID:         e.nextID(),
			EntryTime:  entryTime,
			EntryPrice: entryPrice,
			Side:       "Long",
			Size:       1.0, // Fixed size for now
			Active:     true,
			EntryPattern: patternName,
		}
		e.ActiveLongs = append(e.ActiveLongs, t)
		e.Trades = append(e.Trades, t)
	}

	// Short Entry: Near VAH + Bearish Pattern
	distToVah := math.Abs(curr.Close - vp.VAH)
	isNearVah := distToVah <= (vp.VAH * prox)

	if isNearVah && isBear {
		t := &model.Trade{
			ID:         e.nextID(),
			EntryTime:  entryTime,
			EntryPrice: entryPrice,
			Side:       "Short",
			Size:       1.0,
			Active:     true,
			EntryPattern: patternName,
		}
		e.ActiveShorts = append(e.ActiveShorts, t)
		e.Trades = append(e.Trades, t)
	}
}

func (e *Engine) checkExits(curr model.Candle, vp analysis.VolumeProfile, isBull, isBear bool) {
	// Long Exits
	// 1. Target: UpperPOC (VAH). Spec: "Target UpperPOC 도달 시 90%..." (User: 100%).
	// Check if High >= VAH.
	// 2. StopLoss: Close below VAL (LowerPOC). "LowerPOC 라인을 종가 기준으로 하향 이탈"
	// 3. Pattern: Bearish Pattern implies exit.

	// Note: We iterate backwards to allow removal from slice? Or rebuild slice.
	var activeLongs []*model.Trade
	for _, t := range e.ActiveLongs {
		closed := false

		// Stop Loss: Close < VAL
		// Or "Entry Low" break? "진입 캔들의 저가를 이탈할 경우" -> We don't track entry candle Low in Trade struct.
		// We'll stick to "Close < VAL" for now as it's cleaner, or update Trade to store EntryLow.
		// For now: Stop Loss if Close < VAL.
		if curr.Close < vp.VAL {
			e.closeTrade(t, curr.Close, curr.Timestamp, "StopLoss")
			closed = true
		} else if curr.High >= vp.VAH {
			// Take Profit
			e.closeTrade(t, vp.VAH, curr.Timestamp, "Target") // Limit fill at VAH
			closed = true
		} else if isBear {
			// Pattern Exit
			e.closeTrade(t, curr.Close, curr.Timestamp, "Pattern")
			closed = true
		}

		if !closed {
			activeLongs = append(activeLongs, t)
		}
	}
	e.ActiveLongs = activeLongs

	// Short Exits
	var activeShorts []*model.Trade
	for _, t := range e.ActiveShorts {
		closed := false

		// Stop Loss: Close > VAH
		if curr.Close > vp.VAH {
			e.closeTrade(t, curr.Close, curr.Timestamp, "StopLoss")
			closed = true
		} else if curr.Low <= vp.VAL {
			// Take Profit
			e.closeTrade(t, vp.VAL, curr.Timestamp, "Target") // Limit fill at VAL
			closed = true
		} else if isBull {
			// Pattern Exit
			e.closeTrade(t, curr.Close, curr.Timestamp, "Pattern")
			closed = true
		}

		if !closed {
			activeShorts = append(activeShorts, t)
		}
	}
	e.ActiveShorts = activeShorts
}

func (e *Engine) closeTrade(t *model.Trade, price float64, time int64, reason string) {
	t.ExitPrice = price
	t.ExitTime = time
	t.Reason = reason
	t.Active = false

	if t.Side == "Long" {
		t.PnL = (t.ExitPrice - t.EntryPrice) * t.Size
	} else {
		t.PnL = (t.EntryPrice - t.ExitPrice) * t.Size
	}
}

func (e *Engine) nextID() int {
	e.TradeIDCounter++
	return e.TradeIDCounter
}
