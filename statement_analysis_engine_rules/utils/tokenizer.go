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
		"WHDF":           "BillDesk",
		"PAYU":           "PayU",
		"RAZP":           "Razorpay",
		"RAZORPAY":       "Razorpay",
		"BILLDK":         "BillDesk",
		"BILLDESK":       "BillDesk",
		"UPI":            "UPI",
		"NET BANKING":    "NetBanking",
		"IB ":            "NetBanking",
		"IB-":            "NetBanking",
		"ECS":            "ECS",
		"IMPS":           "IMPS",
		"NEFT":           "NEFT",
		"RTGS":           "RTGS",
		"ACH":            "ACH",
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
		"HDFCLIFE":     "HDFC Life",
		"STANDARDLIFE": "HDFC Standard Life",
		"LIC":          "LIC",
		"SBILIFE":      "SBI Life",
		"MAXLIFE":      "Max Life",
		"ICICIPRUDENTIAL": "ICICI Prudential",
		"BAJAJALLIANZ":   "Bajaj Allianz",
		
		// Utilities
		"IGL":      "Indraprastha Gas",
		"PVVNL":    "PVVNL",
		"AIRTEL":   "Airtel",
		"JIO":      "Jio",
		"VODAFONE": "Vodafone",
		"BSNL":     "BSNL",
		"MSEDCL":   "Maharashtra State Electricity",
		"EL":       "Electricity",
		
		// Travel
		"IRCTC":        "IRCTC",
		"IRCTCIPAY":    "IRCTC",
		"RAZPIRCTCIPAY": "IRCTC",
		"UBER":         "Uber",
		"OLA":          "Ola",
		
		// Food
		"SWIGGY":  "Swiggy",
		"ZOMATO":  "Zomato",
		"DOMINOS": "Dominos",
		
		// Shopping
		"AMAZON":   "Amazon",
		"FLIPKART": "Flipkart",
		"MYNTRA":   "Myntra",
		"SIMPL":    "Simpl",
		"GETSIMPL": "Simpl",
		
		// Investment
		"NSDL": "NSDL",
		"CDSL": "CDSL",
		"NPS":  "National Pension System",
		"ZERODHA": "Zerodha",
		"ZERODHABROKING": "Zerodha Broking",
		"BROKING": "Broking",
		"HSL": "HSL Securities",
		"SEC": "Securities",
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
		"MOBIKWIK":  "Mobikwik",
		"PAYTM":     "Paytm",
		"PHONEPE":   "PhonePe",
		"GPAY":      "GooglePay",
		"AMAZONPAY": "AmazonPay",
		"FREECHARGE": "Freecharge",
		"JIO":       "JioMoney",
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
		"INDIANCLEARINGCORPORATION":        "Indian Clearing Corporation",
		"INDIANCCLEARINGCORPORATION":       "Indian Clearing Corporation",
		"INDIANCCLEARINGCORPORATIONLIMITED": "Indian Clearing Corporation Limited",
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

