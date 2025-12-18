package rules

import (
	"classify/statement_analysis_engine_rules/utils"
	"regexp"
	"strings"
)

// CategoryResult contains category classification with metadata
type CategoryResult struct {
	Category       string
	Confidence     float64
	MatchedKeywords []string
	Gateway        string
	Channel        string
	RuleVersion    string
	Reason         string
}

// ClassifyCategory classifies the transaction category based on narration
// Enhanced with tokenization and gateway detection
func ClassifyCategory(narration string, merchant string) string {
	result := ClassifyCategoryWithMetadata(narration, merchant, 0.0)
	return result.Category
}

// ClassifyCategoryWithAmount classifies the transaction category with amount for charge detection
func ClassifyCategoryWithAmount(narration string, merchant string, amount float64) string {
	result := ClassifyCategoryWithMetadata(narration, merchant, amount)
	return result.Category
}

// ClassifyCategoryWithMetadata classifies category and returns metadata (for explainability)
func ClassifyCategoryWithMetadata(narration string, merchant string, amount float64) CategoryResult {
	result := CategoryResult{
		Category:        "Other",
		Confidence:      0.0,
		MatchedKeywords: make([]string, 0),
		RuleVersion:     utils.RuleVersion,
	}
	
	originalNarration := narration
	narration = strings.ToUpper(narration)
	merchant = strings.ToUpper(merchant)
	combined := narration + " " + merchant

	// Tokenize narration for better pattern matching
	tokens := utils.Tokenize(originalNarration)
	
	// Extract gateway (separate concept from category)
	gateway := utils.ExtractGateway(originalNarration)
	result.Gateway = gateway
	
	// Extract channel (payment method - will be set by caller)
	
	// Track matched keywords for explainability
	matchedKeywords := make([]string, 0)
	
	// Helper function to return CategoryResult with category
	returnCategory := func(category string, confidence float64, reason string, keywords ...string) CategoryResult {
		resultCopy := result
		resultCopy.Category = category
		resultCopy.Confidence = confidence
		resultCopy.Reason = reason
		resultCopy.MatchedKeywords = append(matchedKeywords, keywords...)
		// Calculate final confidence
		hasGateway := gateway != ""
		hasMerchant := merchant != "" && merchant != "UNKNOWN"
		amountPattern, hasAmountPattern := utils.DetectAmountPattern(amount)
		if hasAmountPattern {
			resultCopy.MatchedKeywords = append(resultCopy.MatchedKeywords, amountPattern)
		}
		finalConfidence := utils.CalculateConfidence(
			resultCopy.MatchedKeywords,
			hasGateway,
			hasMerchant,
			hasAmountPattern,
			false, // recurrence would be detected separately
		)
		// Use provided confidence if higher, otherwise use calculated
		if confidence > finalConfidence {
			resultCopy.Confidence = confidence
		} else {
			resultCopy.Confidence = finalConfidence
		}
		return resultCopy
	}

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
				return returnCategory("Food_Delivery", 0.85, "Food delivery app detected via wallet", "SWIGGY", "ZOMATO", wallet)
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
		// Vegetables and Fruits
		"VEGETABLE", "VEGETABLES", "VEG", "FRUIT", "FRUITS",
		"VEGETABLE SHOP", "FRUIT SHOP", "VEGETABLE MARKET",
		"FRUIT MARKET", "VEGETABLE VENDOR", "FRUIT VENDOR",
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
	
	// Loan EMI Patterns (comprehensive - banking-industry-grade)
	// Universal Loan EMI Keywords
	loanKeywords := []string{
		"EMI", "LOAN", "INSTALMENT", "INSTALLMENT",
		"SI", "ECS", "NACH", "AUTO DEBIT", "MANDATE",
	}
	
	// Auto-Debit Modes (Critical Signals)
	autoDebitPatterns := []string{
		"ECS EMI", "NACH EMI", "SI EMI", "AUTO EMI", "MANDATE EMI",
	}
	
	// Bank Loan EMI Narrations
	bankLoanPatterns := []string{
		// HDFC Bank / HDFC Ltd
		"ECS EMI HDFC LTD", "HDFC LOAN EMI", "HDFC HOME LOAN EMI",
		"HDFCBANK EMI", "HDFC LTD EMI", "HDFCLOAN",
		// ICICI Bank
		"ICICI LOAN EMI", "ICICI HOME LOAN EMI", "ECS EMI ICICI",
		"ICICI PERSONAL LOAN EMI", "ICICILOAN",
		// SBI
		"SBI LOAN EMI", "SBI HOME LOAN EMI", "ECS EMI SBI",
		"SBI PERSONAL LOAN EMI", "SBILOAN",
		// Axis Bank
		"AXIS LOAN EMI", "AXIS BANK EMI", "NACH EMI AXIS", "AXISLOAN",
		// Kotak Bank
		"KOTAK LOAN EMI", "KOTAK BANK EMI", "KOTAKLOAN",
		// Other Banks
		"IDFC LOAN EMI", "YES BANK EMI", "PNB LOAN EMI",
		"IDFCLOAN", "YESBANK", "PNBLOAN",
	}
	
	// NBFC Loan EMI Narrations
	nbfcLoanPatterns := []string{
		// Bajaj Finserv
		"BAJAJ FINSERV EMI", "BAJAJ FIN EMI", "BAJAJ FINANCE",
		"BAJAJFINSERV", "BAJAJFIN",
		// Tata Capital
		"TATA CAPITAL EMI", "TATACAPITAL",
		// HDB Financial
		"HDB EMI", "HDB FINANCIAL EMI", "HDBFINANCIAL",
		// Home Credit
		"HOME CREDIT EMI", "HOMECREDIT",
		// Aditya Birla Finance
		"ADITYA BIRLA EMI", "ABFL EMI", "ADITYABIRLA",
		// L&T Finance
		"LT FINANCE EMI", "LTF EMI", "LTFINANCE",
	}
	
	// Loan Type-Specific Narrations
	loanTypePatterns := []string{
		// Home Loan
		"HOME LOAN EMI", "HL EMI", "HOUSING LOAN EMI",
		// Vehicle Loan
		"CAR LOAN EMI", "AUTO LOAN EMI", "VEHICLE LOAN EMI",
		// Personal Loan
		"PERSONAL LOAN EMI", "PL EMI",
		// Education Loan
		"EDUCATION LOAN EMI", "STUDENT LOAN EMI",
		// Business Loan
		"BUSINESS LOAN EMI", "MSME LOAN EMI",
	}
	
	// Overdue / Penalty / Recovery Narrations
	loanOverduePatterns := []string{
		"OVERDUE LOAN RECOVERED", "EMI RECOVERY", "LOAN PENALTY",
		"LATE PAYMENT FEE LOAN", "OVERDUE LOAN", "LOAN RECOVERED",
		"REPAYMENT",
	}
	
	// BillDesk / PayU Based Loan Payments
	loanGatewayPatterns := []string{
		"BILLDKHDFCLOAN", "BILLDKBAJAJFINSERV", "PAYUHDFCHOMELOAN",
		"BILLDKICICILOAN", "BILLDKSBILOAN", "BILLDKAXISLOAN",
	}
	
	// Ambiguous but Real Narrations
	loanAmbiguousPatterns := []string{
		"LOAN PAYMENT", "FINANCE PAYMENT", "INSTALLMENT PAID",
		"MONTHLY INSTALLMENT",
	}
	
	// Combined Loan EMI Patterns (for matching)
	loanEmiPatterns := []string{
		// Universal keywords
		"EMI", "LOAN", "INSTALMENT", "INSTALLMENT",
		// Auto-debit
		"ECS EMI", "NACH EMI", "SI EMI", "AUTO EMI", "MANDATE EMI",
		// Banks
		"HDFC LOAN", "ICICI LOAN", "SBI LOAN", "AXIS LOAN", "KOTAK LOAN",
		"HDFCLOAN", "ICICILOAN", "SBILOAN", "AXISLOAN", "KOTAKLOAN",
		"HDFC BANK EMI", "ICICI BANK EMI", "SBI BANK EMI",
		// NBFCs
		"BAJAJ FINSERV", "BAJAJ FIN", "BAJAJFINSERV", "BAJAJFIN",
		"TATA CAPITAL", "HDB FINANCIAL", "HOME CREDIT",
		"ADITYA BIRLA", "LT FINANCE",
		// Loan types
		"HOME LOAN", "CAR LOAN", "AUTO LOAN", "VEHICLE LOAN",
		"PERSONAL LOAN", "EDUCATION LOAN", "BUSINESS LOAN",
		// Overdue/Recovery
		"OVERDUE LOAN", "EMI RECOVERY", "LOAN RECOVERED", "REPAYMENT",
		// Gateways
		"BILLDKHDFCLOAN", "BILLDKBAJAJFINSERV", "PAYUHDFCHOMELOAN",
		// Ambiguous
		"LOAN PAYMENT", "FINANCE PAYMENT", "INSTALLMENT PAID",
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
		"YOUTUBE", "YOUTUBE PREMIUM", "YOUTUBEPREMIUM", "YT",
		"YOUTUBE MUSIC", "YOUTUBEMUSIC",
	}

	// Investment patterns (Mutual Funds, Stocks, NPS, Insurance, Crypto)
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
		// Crypto Exchanges (Indian)
		"WAZIRX", "WAZIRXIN", "ZANMAI", "ZANMAI LABS", "ZANMAILABS",
		"COINDCX", "NEBULAS", "NEBULAS TECHNOLOGIES", "NEBULASTECHNOLOGIES", "DCX",
		"COINSWITCH", "COINSWITCHKUBER", "BITCIPHER", "BITCIPHER LABS",
		"ZEBPAY", "ZEB IT SERVICE", "ZEBITSERVICE",
		"UNOCOIN", "UNOCOMMERCE",
		// Crypto Exchanges (International)
		"BINANCE", "BINANCEPAY", "BIFINANCE",
		"COINBASE", "CB PAY", "CBPAY",
		"KRAKEN", "PAYWARD",
		"CRYPTOCOM", "FORIS",
		"KUCOIN", "MEK GLOBAL", "MEKGLOBAL",
		"BITSTAMP",
		// Crypto Gateway Combinations
		"PAYUWAZIRX", "PAYUZANMAI", "PAYUCOINDCX", "PAYUNEBULAS",
		"RAZPZANMAILABS", "RAZPNEBULASTECHNOLOGIES", "RAZPCOINSWITCH", "RAZPBITCIPHER",
		"CCAVENUEBITCIPHER", "CCAVENUEZANMAI",
		// Crypto Transfer Patterns
		"FUND TRANSFER TO ZANMAI", "NEFT TO NEBULAS", "IMPS BITCIPHER",
		"IMPS FROM ZANMAI", "NEFT FROM COINDCX", "CRYPTO WITHDRAWAL",
		// Crypto Generic Patterns (only if combined with known entities)
		"CRYPTO", "CRYPTOCURRENCY", "DIGITAL ASSET", "VIRTUAL ASSET",
		// Generic
		"INVESTMENT",
	}

	// Dividend patterns (income from investments)
	dividendPatterns := []string{
		"DIV", "DIVIDEND", "DIVIDEND CREDIT", "DIV CR",
	}

	// Priority 0: Check Loan EMI (HIGHEST PRIORITY - before all other categories)
	// Loan EMI detection has very high confidence and should be checked first
	// First check for simple EMI keyword (catches minimal narrations like "EMI 4452581")
	if strings.Contains(combined, "EMI") {
		// If narration contains "EMI" and looks like a loan account number pattern
		// or has loan-related keywords, classify as Loan
		if strings.Contains(combined, "LOAN") || strings.Contains(combined, "ECS") ||
			strings.Contains(combined, "NACH") || strings.Contains(combined, "SI") ||
			strings.Contains(combined, "MANDATE") || strings.Contains(combined, "INSTALLMENT") ||
			strings.Contains(combined, "INSTALMENT") {
			return returnCategory("Loan", 0.95, "EMI with loan keywords detected", "EMI", "LOAN")
		}
		// Check for loan account number pattern (numbers after EMI)
		// Pattern: "EMI" followed by numbers (like "EMI 4452581")
		emiPattern := regexp.MustCompile(`EMI\s*\d+`)
		if emiPattern.MatchString(combined) {
			return returnCategory("Loan", 0.90, "EMI with account number pattern detected", "EMI")
		}
	}
	
	hasLoanKeyword := false
	for _, keyword := range loanKeywords {
		if strings.Contains(combined, keyword) {
			hasLoanKeyword = true
			break
		}
	}
	
	if hasLoanKeyword {
		// Check for auto-debit patterns (highest confidence)
		for _, pattern := range autoDebitPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Loan", 0.95, "Auto-debit loan pattern detected: "+pattern, pattern)
			}
		}
		
		// Check for bank/NBFC names (high confidence)
		for _, pattern := range bankLoanPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Loan", 0.90, "Bank loan pattern detected: "+pattern, pattern)
			}
		}
		for _, pattern := range nbfcLoanPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Loan", 0.90, "NBFC loan pattern detected: "+pattern, pattern)
			}
		}
		
		// Check for loan type patterns
		for _, pattern := range loanTypePatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Loan", 0.85, "Loan type pattern detected: "+pattern, pattern)
			}
		}
		
		// Check for overdue/recovery patterns
		for _, pattern := range loanOverduePatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Loan", 0.85, "Loan overdue/recovery pattern detected: "+pattern, pattern)
			}
		}
		
		// Check for gateway-based loan payments
		for _, pattern := range loanGatewayPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Loan", 0.90, "Gateway-based loan payment detected: "+pattern, pattern)
			}
		}
		
		// Check for ambiguous loan patterns (if EMI or LOAN keyword present)
		if strings.Contains(combined, "EMI") || strings.Contains(combined, "LOAN") {
			for _, pattern := range loanAmbiguousPatterns {
				if strings.Contains(combined, pattern) {
					return returnCategory("Loan", 0.75, "Ambiguous loan pattern detected: "+pattern, pattern)
				}
			}
			// If EMI or LOAN keyword + ECS/NACH/SI, it's likely a loan
			if strings.Contains(combined, "ECS") || strings.Contains(combined, "NACH") ||
				strings.Contains(combined, "SI") || strings.Contains(combined, "MANDATE") {
				return returnCategory("Loan", 0.85, "EMI/LOAN with auto-debit indicator detected", "EMI", "LOAN", "ECS/NACH/SI")
			}
		}
		
		// Check tokens for loan-related patterns
		for _, token := range tokens {
			if strings.Contains(token, "EMI") || strings.Contains(token, "LOAN") {
				// Check if token contains bank/NBFC name
				if strings.Contains(token, "HDFC") || strings.Contains(token, "ICICI") ||
					strings.Contains(token, "SBI") || strings.Contains(token, "AXIS") ||
					strings.Contains(token, "KOTAK") || strings.Contains(token, "BAJAJ") {
					return returnCategory("Loan", 0.80, "Loan keyword with bank/NBFC name detected in token: "+token, token)
				}
			}
		}
	}

	// Priority 1: Check for POS indicator (dining vs delivery distinction)
	// POS + restaurant name = Dining (NOT Food Delivery)
	hasPOS := strings.Contains(combined, "POS")
	if hasPOS {
		// Check if it's a restaurant/cafe (dining)
		for _, pattern := range diningPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Dining", 0.80, "POS transaction at restaurant/cafe", "POS", pattern)
			}
		}
		// Check if it's grocery (POS GROCERY)
		for _, pattern := range groceriesPatterns {
			if strings.Contains(combined, "POS") && strings.Contains(combined, pattern) {
				return returnCategory("Groceries", 0.80, "POS grocery transaction", "POS", pattern)
			}
		}
		// Check if it's shopping (POS RETAIL, POS STORE, etc.)
		if strings.Contains(combined, "POS RETAIL") || strings.Contains(combined, "POS STORE") ||
			strings.Contains(combined, "POS PURCHASE") || strings.Contains(combined, "POS AMAZON") ||
			strings.Contains(combined, "POS FLIPKART") {
			return returnCategory("Shopping", 0.80, "POS retail/shopping transaction", "POS")
		}
	}

	// Priority 2: Check Food Delivery (ONLINE only, not POS)
	// Exclude POS transactions from food delivery
	if !hasPOS {
		for _, pattern := range foodDeliveryPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Food_Delivery", 0.90, "Food delivery app detected (online)", pattern)
			}
		}
		// Also check tokens for food delivery apps
		for _, token := range tokens {
			if strings.Contains(token, "ZOMATO") || strings.Contains(token, "SWIGGY") ||
				strings.Contains(token, "FAASOS") || strings.Contains(token, "EATSURE") ||
				strings.Contains(token, "DOMINOS") || strings.Contains(token, "MCDONALDS") {
				// Verify it's not POS
				if !strings.Contains(combined, "POS") {
					return returnCategory("Food_Delivery", 0.85, "Food delivery app detected in token (online)", token)
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
				return returnCategory("Dining", 0.75, "Restaurant/cafe detected (non-POS)", pattern)
			}
		}
	}

	// Check Fuel (separate category, before Travel)
	for _, pattern := range fuelPatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Fuel", 0.85, "Fuel expense detected", pattern)
		}
	}
	// Also check tokens for fuel patterns
	for _, token := range tokens {
		if strings.Contains(token, "IOCL") || strings.Contains(token, "BPCL") ||
			strings.Contains(token, "HPCL") || strings.Contains(token, "PETROL") ||
			strings.Contains(token, "DIESEL") {
			return returnCategory("Fuel", 0.85, "Fuel expense detected in token", token)
		}
	}
	
	// Check Travel (comprehensive patterns)
	for _, pattern := range travelPatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Travel", 0.85, "Travel expense detected", pattern)
		}
	}
	// Also check tokens for travel apps
	for _, token := range tokens {
		if strings.Contains(token, "UBER") || strings.Contains(token, "OLA") ||
			strings.Contains(token, "IRCTC") || strings.Contains(token, "MMT") ||
			strings.Contains(token, "GOIBIBO") || strings.Contains(token, "REDBUS") ||
			strings.Contains(token, "OYO") {
			return returnCategory("Travel", 0.80, "Travel app detected in token", token)
		}
	}

	// Check Shopping (enhanced with tokenization)
	for _, pattern := range shoppingPatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Shopping", 0.80, "Shopping expense detected", pattern)
		}
	}
	// Also check tokens for shopping-related merchants
	for _, token := range tokens {
		if strings.Contains(token, "SIMPL") || strings.Contains(token, "GETSIMPL") ||
			strings.Contains(token, "AMAZON") || strings.Contains(token, "FLIPKART") {
			return returnCategory("Shopping", 0.75, "Shopping merchant detected in token", token)
		}
	}

	// Check Groceries
	for _, pattern := range groceriesPatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Groceries", 0.75, "Grocery expense detected", pattern)
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
				return returnCategory("Bills_Utilities", 0.90, "Electricity bill payment detected", pattern)
			}
		}
		
		// Gas
		for _, pattern := range gasPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Bills_Utilities", 0.90, "Gas bill payment detected", pattern)
			}
		}
		
		// Water
		for _, pattern := range waterPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Bills_Utilities", 0.90, "Water bill payment detected", pattern)
			}
		}
		
		// Telecom
		for _, pattern := range telecomPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Bills_Utilities", 0.90, "Telecom bill payment detected", pattern)
			}
		}
		
		// DTH
		for _, pattern := range dthPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Bills_Utilities", 0.90, "DTH bill payment detected", pattern)
			}
		}
		
		// Toll/Fastag
		for _, pattern := range tollPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Bills_Utilities", 0.85, "Toll/Fastag payment detected", pattern)
			}
		}
		
		// Government payments
		for _, pattern := range governmentPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Bills_Utilities", 0.85, "Government payment detected", pattern)
			}
		}
		
		// Insurance
		for _, pattern := range insurancePatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Bills_Utilities", 0.90, "Insurance premium payment detected", pattern)
			}
		}
		
		// Credit Card
		for _, pattern := range creditCardPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Bills_Utilities", 0.90, "Credit card bill payment detected", pattern)
			}
		}
		
		// Loan EMI
		for _, pattern := range loanEmiPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Bills_Utilities", 0.90, "Loan EMI payment detected", pattern)
			}
		}
		
		// Housing/Maintenance
		for _, pattern := range housingPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Bills_Utilities", 0.85, "Housing/maintenance bill detected", pattern)
			}
		}
		
		// Tax payments
		for _, pattern := range taxPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Bills_Utilities", 0.85, "Tax payment detected", pattern)
			}
		}
		
		// Default: if bill payment gateway detected but no specific category, still Bills_Utilities
		return returnCategory("Bills_Utilities", 0.75, "Bill payment gateway detected", "BILL_PAYMENT")
	}
	
	// Legacy check for backward compatibility
	for _, pattern := range billsPatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Bills_Utilities", 0.80, "Bill payment pattern detected", pattern)
		}
	}

	// Check for "MAHARASHTRA STATE EL" pattern (can be split across tokens or have variations)
	if strings.Contains(combined, "MAHARASHTRA") && 
		(strings.Contains(combined, "STATE") || strings.Contains(combined, "EL")) {
		return returnCategory("Bills_Utilities", 0.85, "Maharashtra State Electricity detected", "MAHARASHTRA", "EL")
	}
	
	// Check tokens for compressed utility names
	for _, token := range tokens {
		decoded := utils.DecodeCompressedMerchant(token)
		if decoded != token {
			// If decoding found a match, check if it's a utility
			if strings.Contains(decoded, "Gas") || strings.Contains(decoded, "Electricity") {
				return returnCategory("Bills_Utilities", 0.80, "Utility detected via merchant decoding", decoded)
			}
		}
	}

	// Check Healthcare
	for _, pattern := range healthcarePatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Healthcare", 0.75, "Healthcare expense detected", pattern)
		}
	}

	// Check Education
	for _, pattern := range educationPatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Education", 0.75, "Education expense detected", pattern)
		}
	}

	// Check Entertainment
	for _, pattern := range entertainmentPatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Entertainment", 0.75, "Entertainment expense detected", pattern)
		}
	}

	// Check Dividend (income from investments) - should be classified as income category
	for _, pattern := range dividendPatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Investment", 0.90, "Dividend income detected", pattern)
		}
	}

	// Check Investment
	for _, pattern := range investmentPatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Investment", 0.85, "Investment detected", pattern)
		}
	}

	// Check for charges (small amounts with charge keywords)
	if amount > 0 && utils.IsCharge(originalNarration, amount) {
		return returnCategory("Bills_Utilities", 0.70, "Bank charge detected", "CHARGE")
	}

	// Default category (don't over-classify - keep UNKNOWN/Other for ambiguous cases)
	result.Category = "Other"
	result.Confidence = 0.1
	result.Reason = "No matching patterns found - classified as Other"
	
	// Calculate final confidence
	hasGateway := gateway != ""
	hasMerchant := merchant != "" && merchant != "UNKNOWN"
	amountPattern, hasAmountPattern := utils.DetectAmountPattern(amount)
	if hasAmountPattern {
		matchedKeywords = append(matchedKeywords, amountPattern)
	}
	
	result.MatchedKeywords = matchedKeywords
	result.Confidence = utils.CalculateConfidence(
		matchedKeywords,
		hasGateway,
		hasMerchant,
		hasAmountPattern,
		false, // recurrence pattern would be detected separately
	)
	
	return result
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
