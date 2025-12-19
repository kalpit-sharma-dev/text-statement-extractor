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
	
	// Crypto Exchanges (Indian)
	"WAZIRX": {
		Aliases:  []string{"WAZIRX", "WAZIRXIN", "ZANMAI", "ZANMAI LABS", "ZANMAILABS", "ZANMAI LABS PRIVATE LIMITED"},
		Name:     "WazirX",
		Category: "Investment",
	},
	"COINDCX": {
		Aliases:  []string{"COINDCX", "NEBULAS", "NEBULAS TECHNOLOGIES", "NEBULASTECHNOLOGIES", "DCX"},
		Name:     "CoinDCX",
		Category: "Investment",
	},
	"COINSWITCH": {
		Aliases:  []string{"COINSWITCH", "COINSWITCHKUBER", "BITCIPHER", "BITCIPHER LABS"},
		Name:     "CoinSwitch Kuber",
		Category: "Investment",
	},
	"ZEBPAY": {
		Aliases:  []string{"ZEBPAY", "ZEB IT SERVICE", "ZEBITSERVICE"},
		Name:     "ZebPay",
		Category: "Investment",
	},
	"UNOCOIN": {
		Aliases:  []string{"UNOCOIN", "UNOCOMMERCE"},
		Name:     "Unocoin",
		Category: "Investment",
	},
	
	// Crypto Exchanges (International)
	"BINANCE": {
		Aliases:  []string{"BINANCE", "BINANCEPAY", "BIFINANCE"},
		Name:     "Binance",
		Category: "Investment",
	},
	"COINBASE": {
		Aliases:  []string{"COINBASE", "CB PAY", "CBPAY"},
		Name:     "Coinbase",
		Category: "Investment",
	},
	"KRAKEN": {
		Aliases:  []string{"KRAKEN", "PAYWARD"},
		Name:     "Kraken",
		Category: "Investment",
	},
	"CRYPTOCOM": {
		Aliases:  []string{"CRYPTOCOM", "FORIS"},
		Name:     "Crypto.com",
		Category: "Investment",
	},
	"KUCOIN": {
		Aliases:  []string{"KUCOIN", "MEK GLOBAL", "MEKGLOBAL"},
		Name:     "KuCoin",
		Category: "Investment",
	},
	"BITSTAMP": {
		Aliases:  []string{"BITSTAMP"},
		Name:     "Bitstamp",
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
	
	// Check aliases with stricter matching
	// For short aliases (<=4 chars), require exact word match to avoid false positives
	for _, canonical := range MerchantCanonicalMap {
		for _, alias := range canonical.Aliases {
			// For short aliases, require exact match or word boundary
			if len(alias) <= 4 {
				// Exact match only for short aliases
				if upper == alias {
					return canonical.Name, canonical.Category
				}
				// Or match with word boundaries (space or dash)
				if strings.Contains(" "+upper+" ", " "+alias+" ") ||
				   strings.Contains("-"+upper+"-", "-"+alias+"-") {
					return canonical.Name, canonical.Category
				}
			} else {
				// For longer aliases, allow contains matching
				if strings.Contains(upper, alias) {
					return canonical.Name, canonical.Category
				}
			}
		}
	}
	
	// Partial match (fuzzy) - only for longer keys (>= 5 chars)
	for key, canonical := range MerchantCanonicalMap {
		if len(key) >= 5 && strings.Contains(upper, key) {
			return canonical.Name, canonical.Category
		}
	}
	
	return merchant, "" // Return original if no match
}

