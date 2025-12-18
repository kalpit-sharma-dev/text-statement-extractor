package utils

import (
	"strings"
)

// KnownMerchant represents a known merchant with its default category
type KnownMerchant struct {
	Patterns []string // All known patterns/aliases
	Name     string   // Canonical name
	Category string   // Default category (can be overridden)
	Confidence float64 // Base confidence for this merchant
}

// KnownMerchants is a comprehensive map of known merchants
// Layer 4: Merchant / Entity Identification (MOST IMPORTANT)
var KnownMerchants = []KnownMerchant{
	// Food Delivery
	{Patterns: []string{"ZOMATO", "ZOMATOONLINE", "ZMT"}, Name: "Zomato", Category: "Food_Delivery", Confidence: 0.9},
	{Patterns: []string{"SWIGGY", "SWIGGYINSTAMART"}, Name: "Swiggy", Category: "Food_Delivery", Confidence: 0.9},
	{Patterns: []string{"FAASOS", "EATSURE", "BOX8"}, Name: "Food Delivery Apps", Category: "Food_Delivery", Confidence: 0.85},

	// Travel
	{Patterns: []string{"UBER", "UBERTRIP"}, Name: "Uber", Category: "Travel", Confidence: 0.9},
	{Patterns: []string{"OLA", "OLACABS"}, Name: "Ola", Category: "Travel", Confidence: 0.9},
	{Patterns: []string{"IRCTC", "IRCTCIPAY"}, Name: "IRCTC", Category: "Travel", Confidence: 0.9},
	{Patterns: []string{"MAKEMYTRIP", "MMT"}, Name: "MakeMyTrip", Category: "Travel", Confidence: 0.9},
	{Patterns: []string{"GOIBIBO", "YATRA", "CLEARTRIP"}, Name: "Travel Booking", Category: "Travel", Confidence: 0.85},
	{Patterns: []string{"OYO", "OYOROOMS"}, Name: "Oyo", Category: "Travel", Confidence: 0.85},
	{Patterns: []string{"INDIGO", "INDIGO AIRLINES", "INDIGO."}, Name: "IndiGo Airlines", Category: "Travel", Confidence: 0.9},

	// Fuel
	{Patterns: []string{"IOCL", "INDIANOIL"}, Name: "Indian Oil", Category: "Fuel", Confidence: 0.9},
	{Patterns: []string{"BPCL", "BHARATPETROLEUM"}, Name: "Bharat Petroleum", Category: "Fuel", Confidence: 0.9},
	{Patterns: []string{"HPCL", "HINDUSTANPETROLEUM"}, Name: "Hindustan Petroleum", Category: "Fuel", Confidence: 0.9},
	{Patterns: []string{"DAUJI SERVICE STATIO", "PHOOL SERVICE STATIO", "SERVICE STATIO"}, Name: "Service Station", Category: "Fuel", Confidence: 0.8},

	// Utilities
	{Patterns: []string{"IGL", "INDRAPRASTHAGA"}, Name: "Indraprastha Gas", Category: "Bills_Utilities", Confidence: 0.9},
	{Patterns: []string{"PVVNL", "MSEDCL", "BSES"}, Name: "Electricity Board", Category: "Bills_Utilities", Confidence: 0.9},
	{Patterns: []string{"AIRTEL", "JIO", "VODAFONE", "IDEA", "BSNL"}, Name: "Telecom", Category: "Bills_Utilities", Confidence: 0.9},

	// Investment
	{Patterns: []string{"ZERODHA", "ZERODHA BROKING"}, Name: "Zerodha", Category: "Investment", Confidence: 0.9},
	{Patterns: []string{"GROWW", "COIN", "UPSTOX"}, Name: "Investment Apps", Category: "Investment", Confidence: 0.85},
	{Patterns: []string{"INDIAN CLEARING CORPORATION", "NSDL", "CDSL"}, Name: "Clearing Corporation", Category: "Investment", Confidence: 0.9},

	// Shopping
	{Patterns: []string{"AMAZON", "AMAZONPAY"}, Name: "Amazon", Category: "Shopping", Confidence: 0.9},
	{Patterns: []string{"FLIPKART", "FLIPKARTIN"}, Name: "Flipkart", Category: "Shopping", Confidence: 0.9},
	{Patterns: []string{"MYNTRA", "AJIO", "MEESHO"}, Name: "Fashion E-commerce", Category: "Shopping", Confidence: 0.85},
	{Patterns: []string{"MAHA GANESH TRADERS", "MAHA GANESH"}, Name: "Maha Ganesh Traders", Category: "Groceries", Confidence: 0.8},
	{Patterns: []string{"PANCHAM SUPER MARKET", "PANCHAM SUPERMARKET"}, Name: "Pancham Super Market", Category: "Groceries", Confidence: 0.8},
	{Patterns: []string{"AGGARWAL TRADING COM", "AGGARWAL TRADING"}, Name: "Aggarwal Trading", Category: "Shopping", Confidence: 0.75},
	{Patterns: []string{"BALAJI TRADING COMPA", "BALAJI TRADING"}, Name: "Balaji Trading", Category: "Shopping", Confidence: 0.75},
	{Patterns: []string{"POOJA STATIONERY"}, Name: "Pooja Stationery", Category: "Shopping", Confidence: 0.8},
	{Patterns: []string{"BIKANERVALA", "BIKANERVALA PRIVATE"}, Name: "Bikanervala", Category: "Shopping", Confidence: 0.8},
	{Patterns: []string{"BOMBAY WATCH COMPANY"}, Name: "Bombay Watch Company", Category: "Shopping", Confidence: 0.8},
	{Patterns: []string{"BATTERY", "AUTO BATTERY", "ANIKET BATTERY"}, Name: "Auto Parts Shop", Category: "Shopping", Confidence: 0.75},
	{Patterns: []string{"ENTERPRISES"}, Name: "General Store/Enterprise", Category: "Shopping", Confidence: 0.5},

	// Groceries
	{Patterns: []string{"BIGBASKET", "BBNOW"}, Name: "BigBasket", Category: "Groceries", Confidence: 0.9},
	{Patterns: []string{"GROFERS", "BLINKIT"}, Name: "Grocery Apps", Category: "Groceries", Confidence: 0.85},
	{Patterns: []string{"ZEPTO", "ZEPTO MARKETPLACE"}, Name: "Zepto", Category: "Groceries", Confidence: 0.9},
	{Patterns: []string{"BALAJI KHOA PANEER", "KHOA PANEER"}, Name: "Balaji Khoa Paneer", Category: "Groceries", Confidence: 0.75},

	// Dining
	{Patterns: []string{"BAMRADA SONS", "BAMRADA"}, Name: "Bamrada Sons", Category: "Dining", Confidence: 0.75},
	{Patterns: []string{"SPECIAL CHAT CENTER", "SPECIAL CHAT"}, Name: "Special Chat Center", Category: "Dining", Confidence: 0.75},
	{Patterns: []string{"MUSKAN BAKERS", "MUSKAN BAKERS AND CO"}, Name: "Muskan Bakers", Category: "Dining", Confidence: 0.75},
	{Patterns: []string{"ROSIER FOODS", "ROSIER"}, Name: "Rosier Foods", Category: "Dining", Confidence: 0.75},
	{Patterns: []string{"PANCHAITEA", "PANCHAI TEA"}, Name: "Panchai Tea", Category: "Dining", Confidence: 0.75},
	{Patterns: []string{"CATERERS", "CATERING", "BALAJI CATERERS", "SHRI BALAJI CATERERS"}, Name: "Catering Service", Category: "Dining", Confidence: 0.8},
	{Patterns: []string{"DAGDUSHET COUNTER"}, Name: "Dagdushet Temple Counter", Category: "Dining", Confidence: 0.7},

	// Healthcare
	{Patterns: []string{"NOBLE CHEMISTS", "NOBLE CHEMIST"}, Name: "Noble Chemists", Category: "Healthcare", Confidence: 0.8},
	{Patterns: []string{"KAIWALYA MEDICO", "KAIWALYA"}, Name: "Kaiwalya Medico", Category: "Healthcare", Confidence: 0.8},
	{Patterns: []string{"APOLLO", "FORTIS", "MAX"}, Name: "Hospital Chains", Category: "Healthcare", Confidence: 0.9},
	{Patterns: []string{"WAY2FITNESS", "FITNESS", "GYM"}, Name: "Fitness Center", Category: "Healthcare", Confidence: 0.85},

	// Entertainment
	{Patterns: []string{"SONY PICTURES", "SONYPICTURESNETWORK"}, Name: "Sony Pictures", Category: "Entertainment", Confidence: 0.9},
	{Patterns: []string{"ISTHARA PARKS", "ISTHARA PARKS PRIVAT"}, Name: "Isthara Parks", Category: "Entertainment", Confidence: 0.8},
	{Patterns: []string{"NETFLIX", "AMAZON PRIME", "DISNEY", "HOTSTAR"}, Name: "Streaming Services", Category: "Entertainment", Confidence: 0.9},
	{Patterns: []string{"ZEE5", "ZEE 5"}, Name: "Zee5", Category: "Entertainment", Confidence: 0.85},
	{Patterns: []string{"ARCHAEOLOGICAL SURVE", "ARCHAEOLOGICAL", "MUSEUM"}, Name: "Tourism/Heritage Site", Category: "Entertainment", Confidence: 0.75},

	// Education
	{Patterns: []string{"PHYSICSWALLAH", "PHYSICSWALLAH PVT LT"}, Name: "PhysicsWallah", Category: "Education", Confidence: 0.9},
}

// DetectKnownMerchant detects if narration contains a known merchant
// Returns merchant name, category, and confidence
func DetectKnownMerchant(narration string, merchant string) (string, string, float64) {
	upper := strings.ToUpper(narration + " " + merchant)

	for _, knownMerchant := range KnownMerchants {
		for _, pattern := range knownMerchant.Patterns {
			if strings.Contains(upper, strings.ToUpper(pattern)) {
				return knownMerchant.Name, knownMerchant.Category, knownMerchant.Confidence
			}
		}
	}

	return "", "", 0.0
}

