package models

//easyjson:json
type Profile struct {
	User
	Nickname string  `json:"nickname" example:"Nick"`
	Avatar   *string `json:"avatar,omitempty"`
	Stats
}

//easyjson:json
type RegisterProfile struct {
	Nickname string `json:"nickname" example:"Nick"`
	UserPassword
}

//easyjson:json
type User struct {
	UserID uint `json:"id" db:"user_id"`
	UserPassword
}

//easyjson:json
type UserPassword struct {
	Email    string `json:"email,omitempty" example:"email@email.com" valid:"required~Почта не может быть пустой,email~Невалидная почта"`
	Password string `json:"password,omitempty" example:"password" valid:"stringlength(4|32)~Пароль должен быть не менее 4 символов и не более 32 символов"`
}

type Stats struct {
	Record int `json:"record"`
	Win    int `json:"win"`
	Draws  int `json:"draws"`
	Loss   int `json:"loss"`
}

type ProfileError struct {
	Field string `json:"field" example:"nickname"`
	Text  string `json:"text" example:"Этот никнейм уже занят"`
}

//easyjson:json
type ProfileErrorList struct {
	Errors []ProfileError `json:"error"`
}
