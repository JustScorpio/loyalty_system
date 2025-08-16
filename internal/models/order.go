package models

import (
	"time"
)

type Order struct {
	UserID     string
	Number     string
	Accrual    float32
	Status     Status
	UploadedAt time.Time
}

type Status string

const (
	StatusNew        Status = "NEW"
	StatusProcessing Status = "PROCESSING"
	StatusInvalid    Status = "INVALID"
	StatusProcessed  Status = "PROCESSED"
)

func (order Order) GetID() string {
	return order.Number
}

func NewOrder(userID string, number string) *Order {
	return &Order{
		UserID:     userID,
		Number:     number,
		Accrual:    0,
		Status:     StatusNew,
		UploadedAt: time.Now(),
	}
}
