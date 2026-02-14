package octobot_commons

import "time"

var localTimezone = time.Now().Location()

func ConvertTimestampToDatetime(timestamp float64, timeFormat string, localTimezoneFlag bool) string {
    tz := time.UTC
    if localTimezoneFlag {
        tz = localTimezone
    }
    return time.Unix(int64(timestamp), 0).In(tz).Format(timeFormat)
}

func ConvertTimestampsToDatetime(timestamps []float64, timeFormat string, localTimezoneFlag bool) []string {
    out := make([]string, 0, len(timestamps))
    for _, ts := range timestamps {
        out = append(out, ConvertTimestampToDatetime(ts, timeFormat, localTimezoneFlag))
    }
    return out
}

func IsValidTimestamp(timestamp float64) bool {
    if timestamp == 0 {
        return true
    }
    t := time.Unix(int64(timestamp), 0)
    return !t.IsZero()
}

func GetNowTime(timeFormat string, localTimezoneFlag bool) string {
    tz := time.UTC
    if localTimezoneFlag {
        tz = localTimezone
    }
    return time.Now().In(tz).Format(timeFormat)
}

func DatetimeToTimestamp(dateTimeStr string, dateTimeFormat string, localTimezoneFlag bool) float64 {
    return float64(CreateDatetimeFromString(dateTimeStr, dateTimeFormat, localTimezoneFlag).Unix())
}

func CreateDatetimeFromString(dateTimeStr string, dateTimeFormat string, localTimezoneFlag bool) time.Time {
    t, err := time.Parse(dateTimeFormat, dateTimeStr)
    if err != nil {
        return time.Time{}
    }
    tz := time.UTC
    if localTimezoneFlag {
        tz = localTimezone
    }
    return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}
