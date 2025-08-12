package models

type User struct {
	Login           string
	Password        string
	CurrentPoints   float32
	WithdrawnPoints float32
}

func (user User) GetID() string {
	return user.Login
}

func NewUser(login string, password string) *User {
	return &User{
		Login:           login,
		Password:        password,
		CurrentPoints:   0,
		WithdrawnPoints: 0,
	}
}
