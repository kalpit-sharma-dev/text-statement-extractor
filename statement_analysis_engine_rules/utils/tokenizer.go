package utils

import (
	"regexp"
	"strings"
)

// Tokenize splits narration by known separators
func Tokenize(narration string) []string {
	// Replace separators with space, then split
	separators := regexp.MustCompile(`[/\-_\s]+`)
	normalized := separators.ReplaceAllString(narration, " ")

	// Split by spaces and filter empty strings
	parts := strings.Fields(normalized)

	// Filter out very short tokens (likely noise)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) >= 2 {
			result = append(result, strings.ToUpper(part))
		}
	}

	return result
}

// ExtractGateway identifies payment gateway/channel from narration
func ExtractGateway(narration string) string {
	upper := strings.ToUpper(narration)

	// Gateway patterns (order matters - more specific first)
	gatewayPatterns := map[string]string{
		// Bill Payment Gateways
		"WHDF":     "BillDesk",
		"BILLDK":   "BillDesk",
		"BILLDESK": "BillDesk",
		"PAYU":     "PayU",
		"RAZP":     "Razorpay",
		"RAZORPAY": "Razorpay",
		"CCAVENUE": "CCAvenue",
		"BBPS":     "BBPS",
		"PAYGOV":   "PayGov",
		"SBIPG":    "SBI Payment Gateway",
		"AXISPG":   "Axis Payment Gateway",
		"ICICIPG":  "ICICI Payment Gateway",
		"KOTAKPG":  "Kotak Payment Gateway",
		"YESPG":    "Yes Bank Payment Gateway",
		// Other payment methods
		"UPI":         "UPI",
		"NET BANKING": "NetBanking",
		"IB ":         "NetBanking",
		"IB-":         "NetBanking",
		"ECS":         "ECS",
		"IMPS":        "IMPS",
		"NEFT":        "NEFT",
		"RTGS":        "RTGS",
		"ACH":         "ACH",
	}

	for pattern, gateway := range gatewayPatterns {
		if strings.Contains(upper, pattern) {
			return gateway
		}
	}

	return ""
}

// ExtractMerchant attempts to decode compressed merchant names
func ExtractMerchant(tokens []string) string {
	// Known merchant patterns (compressed names)
	merchantPatterns := map[string]string{
		// Wallets
		"MOBIKWIK":  "Mobikwik",
		"PAYTM":     "Paytm",
		"PHONEPE":   "PhonePe",
		"GPAY":      "GooglePay",
		"AMAZONPAY": "AmazonPay",

		// Insurance
		"HDFCLIFE":        "HDFC Life",
		"STANDARDLIFE":    "HDFC Standard Life",
		"LIC":             "LIC",
		"SBILIFE":         "SBI Life",
		"MAXLIFE":         "Max Life",
		"ICICIPRUDENTIAL": "ICICI Prudential",
		"BAJAJALLIANZ":    "Bajaj Allianz",

		// Utilities
		"IGL":      "Indraprastha Gas",
		"PVVNL":    "PVVNL",
		"AIRTEL":   "Airtel",
		"JIO":      "Jio",
		"VODAFONE": "Vodafone",
		"BSNL":     "BSNL",
		"MSEDCL":   "Maharashtra State Electricity",
		"EL":       "Electricity",

		// Fuel / Petrol / Diesel
		"IOCL":               "Indian Oil",
		"INDIANOIL":          "Indian Oil",
		"BPCL":               "Bharat Petroleum",
		"BHARATPETROLEUM":    "Bharat Petroleum",
		"HPCL":               "Hindustan Petroleum",
		"HINDUSTANPETROLEUM": "Hindustan Petroleum",
		"SHELL":              "Shell",
		"ESSAR":              "Essar",
		"NAYARA":             "Nayara Energy",
		"PETROL":             "Petrol",
		"DIESEL":             "Diesel",

		// Travel - Cab
		"UBER":     "Uber",
		"UBERTRIP": "Uber",
		"OLA":      "Ola",
		"OLACABS":  "Ola Cabs",
		"RAPIDO":   "Rapido",
		// Travel - Railways
		"IRCTC":         "IRCTC",
		"IRCTCIPAY":     "IRCTC",
		"RAZPIRCTCIPAY": "IRCTC",
		// Travel - Bus
		"REDBUS":     "RedBus",
		"ABHIBUS":    "AbhiBus",
		"YATRAGENIE": "Yatra Genie",
		// Travel - Flight
		"MAKEMYTRIP": "MakeMyTrip",
		"MMT":        "MakeMyTrip",
		"GOIBIBO":    "Goibibo",
		"YATRA":      "Yatra",
		"CLEARTRIP":  "Cleartrip",
		// Travel - Hotels
		"OYO":       "OYO",
		"OYOROOMS":  "OYO Rooms",
		"TREEBO":    "Treebo",
		"FABHOTELS": "FabHotels",
		"AIRBNB":    "Airbnb",

		// Food Delivery
		"SWIGGY":          "Swiggy",
		"SWIGGYINSTAMART": "Swiggy Instamart",
		"ZOMATO":          "Zomato",
		"ZOMATOONLINE":    "Zomato",
		"ZOMATOORDER":     "Zomato",
		"FAASOS":          "Faasos",
		"EATSURE":         "EatSure",
		"BOX8":            "Box8",
		"DOMINOS":         "Dominos",
		"MCDONALDS":       "McDonald's",
		"KFC":             "KFC",
		"BURGERKING":      "Burger King",
		"SUBWAY":          "Subway",
		"PIZZAHUT":        "Pizza Hut",

		// Shopping - E-commerce
		"AMAZON":     "Amazon",
		"FLIPKART":   "Flipkart",
		"FLIPKARTIN": "Flipkart",
		"MYNTRA":     "Myntra",
		"AJIO":       "Ajio",
		"MEESHO":     "Meesho",
		"SIMPL":      "Simpl",
		"GETSIMPL":   "Simpl",
		// Shopping - Fashion/Lifestyle
		"ZARA":       "Zara",
		"HNM":        "H&M",
		"PANTALOONS": "Pantaloons",
		"LIFESTYLE":  "Lifestyle",

		// Investment - Mutual Funds
		"GROWW":  "Groww",
		"COIN":   "Coin",
		"UPSTOX": "Upstox",
		"KITE":   "Zerodha Kite",
		// Investment - Stocks/Trading
		"NSE": "NSE",
		"BSE": "BSE",
		// Investment - Clearing Corporations
		"NSDL": "NSDL",
		"CDSL": "CDSL",
		// Investment - NPS/Pension
		"NPS":      "National Pension System",
		"NPSTRUST": "NPS Trust",
		// Investment - Stock Broking
		"ZERODHA":        "Zerodha",
		"ZERODHABROKING": "Zerodha Broking",
		"BROKING":        "Broking",
		"HSL":            "HSL Securities",
		"SEC":            "Securities",
		// Investment - Crypto Exchanges (Indian)
		"WAZIRX":               "WazirX",
		"WAZIRXIN":             "WazirX",
		"ZANMAI":               "WazirX (Zanmai Labs)",
		"ZANMAILABS":           "WazirX (Zanmai Labs)",
		"ZANMAI LABS":          "WazirX (Zanmai Labs)",
		"COINDCX":              "CoinDCX",
		"NEBULAS":              "CoinDCX (Nebulas Technologies)",
		"NEBULASTECHNOLOGIES":  "CoinDCX (Nebulas Technologies)",
		"NEBULAS TECHNOLOGIES": "CoinDCX (Nebulas Technologies)",
		"DCX":                  "CoinDCX",
		"COINSWITCH":           "CoinSwitch Kuber",
		"COINSWITCHKUBER":      "CoinSwitch Kuber",
		"BITCIPHER":            "CoinSwitch Kuber (Bitcipher Labs)",
		"BITCIPHER LABS":       "CoinSwitch Kuber (Bitcipher Labs)",
		"ZEBPAY":               "ZebPay",
		"ZEBITSERVICE":         "ZebPay (Zeb IT Service)",
		"ZEB IT SERVICE":       "ZebPay (Zeb IT Service)",
		"UNOCOIN":              "Unocoin",
		"UNOCOMMERCE":          "Unocoin",
		// Investment - Crypto Exchanges (International)
		"BINANCE":    "Binance",
		"BINANCEPAY": "Binance",
		"BIFINANCE":  "Binance",
		"COINBASE":   "Coinbase",
		"CBPAY":      "Coinbase",
		"CB PAY":     "Coinbase",
		"KRAKEN":     "Kraken",
		"PAYWARD":    "Kraken (Payward)",
		"CRYPTOCOM":  "Crypto.com",
		"FORIS":      "Crypto.com (Foris)",
		"KUCOIN":     "KuCoin",
		"MEKGLOBAL":  "KuCoin (Mek Global)",
		"MEK GLOBAL": "KuCoin (Mek Global)",
		"BITSTAMP":   "Bitstamp",
	}

	// Check each token against known patterns
	for _, token := range tokens {
		if merchant, found := merchantPatterns[token]; found {
			return merchant
		}
	}

	// Try to decode compressed names by checking if token contains known substrings
	for _, token := range tokens {
		// Check for common compressed patterns
		if strings.Contains(token, "NPS") {
			return "National Pension System"
		}
		if strings.Contains(token, "IGL") {
			return "Indraprastha Gas"
		}
		if strings.Contains(token, "INDRAPRASTHA") || strings.Contains(token, "INDRAPRASTHAGA") {
			return "Indraprastha Gas"
		}
		// Crypto exchange detection
		if strings.Contains(token, "ZANMAI") {
			return "WazirX (Zanmai Labs)"
		}
		if strings.Contains(token, "NEBULAS") {
			return "CoinDCX (Nebulas Technologies)"
		}
		if strings.Contains(token, "BITCIPHER") {
			return "CoinSwitch Kuber (Bitcipher Labs)"
		}
		if strings.Contains(token, "WAZIRX") {
			return "WazirX"
		}
		if strings.Contains(token, "COINDCX") {
			return "CoinDCX"
		}
		if strings.Contains(token, "COINSWITCH") {
			return "CoinSwitch Kuber"
		}
		// Vegetables and Fruits detection
		if strings.Contains(token, "VEGETABLE") || strings.Contains(token, "VEG") {
			return "Vegetable Shop"
		}
		if strings.Contains(token, "FRUIT") || strings.Contains(token, "FRUITS") {
			return "Fruit Shop"
		}
	}

	return ""
}

// HasKeyword checks if narration contains semantic keywords
func HasKeyword(narration string, keywords []string) bool {
	upper := strings.ToUpper(narration)
	for _, keyword := range keywords {
		if strings.Contains(upper, keyword) {
			return true
		}
	}
	return false
}

// IsCharge detects if transaction is likely a charge/fee
func IsCharge(narration string, amount float64) bool {
	upper := strings.ToUpper(narration)

	// Charge keywords
	chargeKeywords := []string{
		"CHG", "CHARGE", "FEE", "SMS", "ALERT",
		"INSTA ALERT", "MAINTENANCE", "SERVICE CHARGE",
		"PROCESSING FEE", "CONVENIENCE FEE",
	}

	// Check for charge keywords
	for _, keyword := range chargeKeywords {
		if strings.Contains(upper, keyword) {
			return true
		}
	}

	// Small amounts are often charges
	if amount > 0 && amount < 50 {
		// But exclude known non-charge small amounts (like UPI transactions)
		if !strings.Contains(upper, "UPI") && !strings.Contains(upper, "PAYTM") {
			return true
		}
	}

	return false
}

// DetectWallet identifies wallet apps from tokens
func DetectWallet(tokens []string) string {
	walletPatterns := map[string]string{
		"MOBIKWIK":   "Mobikwik",
		"PAYTM":      "Paytm",
		"PHONEPE":    "PhonePe",
		"GPAY":       "GooglePay",
		"AMAZONPAY":  "AmazonPay",
		"FREECHARGE": "Freecharge",
		"JIO":        "JioMoney",
	}

	for _, token := range tokens {
		if wallet, found := walletPatterns[token]; found {
			return wallet
		}
	}

	return ""
}

// DecodeCompressedMerchant attempts to decode compressed merchant names
// by reading syllables and matching against known patterns
func DecodeCompressedMerchant(compressed string) string {
	upper := strings.ToUpper(compressed)

	// Common compressed patterns
	patterns := map[string]string{
		// Insurance
		"HDFCLIFE":     "HDFC Life Insurance",
		"STANDARDLIFE": "HDFC Standard Life",
		"SBILIFE":      "SBI Life Insurance",
		"MAXLIFE":      "Max Life Insurance",

		// Utilities
		"INDRAPRASTHAGA": "Indraprastha Gas",
		"IGL":            "Indraprastha Gas Limited",
		"PVVNL":          "Purvanchal Vidyut Vitran Nigam",

		// Investment
		"INDIANCLEARINGCORPORATION":         "Indian Clearing Corporation",
		"INDIANCCLEARINGCORPORATION":        "Indian Clearing Corporation",
		"INDIANCCLEARINGCORPORATIONLIMITED": "Indian Clearing Corporation Limited",
		// Crypto Exchanges
		"ZANMAILABS":          "WazirX (Zanmai Labs)",
		"NEBULASTECHNOLOGIES": "CoinDCX (Nebulas Technologies)",
		"BITCIPHERLABS":       "CoinSwitch Kuber (Bitcipher Labs)",
		"ZEBITSERVICE":        "ZebPay (Zeb IT Service)",
	}

	// Direct match
	if merchant, found := patterns[upper]; found {
		return merchant
	}

	// Try partial matches
	for pattern, merchant := range patterns {
		if strings.Contains(upper, pattern) {
			return merchant
		}
	}

	return compressed // Return original if no match
}
