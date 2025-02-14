package model

type InfoResponse struct {
	Balance     int             `json:"coins"`
	Inventory   []InfoInventory `json:"inventory"`
	CoinHistory []CoinHistory   `json:"coinHistory"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Errors string `json:"errors"`
}

type InfoInventory struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []ReceivedCoin `json:"received"`
	Sent     []SentCoin     `json:"sent"`
}

type ReceivedCoin struct {
	FromUser string `json:"fromUser"`
	Amount   int    `json:"amount"`
}

type SentCoin struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}
