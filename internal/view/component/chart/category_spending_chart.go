package chart

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hmdnu/okane/internal/dto"
)

var chartColors = []string{
	"#dc2626",
	"#2563eb",
	"#16a34a",
	"#d97706",
	"#7c3aed",
	"#0891b2",
	"#db2777",
	"#4b5563",
}

func categorySpendingChartState(spendings []dto.DashboardCategorySpending) string {
	labels := make([]string, 0, len(spendings))
	values := make([]float64, 0, len(spendings))
	colors := make([]string, 0, len(spendings))

	for index, spending := range spendings {
		labels = append(labels, spending.CategoryName)
		values = append(values, spending.Total)
		colors = append(colors, chartColors[index%len(chartColors)])
	}

	labelsJSON, _ := json.Marshal(labels)
	valuesJSON, _ := json.Marshal(values)
	colorsJSON, _ := json.Marshal(colors)

	return fmt.Sprintf(`{
		chart: null,
		labels: %s,
		values: %s,
		colors: %s,
		init() {
			if (!this.$refs.chart || typeof Chart === 'undefined') return
			this.chart = new Chart(this.$refs.chart, {
				type: 'doughnut',
				data: {
					labels: this.labels,
					datasets: [{
						data: this.values,
						backgroundColor: this.colors,
						borderColor: '#ffffff',
						borderWidth: 2
					}]
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					cutout: '62%%',
					plugins: {
						legend: { display: false },
						tooltip: {
							callbacks: {
								label(context) {
									return context.label + ': Rp ' + Number(context.raw || 0).toLocaleString('id-ID')
								}
							}
						}
					}
				}
			})
		}
	}`, labelsJSON, valuesJSON, colorsJSON)
}

func chartLegendClass(index int) string {
	return "h-3 w-3 shrink-0 rounded-full " + chartLegendColorClass(index)
}

func chartLegendColorClass(index int) string {
	classes := []string{
		"bg-red-600",
		"bg-blue-600",
		"bg-green-600",
		"bg-amber-600",
		"bg-violet-600",
		"bg-cyan-600",
		"bg-pink-600",
		"bg-gray-600",
	}

	return classes[index%len(classes)]
}

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
