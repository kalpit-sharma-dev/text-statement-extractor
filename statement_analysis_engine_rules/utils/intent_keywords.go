package utils

import (
	"strings"
)

// IntentKeyword represents intent keywords that indicate category
type IntentKeyword struct {
	Keyword    string
	Category   string
	Confidence float64
}

// IntentKeywords is a map of intent keywords to categories
// Layer 5: Intent Keywords (Supporting Evidence)
var IntentKeywords = []IntentKeyword{
	// Bills & Utilities
	{Keyword: "BILL", Category: "Bills_Utilities", Confidence: 0.3},
	{Keyword: "UTILITY PAYMENT", Category: "Bills_Utilities", Confidence: 0.4},
	{Keyword: "RECHARGE", Category: "Bills_Utilities", Confidence: 0.3},
	{Keyword: "PREPAID", Category: "Bills_Utilities", Confidence: 0.3},
	{Keyword: "POSTPAID", Category: "Bills_Utilities", Confidence: 0.3},
	{Keyword: "RENT", Category: "Bills_Utilities", Confidence: 0.4},
	{Keyword: "MAINTENANCE", Category: "Bills_Utilities", Confidence: 0.3},

	// Investment
	{Keyword: "INSTALLMENT", Category: "Investment", Confidence: 0.3},
	{Keyword: "SIP", Category: "Investment", Confidence: 0.4},
	{Keyword: "RD", Category: "Investment", Confidence: 0.4},
	{Keyword: "FD", Category: "Investment", Confidence: 0.4},
	{Keyword: "MUTUAL FUND", Category: "Investment", Confidence: 0.5},
	{Keyword: "STOCK", Category: "Investment", Confidence: 0.4},
	{Keyword: "SHARE", Category: "Investment", Confidence: 0.4},
	{Keyword: "DIVIDEND", Category: "Investment", Confidence: 0.5},

	// Loan
	{Keyword: "EMI", Category: "Loan", Confidence: 0.5},
	{Keyword: "LOAN", Category: "Loan", Confidence: 0.4},
	{Keyword: "OVERDUE", Category: "Loan", Confidence: 0.4},
	{Keyword: "RECOVERED", Category: "Loan", Confidence: 0.3},

	// Fuel
	{Keyword: "FUEL", Category: "Fuel", Confidence: 0.4},
	{Keyword: "PETROL", Category: "Fuel", Confidence: 0.4},
	{Keyword: "DIESEL", Category: "Fuel", Confidence: 0.4},
	{Keyword: "SERVICE STATION", Category: "Fuel", Confidence: 0.3},
	{Keyword: "PETROL PUMP", Category: "Fuel", Confidence: 0.4},

	// Travel
	{Keyword: "TRAVEL", Category: "Travel", Confidence: 0.3},
	{Keyword: "FLIGHT", Category: "Travel", Confidence: 0.4},
	{Keyword: "HOTEL", Category: "Travel", Confidence: 0.3},
	{Keyword: "CAB", Category: "Travel", Confidence: 0.3},
	{Keyword: "TAXI", Category: "Travel", Confidence: 0.3},
	{Keyword: "BOOKING", Category: "Travel", Confidence: 0.3},

	// Food Delivery
	{Keyword: "ORDER", Category: "Food_Delivery", Confidence: 0.2},
	{Keyword: "FOOD DELIVERY", Category: "Food_Delivery", Confidence: 0.4},
	{Keyword: "ONLINE FOOD", Category: "Food_Delivery", Confidence: 0.3},

	// Dining
	{Keyword: "RESTAURANT", Category: "Dining", Confidence: 0.3},
	{Keyword: "CAFE", Category: "Dining", Confidence: 0.3},
	{Keyword: "DINING", Category: "Dining", Confidence: 0.3},
	{Keyword: "EATERY", Category: "Dining", Confidence: 0.3},
	{Keyword: "BAKERY", Category: "Dining", Confidence: 0.3},

	// Shopping
	{Keyword: "SHOPPING", Category: "Shopping", Confidence: 0.2},
	{Keyword: "PURCHASE", Category: "Shopping", Confidence: 0.2},
	{Keyword: "STORE", Category: "Shopping", Confidence: 0.2},
	{Keyword: "SHOP", Category: "Shopping", Confidence: 0.2},

	// Groceries
	{Keyword: "GROCERY", Category: "Groceries", Confidence: 0.3},
	{Keyword: "GROCERIES", Category: "Groceries", Confidence: 0.3},
	{Keyword: "SUPERMARKET", Category: "Groceries", Confidence: 0.3},
	{Keyword: "KIRANA", Category: "Groceries", Confidence: 0.3},
	{Keyword: "VEGETABLE", Category: "Groceries", Confidence: 0.3},
	{Keyword: "FRUIT", Category: "Groceries", Confidence: 0.3},

	// Healthcare
	{Keyword: "MEDICAL", Category: "Healthcare", Confidence: 0.3},
	{Keyword: "PHARMACY", Category: "Healthcare", Confidence: 0.4},
	{Keyword: "HOSPITAL", Category: "Healthcare", Confidence: 0.4},
	{Keyword: "CLINIC", Category: "Healthcare", Confidence: 0.3},
	{Keyword: "DOCTOR", Category: "Healthcare", Confidence: 0.3},
	{Keyword: "HEALTH", Category: "Healthcare", Confidence: 0.2},

	// Entertainment
	{Keyword: "MOVIE", Category: "Entertainment", Confidence: 0.3},
	{Keyword: "CINEMA", Category: "Entertainment", Confidence: 0.3},
	{Keyword: "MUSIC", Category: "Entertainment", Confidence: 0.2},
	{Keyword: "GAME", Category: "Entertainment", Confidence: 0.2},
	{Keyword: "GAMING", Category: "Entertainment", Confidence: 0.3},

	// Education
	{Keyword: "SCHOOL", Category: "Education", Confidence: 0.3},
	{Keyword: "COLLEGE", Category: "Education", Confidence: 0.3},
	{Keyword: "UNIVERSITY", Category: "Education", Confidence: 0.3},
	{Keyword: "TUITION", Category: "Education", Confidence: 0.4},
	{Keyword: "EDUCATION", Category: "Education", Confidence: 0.3},
	{Keyword: "COURSE", Category: "Education", Confidence: 0.3},
}

// DetectIntentKeywords detects intent keywords in narration
// Returns map of category -> confidence score
func DetectIntentKeywords(narration string) map[string]float64 {
	upper := strings.ToUpper(narration)
	scores := make(map[string]float64)

	for _, intent := range IntentKeywords {
		if strings.Contains(upper, intent.Keyword) {
			// Use maximum confidence if multiple keywords match same category
			if currentScore, exists := scores[intent.Category]; !exists || intent.Confidence > currentScore {
				scores[intent.Category] = intent.Confidence
			}
		}
	}

	return scores
}

