package dto


type SignInRequestDto struct {
	Email		string	`json:"email" validate:"required,email"`
	Password	string	`json:"password" validate:"required,min=8"`
}

type SignInResponseDto struct {
	Token			string	`json:"token"`
	RefreshToken	string	`json:"refreshToken"`
}