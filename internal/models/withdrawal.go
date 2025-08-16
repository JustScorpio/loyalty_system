package models

import "time"

type Withdrawal struct {
	UserID      string
	Order       string
	Sum         float32
	ProcessedAt time.Time
}

func (order Withdrawal) GetID() string {
	return order.Order
}

func NewWithdrawal(userID string, order string, sum float32) *Withdrawal {
	return &Withdrawal{
		UserID:      userID,
		Order:       order,
		Sum:         sum,
		ProcessedAt: time.Now(),
	}
}
