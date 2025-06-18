package auth

import (
	"errors"
	"time"

	"foodcourt-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID   uint             `json:"user_id"`
	Username string           `json:"username"`
	Role     models.UserRole  `json:"role"`
	KiosID   *uint            `json:"kios_id,omitempty"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secretKey string
	expiresIn time.Duration
}

func NewJWTService(secretKey string, expiresIn string) (*JWTService, error) {
	duration, err := time.ParseDuration(expiresIn)
	if err != nil {
		return nil, err
	}

	return &JWTService{
		secretKey: secretKey,
		expiresIn: duration,
	}, nil
}

func (j *JWTService) GenerateToken(user *models.User) (string, error) {
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		KiosID:   user.KiosID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "foodcourt-backend",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (j *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Create new token with updated expiration
	newClaims := &JWTClaims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
		KiosID:   claims.KiosID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "foodcourt-backend",
			Subject:   claims.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return token.SignedString([]byte(j.secretKey))
}
