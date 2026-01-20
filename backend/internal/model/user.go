package model

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
}

type AuthRequest struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	CaptchaID        string `json:"captchaId"`
	CaptchaCode      string `json:"captchaCode"`
}

type AuthResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	ID       string `json:"id"`
}
