package model

type Transaction struct {
	Amount             float64 `json:"amount"`
	CreatedAt          string  `json:"createdAt"`
	SenderAddress      string  `json:"from"`
	DestinationAddress string  `json:"to"`
}
