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

	// Food Delivery patterns (comprehensive - ONLINE ONLY, NOT POS)
	foodDeliveryPatterns := []string{
		// Primary Food Apps
		"ZOMATO", "ZOMATOONLINE", "ZOMATOINDIA", "ZOMATOORDER", "ZMT",
		"SWIGGY", "SWIGGYINSTAMART", "SWIGGYONLINE", "SWIGGYORDER",
		"FAASOS", "EATSURE", "BOX8", "REVOLVEEATSURE",
		// Gateway combinations (key indicator for online delivery)
		"PAYUZOMATO", "RAZPZOMATO", "PAYUSWIGGY", "RAZPSWIGGY",
		"PAYUDOMINOS", "RAZPMCDONALDS", "AMAZONPAYZOMATO",
		// Generic
		"UBER EATS", "FOODPANDA", "FOOD DELIVERY", "ONLINE FOOD ORDER",
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

	// Dining patterns (POS signals - restaurants, cafes, NOT delivery)
	diningPatterns := []string{
		// POS indicators (key signal for dining vs delivery)
		"POS RESTAURANT", "POS CAFE", "POS DINING",
		// Restaurant/Cafe names (when not with delivery gateways)
		"RESTAURANT", "CAFE", "HOTEL", "DINING",
		"FOOD COURT", "EATERY", "BAKERY", "COFFEE",
		"STARBUCKS", "CAFE COFFEE DAY", "CCD",
		"BARBEQUENATION",
		// POS + restaurant chain (dining, not delivery)
		"POS DOMINOS", "POS MCDONALDS", "POS KFC",
		"POS PIZZAHUT", "POS BURGERKING", "POS SUBWAY",
	}

	// Travel patterns (comprehensive)
	travelPatterns := []string{
		// Cab & Local Travel
		"UBER", "UBERTRIP", "UBERINDIA", "UBERBV", "PAYUUBER",
		"OLA", "OLACABS", "OLAMONEY", "OLATRIP", "RAPIDO",
		// Railways
		"IRCTC", "IRCTCIPAY", "RAZPIRCTC", "PAYUIRCTC", "RAZPIRCTCIPAY",
		// Bus Booking
		"REDBUS", "ABHIBUS", "YATRAGENIE",
		// Flight Booking
		"MAKEMYTRIP", "MMT", "MMTFLIGHT", "MMTHOTEL",
		"GOIBIBO", "GOIBIBOFLIGHT", "IBIBO",
		"YATRA", "YATRADOTCOM", "CLEARTRIP",
		// Hotels / Stays
		"OYO", "OYOROOMS", "TREEBO", "FABHOTELS", "AIRBNB",
		// Generic
		"TRAVEL", "FLIGHT", "HOTEL", "CAB", "TAXI", "BOOKING",
		"ONLINE TRAVEL PAYMENT",
	}
	
	// Fuel / Petrol / Diesel / EV patterns (separate from travel)
	fuelPatterns := []string{
		// PSU Oil Companies
		"IOCL", "INDIANOIL", "INDIAN OIL",
		"BPCL", "BHARATPETROLEUM", "BHARAT PET",
		"HPCL", "HINDUSTANPETROLEUM", "HIND PET",
		// Private Fuel Companies
		"RELIANCE PETROLEUM", "RELIANCE PETROL", "RELIANCE",
		"SHELL", "ESSAR", "NAYARA ENERGY",
		// EV Charging
		"TATA POWER EV", "ATHER ENERGY",
		// FASTag (Fuel/Toll mix)
		"FASTAG", "ICICIFASTAG", "HDFCBANKFASTAG",
		"PAYTMFASTAG", "NHAI",
		// Generic
		"PETROL", "DIESEL", "FUEL", "PETROL PUMP", "SERVICE STATION",
		"GAS STATION",
	}

	// Shopping patterns (E-commerce & Retail)
	shoppingPatterns := []string{
		// E-commerce platforms
		"AMAZON", "AMAZONPAY", "FLIPKART", "FLIPKARTIN",
		"MYNTRA", "AJIO", "MEESHO", "NYkaa",
		// Retail / POS
		"POS AMAZON", "POS FLIPKART", "POS RETAIL",
		"POS STORE", "POS PURCHASE",
		// Fashion / Lifestyle
		"ZARA", "HNM", "PANTALOONS", "LIFESTYLE",
		// Generic shopping
		"SHOPPING", "MALL", "STORE", "SHOP",
		"JEWELLERY", "TANISHQ", "MALABAR", "PC JEWELLER",
		"ELECTRONICS", "CROMA", "RELIANCE DIGITAL",
		"Vijay Sales", "GREAT EASTERN", "SHOPPERS STOP",
		"SIMPL", "SIMPL TECHNOLOGI", "GETSIMPL", // Simpl buy now pay later
	}

	// Groceries patterns (Online & Offline)
	groceriesPatterns := []string{
		// Online Groceries
		"BIGBASKET", "BBNOW", "GROFERS", "BLINKIT",
		"JIO MART", "JIOMART", "AMAZONFRESH",
		// Offline Grocery / Kirana
		"POS GROCERY", "POS SUPERMARKET",
		"DMART", "RELIANCE SMART", "MORE SUPERMARKET",
		"RELIANCE FRESH", "SPENCERS", "BIG BAZAAR",
		// Generic
		"GROCERY", "GROCERIES", "SUPERMARKET", "KIRANA", "GENERAL STORE",
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

	// Investment patterns (Mutual Funds, Stocks, NPS, Insurance)
	investmentPatterns := []string{
		// Mutual Funds
		"MUTUAL FUND", "MF SIP", "SIP INSTALLMENT",
		"GROWW", "COIN", "UPSTOX", "KITE",
		// Stocks / Trading
		"NSE", "BSE", "STOCK PURCHASE", "SECURITIES BUY",
		"STOCK", "SHARE", "DEMAT",
		// NPS / Pension
		"BILLDKNPSTRUST", "NATIONAL PENSION SYSTEM",
		"NPS CONTRIBUTION", "NPS", "PPF", "ELSS",
		// Recurring Deposits / Fixed Deposits
		"RD", "FD", "SIP", "RD INSTALLMENT",
		// Clearing Corporations
		"INDIAN CLEARING CORPORATION", "INDIAN CLEARING CORPORATION LIMITED",
		"INDIAN C LEARING CORPORATION", "INDIAN C LEARING CORPORATION LIMITED", // Handle typo with space
		"NSDL", "CDSL", "CLEARING CORPORATION",
		// Stock Broking Companies
		"ZERODHA", "ZERODHA BROKING", "ZERODHA BROKING LTD", "ZERODHABROKING",
		"BROKING", "BROKING LTD", "HSL SEC", "HSL", "SEC",
		"ANGEL BROKING", "ICICI SECURITIES", "HDFC SECURITIES", "KOTAK SECURITIES",
		"SHAREKHAN", "MOTILAL OSWAL", "IIFL", "5PAISA",
		// Insurance (Investment-type)
		"HDFCLIFE", "ICICIPRULIFE", "SBILIFE", "LIC",
		"MAXLIFE", "INSURANCE PREMIUM",
		// Generic
		"INVESTMENT",
	}

	// Dividend patterns (income from investments)
	dividendPatterns := []string{
		"DIV", "DIVIDEND", "DIVIDEND CREDIT", "DIV CR",
	}

	// Priority 1: Check for POS indicator (dining vs delivery distinction)
	// POS + restaurant name = Dining (NOT Food Delivery)
	hasPOS := strings.Contains(combined, "POS")
	if hasPOS {
		// Check if it's a restaurant/cafe (dining)
		for _, pattern := range diningPatterns {
			if strings.Contains(combined, pattern) {
				return "Dining"
			}
		}
		// Check if it's grocery (POS GROCERY)
		for _, pattern := range groceriesPatterns {
			if strings.Contains(combined, "POS") && strings.Contains(combined, pattern) {
				return "Groceries"
			}
		}
		// Check if it's shopping (POS RETAIL, POS STORE, etc.)
		if strings.Contains(combined, "POS RETAIL") || strings.Contains(combined, "POS STORE") ||
			strings.Contains(combined, "POS PURCHASE") || strings.Contains(combined, "POS AMAZON") ||
			strings.Contains(combined, "POS FLIPKART") {
			return "Shopping"
		}
	}

	// Priority 2: Check Food Delivery (ONLINE only, not POS)
	// Exclude POS transactions from food delivery
	if !hasPOS {
		for _, pattern := range foodDeliveryPatterns {
			if strings.Contains(combined, pattern) {
				return "Food_Delivery"
			}
		}
		// Also check tokens for food delivery apps
		for _, token := range tokens {
			if strings.Contains(token, "ZOMATO") || strings.Contains(token, "SWIGGY") ||
				strings.Contains(token, "FAASOS") || strings.Contains(token, "EATSURE") ||
				strings.Contains(token, "DOMINOS") || strings.Contains(token, "MCDONALDS") {
				// Verify it's not POS
				if !strings.Contains(combined, "POS") {
					return "Food_Delivery"
				}
			}
		}
	}

	// Priority 3: Check Dining (non-POS restaurants/cafes)
	for _, pattern := range diningPatterns {
		if strings.Contains(combined, pattern) {
			// Make sure it's not a delivery gateway
			if !strings.Contains(combined, "PAYU") && !strings.Contains(combined, "RAZP") &&
				!strings.Contains(combined, "ZOMATO") && !strings.Contains(combined, "SWIGGY") {
				return "Dining"
			}
		}
	}

	// Check Fuel (separate category, before Travel)
	for _, pattern := range fuelPatterns {
		if strings.Contains(combined, pattern) {
			return "Fuel" // Fuel is a separate category
		}
	}
	// Also check tokens for fuel patterns
	for _, token := range tokens {
		if strings.Contains(token, "IOCL") || strings.Contains(token, "BPCL") ||
			strings.Contains(token, "HPCL") || strings.Contains(token, "PETROL") ||
			strings.Contains(token, "DIESEL") {
			return "Fuel" // Fuel is a separate category
		}
	}
	
	// Check Travel (comprehensive patterns)
	for _, pattern := range travelPatterns {
		if strings.Contains(combined, pattern) {
			return "Travel"
		}
	}
	// Also check tokens for travel apps
	for _, token := range tokens {
		if strings.Contains(token, "UBER") || strings.Contains(token, "OLA") ||
			strings.Contains(token, "IRCTC") || strings.Contains(token, "MMT") ||
			strings.Contains(token, "GOIBIBO") || strings.Contains(token, "REDBUS") ||
			strings.Contains(token, "OYO") {
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
