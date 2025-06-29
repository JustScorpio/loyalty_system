package models

type Order struct {
	ID            string
	UserID        string
	Number        string
	LoyaltyPoints int32
}

func (order Order) GetID() string {
	return order.ID
}
