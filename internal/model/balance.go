package model

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type ResponseWithdraw struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
