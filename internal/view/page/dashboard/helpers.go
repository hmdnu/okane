package dashboard

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func formatCurrency(amount float64) string {
	negative := amount < 0
	amount = math.Abs(amount)

	whole := int64(amount)
	cents := int64(math.Round((amount - float64(whole)) * 100))
	if cents == 100 {
		whole++
		cents = 0
	}

	formatted := "Rp " + formatThousands(whole)
	if cents > 0 {
		formatted += fmt.Sprintf(",%02d", cents)
	}
	if negative {
		return "-" + formatted
	}

	return formatted
}

func transactionModalState(hasErrors bool) string {
	if hasErrors {
		return "{ transactionModalOpen: true, deleteModalOpen: false, deleteAction: '' }"
	}

	return "{ transactionModalOpen: false, deleteModalOpen: false, deleteAction: '' }"
}

func formatThousands(value int64) string {
	digits := strconv.FormatInt(value, 10)
	if len(digits) <= 3 {
		return digits
	}

	groups := make([]string, 0, len(digits)/3+1)
	for len(digits) > 3 {
		groups = append([]string{digits[len(digits)-3:]}, groups...)
		digits = digits[:len(digits)-3]
	}
	groups = append([]string{digits}, groups...)

	return strings.Join(groups, ".")
}

func formatDateInput(year int, month int, day int) string {
	if year == 0 || month == 0 || day == 0 {
		return ""
	}

	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func todayDateInput() string {
	now := time.Now()
	return formatDateInput(now.Year(), int(now.Month()), now.Day())
}

func formatUint(value uint) string {
	return fmt.Sprintf("%d", value)
}

func formatDate(value time.Time) string {
	return value.Format("02/01/2006")
}

func formatDateTime(value time.Time) string {
	return value.Format("02 January 2006 | 15:04")
}

func formatTime(value time.Time) string {
	return value.Format("15:04")
}

func formatType(value string) string {
	switch value {
	case "income":
		return "Income"
	case "expense":
		return "Expense"
	case "initial_balance":
		return "Initial Balance"
	default:
		return value
	}
}

func transactionTypeClass(value string) string {
	switch value {
	case "income", "initial_balance":
		return "rounded-full bg-green-100 px-2 py-1 text-xs font-medium text-green-700"
	case "expense":
		return "rounded-full bg-red-100 px-2 py-1 text-xs font-medium text-red-700"
	default:
		return "rounded-full bg-gray-100 px-2 py-1 text-xs font-medium text-gray-700"
	}
}
