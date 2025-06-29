package models

type User struct {
	ID            string
	Login         string
	Password      string
	LoyaltyPoints int32
}

func (user User) GetID() string {
	return user.ID
}
