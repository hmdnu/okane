package category

import (
	"fmt"
	"time"
)

func formatUint(value uint) string {
	return fmt.Sprintf("%d", value)
}

func formatDate(value time.Time) string {
	return value.Format("02/01/2006")
}
