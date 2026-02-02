package models

import (
	"time"
)

type LotStatus string

const (
	LotStatusActive   LotStatus = "active"
	LotStatusClosed   LotStatus = "closed"
	LotStatusCanceled LotStatus = "canceled"
)

type Lot struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	StartPrice   float64   `json:"start_price"`
	MinStep      float64   `json:"min_step"`
	CurrentPrice float64   `json:"current_price"`
	Status       LotStatus `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	ClosingAt    time.Time `json:"closing_at"`
	Version      int       `json:"version"`
	WinnerID     *string   `json:"winner_id,omitempty"`
}

type Bid struct {
	ID        string    `json:"id"`
	LotID     string    `json:"lot_id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}
