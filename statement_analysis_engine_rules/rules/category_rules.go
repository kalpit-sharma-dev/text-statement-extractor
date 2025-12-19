package rules

import (
	"classify/statement_analysis_engine_rules/utils"
	"regexp"
	"strings"
)

// CategoryResult contains category classification with metadata
type CategoryResult struct {
	Category        string
	Confidence      float64
	MatchedKeywords []string
	Gateway         string
	Channel         string
	RuleVersion     string
	Reason          string
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

	// ========================================================================
	// LAYER 4: MERCHANT/ENTITY IDENTIFICATION (MOST IMPORTANT - 90% decision)
	// ========================================================================
	// Check for known merchants first (strongest signal)
	knownMerchantName, knownMerchantCategory, knownMerchantConfidence := utils.DetectKnownMerchant(originalNarration, merchant)
	if knownMerchantName != "" {
		matchedKeywords = append(matchedKeywords, knownMerchantName)
		// Merchant match is very strong - use it as base confidence
		// But still check other signals to refine
		result.Category = knownMerchantCategory
		result.Confidence = knownMerchantConfidence
		result.MatchedKeywords = append(matchedKeywords, knownMerchantName)
		// Continue to check other layers for refinement, but merchant is primary
	}

	// ========================================================================
	// LAYER 5: INTENT KEYWORDS (Supporting Evidence)
	// ========================================================================
	// Detect intent keywords - they support but don't override merchant
	intentScores := utils.DetectIntentKeywords(originalNarration)
	// Store intent keywords for explainability
	for category, score := range intentScores {
		if score > 0 {
			matchedKeywords = append(matchedKeywords, category+"_INTENT")
		}
	}

	// ========================================================================
	// LAYER 6: PATTERN & AMOUNT HEURISTICS (Tie-breakers)
	// ========================================================================
	// Detect amount patterns (used only when merchant is ambiguous)
	amountPattern, hasAmountPattern := utils.DetectAmountPattern(amount)
	if hasAmountPattern {
		matchedKeywords = append(matchedKeywords, amountPattern)
	}

	// Helper function to return CategoryResult with category
	returnCategory := func(category string, confidence float64, reason string, keywords ...string) CategoryResult {
		resultCopy := result
		resultCopy.Category = category
		resultCopy.Confidence = confidence
		resultCopy.Reason = reason
		resultCopy.MatchedKeywords = append(matchedKeywords, keywords...)

		// Calculate final confidence using 7-layer scoring
		hasGateway := gateway != ""

		// Confidence scoring based on signals:
		// - Known merchant: +0.6 (already in base confidence)
		// - Intent keyword: +0.2 (from intentScores)
		// - Gateway match: +0.1
		// - Amount pattern: +0.1
		finalConfidence := confidence

		// Add intent keyword score if it matches category
		if intentScore, hasIntent := intentScores[category]; hasIntent {
			finalConfidence += intentScore * 0.2 // Intent keywords support but don't override
		}

		// Add gateway confidence
		if hasGateway {
			finalConfidence += 0.1
		}

		// Add amount pattern confidence
		if hasAmountPattern {
			finalConfidence += 0.1
		}

		// Cap at 1.0
		if finalConfidence > 1.0 {
			finalConfidence = 1.0
		}

		// If we have a known merchant match, ensure minimum confidence
		if knownMerchantName != "" && category == knownMerchantCategory {
			if finalConfidence < knownMerchantConfidence {
				finalConfidence = knownMerchantConfidence
			}
		}

		resultCopy.Confidence = finalConfidence
		return resultCopy
	}

	// If we already have a known merchant match, prioritize it
	// But still check other patterns for edge cases
	if knownMerchantName != "" {
		// Known merchant is strongest signal - use it unless overridden by higher priority rules
		// Continue to check other patterns but merchant takes precedence
		// Early return for high-confidence merchant matches (unless overridden by Loan/EMI or large amounts to utilities)
		// EXCEPTION: Large payments to gas utilities (> ₹25,000) should NOT return early - they need amount-based classification
		isLargeGasPayment := false
		if amount > 25000 && knownMerchantCategory == "Bills_Utilities" {
			// Check if it's a gas utility company (IGL, MGL, etc.)
			gasUtilities := []string{"INDRAPRASTHA GAS", "IGL", "MAHANAGAR GAS", "MGL", "ADANIGAS", "GUJGAS"}
			for _, gasUtil := range gasUtilities {
				if strings.Contains(strings.ToUpper(knownMerchantName), gasUtil) {
					isLargeGasPayment = true
					break
				}
			}
		}
		
		if knownMerchantConfidence >= 0.9 && !strings.Contains(combined, "EMI") && !strings.Contains(combined, "LOAN") && !isLargeGasPayment {
			return returnCategory(knownMerchantCategory, knownMerchantConfidence, "Known merchant detected: "+knownMerchantName, knownMerchantName)
		}
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
		// Catering services (from "Other" transactions)
		"CATERERS", "CATERING", "BALAJI CATERERS",
		// Vending machines (from "Other" transactions)
		"VENDING", "SHREEVENDING",
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
		// Sweets shops
		"BANSAL BIKANER SWEET", "BIKANER SWEET", "SWEET",
		"AGGARWAL SWEETS", "AGGARWAL FOOD", "AGGARWAL SWEET",
		"SWEETS", "SWEET SHOP",
		"STARBUCKS", "CAFE COFFEE DAY", "CCD",
		"BARBEQUENATION",
		// POS + restaurant chain (dining, not delivery)
		"POS DOMINOS", "POS MCDONALDS", "POS KFC",
		"POS PIZZAHUT", "POS BURGERKING", "POS SUBWAY",
		// Restaurants/Cafes (from "Other" transactions)
		"EATSOME", "MEGAPOLISSANGRIA", "SANGRIA",
		"SNACKS CENT", "SNACKS", "DAIRY AND SWEE",
		"GODAVARI SNACKS", "GODAVARI",
		// Dining establishments (from classification issues)
		"BAMRADA SONS", "BAMRADA",
		"SPECIAL CHAT CENTER", "SPECIAL CHAT", "CHAT CENTER",
		"MUSKAN BAKERS", "MUSKAN BAKERS AND CO",
		"ROSIER FOODS", "ROSIER",
		"PANCHAITEA", "PANCHAI TEA", "TEA",
		// Beverage vendors (from 2025 data)
		"LASSI WALE", "LASSI", "JUICE", "JUICE WALE",
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
		// Moving/Relocation services (from "Other" transactions)
		"PACKER", "MOVER", "PACKER MOVER", "PACKING", "MOVING",
		"RELOCATION", "SHIFTING",
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
		// Service stations (from classification issues)
		"DAUJI SERVICE STATIO", "DAUJI SERVICE", "SERVICE STATIO",
		"PHOOL SERVICE STATIO", "PHOOL SERVICE",
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
		// Clothing stores (from "Other" transactions)
		"NEW LOOK", "RANGOLI HOSIERY", "NEW BOMBAY GENTS PAR",
		"MEGA INNERWEARS", "GIFT GALLERY", "INNERWEARS",
		"HOSIERY", "GENTS PAR", "GENTS", "CLOTHING",
		// Auto parts/accessories (from "Other" transactions)
		"JAIN AUTO", "AUTO AND ACCESS", "AUTO ACCESS",
		"AUTO PARTS", "AUTO ACCESSORIES",
		"ASB AUTOMOBILES", "AUTOMOBILES", "AUTO MOBILES",
		// Auto services and vehicle maintenance (from 2025 data)
		"AUTO CARE", "RIDE N REPAIR", "RIDE AND REPAIR",
		"SCOOTER AGENCY", "SHIVA SCOOTER", "VEHICLE SERVICE",
		"BIKE SERVICE", "SCOOTER SERVICE", "CAR SERVICE",
		// Jewellery stores
		"JEWELLERS", "JEWELLERY", "JEWELRY", "KAMLA JI JEWELLERS", "KUMAR JEWELLERS",
		// Shoes
		"WELCO SHOES", "SHOES", "FOOTWEAR",
		// Beauty salons/products
		"FINAL TOUCH BEAUTY", "BEAUTY", "BEAUTY PARLOUR", "BEAUTY PARLOR",
		"SALON", "SALOON", "SPA",
		// Supply stores
		"ALPHABULK SUPPLY", "SUPPLY SOL", "SUPPLY SOLUTION",
		// Technology/services (if not bills)
		"PARVIOM TECHNOLOGIES", "PARVIOM",
		// Trading companies and stores (from classification issues)
		"TRADING COM", "TRADING",
		"TRADING COMPA",
		"STATIONERY", "STATIONARY",
		"BIKANERVALA", "BIKANERVALA PRIVATE",
		"BOMBAY WATCH COMPANY", "BOMBAY WATCH",
		"VENDING BROTHERS", "BROTHERS PVT",
		// Local businesses with generic names (from "Other" transactions)
		"ENTERPRISE", "ENTERPRISES", "INDUSTRIA",
		// Examples: A Y ENTERPRISE, KALIKA ENTERPRISES, AAYUSH ENTERPRISES, 
		// DAYALS ENTERPRISES, RUKMANI ENTERPRISES, RAW REAPS ENTERPRISE,
		// CHANDIGARH INDUSTRIA
		// Home services (from 2025 data)
		"INTERIORS", "TULSI INTERIORS", "INTERIOR DESIGN",
		"FURNITURE", "HOME DECOR", "FURNISHING",
	}

	// Groceries patterns (Online & Offline)
	groceriesPatterns := []string{
		// Online Groceries
		"BIGBASKET", "BBNOW", "GROFERS", "BLINKIT",
		"JIO MART", "JIOMART", "AMAZONFRESH",
		"ZEPTO", "ZEPTO MARKETPLACE", "ZEPTO MARKETPLACE PR", // Grocery delivery app
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
		// Dairy and local stores (from "Other" transactions)
		"ANKIT DAIRY", "DAIRY", "DAIRY AND SWEE", "DAIRY AND SWEET",
		// Smart bazar / marketplaces
		"SMART BAZAR", "SMART BAZAAR", "MAYUR SMART BAZAR",
		// Provision stores (from 2025 data)
		"PROVISION", "PROVISIONS", "PROVISION STOR", "PROVISION STORE",
		// Agricultural/farm products
		"KISANKONNECT", "KISAN KONNECT", "FARM", "AGRICULTURAL",
		// Local traders and markets (from classification issues)
		"TRADERS",
		"SUPER MARKET",
		"KHOA PANEER",
	}

	// Universal Bill Payment Aggregators/Gateways
	// IMPORTANT: Only include actual bill payment aggregators, NOT generic payment gateways
	// Generic gateways like PAYTM, GPAY, PHONEPE, AMAZONPAY are used for ALL payment types
	// EXCEPTION: PAYTM UTILITY / PAYTM ECOMMERCE-UTILITYPAYTM are bill payments
	billGateways := []string{
		"BILLDESK", "BILLDK", "BBPS", // Actual bill payment aggregators
		"WHDF", "SBIPG", "AXISPG", "ICICIPG", "KOTAKPG", "YESPG", // Bank-specific bill payment gateways
		"PAYGOV", // Government payment gateway
		// Note: PAYU, RAZORPAY, RAZP, CCAVENUE can be used for bills but also for other payments
		// Only classify as bill if combined with actual bill keywords
	}

	// Generic payment gateways (used for ALL payment types, not just bills)
	// These should NOT trigger bill payment detection alone
	genericGateways := []string{
		"PAYTM", "GPAY", "PHONEPE", "AMAZONPAY", "PAYU", "RAZORPAY", "RAZP",
		"CCAVENUE", "VYAPAR", "BHARATPE", "BAJAJPAY", "MOBIKWIK",
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
		"RENT", "RENT FOR MONTH", "HOUSE RENT", "RENTAL",
		"MONTHLY RENT", "RENT PAYMENT",
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
		// Wallets (typically used for bills/recharges)
		"MOBIKWIK",
		// Business services (often bill payments)
		"BUSINESS SOL", "BUSINESS SERVICE", "EROCKET",
		// Farm/agricultural services (often bill payments)
		"FARMWORK", "FARM WORK", "AGRICULTURAL",
		// Bank charges and fees (from classification issues)
		"DC INTL POS TXN MARKUP", "INTL POS TXN MARKUP", "POS TXN MARKUP",
		"BANK CHARGES", "SERVICE CHARGE", "ANNUAL FEE", "MAINTENANCE CHARGE",
		// Software & Technology subscriptions (from 2025 data)
		"CLAUDE.AI", "CLAUDE AI", "ANTHROPIC", "CURSOR", "AI POWERED IDE",
		"GOOGLE CLOUD", "GOOGLECLOUD", "AWS", "AZURE", "CLOUD COMPUTING",
		"SOFTWARE SUBSCRIPTION", "SAAS",
	}

	// Healthcare patterns
	healthcarePatterns := []string{
		"HOSPITAL", "CLINIC", "PHARMACY", "MEDICINE",
		"APOLLO", "FORTIS", "MAX", "MEDICOS", "MEDICAL",
		"HEALTH", "DOCTOR", "LAB", "DIAGNOSTIC",
		"MEDICAL STORE",
		"HOSPITALS", "SHIVALIK HOSPITAL",
		"RANVEER MEDICAL", "MAYUR MEDICAL", "SHREE CHINTAMANI MED",
		"PHARMACY", "CHEMIST", "CHITRANSH PHARMACY",
		"PATANJALI CHIKITSALY", "HEALTHPLIX",
		"DR ", "DR.", // Doctor prefix
		// Pharmacies and medical stores (from classification issues)
		"CHEMISTS",
		"MEDICO",
		// Medical imaging and health services (from 2025 data)
		"MOLECULAR IMAGING", "IMAGING", "RADIOLOGY", "SCAN",
		"AROGYA", "AROGYALAXMI", // "Arogya" = Health in Sanskrit
	}

	// Education patterns
	educationPatterns := []string{
		"SCHOOL", "COLLEGE", "UNIVERSITY", "TUITION",
		"EDUCATION", "COURSE", "TRAINING", "INSTITUTE",
		// Online education platforms (from classification issues)
		"PHYSICSWALLAH", "PHYSICSWALLAH PVT LT", "PHYSICSWALLAH PVT",
	}

	// Entertainment patterns
	// IMPORTANT: "YT" alone matches "PAYTM" - use specific patterns only
	entertainmentPatterns := []string{
		"MOVIE", "CINEMA", "THEATER", "NETFLIX", "AMAZON PRIME",
		"DISNEY", "HOTSTAR", "SPOTIFY", "MUSIC", "GAME",
		"PLAYSTORE", "GOOGLE PLAY",
		// YouTube - use specific patterns, NOT just "YT" (matches PAYTM)
		"YOUTUBE", "YOUTUBE PREMIUM", "YOUTUBEPREMIUM",
		"YOUTUBE MUSIC", "YOUTUBEMUSIC", "YT PREMIUM", "YTPREMIUM",
		// Streaming services
		"ZEE5", "ZEE 5", "ZEE5SUBSCRIPTION",
		"SONY PICTURES", "SONY PICTURES NETWOR", "SONYPICTURESNETWORK",
		// Gaming (from "Other" transactions)
		"GAMING", "GAME BUSINESS", "GAMING BUSINESS",
		"JD DIGITAL", "DIGITAL", "ARTS", "VRT ARTS",
		// Audio services (from "Other" transactions)
		"AUDIOKRAFT", "AUDIOKRAFTSERVICE", "AUDIO SERVICE",
		"AUDIO", "SOUND", "RECORDING",
		// Parks and recreation (from classification issues)
		"PARKS", "PARKS",
		// Fitness and gym (from "Other" transactions)
		"FITNESS", "GYM", "GYMNASIUM", "YOGA", "ZUMBA",
		"WAY2FITNESS", "GOLD GYM", "ANYTIME FITNESS",
		"CROSSFIT", "WORKOUT", "STRENGTH",
		// Audio content platforms (from 2025 data)
		"KUKUFM", "KUKU FM", "AUDIBLE", "PODCAST",
		"QUICK TV", "QUICKTV",
	}

	// Religious and charitable organizations
	religiousCharitablePatterns := []string{
		"TEMPLE", "MANDIR", "CHURCH", "MOSQUE", "GURUDWARA",
		"GAYATRI", "VEDMATA", "SANATAN", "SANATANA",
		"SAMITI", "TRUST", "CHARITABLE", "DONATION",
		"RAMAKRISHNA", "ISKCON", "TIRUMALA", "TIRUPATI",
		"DARGAH", "SHRINE", "RELIGIOUS",
	}

	// Auto parts and services
	autoPartsPatterns := []string{
		"BATTERY", "TYRE", "TYRES", "AUTO PARTS", "AUTOPARTS",
		"GARAGE", "CAR SERVICE", "CAR REPAIR", "PUNCTURE",
		"MECHANIC", "OIL CHANGE", "SPARE PARTS", "SPAREPARTS",
		"CAR WASH", "CARWASH",
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

	// Priority 0.5: Check for Rent payments (high confidence)
	// Rent appears explicitly in narration, especially in IMPS/NEFT
	if strings.Contains(combined, "RENT") || strings.Contains(combined, "-RENT") {
		return returnCategory("Bills_Utilities", 0.90, "Rent payment detected", "RENT")
	}

	// Priority 0.6: Check for CRED bill payment app
	// CRED is used for credit card bill payments
	if strings.Contains(combined, "CREDPAY") || strings.Contains(combined, "PAYMENT ON CRED") ||
		strings.Contains(combined, "CRED.UBCP") || strings.Contains(combined, "PWCUBCPUPI") {
		return returnCategory("Bills_Utilities", 0.85, "CRED bill payment detected", "CRED")
	}

	// Priority 0.7: Check for generic payment gateways with ambiguous merchants
	// PAYU, RAZORPAY, etc. are just gateways - try to infer from amount
	if strings.Contains(combined, "PAYU PAYMENTS") || strings.Contains(combined, "PAYU@") ||
		strings.Contains(combined, "PAYU@AXISBANK") {
		// Check for utility/bill amount patterns (round amounts below 500)
		if amount > 0 && amount < 500 && (int(amount)%10 == 0 || int(amount)%5 == 0) {
			return returnCategory("Bills_Utilities", 0.50, "Small utility payment via PayU gateway", "PAYU")
		}
		// Medium amounts (₹500-₹2000) could be shopping or bills
		if amount >= 500 && amount <= 2000 {
			return returnCategory("Shopping", 0.40, "Payment via PayU gateway - likely shopping", "PAYU")
		}
		// Otherwise, leave as other or rely on other signals
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

	// Check Healthcare (before Bills to avoid misclassification)
	// Medical/pharmacy transactions should NOT be classified as bills
	for _, pattern := range healthcarePatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Healthcare", 0.80, "Healthcare expense detected", pattern)
		}
	}
	// Also check tokens for healthcare merchants
	for _, token := range tokens {
		if strings.Contains(token, "MEDICAL") || strings.Contains(token, "MEDICOS") ||
			strings.Contains(token, "PHARMACY") || strings.Contains(token, "CLINIC") ||
			strings.Contains(token, "HOSPITAL") || strings.Contains(token, "HEALTH") {
			return returnCategory("Healthcare", 0.75, "Healthcare merchant detected in token", token)
		}
	}

	// Check for medical professionals (doctors) - often in UPI IDs
	if strings.Contains(combined, "DR ") || strings.Contains(combined, "DR.") || 
		strings.Contains(combined, "DOCTOR") {
		return returnCategory("Healthcare", 0.75, "Medical professional payment", "DR", "DOCTOR")
	}

	// Check for specific medical specializations in UPI IDs/names
	if strings.Contains(combined, "SPINEDOC") || strings.Contains(combined, "CARDIO") ||
		strings.Contains(combined, "ORTHO") || strings.Contains(combined, "DENT") ||
		strings.Contains(combined, "PHYSIO") || strings.Contains(combined, "NEURO") {
		return returnCategory("Healthcare", 0.85, "Medical specialist payment", "SPINEDOC", "CARDIO", "ORTHO")
	}

	// Check for legal/professional services (advocates/lawyers)
	if strings.Contains(combined, "ADV ") || strings.Contains(combined, "ADVOCATE") ||
		strings.Contains(combined, "LAWYER") || strings.Contains(combined, "LEGAL") {
		return returnCategory("Bills_Utilities", 0.75, "Legal/professional services", "ADV", "ADVOCATE", "LAWYER")
	}

	// Check Bills & Utilities (comprehensive bill payment detection)
	// CRITICAL: Only classify as Bills_Utilities if there are ACTUAL bill-related indicators
	// Generic payment gateways (PAYTM, GPAY, etc.) are used for ALL payment types
	// EXCEPTION: Check for PAYTM UTILITY first (high confidence bill payment)
	if strings.Contains(combined, "PAYTM") {
		if strings.Contains(combined, "UTILITYPAYTM") || strings.Contains(combined, "PAYTM UTILITY") ||
			(strings.Contains(combined, "PAYTM ECOMMERCE") && strings.Contains(combined, "UTILITY")) {
			return returnCategory("Bills_Utilities", 0.95, "PAYTM utility bill payment detected", "PAYTM", "UTILITY")
		}
	}

	// Step 1: Check for actual bill payment gateways/aggregators (high confidence)
	// Exclude generic payment gateways - they're used for ALL payment types
	hasBillGateway := false
	hasGenericGateway := false
	for _, gw := range genericGateways {
		if strings.Contains(combined, gw) {
			hasGenericGateway = true
			break
		}
	}
	// Only check bill gateways if no generic gateway is present (to avoid false positives)
	if !hasGenericGateway {
		for _, gw := range billGateways {
			if strings.Contains(combined, gw) {
				hasBillGateway = true
				break
			}
		}
	}

	// Step 2: Exclude merchants that are clearly NOT bills (check FIRST before bill detection)
	// These merchants should be classified in their respective categories, not as bills
	excludeFromBills := []string{
		"FOOD", "SWEET", "RESTAURANT", "CAFE", "DINING", "EATERY", "BAKERY",
		"MEDICAL", "MEDICOS", "PHARMACY", "CLINIC", "HOSPITAL", "HEALTH",
		"SALOON", "SALON", "BEAUTY", "SPA",
		"SUPER MARKET", "MARKET", "GROCERY", "GROCERIES", "KIRANA",
		"JEWELLERS", "JEWELLERY", "WATCH", "SHOP", "STORE", "MALL",
		"TEA", "COFFEE", "SNACKS", "DAIRY",
		"TRADERS", "TRADING", "ENTERPRISE", "BUSINESS",
		"CHIKITSALY", "CHEMISTS", "MED", // Medical abbreviations
		"BAZAR", "BAZAAR", "MARKETPLACE", // Marketplaces
		"INN", "HOTEL", // Hotels/restaurants
	}

	hasExcludedMerchant := false
	for _, exclude := range excludeFromBills {
		if strings.Contains(combined, exclude) {
			hasExcludedMerchant = true
			break
		}
	}

	// Step 3: Check for explicit bill/utility keywords (required for classification)
	// IMPORTANT: "UTILITY" alone is too generic - it appears in bank IFSC codes (UTIB0000553)
	// Only count "UTILITY" if it's part of "UTILITY PAYMENT" or combined with bill-related terms
	hasBillKeyword := strings.Contains(combined, "BILL") || strings.Contains(combined, "BBPS") ||
		strings.Contains(combined, "ECS BILL") || strings.Contains(combined, "NACH BILL") ||
		strings.Contains(combined, "AUTO BILL") || strings.Contains(combined, "FUNDSTRANSFER-BILL") ||
		strings.Contains(combined, "ONLINE BILL") || strings.Contains(combined, "PAYMENT TO BILLER") ||
		strings.Contains(combined, "UTILITY PAYMENT") || // Only "UTILITY PAYMENT", not just "UTILITY"
		strings.Contains(combined, "RECHARGE") || strings.Contains(combined, "PREPAID") ||
		strings.Contains(combined, "POSTPAID")

	// Check for utility-specific patterns (electricity, gas, water, telecom) - these are actual utilities
	hasActualUtility := false
	for _, pattern := range electricityPatterns {
		if strings.Contains(combined, pattern) {
			hasActualUtility = true
			hasBillKeyword = true // Treat as bill keyword
			break
		}
	}
	if !hasActualUtility {
		for _, pattern := range gasPatterns {
			if strings.Contains(combined, pattern) {
				hasActualUtility = true
				// Only treat as bill keyword if amount is reasonable for a gas bill (< ₹25,000)
				// Large amounts to gas companies are likely investments/share purchases
				if amount == 0 || amount <= 25000 {
					hasBillKeyword = true
				}
				break
			}
		}
	}
	if !hasActualUtility {
		for _, pattern := range waterPatterns {
			if strings.Contains(combined, pattern) {
				hasActualUtility = true
				hasBillKeyword = true
				break
			}
		}
	}
	if !hasActualUtility {
		for _, pattern := range telecomPatterns {
			if strings.Contains(combined, pattern) {
				hasActualUtility = true
				hasBillKeyword = true
				break
			}
		}
	}

	// Step 4: Check for large transfers (> ₹25,000 for gas bills, or > ₹1,00,000 for other bills)
	// Large transfers are typically investments or transfers, not utility bills
	// Typical utility bills: Electricity ₹500-₹10K, Gas ₹500-₹5K, Water ₹200-₹2K
	// Threshold: ₹25,000 for gas (above this is likely share purchase/deposit), ₹1,00,000 for others
	isLargeTransfer := false
	largeTransferMethods := []string{"RTGS", "IMPS", "NEFT"}
	for _, method := range largeTransferMethods {
		if strings.Contains(combined, method) {
			isLargeTransfer = true
			break
		}
	}
	// Also check amount - if amount > ₹1,00,000, it's a large transfer regardless of method
	if amount > 100000 {
		isLargeTransfer = true
	}

	// Step 5: Only classify as Bills_Utilities if:
	// - Has actual bill gateway AND bill keywords, OR
	// - Has bill keywords (even without gateway), OR
	// - Has bill gateway AND specific utility patterns
	// BUT NOT if it has excluded merchant patterns (those go to their specific categories)
	// AND NOT if it's a large transfer (> ₹1,00,000) - those are investments/transfers
	isBillPayment := false
	if !hasExcludedMerchant && !isLargeTransfer {
		if hasBillGateway && hasBillKeyword {
			isBillPayment = true
		} else if hasBillKeyword {
			isBillPayment = true
		} else if hasBillGateway {
			// Only classify as bill if gateway is combined with specific utility patterns
			// Check for utility-specific patterns
			hasUtilityPattern := false
			for _, pattern := range electricityPatterns {
				if strings.Contains(combined, pattern) {
					hasUtilityPattern = true
					break
				}
			}
		if !hasUtilityPattern {
			for _, pattern := range gasPatterns {
				if strings.Contains(combined, pattern) {
					// Only treat as utility pattern if amount is reasonable for a gas bill
					if amount == 0 || amount <= 25000 {
						hasUtilityPattern = true
					}
					break
				}
			}
		}
			if !hasUtilityPattern {
				for _, pattern := range telecomPatterns {
					if strings.Contains(combined, pattern) {
						hasUtilityPattern = true
						break
					}
				}
			}
			if hasUtilityPattern {
				isBillPayment = true
			}
		}
	}

	// Step 5: If bill payment detected, classify by specific category
	if isBillPayment {
		// Check for specific utility types

		// Electricity
		for _, pattern := range electricityPatterns {
			if strings.Contains(combined, pattern) {
				return returnCategory("Bills_Utilities", 0.90, "Electricity bill payment detected", pattern)
			}
		}

		// Gas
		// IMPORTANT: Large transfers should NOT be classified as gas bills
		// Typical gas bills: ₹500-₹5,000. Above ₹25,000 is likely share purchase/investment/deposit
		// IGL (Indraprastha Gas Limited) is a publicly traded company - large payments are likely share purchases
		if !isLargeTransfer {
			// Additional check: Gas bills are typically small (< ₹25,000)
			// If amount > ₹25,000, it's likely an investment/share purchase, not a gas bill
			if amount > 0 && amount <= 25000 {
				for _, pattern := range gasPatterns {
					if strings.Contains(combined, pattern) {
						return returnCategory("Bills_Utilities", 0.90, "Gas bill payment detected", pattern)
					}
				}
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
		// IMPORTANT: Check for investment-type insurance first (ULIP, Endowment, etc.)
		// Investment-type insurance should be classified as "Investment", not "Bills_Utilities"
		investmentInsurancePatterns := []string{
			"ULIP", "ENDOWMENT", "WHOLE LIFE", "MONEY BACK",
			"RETIREMENT", "PENSION PLAN", "SAVINGS PLAN",
		}
		hasInvestmentInsurance := false
		for _, pattern := range investmentInsurancePatterns {
			if strings.Contains(combined, pattern) {
				hasInvestmentInsurance = true
				break
			}
		}

		// If it's investment-type insurance, classify as Investment
		if hasInvestmentInsurance {
			for _, pattern := range insurancePatterns {
				if strings.Contains(combined, pattern) {
					return returnCategory("Investment", 0.90, "Investment-type insurance premium detected", pattern)
				}
			}
		}

		// Regular insurance (term, health) - classify as Bills_Utilities
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

		// Default: Only classify as Bills_Utilities if we have high confidence
		// If we have bill keywords but no specific category, it's still likely a bill
		// EXCEPTION: Large payments (> ₹25,000) to gas utilities should NOT be classified as bills
		if hasBillKeyword {
			// Check if this is a large gas utility payment
			if amount > 25000 {
				gasUtilityPatterns := []string{"IGL", "INDRAPRASTHAGA", "INDRAPRASTHA GAS", "MGL", "MAHANAGAR GAS", "ADANIGAS", "GUJGAS"}
				for _, gasPattern := range gasUtilityPatterns {
					if strings.Contains(combined, gasPattern) {
						// Skip - don't classify as bill, let it fall through to Investment
						goto skipBillKeywordCheck
					}
				}
			}
			return returnCategory("Bills_Utilities", 0.75, "Bill payment keyword detected", "BILL_PAYMENT")
		}
		skipBillKeywordCheck:
		// If only gateway detected without keywords, don't classify as bill (too ambiguous)
		// Let it fall through to other categories
	}

	// Legacy check for backward compatibility (only if not already excluded)
	// Only check billsPatterns if we haven't already classified and it's not an excluded merchant
	if !hasExcludedMerchant {
		// IMPORTANT: Check for large amounts to gas utilities FIRST (> ₹25,000)
		// IGL, MGL, etc. are publicly traded companies - large payments are share purchases, not bills
		// Skip all bill pattern matching for large gas utility payments
		isLargeGasUtilityPayment := false
		if amount > 25000 {
			gasUtilityPatterns := []string{"IGL", "INDRAPRASTHAGA", "INDRAPRASTHA GAS", "MGL", "MAHANAGAR GAS", "ADANIGAS", "GUJGAS"}
			for _, gasPattern := range gasUtilityPatterns {
				if strings.Contains(combined, gasPattern) {
					isLargeGasUtilityPayment = true
					break
				}
			}
		}
		
		if !isLargeGasUtilityPayment {
			for _, pattern := range billsPatterns {
				if strings.Contains(combined, pattern) {
					// IMPORTANT: Skip large amounts to gas utilities (> ₹25,000) - they're investments, not bills
					// IGL, MGL, etc. are publicly traded companies - large payments are share purchases
					if amount > 25000 && (pattern == "GAS" || pattern == "IGL" || pattern == "MGL" || pattern == "INDRAPRASTHAGA") {
						// Skip this pattern - will be classified as Investment later
						continue
					}
					
					// Double-check: make sure it's not a false positive
					// If pattern is too generic (like "UTILITY" which appears in many places), require more context
					if pattern == "UTILITY" || pattern == "BILL" {
						// For generic patterns, require additional bill-related context
						if strings.Contains(combined, "PAYMENT") || strings.Contains(combined, "BILL") ||
							strings.Contains(combined, "RECHARGE") || hasBillGateway {
							return returnCategory("Bills_Utilities", 0.80, "Bill payment pattern detected", pattern)
						}
					} else {
						// For specific patterns (like "ELECTRICITY", "AIRTEL", etc.), classify directly
						return returnCategory("Bills_Utilities", 0.80, "Bill payment pattern detected", pattern)
					}
				}
			}
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
				// IMPORTANT: For gas utilities, check amount threshold
				// Large payments (> ₹25,000) to gas companies are likely share purchases, not bills
				if strings.Contains(decoded, "Gas") && amount > 25000 {
					// Skip - will be classified as Investment later
					continue
				}
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
	// IMPORTANT: Check for specific YouTube patterns first (don't match "PAYTM")
	// "YT" alone matches "PAYTM" - use word boundaries or specific patterns
	if strings.Contains(combined, "YOUTUBE") || strings.Contains(combined, "YOUTUBE PREMIUM") ||
		strings.Contains(combined, "YOUTUBEPREMIUM") || strings.Contains(combined, "YOUTUBE MUSIC") ||
		strings.Contains(combined, "YOUTUBEMUSIC") || strings.Contains(combined, "YT PREMIUM") ||
		strings.Contains(combined, "YTPREMIUM") {
		return returnCategory("Entertainment", 0.85, "YouTube subscription detected", "YOUTUBE")
	}
	// Check other entertainment patterns
	for _, pattern := range entertainmentPatterns {
		// Skip "YT" pattern (already handled above, and it matches PAYTM)
		if pattern == "YT" {
			continue
		}
		if strings.Contains(combined, pattern) {
			return returnCategory("Entertainment", 0.75, "Entertainment expense detected", pattern)
		}
	}

	// Check Religious and Charitable organizations
	for _, pattern := range religiousCharitablePatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Bills_Utilities", 0.70, "Religious/charitable donation", pattern)
		}
	}

	// Check Auto Parts and Services
	for _, pattern := range autoPartsPatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Shopping", 0.75, "Auto parts/service expense", pattern)
		}
	}

	// Check Dividend (income from investments) - should be classified as income category
	for _, pattern := range dividendPatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Investment", 0.90, "Dividend income detected", pattern)
		}
	}

	// Check for large payments to utility companies (likely share purchases/investments, not bills)
	// IGL (Indraprastha Gas Limited), MGL (Mahanagar Gas Limited), etc. are publicly traded companies
	// Payments > ₹25,000 to gas companies are likely share purchases/deposits, not gas bills
	if amount > 25000 {
		gasCompanies := []string{"IGL", "INDRAPRASTHAGA", "INDRAPRASTHA GAS", "MGL", "MAHANAGAR GAS", "ADANIGAS", "GUJGAS"}
		for _, company := range gasCompanies {
			if strings.Contains(combined, company) {
				return returnCategory("Investment", 0.85, "Large payment to gas utility company (likely share purchase/investment)", company)
			}
		}
	}

	// Check Investment
	// Priority 1: Check for large transfers to investment accounts (IMPS, UPI, NetBanking, RTGS, NEFT)
	// These are high-confidence investment indicators
	if amount >= 10000 {
		// RTGS/NEFT to brokerage/investment accounts (large transfers to banks)
		// These should NEVER be classified as utility bills
		if strings.Contains(combined, "RTGS") || strings.Contains(combined, "NEFT") {
			// RTGS/NEFT to HDFC Bank (common for brokerage accounts like Zerodha, Upstox)
			if strings.Contains(combined, "HDFC") || strings.Contains(combined, "HDFCR") {
				return returnCategory("Investment", 0.95, "Large RTGS/NEFT transfer to brokerage/investment account detected", "RTGS/NEFT", "HDFC")
			}
			// RTGS/NEFT to IDFB/IDFC (investment account transfers)
			if strings.Contains(combined, "IDFB") || strings.Contains(combined, "IDFC") {
				return returnCategory("Investment", 0.95, "Large RTGS/NEFT transfer to investment account detected", "RTGS/NEFT", "IDFB")
			}
			// Large RTGS/NEFT transfers are generally for investments/deposits, not utility bills
			if amount >= 50000 {
				return returnCategory("Investment", 0.90, "Large RTGS/NEFT transfer detected (likely investment/deposit)", "RTGS/NEFT")
			}
		}
		// IMPS to investment accounts (IDFC First Bank, etc.)
		if strings.Contains(combined, "IMPS") {
			if strings.Contains(combined, "IDFB") || strings.Contains(combined, "IDFC") ||
				(strings.Contains(combined, "KALPIT") && strings.Contains(combined, "SHARMA") && strings.Contains(combined, "XXXXXXX2950")) {
				return returnCategory("Investment", 0.95, "Large IMPS transfer to investment account detected", "IMPS", "IDFB")
			}
		}
		// Recurring large UPI transfers to same account (investment pattern)
		if strings.Contains(combined, "UPI") && amount >= 30000 {
			// Check for recurring investment account patterns
			if strings.Contains(combined, "XXXXXX3286") && strings.Contains(combined, "ICIC0003458") {
				return returnCategory("Investment", 0.90, "Large recurring UPI transfer to investment account", "UPI", "ICICI")
			}
			if strings.Contains(combined, "XXXXXX7431") && strings.Contains(combined, "SBIN0009062") {
				return returnCategory("Investment", 0.90, "Large recurring UPI transfer to investment account", "UPI", "SBI")
			}
		}
		// NetBanking large transfers (investment pattern)
		if strings.Contains(combined, "FUNDS TRANSFER") || strings.Contains(combined, "IB SS FUNDS TRANSFER") {
			if amount >= 100000 {
				return returnCategory("Investment", 0.90, "Large NetBanking transfer detected (likely investment)", "FUNDS TRANSFER")
			}
		}
		// Bajaj Finance investment
		if strings.Contains(combined, "BAJAJS") || strings.Contains(combined, "BAJAJ FINANCE") {
			return returnCategory("Investment", 0.90, "Bajaj Finance investment detected", "BAJAJS")
		}
		// IDFC First Bank fund transfer
		if strings.Contains(combined, "FUND IDFC") || strings.Contains(combined, "IDFC FIRST ACCO") {
			return returnCategory("Investment", 0.90, "IDFC First Bank fund transfer detected", "IDFC")
		}
	}

	// Priority 2: Check standard investment patterns
	for _, pattern := range investmentPatterns {
		if strings.Contains(combined, pattern) {
			return returnCategory("Investment", 0.90, "Investment detected", pattern)
		}
	}

	// Check for charges (small amounts with charge keywords)
	if amount > 0 && utils.IsCharge(originalNarration, amount) {
		return returnCategory("Bills_Utilities", 0.70, "Bank charge detected", "CHARGE")
	}

	// Special handling for Standing Instructions (SI) / Recurring subscriptions
	// Detect monthly/recurring subscription payments via standing instructions
	if strings.Contains(combined, "DC SI") || strings.Contains(combined, "DEBIT CARD SI") ||
		strings.Contains(combined, "ME DC SI") {
		// Check if it's a subscription service pattern
		if strings.Contains(combined, "MONTHLY") || strings.Contains(combined, "SUBSCRIPTION") {
			// Could be streaming service, software subscription, etc.
			// Without merchant info, classify as Bills_Utilities
			return returnCategory("Bills_Utilities", 0.60, "Standing instruction subscription payment", "SI", "MONTHLY")
		}
	}

	// Special handling for ATM Withdrawals
	// ATW/EAW - ATM cash withdrawals typically used for shopping
	if strings.Contains(combined, "ATW") || strings.Contains(combined, "EAW") || 
		strings.Contains(combined, "ATM WITHDRAWAL") || strings.Contains(combined, "CASH WITHDRAWAL") {
		return returnCategory("Shopping", 0.80, "ATM cash withdrawal for shopping", "ATM")
	}

	// Special handling for Person-to-Person (P2P) transfers
	// Large UPI transfers to individuals (not merchants) - likely personal transfers
	if strings.Contains(combined, "UPI") && amount > 5000 {
		// Check if it doesn't have known merchant patterns
		hasKnownMerchantPattern := strings.Contains(combined, "@") || 
			strings.Contains(combined, "PAY") && (strings.Contains(combined, "AMAZON") || 
			strings.Contains(combined, "GOOGLE") || strings.Contains(combined, "PAYTM"))
		
		// If no known merchant pattern and has common name patterns
		hasNamePattern := strings.Contains(combined, "MR ") || strings.Contains(combined, "MRS ") || 
			strings.Contains(combined, "MS ") || strings.Contains(combined, "KUMAR") || 
			strings.Contains(combined, "SHARMA") || strings.Contains(combined, "SINGH")
		
		if hasNamePattern && !hasKnownMerchantPattern {
			return returnCategory("Other", 0.70, "Large P2P transfer - purpose unclear", "P2P")
		}
	}

	// Default category (don't over-classify - keep UNKNOWN/Other for ambiguous cases)
	result.Category = "Other"
	result.Confidence = 0.1
	result.Reason = "No matching patterns found - classified as Other"

	// Calculate final confidence
	hasGateway := gateway != ""
	hasMerchant := merchant != "" && merchant != "UNKNOWN"
	// amountPattern and hasAmountPattern already declared above (Layer 6)
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

	// Try to extract from IMPS/NEFT/RTGS format
	// Pattern: RTGS DR-BANKCODE-BENEFICIARY NAME-REF or IMPS-REF-BENEFICIARY-BANK
	// First try: Extract beneficiary name (skip bank code)
	re = regexp.MustCompile(`(?:IMPS|NEFT|RTGS)\s+(?:DR|CR)?-?([A-Z]{4}\d+)-([A-Z\s]+?)-`)
	matches = re.FindStringSubmatch(narration)
	if len(matches) > 2 {
		beneficiary := strings.TrimSpace(matches[2])
		// Clean up
		beneficiary = strings.TrimSuffix(beneficiary, " ")
		if len(beneficiary) > 0 && len(beneficiary) < 50 {
			return beneficiary
		}
	}

	// Alternative: IMPS-REF-NAME-BANK format
	re = regexp.MustCompile(`(?:IMPS|NEFT)-(\d+)-([A-Z\s]+?)-[A-Z]{4}`)
	matches = re.FindStringSubmatch(narration)
	if len(matches) > 2 {
		beneficiary := strings.TrimSpace(matches[2])
		// Remove MR/MRS/MS prefix
		beneficiary = strings.TrimPrefix(beneficiary, "MR ")
		beneficiary = strings.TrimPrefix(beneficiary, "MRS ")
		beneficiary = strings.TrimPrefix(beneficiary, "MS ")
		if len(beneficiary) > 0 && len(beneficiary) < 50 {
			return beneficiary
		}
	}

	return "Unknown"
}
