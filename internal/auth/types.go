package auth

type User struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password,omitempty"`
}

type SigninCreds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
