package dto

type SignUpRequestDto struct {
	Email    	string `json:"email" validate:"required,email"`
	Password 	string `json:"password" validate:"required,min=8"`
	Name		string `json:"name" validate:"required"`
	LastName	string `json:"lastname" validate:"required"`
	Role		string `json:"role" validate:"required,oneof=manager waiter"`
}

type SignInRequestDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type SignInResponseDto struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}
