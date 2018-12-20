package models

//easyjson:json
type Profile struct {
	User
	Nickname string  `json:"nickname" example:"Nick"`
	Avatar   *string `json:"avatar,omitempty"`
	Stats
	Store
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
	Email    string `json:"email,omitempty" example:"email@email.com" valid:"required~Email can not be empty,email~Invalid email"`
	Password string `json:"password,omitempty" example:"password" valid:"stringlength(4|32)~Password must be at least 4 characters and no more than 32 characters"`
}

//easyjson:json
type Stats struct {
	Record int `json:"record"`
	Win    int `json:"win"`
	Draws  int `json:"draws"`
	Loss   int `json:"loss"`
}

//easyjson:json
type ProfileError struct {
	Field string `json:"field" example:"nickname"`
	Text  string `json:"text" example:"This nickname is already taken."`
}

//easyjson:json
type ProfileErrorList struct {
	Errors []ProfileError `json:"error"`
}
