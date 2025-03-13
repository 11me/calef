package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

type Portfolio struct {
	ID        string    `json:"id"`
	Symbols   []string  `json:"symbols"`
	Formula   string    `json:"formula"`
	Timeframe Timeframe `json:"timeframe"`
}

type Timeframe time.Duration

const (
	M1  Timeframe = Timeframe(time.Minute)
	M5  Timeframe = Timeframe(5 * time.Minute)
	M10 Timeframe = Timeframe(10 * time.Minute)
	M15 Timeframe = Timeframe(15 * time.Minute)
)

// UnmarshalJSON implements the json.Unmarshaler interface for Timeframe.
// It accepts strings like "1m", "5m", "1h", "1d", "1w".
func (tf *Timeframe) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("empty timeframe string")
	}

	// Support days ("d") and weeks ("w") which time.ParseDuration doesn't handle.
	if strings.HasSuffix(s, "d") {
		numStr := strings.TrimSuffix(s, "d")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return fmt.Errorf("invalid day duration %q: %w", s, err)
		}
		*tf = Timeframe(time.Duration(num) * 24 * time.Hour)
		return nil
	}

	if strings.HasSuffix(s, "w") {
		numStr := strings.TrimSuffix(s, "w")

		num, err := strconv.Atoi(numStr)
		if err != nil {
			return fmt.Errorf("invalid week duration %q: %w", s, err)
		}

		*tf = Timeframe(time.Duration(num) * 7 * 24 * time.Hour)

		return nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}

	*tf = Timeframe(d)

	return nil
}

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
