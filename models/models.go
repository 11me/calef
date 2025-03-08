package models

import "time"

const (
	BinanceTicksSubject = "binancef.ticks"
)

type TradeTick struct {
	Symbol      string `json:"s"`
	Price       string `json:"p"`
	Quantity    string `json:"q"`
	EventTimeMs int64  `json:"E"`
}

type Bar struct {
	High  float64 `json:"h"`
	Low   float64 `json:"l"`
	Open  float64 `json:"o"`
	Close float64 `json:"c"`

	Volume   float64   `json:"v"`
	IsClosed bool      `json:"isClosed"`
	Time     time.Time `json:"time"`
}
