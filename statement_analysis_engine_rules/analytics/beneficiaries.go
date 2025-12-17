package analytics

import (
	"sort"
	"statement_analysis_engine_rules/models"
)

// CalculateTopBeneficiaries calculates top beneficiaries
func CalculateTopBeneficiaries(transactions []models.ClassifiedTransaction, limit int) []models.TopBeneficiary {
	beneficiaryMap := make(map[string]map[string]float64) // beneficiary -> method -> amount

	for _, txn := range transactions {
		if txn.Beneficiary == "" || txn.IsIncome {
			continue
		}

		if beneficiaryMap[txn.Beneficiary] == nil {
			beneficiaryMap[txn.Beneficiary] = make(map[string]float64)
		}
		beneficiaryMap[txn.Beneficiary][txn.Method] += txn.WithdrawalAmt
	}

	// Convert to slice and sort
	type beneficiaryData struct {
		name   string
		method string
		amount float64
	}

	beneficiaries := make([]beneficiaryData, 0)
	for name, methods := range beneficiaryMap {
		totalAmount := 0.0
		primaryMethod := ""
		for method, amount := range methods {
			totalAmount += amount
			if amount > 0 {
				primaryMethod = method
			}
		}
		beneficiaries = append(beneficiaries, beneficiaryData{
			name:   name,
			method: primaryMethod,
			amount: totalAmount,
		})
	}

	// Sort by amount descending
	sort.Slice(beneficiaries, func(i, j int) bool {
		return beneficiaries[i].amount > beneficiaries[j].amount
	})

	// Take top N
	if limit > len(beneficiaries) {
		limit = len(beneficiaries)
	}

	result := make([]models.TopBeneficiary, limit)
	for i := 0; i < limit; i++ {
		result[i] = models.TopBeneficiary{
			Name:   beneficiaries[i].name,
			Amount: beneficiaries[i].amount,
			Type:   beneficiaries[i].method,
		}
	}

	return result
}
