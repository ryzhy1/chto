package dto

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
}
