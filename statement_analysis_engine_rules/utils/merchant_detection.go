package utils

import (
	"strings"
)

// KnownMerchant represents a known merchant with its default category
type KnownMerchant struct {
	Patterns   []string // All known patterns/aliases
	Name       string   // Canonical name
	Category   string   // Default category (can be overridden)
	Confidence float64  // Base confidence for this merchant
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
	// Generic pattern from local merchants (work for all customers)
	{Patterns: []string{"SERVICE STATIO", "SERVICE STATION"}, Name: "Service Station", Category: "Fuel", Confidence: 0.8},

	// Utilities
	{Patterns: []string{"IGL", "INDRAPRASTHAGA"}, Name: "Indraprastha Gas", Category: "Bills_Utilities", Confidence: 0.9},
	{Patterns: []string{"PVVNL", "MSEDCL", "BSES"}, Name: "Electricity Board", Category: "Bills_Utilities", Confidence: 0.9},
	{Patterns: []string{"AIRTEL", "JIO", "VODAFONE", "IDEA", "BSNL"}, Name: "Telecom", Category: "Bills_Utilities", Confidence: 0.9},

	// Investment
	{Patterns: []string{"ZERODHA", "ZERODHA BROKING"}, Name: "Zerodha", Category: "Investment", Confidence: 0.9},
	{Patterns: []string{"GROWW", "COIN", "UPSTOX"}, Name: "Investment Apps", Category: "Investment", Confidence: 0.85},
	{Patterns: []string{"INDIAN CLEARING CORPORATION", "NSDL", "CDSL"}, Name: "Clearing Corporation", Category: "Investment", Confidence: 0.9},

	// Fintech/Payment Services (classify as Bills_Utilities for bill payments)
	{Patterns: []string{"EFPI TECHNOLOGIES", "EFPI@RBL"}, Name: "EFPI Technologies", Category: "Bills_Utilities", Confidence: 0.75},
	{Patterns: []string{"ALZAPAY TECHNOLOGY", "LYRA@RBL"}, Name: "AlzaPay", Category: "Bills_Utilities", Confidence: 0.75},

	// Shopping
	{Patterns: []string{"AMAZON", "AMAZONPAY"}, Name: "Amazon", Category: "Shopping", Confidence: 0.9},
	{Patterns: []string{"FLIPKART", "FLIPKARTIN"}, Name: "Flipkart", Category: "Shopping", Confidence: 0.9},
	{Patterns: []string{"MYNTRA", "AJIO", "MEESHO"}, Name: "Fashion E-commerce", Category: "Shopping", Confidence: 0.85},
	// Generic patterns from local merchants (work for all customers)
	{Patterns: []string{"TRADERS", "TRADING"}, Name: "Trading Company", Category: "Shopping", Confidence: 0.75},
	{Patterns: []string{"SUPER MARKET", "SUPERMARKET"}, Name: "Super Market", Category: "Groceries", Confidence: 0.8},
	{Patterns: []string{"STATIONERY", "STATIONARY"}, Name: "Stationery Shop", Category: "Shopping", Confidence: 0.8},
	{Patterns: []string{"WATCH COMPANY", "WATCH"}, Name: "Watch Shop", Category: "Shopping", Confidence: 0.75},
	{Patterns: []string{"BATTERY", "AUTO BATTERY"}, Name: "Auto Parts Shop", Category: "Shopping", Confidence: 0.75},

	// Groceries
	{Patterns: []string{"BIGBASKET", "BBNOW"}, Name: "BigBasket", Category: "Groceries", Confidence: 0.9},
	{Patterns: []string{"GROFERS", "BLINKIT"}, Name: "Grocery Apps", Category: "Groceries", Confidence: 0.85},
	{Patterns: []string{"ZEPTO", "ZEPTO MARKETPLACE"}, Name: "Zepto", Category: "Groceries", Confidence: 0.9},
	// Generic patterns from local merchants (work for all customers)
	{Patterns: []string{"PANEER", "KHOA PANEER"}, Name: "Dairy Product", Category: "Groceries", Confidence: 0.75},

	// Dining
	// Generic patterns from local merchants (work for all customers)
	{Patterns: []string{"CATERERS", "CATERING"}, Name: "Catering Service", Category: "Dining", Confidence: 0.8},
	{Patterns: []string{"BAKERS", "BAKERY"}, Name: "Bakery", Category: "Dining", Confidence: 0.8},
	{Patterns: []string{"CHAT", "CHAT CENTER", "CHAT CENTRE"}, Name: "Chat Center", Category: "Dining", Confidence: 0.75},
	{Patterns: []string{"TEA", "TEA SHOP", "TEA STALL"}, Name: "Tea Shop", Category: "Dining", Confidence: 0.75},

	// Dairy Shops (Groceries, NOT Dining - they sell milk, paneer, curd etc.)
	// Keep generic patterns only, removed specific local dairy names
	{Patterns: []string{"DAIRY", "DAIRY AND SWEE", "DAIRY AND SWEET"}, Name: "Dairy Shop", Category: "Groceries", Confidence: 0.85},
	{Patterns: []string{"MILK SHOP", "MILK STORE", "DOODH"}, Name: "Milk Shop", Category: "Groceries", Confidence: 0.85},

	// Healthcare
	// Keep major hospital chains and generic patterns
	{Patterns: []string{"APOLLO", "FORTIS", "MAX"}, Name: "Hospital Chains", Category: "Healthcare", Confidence: 0.9},
	{Patterns: []string{"WAY2FITNESS", "FITNESS", "GYM"}, Name: "Fitness Center", Category: "Healthcare", Confidence: 0.85},
	// Generic patterns from local merchants (work for all customers)
	{Patterns: []string{"CHEMISTS", "CHEMIST"}, Name: "Pharmacy", Category: "Healthcare", Confidence: 0.8},
	{Patterns: []string{"MEDICO", "MEDICAL"}, Name: "Medical Store", Category: "Healthcare", Confidence: 0.75},

	// Entertainment
	{Patterns: []string{"SONY PICTURES", "SONYPICTURESNETWORK"}, Name: "Sony Pictures", Category: "Entertainment", Confidence: 0.9},
	{Patterns: []string{"NETFLIX", "AMAZON PRIME", "DISNEY", "HOTSTAR"}, Name: "Streaming Services", Category: "Entertainment", Confidence: 0.9},
	{Patterns: []string{"ZEE5", "ZEE 5"}, Name: "Zee5", Category: "Entertainment", Confidence: 0.85},
	// Generic patterns for tourism/heritage and parks (work for all customers)
	{Patterns: []string{"ARCHAEOLOGICAL SURVE", "ARCHAEOLOGICAL", "MUSEUM"}, Name: "Tourism/Heritage Site", Category: "Entertainment", Confidence: 0.75},
	{Patterns: []string{"PARKS", "PARK"}, Name: "Park/Recreation", Category: "Entertainment", Confidence: 0.75},

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
