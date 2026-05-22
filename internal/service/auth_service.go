package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService struct {
	secretKey []byte
}

type Claims struct {
	UserID uuid.UUID `json:"id"`
	jwt.RegisteredClaims
}

func NewAuthService(secretKey string) *AuthService {
	return &AuthService{
		secretKey: []byte(secretKey),
	}
}

func (s *AuthService) GenerateToken(userID uuid.UUID) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *AuthService) GetUserIDFromToken(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return s.secretKey, nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, err
	}

	return claims.UserID, nil
}
