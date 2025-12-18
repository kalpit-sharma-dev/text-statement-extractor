package utils

import "strings"

// CanonicalMerchant represents a canonical merchant with aliases
type CanonicalMerchant struct {
	Aliases  []string // All known aliases/variations
	Name     string   // Canonical name
	Category string   // Default category (can be overridden)
}

// MerchantCanonicalMap maps aliases to canonical merchants
var MerchantCanonicalMap = map[string]CanonicalMerchant{
	// Gas Utilities
	"IGL": {
		Aliases:  []string{"IGL", "INDRAPRASTHAGA", "INDRAPRASTHA GAS", "INDRAP GAS LTD", "INDRAPRASTHAGAS"},
		Name:     "Indraprastha Gas Limited",
		Category: "UTILITY_GAS",
	},
	"MGL": {
		Aliases:  []string{"MGL", "MAHANAGAR GAS", "MAHANAGAR GAS LIMITED"},
		Name:     "Mahanagar Gas Limited",
		Category: "UTILITY_GAS",
	},
	
	// Electricity
	"MSEDCL": {
		Aliases:  []string{"MSEDCL", "MAHARASHTRA STATE EL", "MAHARASHTRA STATE ELECTRICITY", "MAHARASHTRA STATE", "EL"},
		Name:     "Maharashtra State Electricity Distribution Company",
		Category: "UTILITY_ELECTRICITY",
	},
	"BSES": {
		Aliases:  []string{"BSES", "BSESR", "BSESRAJDHANI", "BSESYAMUNA"},
		Name:     "BSES",
		Category: "UTILITY_ELECTRICITY",
	},
	
	// Food Delivery
	"ZOMATO": {
		Aliases:  []string{"ZOMATO", "ZOMATOONLINE", "ZOMATOINDIA", "ZOMATOORDER", "ZMT"},
		Name:     "Zomato",
		Category: "Food_Delivery",
	},
	"SWIGGY": {
		Aliases:  []string{"SWIGGY", "SWIGGYINSTAMART", "SWIGGYONLINE", "SWIGGYORDER"},
		Name:     "Swiggy",
		Category: "Food_Delivery",
	},
	
	// Investment
	"ZERODHA": {
		Aliases:  []string{"ZERODHA", "ZERODHA BROKING", "ZERODHA BROKING LTD", "ZERODHABROKING", "BROKING", "BROKING LTD"},
		Name:     "Zerodha",
		Category: "Investment",
	},
}

// CanonicalizeMerchant normalizes merchant name using canonicalization map
func CanonicalizeMerchant(merchant string) (string, string) {
	upper := strings.ToUpper(strings.TrimSpace(merchant))
	if upper == "" {
		return "", ""
	}
	
	// Direct match
	if canonical, found := MerchantCanonicalMap[upper]; found {
		return canonical.Name, canonical.Category
	}
	
	// Check aliases
	for _, canonical := range MerchantCanonicalMap {
		for _, alias := range canonical.Aliases {
			if strings.Contains(upper, alias) || strings.Contains(alias, upper) {
				return canonical.Name, canonical.Category
			}
		}
	}
	
	// Partial match (fuzzy)
	for key, canonical := range MerchantCanonicalMap {
		if strings.Contains(upper, key) || strings.Contains(key, upper) {
			return canonical.Name, canonical.Category
		}
	}
	
	return merchant, "" // Return original if no match
}

