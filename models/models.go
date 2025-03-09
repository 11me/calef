package models

import (
	"time"
)

const (
	BinanceTicksSubject = "binancef.ticks"
)

type BinanceAggTrade struct {
	Symbol      string `json:"s"`
	Price       string `json:"p"`
	Quantity    string `json:"q"`
	EventTimeMs int64  `json:"E"`
}

type Candle struct {
	High  float64 `json:"h"`
	Low   float64 `json:"l"`
	Open  float64 `json:"o"`
	Close float64 `json:"c"`

	Volume    float64   `json:"v"`
	IsClosed  bool      `json:"isClosed"`
	StartTime time.Time `json:"startTime"`
}

// Websockets.

type WsCommand struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
}

type Method string

const (
	SubscribeBars     Method = "subscribeBars"
	SubscribeSynthBar Method = "subscribeSynthBar"
)

func (m Method) IsValid() bool {
	switch m {
	case
		SubscribeBars,
		SubscribeSynthBar:

		return true
	}

	return false
}
