package timehelper

import "time"

const layout = "2006-01-02 15:04:05"

func ParseTimeInUTC(value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, time.UTC)
}
