package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string   `json:"kid"`
	Kty string   `json:"kty"`
	Alg string   `json:"alg"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c,omitempty"`
}

// JWKSProvider fetches and caches JWKS from Zitadel
type JWKSProvider struct {
	jwksURL    string
	httpClient *http.Client
	cacheTTL   time.Duration

	mu         sync.RWMutex
	cachedJWKS *JWKS
	cacheTime  time.Time
}

// NewJWKSProvider creates a new JWKS provider with caching
func NewJWKSProvider(jwksURL string, cacheTTL time.Duration) *JWKSProvider {
	return &JWKSProvider{
		jwksURL: jwksURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cacheTTL: cacheTTL,
	}
}

// GetJWKS fetches the JWKS, using cache if available and not expired
func (p *JWKSProvider) GetJWKS(ctx context.Context) (*JWKS, error) {
	p.mu.RLock()
	if p.cachedJWKS != nil && time.Since(p.cacheTime) < p.cacheTTL {
		cached := p.cachedJWKS
		p.mu.RUnlock()
		return cached, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cachedJWKS != nil && time.Since(p.cacheTime) < p.cacheTTL {
		return p.cachedJWKS, nil
	}

	jwks, err := p.fetchJWKS(ctx)
	if err != nil {
		if p.cachedJWKS != nil {
			return p.cachedJWKS, nil
		}
		return nil, err
	}

	p.cachedJWKS = jwks
	p.cacheTime = time.Now()

	return jwks, nil
}

func (p *JWKSProvider) fetchJWKS(ctx context.Context) (*JWKS, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	return &jwks, nil
}

// GetKey retrieves a specific key by kid from the JWKS
func (p *JWKSProvider) GetKey(ctx context.Context, kid string) (*JWK, error) {
	jwks, err := p.GetJWKS(ctx)
	if err != nil {
		return nil, err
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid {
			return &key, nil
		}
	}

	return nil, fmt.Errorf("key with kid %s not found in JWKS", kid)
}

// KeyFunc returns a jwt.Keyfunc for use with jwt.ParseWithClaims
func (p *JWKSProvider) KeyFunc(ctx context.Context) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid header not found in token")
		}

		jwk, err := p.GetKey(ctx, kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get JWK: %w", err)
		}

		return parseRSAPublicKey(jwk)
	}
}

func parseRSAPublicKey(jwk *JWK) (interface{}, error) {
	if len(jwk.X5c) > 0 {
		cert := "-----BEGIN CERTIFICATE-----\n" + jwk.X5c[0] + "\n-----END CERTIFICATE-----"
		return jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
	}

	if jwk.N == "" || jwk.E == "" {
		return nil, fmt.Errorf("JWK missing required RSA components (n and e)")
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWK modulus (n): %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWK exponent (e): %w", err)
	}

	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}

	pubKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}

	return pubKey, nil
}

// InvalidateCache clears the cached JWKS
func (p *JWKSProvider) InvalidateCache() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cachedJWKS = nil
	p.cacheTime = time.Time{}
}
