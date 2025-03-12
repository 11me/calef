package models

import (
	"fmt"
	"time"
)

type BinanceAggTrade struct {
	Symbol      string `json:"s"`
	Price       string `json:"p"`
	Quantity    string `json:"q"`
	EventTimeMs int64  `json:"E"`
}

type Bar struct {
	Symbol string  `json:"symbol"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`

	Volume    float64   `json:"volume"`
	IsClosed  bool      `json:"isClosed"`
	StartTime time.Time `json:"startTime"`
}

type Timeframe time.Duration

const (
	M1  Timeframe = Timeframe(time.Minute)
	M5  Timeframe = Timeframe(5 * time.Minute)
	M10 Timeframe = Timeframe(10 * time.Minute)
	M15 Timeframe = Timeframe(15 * time.Minute)
)

func (tf Timeframe) String() string {
	d := time.Duration(tf)

	// Weeks
	if d >= 7*24*time.Hour {
		weeks := d / (7 * 24 * time.Hour)
		return fmt.Sprintf("%dw", weeks)
	}

	// Days
	if d >= 24*time.Hour {
		days := d / (24 * time.Hour)
		return fmt.Sprintf("%dd", days)
	}

	// Hours
	if d >= time.Hour {
		hours := d / time.Hour
		return fmt.Sprintf("%dh", hours)
	}

	// Minutes
	if d >= time.Minute {
		minutes := d / time.Minute
		return fmt.Sprintf("%dm", minutes)
	}

	// Seconds
	if d >= time.Second {
		seconds := d / time.Second
		return fmt.Sprintf("%ds", seconds)
	}

	return d.String()
}
