package strategy

import "fmt"

// longEntryConditions holds the registry for long entry condition functions.
var longEntryConditions = map[string]EntryCondition{
	"default": DefaultLongCondition,
	"macd":    MACDLongCondition,
}

// shortEntryConditions holds the registry for short entry condition functions.
var shortEntryConditions = map[string]EntryCondition{
	"default": DefaultShortCondition,
	"macd":    MACDShortCondition,
}

// GetEntryCondition retrieves a long or short entry condition function from the registry.
// direction must be "long" or "short".
func GetEntryCondition(name string, direction string) (EntryCondition, error) {
	var registry map[string]EntryCondition
	if direction == "long" {
		registry = longEntryConditions
	} else if direction == "short" {
		registry = shortEntryConditions
	} else {
		return nil, fmt.Errorf("invalid direction for entry condition: %s", direction)
	}

	condition, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("no %s entry condition found for name: %s", direction, name)
	}
	return condition, nil
}
