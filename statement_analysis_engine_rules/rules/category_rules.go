package rules

import (
	"classify/statement_analysis_engine_rules/utils"
	"regexp"
	"strings"
)

// ClassifyCategory classifies the transaction category based on narration
// Enhanced with tokenization and gateway detection
func ClassifyCategory(narration string, merchant string) string {
	return ClassifyCategoryWithAmount(narration, merchant, 0.0)
}

// ClassifyCategoryWithAmount classifies the transaction category with amount for charge detection
func ClassifyCategoryWithAmount(narration string, merchant string, amount float64) string {
	originalNarration := narration
	narration = strings.ToUpper(narration)
	merchant = strings.ToUpper(merchant)
	combined := narration + " " + merchant

	// Tokenize narration for better pattern matching
	tokens := utils.Tokenize(originalNarration)

	// Food Delivery patterns
	foodDeliveryPatterns := []string{
		"SWIGGY", "ZOMATO", "UBER EATS", "FOODPANDA",
		"FAASOS", "DOMINOS", "PIZZA HUT", "FOOD DELIVERY",
	}

	// Check tokens for wallet-based food delivery
	wallet := utils.DetectWallet(tokens)
	if wallet != "" {
		// If narration suggests food delivery through wallet
		for _, token := range tokens {
			if strings.Contains(token, "SWIGGY") || strings.Contains(token, "ZOMATO") {
				return "Food_Delivery"
			}
		}
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
		"GOIBIBO", "CLEARTRIP", "IRCTC", "IRCTCIPAY", "RAZPIRCTCIPAY", "BOOKING",
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
		"SIMPL", "SIMPL TECHNOLOGI", "GETSIMPL", // Simpl buy now pay later
	}

	// Groceries patterns
	groceriesPatterns := []string{
		"GROCERY", "BIG BAZAAR", "DMART", "RELIANCE FRESH",
		"SPENCERS", "MORE", "FOOD BAZAAR", "HYPERCITY",
		"SUPERMARKET", "KIRANA", "GENERAL STORE",
	}

	// Universal Bill Payment Aggregators/Gateways
	billGateways := []string{
		"BILLDESK", "BILLDK", "PAYU", "RAZORPAY", "RAZP",
		"CCAVENUE", "PAYTM", "AMAZONPAY", "PHONEPE", "GPAY",
		"PAYGOV", "BBPS", "WHDF", "SBIPG", "AXISPG", "ICICIPG",
		"KOTAKPG", "YESPG",
	}
	
	// Electricity Bill Patterns
	electricityPatterns := []string{
		"ELECTRICITY", "BSESR", "BSES", "TATAPOWER", "TORRENTPOWER",
		"MSEB", "MSEDCL", "UPPCL", "DVVNL", "BSESYAMUNA", "BSESRAJDHANI",
		"MAHARASHTRA STATE EL", "MAHARASHTRA STATE ELECTRICITY",
		"EL", "POWER", "DISCOM",
	}
	
	// Gas (PNG/LPG) Patterns
	gasPatterns := []string{
		"GAS", "INDRAPRASTHAGA", "IGL", "MGL", "ADANIGAS", "GUJGAS",
		"HPGAS", "BPCL GAS", "LPG",
	}
	
	// Water Bill Patterns
	waterPatterns := []string{
		"WATER", "DELHIJALBOARD", "BWSSB", "MCGM", "JAL BOARD",
		"WATER BOARD", "WATER SUPPLY",
	}
	
	// Telecom & Internet Patterns
	telecomPatterns := []string{
		"PHONE", "MOBILE", "BROADBAND", "INTERNET", "AIRTEL", "JIO",
		"VODAFONE", "IDEA", "BSNL", "ACTFIBERNET", "HATHWAY", "TIKONA",
		"RECHARGE", "PREPAID", "POSTPAID", "TELECOM",
	}
	
	// DTH/TV Patterns
	dthPatterns := []string{
		"DTH", "CABLE", "TATASKY", "AIRTELDTH", "DISH", "SUNTV",
		"VIDEOCON D2H", "D2H",
	}
	
	// Transport & Toll Patterns
	tollPatterns := []string{
		"FASTAG", "NHAI", "TOLL", "PAYTMFASTAG", "ICICIFASTAG",
		"HDFCBANKFASTAG", "AXISFASTAG", "SBIFASTAG",
	}
	
	// Government Payment Patterns
	governmentPatterns := []string{
		"PAYGOV", "GOVT", "GOVERNMENT", "GST", "INCOMETAX", "PASSPORT",
		"CHALLAN", "TRAFFIC CHALLAN", "ROAD TAX", "PROPERTY TAX",
		"PROFESSIONAL TAX",
	}
	
	// Insurance Premium Patterns
	insurancePatterns := []string{
		"INSURANCE", "PREMIUM", "LIC", "HDFC LIFE", "HLIC", "HLIC_INST", "HLIC INST",
		"MAXLIFE", "SBI LIFE", "ICICI PRUDENTIAL", "BAJAJ ALLIANZ",
		"STANDARDLIFE", "SBILIFE", "ICICIPRULIFE",
	}
	
	// Credit Card Payment Patterns
	creditCardPatterns := []string{
		"CREDITCARD", "CREDIT CARD", "CARDBILL", "CARDPAYMENT",
		"HDFCCARD", "SBICARD", "AXISCARD", "ICICICARD", "KOTAKCARD",
	}
	
	// Loan EMI Patterns
	loanEmiPatterns := []string{
		"LOAN", "EMI", "HDFCLOAN", "SBILOAN", "BAJAJFINSERV", "BAJAJFIN",
		"OVERDUE LOAN", "LOAN RECOVERED", "EMI RECOVERY", "REPAYMENT",
	}
	
	// Housing/Maintenance Patterns
	housingPatterns := []string{
		"MAINTENANCE", "SOCIETY", "APARTMENT", "ASSOCIATION",
		"HOUSING", "SOCIETY MAINTENANCE",
	}
	
	// Tax Payment Patterns
	taxPatterns := []string{
		"TAX", "GST PAYMENT", "INCOME TAX", "PROPERTY TAX",
		"PROFESSIONAL TAX", "ROAD TAX", "TRAFFIC CHALLAN",
	}
	
	// Combined Bills & Utilities patterns (for backward compatibility)
	billsPatterns := []string{
		"ELECTRICITY", "WATER", "GAS", "PHONE", "INTERNET",
		"MOBILE", "BROADBAND", "DTH", "CABLE", "INSURANCE",
		"PREMIUM", "LIC", "HDFC LIFE", "HLIC", "HLIC_INST", "HLIC INST",
		"MAXLIFE", "SBI LIFE", "ICICI PRUDENTIAL", "BAJAJ ALLIANZ",
		"PVVNL", "IGL", "AIRTEL", "JIO", "VODAFONE", "BSNL",
		"RECHARGE", "PREPAID", "POSTPAID", "BILL",
		"BILLDK", "WHDF", // BillDesk gateway indicators
		"MAHARASHTRA STATE EL", "MAHARASHTRA STATE ELECTRICITY",
		"MSEDCL", "MAHARASHTRA STATE", "EL", // Electricity board patterns
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
		"ZERODHA", "ZERODHA BROKING", "ZERODHA BROKING LTD", "ZERODHABROKING",
		"BROKING", "BROKING LTD", "HSL SEC", "HSL", "SEC", // Stock broking companies
		"ANGEL BROKING", "ICICI SECURITIES", "HDFC SECURITIES", "KOTAK SECURITIES",
		"SHAREKHAN", "MOTILAL OSWAL", "IIFL", "5PAISA",
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

	// Check Shopping (enhanced with tokenization)
	for _, pattern := range shoppingPatterns {
		if strings.Contains(combined, pattern) {
			return "Shopping"
		}
	}
	// Also check tokens for shopping-related merchants
	for _, token := range tokens {
		if strings.Contains(token, "SIMPL") || strings.Contains(token, "GETSIMPL") ||
			strings.Contains(token, "AMAZON") || strings.Contains(token, "FLIPKART") {
			return "Shopping"
		}
	}

	// Check Groceries
	for _, pattern := range groceriesPatterns {
		if strings.Contains(combined, pattern) {
			return "Groceries"
		}
	}

	// Check Bills & Utilities (comprehensive bill payment detection)
	// Step 1: Check for bill payment gateways/aggregators
	isBillPayment := false
	for _, gw := range billGateways {
		if strings.Contains(combined, gw) {
			isBillPayment = true
			break
		}
	}
	
	// Step 2: Check for generic bill keywords
	if strings.Contains(combined, "BILL") || strings.Contains(combined, "BBPS") ||
		strings.Contains(combined, "ECS BILL") || strings.Contains(combined, "NACH BILL") ||
		strings.Contains(combined, "AUTO BILL") || strings.Contains(combined, "FUNDSTRANSFER-BILL") ||
		strings.Contains(combined, "ONLINE BILL") || strings.Contains(combined, "PAYMENT TO BILLER") ||
		strings.Contains(combined, "UTILITY PAYMENT") {
		isBillPayment = true
	}
	
	// Step 3: If bill payment detected, classify by specific category
	if isBillPayment {
		// Check for specific utility types
		
		// Electricity
		for _, pattern := range electricityPatterns {
			if strings.Contains(combined, pattern) {
				return "Bills_Utilities"
			}
		}
		
		// Gas
		for _, pattern := range gasPatterns {
			if strings.Contains(combined, pattern) {
				return "Bills_Utilities"
			}
		}
		
		// Water
		for _, pattern := range waterPatterns {
			if strings.Contains(combined, pattern) {
				return "Bills_Utilities"
			}
		}
		
		// Telecom
		for _, pattern := range telecomPatterns {
			if strings.Contains(combined, pattern) {
				return "Bills_Utilities"
			}
		}
		
		// DTH
		for _, pattern := range dthPatterns {
			if strings.Contains(combined, pattern) {
				return "Bills_Utilities"
			}
		}
		
		// Toll/Fastag
		for _, pattern := range tollPatterns {
			if strings.Contains(combined, pattern) {
				return "Bills_Utilities"
			}
		}
		
		// Government payments
		for _, pattern := range governmentPatterns {
			if strings.Contains(combined, pattern) {
				return "Bills_Utilities"
			}
		}
		
		// Insurance
		for _, pattern := range insurancePatterns {
			if strings.Contains(combined, pattern) {
				return "Bills_Utilities"
			}
		}
		
		// Credit Card
		for _, pattern := range creditCardPatterns {
			if strings.Contains(combined, pattern) {
				return "Bills_Utilities"
			}
		}
		
		// Loan EMI
		for _, pattern := range loanEmiPatterns {
			if strings.Contains(combined, pattern) {
				return "Bills_Utilities"
			}
		}
		
		// Housing/Maintenance
		for _, pattern := range housingPatterns {
			if strings.Contains(combined, pattern) {
				return "Bills_Utilities"
			}
		}
		
		// Tax payments
		for _, pattern := range taxPatterns {
			if strings.Contains(combined, pattern) {
				return "Bills_Utilities"
			}
		}
		
		// Default: if bill payment gateway detected but no specific category, still Bills_Utilities
		return "Bills_Utilities"
	}
	
	// Legacy check for backward compatibility
	for _, pattern := range billsPatterns {
		if strings.Contains(combined, pattern) {
			return "Bills_Utilities"
		}
	}

	// Check for "MAHARASHTRA STATE EL" pattern (can be split across tokens or have variations)
	if strings.Contains(combined, "MAHARASHTRA") && 
		(strings.Contains(combined, "STATE") || strings.Contains(combined, "EL")) {
		return "Bills_Utilities"
	}
	
	// Check tokens for compressed utility names
	for _, token := range tokens {
		decoded := utils.DecodeCompressedMerchant(token)
		if decoded != token {
			// If decoding found a match, check if it's a utility
			if strings.Contains(decoded, "Gas") || strings.Contains(decoded, "Electricity") {
				return "Bills_Utilities"
			}
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

	// Check for charges (small amounts with charge keywords)
	if amount > 0 && utils.IsCharge(originalNarration, amount) {
		return "Bills_Utilities" // Charges are typically utility-related
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
