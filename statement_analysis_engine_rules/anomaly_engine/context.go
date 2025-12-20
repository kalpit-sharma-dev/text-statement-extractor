package anomaly_engine

import (
	"time"

	"classify/statement_analysis_engine_rules/anomaly_engine/types"
	"classify/statement_analysis_engine_rules/models"
)

// TransactionContext is an alias for types.TransactionContext
type TransactionContext = types.TransactionContext

// NewTransactionContext creates a new transaction context
func NewTransactionContext(txn models.ClassifiedTransaction, userID string) TransactionContext {
	return types.TransactionContext{
		Txn:       txn,
		UserID:    userID,
		Timestamp: parseTransactionTimestamp(txn.Date),
		Location:  "",
		DeviceID:  "",
	}
}

func parseTransactionTimestamp(dateStr string) time.Time {
	layouts := []string{
		"02/01/2006",
		"2006-01-02",
		"02-01-2006",
		"01/02/2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t
		}
	}

	return time.Now()
}

