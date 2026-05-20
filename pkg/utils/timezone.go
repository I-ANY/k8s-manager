package utils

import "time"

var timezone = "Asia/Shanghai"

// SetTimezone sets the global timezone used by TimenowInTimezone.
func SetTimezone(tz string) {
	if tz != "" {
		timezone = tz
	}
}

// TimenowInTimezone returns the current time in the configured timezone.
func TimenowInTimezone() time.Time {
	loc, _ := time.LoadLocation(timezone)
	return time.Now().In(loc)
}
