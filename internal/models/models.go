package models

import "time"

type RangeOption struct {
	Label string
	Value string
}

var RangeDurations = map[string]time.Duration{
	"5m": 5 * time.Minute, "1h": time.Hour,
	"6h": 6 * time.Hour, "24h": 24 * time.Hour, "7d": 7 * 24 * time.Hour,
}

var TimeRanges = []RangeOption{
	{"5m", "5m"}, {"1h", "1h"}, {"6h", "6h"}, {"24h", "24h"}, {"7d", "7d"},
}
