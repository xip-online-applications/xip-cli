package helpers

import (
	"strconv"
	"time"
)

func IntToTime(unix int) time.Time {
	t1, _ := strconv.ParseInt(strconv.Itoa(unix), 10, 64)
	return time.Unix(t1, 0)
}
