package models

import (
	"time"
)

type PageData struct {
	Vhosts   []string
	Selected string
	Metrics  *VhostMetrics
	Limits   Limits
}

type RangeOption struct {
	Label string
	Value string
}

type NotifPageData struct {
	Vhosts     []string
	Selected   string
	Recipients []Recipient
	Rules      []AlarmRule
}

var RangeDurations = map[string]time.Duration{
	"5m": 5 * time.Minute, "1h": time.Hour,
	"6h": 6 * time.Hour, "24h": 24 * time.Hour, "7d": 7 * 24 * time.Hour,
}

var TimeRanges = []RangeOption{
	{"5m", "5m"}, {"1h", "1h"}, {"6h", "6h"}, {"24h", "24h"}, {"7d", "7d"},
}
