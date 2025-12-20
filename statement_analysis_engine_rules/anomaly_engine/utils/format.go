package utils

import (
	"fmt"
	"strings"
)

// FormatAmount formats amount in Indian currency style
func FormatAmount(amount float64) string {
	if amount >= 100000 {
		lakhs := amount / 100000
		if lakhs >= 100 {
			crores := lakhs / 100
			return fmt.Sprintf("₹%.1fCr", crores)
		}
		return fmt.Sprintf("₹%.1fL", lakhs)
	}
	if amount >= 1000 {
		thousands := amount / 1000
		return fmt.Sprintf("₹%.1fK", thousands)
	}
	return fmt.Sprintf("₹%.0f", amount)
}

// FormatFloat formats float to string, removing trailing zeros
func FormatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(strings.TrimRight(
		strings.ReplaceAll(fmt.Sprintf("%.2f", f), ".00", ""), "0"), "."), "")
}

// FormatRatio formats ratio as string
func FormatRatio(ratio float64) string {
	return FormatFloat(ratio)
}

