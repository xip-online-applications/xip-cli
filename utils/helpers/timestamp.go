package helpers

import (
	"strconv"
	"time"
)

func IntToTime(unix int) time.Time {
	t1, _ := strconv.ParseInt(strconv.Itoa(unix), 10, 64)
	return time.Unix(t1, 0)
}

func StringToTime(timeRaw string) time.Time {
	_timeParsed, _ := time.Parse(time.RFC3339, timeRaw)
	timeParsed := _timeParsed

	return timeParsed
}
