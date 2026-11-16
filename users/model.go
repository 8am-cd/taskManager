package users

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewUser(email string, password string) *User {
	return &User{
		Email:    email,
		Password: password,
	}
}
