package rules

import (
	"regexp"
	"strings"
)

// ClassifyCategory classifies the transaction category based on narration
func ClassifyCategory(narration string, merchant string) string {
	narration = strings.ToUpper(narration)
	merchant = strings.ToUpper(merchant)
	combined := narration + " " + merchant

	// Food Delivery patterns
	foodDeliveryPatterns := []string{
		"SWIGGY", "ZOMATO", "UBER EATS", "FOODPANDA",
		"FAASOS", "DOMINOS", "PIZZA HUT", "FOOD DELIVERY",
	}

	// Dining patterns
	diningPatterns := []string{
		"RESTAURANT", "CAFE", "HOTEL", "DINING",
		"FOOD COURT", "EATERY", "BAKERY", "COFFEE",
		"STARBUCKS", "CAFE COFFEE DAY", "CCD",
	}

	// Travel patterns
	travelPatterns := []string{
		"UBER", "OLA", "RAPIDO", "MAKE MY TRIP", "MMT",
		"GOIBIBO", "CLEARTRIP", "IRCTC", "BOOKING",
		"TRAVEL", "FLIGHT", "HOTEL", "CAB", "TAXI",
		"FUEL", "PETROL", "DIESEL", "BPCL", "HPCL",
		"IOCL", "SHELL", "RELIANCE", "PETROL PUMP",
	}

	// Shopping patterns
	shoppingPatterns := []string{
		"AMAZON", "FLIPKART", "MYNTRA", "NYkaa",
		"SHOPPING", "MALL", "STORE", "SHOP",
		"JEWELLERY", "TANISHQ", "MALABAR", "PC JEWELLER",
		"ELECTRONICS", "CROMA", "RELIANCE DIGITAL",
		"Vijay Sales", "GREAT EASTERN", "SHOPPERS STOP",
	}

	// Groceries patterns
	groceriesPatterns := []string{
		"GROCERY", "BIG BAZAAR", "DMART", "RELIANCE FRESH",
		"SPENCERS", "MORE", "FOOD BAZAAR", "HYPERCITY",
		"SUPERMARKET", "KIRANA", "GENERAL STORE",
	}

	// Bills & Utilities patterns
	billsPatterns := []string{
		"ELECTRICITY", "WATER", "GAS", "PHONE", "INTERNET",
		"MOBILE", "BROADBAND", "DTH", "CABLE", "INSURANCE",
		"PREMIUM", "LIC", "HDFC LIFE", "HLIC", "HLIC_INST", "HLIC INST",
		"MAXLIFE", "SBI LIFE", "ICICI PRUDENTIAL", "BAJAJ ALLIANZ",
		"PVVNL", "IGL", "AIRTEL", "JIO", "VODAFONE", "BSNL",
		"RECHARGE", "PREPAID", "POSTPAID", "BILL",
	}

	// Healthcare patterns
	healthcarePatterns := []string{
		"HOSPITAL", "CLINIC", "PHARMACY", "MEDICINE",
		"APOLLO", "FORTIS", "MAX", "MEDICOS", "MEDICAL",
		"HEALTH", "DOCTOR", "LAB", "DIAGNOSTIC",
	}

	// Education patterns
	educationPatterns := []string{
		"SCHOOL", "COLLEGE", "UNIVERSITY", "TUITION",
		"EDUCATION", "COURSE", "TRAINING", "INSTITUTE",
	}

	// Entertainment patterns
	entertainmentPatterns := []string{
		"MOVIE", "CINEMA", "THEATER", "NETFLIX", "AMAZON PRIME",
		"DISNEY", "HOTSTAR", "SPOTIFY", "MUSIC", "GAME",
		"PLAYSTORE", "GOOGLE PLAY",
	}

	// Investment patterns
	investmentPatterns := []string{
		"RD", "FD", "SIP", "MUTUAL FUND", "STOCK", "SHARE",
		"DEMAT", "INVESTMENT", "NPS", "PPF", "ELSS", "RD INSTALLMENT",
		"INDIAN CLEARING CORPORATION", "INDIAN CLEARING CORPORATION LIMITED",
		"INDIAN C LEARING CORPORATION", "INDIAN C LEARING CORPORATION LIMITED", // Handle typo with space
		"NSDL", "CDSL", "CLEARING CORPORATION",
	}

	// Dividend patterns (income from investments)
	dividendPatterns := []string{
		"DIV", "DIVIDEND", "DIVIDEND CREDIT", "DIV CR",
	}

	// Check Food Delivery
	for _, pattern := range foodDeliveryPatterns {
		if strings.Contains(combined, pattern) {
			return "Food_Delivery"
		}
	}

	// Check Dining
	for _, pattern := range diningPatterns {
		if strings.Contains(combined, pattern) {
			return "Dining"
		}
	}

	// Check Travel
	for _, pattern := range travelPatterns {
		if strings.Contains(combined, pattern) {
			return "Travel"
		}
	}

	// Check Shopping
	for _, pattern := range shoppingPatterns {
		if strings.Contains(combined, pattern) {
			return "Shopping"
		}
	}

	// Check Groceries
	for _, pattern := range groceriesPatterns {
		if strings.Contains(combined, pattern) {
			return "Groceries"
		}
	}

	// Check Bills & Utilities
	for _, pattern := range billsPatterns {
		if strings.Contains(combined, pattern) {
			return "Bills_Utilities"
		}
	}

	// Check Healthcare
	for _, pattern := range healthcarePatterns {
		if strings.Contains(combined, pattern) {
			return "Healthcare"
		}
	}

	// Check Education
	for _, pattern := range educationPatterns {
		if strings.Contains(combined, pattern) {
			return "Education"
		}
	}

	// Check Entertainment
	for _, pattern := range entertainmentPatterns {
		if strings.Contains(combined, pattern) {
			return "Entertainment"
		}
	}

	// Check Dividend (income from investments) - should be classified as income category
	for _, pattern := range dividendPatterns {
		if strings.Contains(combined, pattern) {
			return "Investment" // Dividends are investment income
		}
	}

	// Check Investment
	for _, pattern := range investmentPatterns {
		if strings.Contains(combined, pattern) {
			return "Investment"
		}
	}

	// Default category
	return "Other"
}

// ExtractMerchantName extracts merchant name from narration
func ExtractMerchantName(narration string) string {
	narration = strings.TrimSpace(narration)

	// Common merchant patterns
	merchants := map[string][]string{
		"Amazon":      {"AMAZON", "AMZN"},
		"Flipkart":    {"FLIPKART", "FKRT"},
		"Swiggy":      {"SWIGGY"},
		"Zomato":      {"ZOMATO"},
		"Uber":        {"UBER"},
		"Ola":         {"OLA"},
		"Netflix":     {"NETFLIX"},
		"Spotify":     {"SPOTIFY"},
		"Google Play": {"GOOGLE PLAY", "PLAYSTORE", "PLAY STORE"},
		"MakeMyTrip":  {"MAKE MY TRIP", "MMT", "MAKEMYTRIP"},
		"Croma":       {"CROMA"},
		"Tanishq":     {"TANISHQ"},
		"Apollo":      {"APOLLO"},
		"Reliance":    {"RELIANCE"},
		"DMart":       {"DMART", "D MART"},
		"Big Bazaar":  {"BIG BAZAAR"},
	}

	upperNarration := strings.ToUpper(narration)
	for merchant, patterns := range merchants {
		for _, pattern := range patterns {
			if strings.Contains(upperNarration, pattern) {
				return merchant
			}
		}
	}

	// Try to extract from UPI format: UPI-MERCHANT NAME-...
	re := regexp.MustCompile(`UPI-([^-@]+)`)
	matches := re.FindStringSubmatch(narration)
	if len(matches) > 1 {
		merchant := strings.TrimSpace(matches[1])
		// Clean up common suffixes
		merchant = strings.TrimSuffix(merchant, " -")
		merchant = strings.TrimSuffix(merchant, "-")
		return merchant
	}

	// Try to extract from IMPS/NEFT format
	re = regexp.MustCompile(`(?:IMPS|NEFT|RTGS)[- ]+([^-]+)`)
	matches = re.FindStringSubmatch(narration)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return "Unknown"
}
