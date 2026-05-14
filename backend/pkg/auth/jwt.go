package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ============================================
// JWT Authentication Manager
// ============================================
// Handles JWT token generation and validation
// with HttpOnly cookie strategy for XSS protection

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidSignature = errors.New("invalid signature")
)

// JWTManager handles JWT operations
type JWTManager struct {
	secret   []byte
	issuer   string
	audience string
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret, issuer, audience string) *JWTManager {
	return &JWTManager{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
	}
}

// GenerateToken creates a new JWT access token
func (m *JWTManager) GenerateToken(userID, username, role string, duration time.Duration) (string, error) {
	now := time.Now()

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateToken validates and parses a JWT token
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Verify issuer and audience
	if claims.Issuer != m.issuer {
		return nil, ErrInvalidToken
	}

	if len(claims.Audience) == 0 || claims.Audience[0] != m.audience {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken generates a new access token from a refresh token
func (m *JWTManager) RefreshToken(refreshTokenString string, accessDuration time.Duration) (string, error) {
	// Validate refresh token
	claims, err := m.ValidateToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	// Generate new access token with same user info
	return m.GenerateToken(claims.UserID, claims.Username, claims.Role, accessDuration)
}

// ExtractUserID extracts user ID from token without full validation (for logging)
func (m *JWTManager) ExtractUserID(tokenString string) string {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})

	if err != nil || !token.Valid {
		return "unknown"
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "unknown"
	}

	return claims.UserID
}
