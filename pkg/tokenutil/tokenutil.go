package tokenutil

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	publicKey *rsa.PublicKey
}

type Config struct {
	PublicKeyPath string `yaml:"public_key"`
}

type JWTCustomClaims struct {
	UserID int
	jwt.RegisteredClaims
}

func New(cfg *Config) (*TokenManager, error) {
	keyData, err := os.ReadFile(cfg.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	// Парсим публичный ключ (PKIX формат)
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	// Приводим к типу *rsa.PublicKey
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return &TokenManager{publicKey: rsaPub}, nil
}

func (tm *TokenManager) ParseAccessToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTCustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected singing method")
		}
		return tm.publicKey, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*JWTCustomClaims); ok && token.Valid {
		return claims.UserID, nil
	}
	return 0, fmt.Errorf("access token not valid")
}
