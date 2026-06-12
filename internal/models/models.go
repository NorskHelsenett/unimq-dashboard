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

func NewPageData(vhosts []string, selected string, metrics *VhostMetrics, limits Limits) PageData {
	return PageData{Vhosts: vhosts, Selected: selected, Metrics: metrics, Limits: limits}
}

type RangeOption struct {
	Label string
	Value string
}

// NotifPageData is the data structure for the notifications page, containing the list of vhosts, the selected vhost, and its associated recipients and alarm rules.
// I'm unsure why the structure is like this, the original is from Claude.
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
