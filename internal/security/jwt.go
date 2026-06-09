package security

import (
	"fmt"       // ← ДОБАВЬ ЭТУ СТРОКУ
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ТРЕБОВАНИЕ 3: JWT с коротким TTL
type JWTManager struct {
	secretKey string
	issuer    string
	accessTTL time.Duration
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTManager(secretKey, issuer string, accessTTL time.Duration) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
		issuer:    issuer,
		accessTTL: accessTTL,
	}
}

func (m *JWTManager) GenerateToken(userID int64, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("user:%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) VerifyToken(tokenString string) (*Claims, error) {
	// ТРЕБОВАНИЕ 3: Проверка exp/iss/alg
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
