// Package financeseed holds the one-time import of the SAT 1111 sheet
// (Kassa ledger, CBU rates, contracts and balances) into the finance tables.
package financeseed

import _ "embed"

//go:embed seed.sql
var SQL string
