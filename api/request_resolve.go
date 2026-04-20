package api

// Request field resolvers: prefer explicit *Sat / *Msat JSON fields when set,
// otherwise fall back to deprecated names (issue #2223).

func (r *PayInvoiceRequest) ResolvedAmountMsat() *uint64 {
	if r == nil {
		return nil
	}
	if r.AmountMsat != nil {
		return r.AmountMsat
	}
	return r.Amount
}

func (r *MakeInvoiceRequest) ResolvedAmountMsat() uint64 {
	if r == nil {
		return 0
	}
	if r.AmountMsat != nil {
		return *r.AmountMsat
	}
	return r.Amount
}

func (r *SendSpontaneousPaymentProbesRequest) ResolvedAmountMsat() uint64 {
	if r == nil {
		return 0
	}
	if r.AmountMsat != 0 {
		return r.AmountMsat
	}
	return r.Amount
}

func (r *LSPOrderRequest) ResolvedAmountSat() uint64 {
	if r == nil {
		return 0
	}
	if r.AmountSat != 0 {
		return r.AmountSat
	}
	return r.Amount
}

func (r *RedeemOnchainFundsRequest) ResolvedAmountSat() uint64 {
	if r == nil {
		return 0
	}
	if r.AmountSat != 0 {
		return r.AmountSat
	}
	return r.Amount
}

func (r *RedeemOnchainFundsRequest) ResolvedFeeRateSatPerVbyte() *uint64 {
	if r == nil {
		return nil
	}
	if r.FeeRateSatPerVbyte != nil {
		return r.FeeRateSatPerVbyte
	}
	return r.FeeRate
}

func (r *InitiateSwapRequest) ResolvedSwapAmountSat() uint64 {
	if r == nil {
		return 0
	}
	if r.SwapAmountSat != 0 {
		return r.SwapAmountSat
	}
	return r.SwapAmount
}

func (r *EnableAutoSwapRequest) ResolvedBalanceThresholdSat() uint64 {
	if r == nil {
		return 0
	}
	if r.BalanceThresholdSat != 0 {
		return r.BalanceThresholdSat
	}
	return r.BalanceThreshold
}

func (r *EnableAutoSwapRequest) ResolvedSwapAmountSat() uint64 {
	if r == nil {
		return 0
	}
	if r.SwapAmountSat != 0 {
		return r.SwapAmountSat
	}
	return r.SwapAmount
}
