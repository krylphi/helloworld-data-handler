package utils

import "time"

// DateFromUnixMillis returns date from Unix timestamp with ms precision
func DateFromUnixMillis(ts int64) string {
	t := time.Unix(0, ts*1000000)
	return t.Format("2006-01-02")
}
