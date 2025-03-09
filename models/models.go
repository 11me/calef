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

type Bar struct {
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Open  float64 `json:"open"`
	Close float64 `json:"close"`

	Volume    float64   `json:"volume"`
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
