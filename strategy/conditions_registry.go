package strategy

import "fmt"

// longEntryConditions holds the registry for long entry condition functions.
var longEntryConditions = map[string]EntryCondition{
	"default":        DefaultLongCondition,
	"macd":           MACDLongCondition,
	"bbw":            BBWLongCondition,
	"combined":       CombinedLongCondition,
	"inverse":        InverseLongCondition,
	"dmi":            DMILongCondition,
	"volume_cluster": VolumeClusterLongEntry,
}

// shortEntryConditions holds the registry for short entry condition functions.
var shortEntryConditions = map[string]EntryCondition{
	"default":        DefaultShortCondition,
	"macd":           MACDShortCondition,
	"bbw":            BBWShortCondition,
	"combined":       CombinedShortCondition,
	"inverse":        InverseShortCondition,
	"dmi":            DMIShortCondition,
	"volume_cluster": VolumeClusterShortEntry,
}

// GetEntryCondition retrieves a long or short entry condition function from the registry.
// direction must be "long" or "short".
func GetEntryCondition(name string, direction string) (EntryCondition, error) {
	var registry map[string]EntryCondition
	switch direction {
	case "long":
		registry = longEntryConditions
	case "short":
		registry = shortEntryConditions
	default:
		return nil, fmt.Errorf("invalid direction for entry condition: %s", direction)
	}

	condition, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("no %s entry condition found for name: %s", direction, name)
	}
	return condition, nil
}
