package analytics

import (
	"classify/statement_analysis_engine_rules/models"
	"sort"
)

// CalculateTopBeneficiaries calculates top beneficiaries
func CalculateTopBeneficiaries(transactions []models.ClassifiedTransaction, limit int) []models.TopBeneficiary {
	beneficiaryMap := make(map[string]map[string]float64) // beneficiary -> method -> amount

	for _, txn := range transactions {
		// Only count withdrawals (expenses) with beneficiaries
		// Skip if no beneficiary, or if it's a deposit (income), or if no withdrawal amount
		if txn.Beneficiary == "" || txn.DepositAmt > 0 || txn.WithdrawalAmt == 0 {
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
		maxMethodAmount := 0.0
		// Find the method with the highest amount (primary method)
		for method, amount := range methods {
			totalAmount += amount
			if amount > maxMethodAmount {
				maxMethodAmount = amount
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
	if limit <= 0 {
		return []models.TopBeneficiary{} // Return empty if invalid limit
	}
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
