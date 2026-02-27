package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTValidator validates JWT tokens from Zitadel
type JWTValidator struct {
	jwksProvider *JWKSProvider
	issuer       string
	audience     string
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(issuer, audience, jwksURL string, jwksCacheTTL time.Duration) *JWTValidator {
	return &JWTValidator{
		jwksProvider: NewJWKSProvider(jwksURL, jwksCacheTTL),
		issuer:       issuer,
		audience:     audience,
	}
}

// ValidateToken validates a JWT token and returns the parsed claims
func (v *JWTValidator) ValidateToken(ctx context.Context, tokenString string) (*ZitadelClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&ZitadelClaims{},
		v.jwksProvider.KeyFunc(ctx),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*ZitadelClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ExtractAuthContext validates a JWT token and extracts the authentication context
func (v *JWTValidator) ExtractAuthContext(ctx context.Context, tokenString string) (*AuthContext, error) {
	claims, err := v.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	roles := ParseRoles(claims.Roles)

	authCtx := &AuthContext{
		UserID:            claims.Subject,
		Email:             claims.Email,
		ResourceOwnerID:   claims.ResourceOwnerID,
		ResourceOwnerName: claims.ResourceOwnerName,
		Roles:             roles,
	}

	return authCtx, nil
}

// ExtractToken extracts the JWT token from Authorization or X-User-Token headers
func ExtractToken(authHeader, userTokenHeader string) (string, error) {
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer "), nil
		}
		return authHeader, nil
	}

	if userTokenHeader != "" {
		return userTokenHeader, nil
	}

	return "", fmt.Errorf("no token found in Authorization or X-User-Token headers")
}

// contextKey is an unexported type for context keys in this package
type contextKey string

const authContextKey contextKey = "authContext"

// SetAuthContext stores the AuthContext in the request context
func SetAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, authCtx)
}

// GetAuthContext retrieves the AuthContext from the request context
func GetAuthContext(ctx context.Context) (*AuthContext, bool) {
	authCtx, ok := ctx.Value(authContextKey).(*AuthContext)
	return authCtx, ok
}
