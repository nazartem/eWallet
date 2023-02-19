package model

type Wallet struct {
	Address string  `json:"address"`
	Balance float64 `json:"amount"`
}
