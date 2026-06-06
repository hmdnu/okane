package setting

import (
	"fmt"
	"strconv"
	"time"
)

func formatDateTime(value string) string {
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.Format("02 January 2006 | 15:04")
		}
	}

	return value
}

func formatInt(value int) string {
	return strconv.Itoa(value)
}

func formatUint(value uint) string {
	return fmt.Sprintf("%d", value)
}

func formatFloat(value float64) string {
	if value == 0 {
		return ""
	}

	return strconv.FormatFloat(value, 'f', -1, 64)
}
