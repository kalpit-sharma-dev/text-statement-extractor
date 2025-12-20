package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// AccountInfo represents the account holder and account details
type AccountInfo struct {
	BankName          string
	AccountHolderName string
	Address           []string // Multiple address lines
	City              string
	State             string
	PhoneNo           string
	Email             string
	ODLimit           string
	Currency          string
	CustID            string
	AccountNo         string
	AccountType       string
	AccountOpenDate   string
	AccountStatus     string
	BranchName        string
	BranchAddress     []string
	BranchCode        string
	IFSC              string
	MICR              string
	JointHolders      string
	Nomination        string
	PrimePotential    string
}

// StatementPeriod represents the statement date range
type StatementPeriod struct {
	FromDate string
	ToDate   string
}

// TxtTransaction represents a single transaction entry from TXT file
type TxtTransaction struct {
	Date           string
	Narration      string // Can be multi-line
	ChequeRefNo    string
	ValueDate      string
	WithdrawalAmt  float64
	DepositAmt     float64
	ClosingBalance float64
}

// StatementSummary represents the summary at the end of the statement
type StatementSummary struct {
	OpeningBalance          float64
	TotalDebits             float64
	TotalCredits            float64
	ClosingBalance          float64
	DebitCount              int
	CreditCount             int
	GeneratedOn             string
	GeneratedBy             string
	RequestingBranchCode    string
	GSTN                    string
	RegisteredOfficeAddress string
}

// TxtAccountStatement represents the complete extracted statement from TXT file
type TxtAccountStatement struct {
	AccountInfo     AccountInfo
	StatementPeriod StatementPeriod
	Transactions    []TxtTransaction
	Summary         StatementSummary
}

// Helper function to parse amount strings (removes commas and converts to float)
func parseAmount(amountStr string) float64 {
	amountStr = strings.TrimSpace(amountStr)
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	if amountStr == "" {
		return 0.0
	}
	val, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0.0
	}
	return val
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Helper function to extract value after colon
func extractAfterColon(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// Helper function to check if line is a transaction line (starts with date pattern DD/MM/YY)
func isTransactionLine(line string) bool {
	matched, _ := regexp.MatchString(`^\d{2}/\d{2}/\d{2}`, strings.TrimSpace(line))
	return matched
}

// Helper function to convert 2-digit year to 4-digit year based on statement period
// Example: "24" -> "2024", "25" -> "2025"
// Uses the statement period to determine the correct century
func convertDateToFullYear(dateStr string, statementPeriod StatementPeriod) string {
	// dateStr format: DD/MM/YY
	if len(dateStr) != 8 {
		return dateStr
	}
	
	parts := strings.Split(dateStr, "/")
	if len(parts) != 3 {
		return dateStr
	}
	
	day := parts[0]
	month := parts[1]
	year := parts[2]
	
	// Convert 2-digit year to 4-digit
	// Determine century from statement period
	// If statement period starts with 2024, use 2000s
	// If year is 00-50, assume 2000-2050
	// If year is 51-99, assume 1951-1999 (for old statements)
	var fullYear string
	yearInt, err := strconv.Atoi(year)
	if err != nil {
		return dateStr
	}
	
	// Extract year from statement period (if available)
	// Format: "01/04/2024"
	var statementStartYear int
	if statementPeriod.FromDate != "" && len(statementPeriod.FromDate) == 10 {
		// Extract year from DD/MM/YYYY format
		periodParts := strings.Split(statementPeriod.FromDate, "/")
		if len(periodParts) == 3 {
			statementStartYear, _ = strconv.Atoi(periodParts[2])
		}
	}
	
	// Determine the century
	if statementStartYear > 0 {
		// Use the statement year's century
		century := (statementStartYear / 100) * 100 // 2024 -> 2000
		fullYear = fmt.Sprintf("%d", century+yearInt)
		
		// Handle year wraparound (e.g., statement from Apr 2024 to Mar 2025)
		// If the resulting year is more than 1 year before statement start, it's next century
		resultYear, _ := strconv.Atoi(fullYear)
		if statementStartYear-resultYear > 50 {
			fullYear = fmt.Sprintf("%d", century+100+yearInt)
		}
	} else {
		// Fallback: use 2000-2099 for years 00-99
		if yearInt <= 99 {
			fullYear = fmt.Sprintf("20%02d", yearInt)
		} else {
			fullYear = year
		}
	}
	
	return fmt.Sprintf("%s/%s/%s", day, month, fullYear)
}

// Helper function to check if line is a continuation of narration (no date, but has content)
func isNarrationContinuation(line string) bool {
	trimmed := strings.TrimSpace(line)
	
	// Empty lines are not continuations
	if trimmed == "" {
		return false
	}
	
	// Skip common header/footer markers
	if strings.HasPrefix(trimmed, "**Continue**") ||
		strings.HasPrefix(trimmed, "--------") ||
		strings.HasPrefix(trimmed, "********") {
		return false
	}
	
	// Skip page headers - account holder information repeated on each page
	if strings.HasPrefix(trimmed, "MR.") ||
		strings.HasPrefix(trimmed, "MRS.") ||
		strings.HasPrefix(trimmed, "MS.") {
		return false
	}
	
	// Skip common header field labels
	headerKeywords := []string{
		"Account Branch :",
		"Address        :",
		"City           :",
		"State          :",
		"Phone no.      :",
		"Email          :",
		"OD Limit       :",
		"Cust ID        :",
		"Account No     :",
		"A/C Open Date  :",
		"Account Status :",
		"JOINT HOLDERS :",
		"Nomination :",
		"Statement From",
		"RTGS/NEFT IFSC :",
		"Branch Code    :",
		"Account Type   :",
		"HDFC BANK",
		"Page No",
	}
	
	for _, keyword := range headerKeywords {
		if strings.Contains(trimmed, keyword) {
			return false
		}
	}
	
	// Skip footer section - statement summary and closing information
	footerKeywords := []string{
		"STATEMENT SUMMARY",
		"Opening Balance",
		"Closing Bal",
		"Debits",
		"Credits",
		"Dr Count",
		"Cr Count",
		"Generated On:",
		"Generated By:",
		"Requesting Branch Code:",
		"This is a computer generated statement",
		"HDFC BANK LIMITED",
		"Closing balance includes funds earmarked",
		"Contents of this statement will be considered correct",
		"State account branch GSTN:",
		"HDFC Bank GSTIN number",
		"Registered Office Address:",
		"End Of Statement",
		"GSTIN",
		"---  End",
	}
	
	for _, keyword := range footerKeywords {
		if strings.Contains(trimmed, keyword) {
			return false
		}
	}
	
	// Skip lines that look like city names or addresses (common in headers)
	// These typically are in ALL CAPS and short
	if len(trimmed) < 50 && strings.ToUpper(trimmed) == trimmed {
		// Check if it looks like an address field (contains common address words)
		addressWords := []string{"GHAZIABAD", "UTTAR PRADESH", "GOVINDPURAM", "PUNE", "MAHARASHTRA", 
			"VIMAN NAGAR", "FLORENCE BUILDING", "HNO", "BLOCK"}
		for _, word := range addressWords {
			if strings.Contains(trimmed, word) {
				return false
			}
		}
	}
	
	// If it doesn't start with a date and has content, and passed all filters above, 
	// it's likely a continuation
	if !isTransactionLine(line) && len(trimmed) > 0 {
		return true
	}
	
	return false
}

// Extract account information from header section
func extractAccountInfo(lines []string) AccountInfo {
	info := AccountInfo{}

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Extract Bank Name
		if strings.Contains(line, "HDFC BANK") && info.BankName == "" {
			info.BankName = "HDFC BANK Ltd."
		}

		// Extract Account Holder Name
		if (strings.HasPrefix(line, "MR.") || strings.HasPrefix(line, "MRS.") || strings.HasPrefix(line, "MS.")) && info.AccountHolderName == "" {
			// Extract just the name part (before any long spaces or special markers)
			// The name is typically on the left side, before the branch info starts
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				// Take first few words as name (MR./MRS./MS. + name parts)
				nameParts := []string{}
				for _, part := range parts {
					if strings.Contains(part, "OFF") || strings.Contains(part, "PUNE") {
						break
					}
					nameParts = append(nameParts, part)
				}
				info.AccountHolderName = strings.Join(nameParts, " ")
			} else {
				info.AccountHolderName = strings.TrimSpace(line)
			}
		}

		// Extract Address (lines after account holder name, before JOINT HOLDERS)
		if info.AccountHolderName != "" && len(info.Address) == 0 {
			// Collect address lines (usually 3-5 lines)
			// Address is on the left side, branch info is on the right
			for j := i + 1; j < i+7 && j < len(lines); j++ {
				originalLine := lines[j]
				// Extract left part (before the branch info which starts around column 80-90)
				// Branch info typically starts with "City", "State", etc.
				addrLine := ""
				if len(originalLine) > 80 {
					// Check if this line has branch info on the right
					rightPart := strings.TrimSpace(originalLine[70:])
					if strings.Contains(rightPart, "City") ||
						strings.Contains(rightPart, "State") ||
						strings.Contains(rightPart, "Phone") ||
						strings.Contains(rightPart, "Email") ||
						strings.Contains(rightPart, "OD Limit") ||
						strings.Contains(rightPart, "Cust ID") ||
						strings.Contains(rightPart, "Account No") {
						// This line has branch info, extract left part as address
						addrLine = strings.TrimSpace(originalLine[0:70])
					} else {
						// No branch info, might be a full address line
						addrLine = strings.TrimSpace(originalLine)
					}
				} else {
					addrLine = strings.TrimSpace(originalLine)
				}

				// Stop if we hit JOINT HOLDERS, Nomination, or Statement From
				if strings.Contains(addrLine, "JOINT HOLDERS") ||
					strings.Contains(addrLine, "Nomination") ||
					strings.Contains(addrLine, "Statement From") ||
					strings.Contains(addrLine, "--------") ||
					strings.Contains(addrLine, "Date      Narration") ||
					strings.Contains(addrLine, "OFF PUNE NAGAR HIGHWAY") {
					break
				}
				// Add non-empty address lines
				if addrLine != "" && !strings.Contains(addrLine, ":") {
					info.Address = append(info.Address, addrLine)
				}
			}
		}

		// Extract fields with colons
		if strings.Contains(line, "Account Branch :") {
			info.BranchName = extractAfterColon(line)
		}
		if strings.Contains(line, "Address        :") {
			// Address can span multiple lines
			addr := extractAfterColon(line)
			if addr != "" {
				info.BranchAddress = append(info.BranchAddress, addr)
			}
			// Check next lines for continuation
			for j := i + 1; j < i+3 && j < len(lines); j++ {
				nextLine := strings.TrimSpace(lines[j])
				if nextLine != "" && !strings.Contains(nextLine, ":") {
					info.BranchAddress = append(info.BranchAddress, nextLine)
				} else {
					break
				}
			}
		}
		if strings.Contains(line, "City           :") {
			info.City = extractAfterColon(line)
		}
		if strings.Contains(line, "State          :") {
			info.State = extractAfterColon(line)
		}
		if strings.Contains(line, "Phone no.      :") {
			info.PhoneNo = extractAfterColon(line)
		}
		if strings.Contains(line, "Email          :") {
			info.Email = extractAfterColon(line)
		}
		if strings.Contains(line, "OD Limit       :") {
			parts := strings.Split(line, "OD Limit       :")
			if len(parts) == 2 {
				rest := strings.TrimSpace(parts[1])
				limitParts := strings.Fields(rest)
				if len(limitParts) >= 2 {
					info.ODLimit = limitParts[0]
					info.Currency = limitParts[1]
				}
			}
		}
		if strings.Contains(line, "Cust ID        :") {
			info.CustID = extractAfterColon(line)
		}
		if strings.Contains(line, "Account No     :") {
			parts := strings.Split(line, "Account No     :")
			if len(parts) == 2 {
				rest := strings.TrimSpace(parts[1])
				fields := strings.Fields(rest)
				if len(fields) >= 1 {
					info.AccountNo = fields[0]
					if len(fields) > 1 {
						info.PrimePotential = strings.Join(fields[1:], " ")
					}
				}
			}
		}
		if strings.Contains(line, "A/C Open Date  :") {
			info.AccountOpenDate = extractAfterColon(line)
		}
		if strings.Contains(line, "Account Status :") {
			info.AccountStatus = extractAfterColon(line)
		}
		if strings.Contains(line, "RTGS/NEFT IFSC :") {
			parts := strings.Split(line, "RTGS/NEFT IFSC :")
			if len(parts) == 2 {
				rest := strings.TrimSpace(parts[1])
				fields := strings.Fields(rest)
				if len(fields) >= 1 {
					info.IFSC = fields[0]
					if len(fields) >= 3 && fields[1] == "MICR" {
						info.MICR = fields[2]
					}
				}
			}
		}
		if strings.Contains(line, "Branch Code    :") {
			info.BranchCode = extractAfterColon(line)
		}
		if strings.Contains(line, "Account Type   :") {
			info.AccountType = extractAfterColon(line)
		}
		if strings.Contains(line, "JOINT HOLDERS :") {
			info.JointHolders = extractAfterColon(line)
		}
		if strings.Contains(line, "Nomination :") {
			parts := strings.Split(line, "Nomination :")
			if len(parts) == 2 {
				info.Nomination = strings.TrimSpace(parts[1])
			}
		}
	}

	return info
}

// Extract statement period
func extractStatementPeriod(lines []string) StatementPeriod {
	period := StatementPeriod{}

	for _, line := range lines {
		if strings.Contains(line, "Statement From") {
			// Format: Statement From      : 01/04/2024  To: 31/03/2025
			re := regexp.MustCompile(`Statement From\s+:\s+(\d{2}/\d{2}/\d{4})\s+To:\s+(\d{2}/\d{2}/\d{4})`)
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				period.FromDate = matches[1]
				period.ToDate = matches[2]
			}
		}
	}

	return period
}

// Parse a transaction line with context (previous balance and statement period)
func parseTransactionLineWithContext(line string, previousBalance float64, statementPeriod StatementPeriod) *TxtTransaction {
	return parseTransactionLine(line, previousBalance, statementPeriod)
}

// Parse a transaction line
func parseTransactionLine(line string, previousBalance float64, statementPeriod StatementPeriod) *TxtTransaction {
	// Use original line (not trimmed) to preserve fixed-width positions
	if !isTransactionLine(line) {
		return nil
	}

	// Extract date (first 8 characters in DD/MM/YY format)
	if len(line) < 8 {
		return nil
	}
	date := strings.TrimSpace(line[0:8])
	if date == "" {
		return nil
	}
	
	// Convert 2-digit year to 4-digit year
	date = convertDateToFullYear(date, statementPeriod)

	// Find the reference number (typically 14-16 digits, but can vary)
	// Reference number is usually after narration, before value date
	// Look for patterns like: 16 digits, or alphanumeric codes
	refRe := regexp.MustCompile(`([A-Z0-9]{14,16})`)
	refMatches := refRe.FindAllString(line, -1)
	chequeRef := ""
	refIndex := -1

	// Find the reference number that appears after the date and before value date
	// Usually around position 60-80
	valueDateRe := regexp.MustCompile(`(\d{2}/\d{2}/\d{2})`)
	valueDateMatches := valueDateRe.FindAllString(line, -1)
	valueDatePos := -1
	if len(valueDateMatches) > 1 {
		valueDatePos = strings.Index(line, valueDateMatches[1])
	}

	// Look for reference number between position 50 and value date position
	for _, match := range refMatches {
		pos := strings.Index(line, match)
		// Reference should be after date (position 8+) and before value date
		if pos > 50 && pos < 100 {
			// Check if it's not part of narration (should be standalone)
			// Reference numbers are usually right-aligned or have spaces around them
			if valueDatePos == -1 || pos < valueDatePos {
				chequeRef = match
				refIndex = pos
				break
			}
		}
	}

	// Fallback: if no reference found, try to find 16-digit number
	if chequeRef == "" {
		refRe16 := regexp.MustCompile(`(\d{16})`)
		refMatches16 := refRe16.FindAllString(line, -1)
		if len(refMatches16) > 0 {
			for _, match := range refMatches16 {
				pos := strings.Index(line, match)
				if pos > 50 && pos < 100 {
					chequeRef = match
					refIndex = pos
					break
				}
			}
		}
	}

	// Find value date (DD/MM/YY format) - should be after reference number
	// Reuse valueDateMatches already found above
	valueDate := ""
	if len(valueDateMatches) > 1 {
		// Second date is value date (first is transaction date)
		valueDate = valueDateMatches[1]
		// Convert 2-digit year to 4-digit year
		valueDate = convertDateToFullYear(valueDate, statementPeriod)
	}

	// Find all amounts (numbers with commas and decimals)
	// But exclude amounts that are clearly in the narration (before position 85)
	// Amounts should be in the transaction columns (position 85+)
	amountRe := regexp.MustCompile(`([\d,]+\.\d{2})`)
	allAmountMatches := amountRe.FindAllString(line, -1)

	// Filter amounts to only include those in the transaction amount columns (position 85+)
	// This excludes amounts that appear in narration text
	amountMatches := make([]string, 0)
	for _, match := range allAmountMatches {
		pos := strings.Index(line, match)
		// Only include amounts that are in the transaction columns (after position 85)
		// This is where withdrawal/deposit/balance columns are located
		if pos >= 85 {
			amountMatches = append(amountMatches, match)
		}
	}

	// Extract narration (between date and reference number)
	narration := ""
	if refIndex > 0 {
		// Narration is between position 10 (after date) and reference number
		if refIndex > 10 {
			narration = strings.TrimSpace(line[10:refIndex])
			// Clean up narration - remove any trailing reference numbers or dates that might have been included
			// Remove any 16-digit numbers or date patterns at the end
			narration = regexp.MustCompile(`\s+\d{16}\s*$`).ReplaceAllString(narration, "")
			narration = regexp.MustCompile(`\s+\d{2}/\d{2}/\d{2}\s*$`).ReplaceAllString(narration, "")
			narration = strings.TrimSpace(narration)
		}
	} else if len(amountMatches) > 0 {
		// Fallback: narration is between date and first amount
		firstAmountIndex := strings.Index(line, amountMatches[0])
		if firstAmountIndex > 10 {
			narration = strings.TrimSpace(line[10:firstAmountIndex])
			// Clean up narration
			narration = regexp.MustCompile(`\s+\d{16}\s*$`).ReplaceAllString(narration, "")
			narration = regexp.MustCompile(`\s+\d{2}/\d{2}/\d{2}\s*$`).ReplaceAllString(narration, "")
			narration = strings.TrimSpace(narration)
		}
	}

	// Parse amounts - typically we have 1-3 amounts
	// Pattern: [withdrawal] [deposit] balance (balance is always last)
	withdrawal := 0.0
	deposit := 0.0
	balance := 0.0

	if len(amountMatches) == 0 {
		return nil // No amounts found, invalid transaction
	}

	// The last amount is always the closing balance
	balance = parseAmount(amountMatches[len(amountMatches)-1])

	// Determine withdrawal and deposit based on positions
	// In the fixed-width format:
	// - Withdrawal column is around position 90-110 (left-aligned, less spacing)
	// - Deposit column is around position 110-130 (right-aligned, more spacing before)
	// - Balance column is around position 130-150

	if len(amountMatches) == 3 {
		// Three amounts: withdrawal, deposit, balance
		withdrawal = parseAmount(amountMatches[0])
		deposit = parseAmount(amountMatches[1])

		// Validate: if both withdrawal and deposit are set, they should be reasonable
		// If one is extremely large (like millions) and doesn't match balance change, it's likely wrong
		if previousBalance > 0 {
			expectedBalanceChange := deposit - withdrawal
			actualBalanceChange := balance - previousBalance
			// If the difference is huge (more than 1M), likely one amount is wrong
			if abs(expectedBalanceChange-actualBalanceChange) > 1000000 {
				// One of the amounts is likely wrong - use balance change to determine
				if actualBalanceChange > 0 {
					// Balance increased, so it's a deposit
					deposit = actualBalanceChange
					withdrawal = 0
				} else {
					// Balance decreased, so it's a withdrawal
					withdrawal = -actualBalanceChange
					deposit = 0
				}
			}
		}
	} else if len(amountMatches) == 2 {
		// Two amounts: either withdrawal+balance or deposit+balance
		firstAmountPos := strings.Index(line, amountMatches[0])
		secondAmountPos := strings.Index(line, amountMatches[1])
		firstAmount := parseAmount(amountMatches[0])

		// Use balance change to determine if it's deposit or withdrawal
		// If we have previous balance, use it to verify
		balanceChange := balance - previousBalance

		// Check the spacing pattern
		// Deposits have more spacing before them (they're in the deposit column which is right-aligned)
		// Withdrawals have less spacing (they're in the withdrawal column which is left-aligned)

		// Calculate spacing before first amount
		spacingBeforeFirst := 0
		if firstAmountPos > 0 {
			// Count spaces before the amount
			for i := firstAmountPos - 1; i >= 0 && line[i] == ' '; i-- {
				spacingBeforeFirst++
			}
		}

		// Primary method: Use balance change if we have previous balance
		if previousBalance > 0 {
			if balanceChange > 0 {
				// Balance increased - this is a deposit
				deposit = firstAmount
			} else if balanceChange < 0 {
				// Balance decreased - this is a withdrawal
				withdrawal = firstAmount
			} else {
				// Balance unchanged - use position-based logic
				if spacingBeforeFirst > 15 || firstAmountPos > 110 {
					deposit = firstAmount
				} else {
					withdrawal = firstAmount
				}
			}
		} else {
			// No previous balance - use position and spacing
			// Deposits typically have more spacing (20+ spaces) and are positioned after column 100
			// Withdrawals typically have less spacing (<20 spaces) and are positioned before column 110

			// Also check narration for deposit indicators
			narrationUpper := strings.ToUpper(narration)
			isLikelyDeposit := strings.Contains(narrationUpper, "SALARY") ||
				strings.Contains(narrationUpper, "SAL FOR") ||
				strings.Contains(narrationUpper, "CR-") ||
				strings.Contains(narrationUpper, "CREDIT") ||
				strings.Contains(narrationUpper, "IMPS") && strings.Contains(narrationUpper, "MR") ||
				strings.Contains(narrationUpper, "NEFT CR") ||
				strings.Contains(narrationUpper, "RTGS CR")

			if isLikelyDeposit {
				// Narration suggests deposit
				deposit = firstAmount
			} else if spacingBeforeFirst > 20 || firstAmountPos > 110 {
				// Large spacing or position suggests deposit column
				deposit = firstAmount
			} else if firstAmountPos >= 85 && firstAmountPos < 110 && spacingBeforeFirst < 20 {
				// In withdrawal column range with less spacing
				withdrawal = firstAmount
			} else {
				// Fallback: check gap between amounts
				gapBetweenAmounts := secondAmountPos - firstAmountPos
				if gapBetweenAmounts > 30 {
					// Very large gap suggests withdrawal column then balance
					withdrawal = firstAmount
				} else if spacingBeforeFirst > 15 {
					// More spacing suggests deposit
					deposit = firstAmount
				} else {
					// Default: check if amount matches typical withdrawal patterns
					// If narration suggests withdrawal, use withdrawal
					isLikelyWithdrawal := strings.Contains(narrationUpper, "DR-") ||
						strings.Contains(narrationUpper, "DEBIT") ||
						strings.Contains(narrationUpper, "UPI-") && firstAmountPos < 100

					if isLikelyWithdrawal {
						withdrawal = firstAmount
					} else if spacingBeforeFirst > 10 {
						// More spacing suggests deposit
						deposit = firstAmount
					} else {
						// Default to withdrawal for safety (most transactions are withdrawals)
						withdrawal = firstAmount
					}
				}
			}
		}
	} else if len(amountMatches) == 1 {
		// Only balance - no withdrawal or deposit (unlikely but possible)
		balance = parseAmount(amountMatches[0])
	}

	return &TxtTransaction{
		Date:           date,
		Narration:      narration,
		ChequeRefNo:    chequeRef,
		ValueDate:      valueDate,
		WithdrawalAmt:  withdrawal,
		DepositAmt:     deposit,
		ClosingBalance: balance,
	}
}

// Extract transactions from the file
func extractTransactions(lines []string, statementPeriod StatementPeriod) []TxtTransaction {
	return extractTransactionsWithOpeningBalance(lines, 0.0, statementPeriod)
}

// Extract transactions from the file with opening balance
func extractTransactionsWithOpeningBalance(lines []string, openingBalance float64, statementPeriod StatementPeriod) []TxtTransaction {
	var transactions []TxtTransaction
	var currentTxn *TxtTransaction
	var previousBalance float64 = openingBalance // Track previous balance to determine deposit vs withdrawal

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip header lines and separators
		if strings.HasPrefix(trimmed, "--------") ||
			strings.HasPrefix(trimmed, "Date      Narration") ||
			trimmed == "" ||
			strings.Contains(trimmed, "**Continue**") ||
			strings.Contains(trimmed, "Page No") {
			continue
		}

		// Check if we've reached the statement summary section (end of transactions)
		if strings.HasPrefix(trimmed, "********") ||
			strings.Contains(trimmed, "STATEMENT SUMMARY") ||
			strings.Contains(trimmed, "Opening Balance") && strings.Contains(trimmed, "Debits") && strings.Contains(trimmed, "Credits") {
			// Reached summary section - save last transaction and stop
			if currentTxn != nil {
				transactions = append(transactions, *currentTxn)
				currentTxn = nil
			}
			break
		}

		// Check if this is a transaction line
		if isTransactionLine(trimmed) {
			// Save previous transaction if exists
			if currentTxn != nil {
				previousBalance = currentTxn.ClosingBalance
				transactions = append(transactions, *currentTxn)
			}

			// Parse new transaction with previous balance context and statement period
			currentTxn = parseTransactionLineWithContext(trimmed, previousBalance, statementPeriod)
		} else if currentTxn != nil && isNarrationContinuation(trimmed) {
			// This is a continuation of the narration (filtered by isNarrationContinuation)
			currentTxn.Narration += " " + trimmed
		}
	}

	// Don't forget the last transaction
	if currentTxn != nil {
		transactions = append(transactions, *currentTxn)
	}

	return transactions
}

// Extract statement summary
func extractSummary(lines []string) StatementSummary {
	summary := StatementSummary{}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Extract summary values
		if strings.Contains(trimmed, "Opening Balance") && i+1 < len(lines) {
			// Next line has the values
			nextLine := strings.TrimSpace(lines[i+1])
			// Format: 379,562.39    6,770,007.52    6,431,384.97    40,939.84
			amountRe := regexp.MustCompile(`([\d,]+\.\d{2})`)
			amounts := amountRe.FindAllString(nextLine, -1)
			if len(amounts) >= 4 {
				summary.OpeningBalance = parseAmount(amounts[0])
				summary.TotalDebits = parseAmount(amounts[1])
				summary.TotalCredits = parseAmount(amounts[2])
				summary.ClosingBalance = parseAmount(amounts[3])
			}
		}

		if strings.Contains(trimmed, "Dr Count") && i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			countRe := regexp.MustCompile(`(\d+)`)
			counts := countRe.FindAllString(nextLine, -1)
			if len(counts) >= 2 {
				summary.DebitCount, _ = strconv.Atoi(counts[0])
				summary.CreditCount, _ = strconv.Atoi(counts[1])
			}
		}

		if strings.Contains(trimmed, "Generated On:") {
			// Format: Generated On: 17-DEC-2025 10:11:33
			re := regexp.MustCompile(`Generated On:\s+([\d\-A-Z\s:]+?)(?:\s+Generated By|$)`)
			matches := re.FindStringSubmatch(trimmed)
			if len(matches) >= 2 {
				summary.GeneratedOn = strings.TrimSpace(matches[1])
			}

			re = regexp.MustCompile(`Generated By:\s+(\S+)`)
			matches = re.FindStringSubmatch(trimmed)
			if len(matches) >= 2 {
				summary.GeneratedBy = matches[1]
			}

			re = regexp.MustCompile(`Requesting Branch Code:\s+(\S+)`)
			matches = re.FindStringSubmatch(trimmed)
			if len(matches) >= 2 {
				summary.RequestingBranchCode = matches[1]
			}
		}

		if strings.Contains(trimmed, "GSTN:") {
			re := regexp.MustCompile(`GSTN:(\S+)`)
			matches := re.FindStringSubmatch(trimmed)
			if len(matches) >= 2 {
				summary.GSTN = matches[1]
			}
		}

		if strings.Contains(trimmed, "Registered Office Address:") {
			parts := strings.SplitN(trimmed, "Registered Office Address:", 2)
			if len(parts) == 2 {
				summary.RegisteredOfficeAddress = strings.TrimSpace(parts[1])
			}
		}
	}

	return summary
}

// ReadAccountStatementFromTxt reads and parses the account statement from a text file
func ReadAccountStatementFromTxt(filePath string) (*TxtAccountStatement, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Process the lines using the shared function
	statement := processStatementLines(lines)

	return statement, nil
}

// processStatementLines processes the statement lines and returns the parsed statement
func processStatementLines(lines []string) *TxtAccountStatement {
	// Extract account info from first page (usually first 25 lines)
	headerLines := lines
	if len(lines) > 25 {
		headerLines = lines[0:25]
	}

	accountInfo := extractAccountInfo(headerLines)
	statementPeriod := extractStatementPeriod(headerLines)

	// Extract summary first to get opening balance
	summary := extractSummary(lines)

	// Extract transactions with opening balance context and statement period for date conversion
	transactions := extractTransactionsWithOpeningBalance(lines, summary.OpeningBalance, statementPeriod)

	statement := &TxtAccountStatement{
		AccountInfo:     accountInfo,
		StatementPeriod: statementPeriod,
		Transactions:    transactions,
		Summary:         summary,
	}

	return statement
}

// ReadAccountStatementFromBase64 reads and parses the account statement from a base64 encoded string
func ReadAccountStatementFromBase64(base64String string) (*TxtAccountStatement, error) {
	// Decode base64 string
	decodedBytes, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 string: %w", err)
	}

	// Convert decoded bytes to string
	decodedText := string(decodedBytes)

	// Split into lines
	lines := strings.Split(decodedText, "\n")

	// Process the lines
	statement := processStatementLines(lines)

	return statement, nil
}
