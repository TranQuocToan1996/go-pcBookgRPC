package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

type UserClaims struct {
	jwt.StandardClaims
	Username string
	Role     string
}

func (claims *UserClaims) IsValid() bool {
	return claims.Valid() == nil && len(claims.Role) != 0 && len(claims.Username) != 0
}

func NewJWTManager(privKey string, expire time.Duration) *JWTManager {
	return &JWTManager{privKey, expire}
}

func (j *JWTManager) Generate(user *User) (string, error) {
	claims := UserClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(j.tokenDuration).Unix(),
		},
		user.UserName,
		user.Role,
	}

	// TODO: change to RSA or Elliptic-Curve based Digital signature
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return tokenObj.SignedString([]byte(j.secretKey))
}

func (j *JWTManager) Verify(token string) (*UserClaims, error) {
	claims := &UserClaims{}
	_, err := jwt.ParseWithClaims(token,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			_, ok := t.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("got wrong algo, actually got %v", t.Method.Alg())
			}

			return []byte(j.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token with err: %w", err)
	}

	if !claims.IsValid() {
		return nil, fmt.Errorf("invalid info of token %v", claims)
	}

	return claims, nil
}
