package auth

type Manager interface {
	Generate(userID int64, login string) (string, error)
	Validate(tokenString string) (*UserInfo, error)
}

type UserInfo struct {
	UserID int64
	Login  string
}
