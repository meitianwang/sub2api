package service

import "github.com/shopspring/decimal"

var (
	hundred  = decimal.NewFromInt(100)
	zeroRate = decimal.Zero
)

// CalculateFee computes the fee for a recharge order.
// feeRate is a percentage (e.g. 2.5 means 2.5%).
// feeAmount = ceil(amount * feeRate / 100, 2 decimals)
// payAmount = amount + feeAmount
func CalculateFee(amount, feeRate decimal.Decimal) (feeAmount, payAmount decimal.Decimal) {
	if feeRate.LessThanOrEqual(zeroRate) {
		return decimal.Zero, amount
	}
	// fee = amount * rate / 100, rounded up to 2 decimals
	feeAmount = amount.Mul(feeRate).Div(hundred).RoundCeil(2)
	payAmount = amount.Add(feeAmount)
	return feeAmount, payAmount
}
