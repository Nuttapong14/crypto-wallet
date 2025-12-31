package security

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrTokenInvalid indicates the token failed signature or claims validation.
	ErrTokenInvalid = errors.New("security: token is invalid")
	// ErrTokenExpired indicates the token has expired.
	ErrTokenExpired = errors.New("security: token has expired")
)

// JWTConfig defines configuration required to initialise the JWT service.
type JWTConfig struct {
	Secret        string
	Issuer        string
	Audience      []string
	Leeway        time.Duration
	SigningMethod string
	Clock         func() time.Time
}

// Claims wraps jwt.RegisteredClaims with custom metadata.
type Claims struct {
	Metadata map[string]any `json:"metadata,omitempty"`
	jwt.RegisteredClaims
}

// JWTService provides helpers for issuing and validating JWT tokens.
type JWTService struct {
	secret        []byte
	issuer        string
	audience      []string
	leeway        time.Duration
	signingMethod jwt.SigningMethod
	clock         func() time.Time
}

// NewJWTService builds a JWTService from configuration.
func NewJWTService(cfg JWTConfig) (*JWTService, error) {
	secret := strings.TrimSpace(cfg.Secret)
	if secret == "" {
		return nil, errors.New("security: JWT secret is required")
	}

	method := strings.TrimSpace(strings.ToUpper(cfg.SigningMethod))
	if method == "" {
		method = jwt.SigningMethodHS256.Alg()
	}

	signingMethod := jwt.GetSigningMethod(method)
	if signingMethod == nil {
		return nil, fmt.Errorf("security: unsupported signing method %s", method)
	}

	clock := cfg.Clock
	if clock == nil {
		clock = time.Now
	}

	service := &JWTService{
		secret:        []byte(secret),
		issuer:        cfg.Issuer,
		audience:      cfg.Audience,
		leeway:        cfg.Leeway,
		signingMethod: signingMethod,
		clock:         clock,
	}

	if service.leeway < 0 {
		service.leeway = 0
	}

	return service, nil
}

// Sign issues a token based on the provided claims, applying defaults when necessary.
func (s *JWTService) Sign(_ context.Context, claims Claims) (string, error) {
	if s == nil {
		return "", errors.New("security: JWT service not initialised")
	}

	if claims.RegisteredClaims.Issuer == "" && s.issuer != "" {
		claims.RegisteredClaims.Issuer = s.issuer
	}

	if len(s.audience) > 0 && len(claims.RegisteredClaims.Audience) == 0 {
		claims.RegisteredClaims.Audience = jwt.ClaimStrings(s.audience)
	}

	now := s.clock().UTC()
	if claims.RegisteredClaims.IssuedAt == nil {
		claims.RegisteredClaims.IssuedAt = jwt.NewNumericDate(now)
	}
	if claims.RegisteredClaims.NotBefore == nil {
		claims.RegisteredClaims.NotBefore = jwt.NewNumericDate(now)
	}

	token := jwt.NewWithClaims(s.signingMethod, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("security: sign token: %w", err)
	}
	return signed, nil
}

// GenerateToken creates a token for the supplied subject with the provided TTL and optional metadata.
func (s *JWTService) GenerateToken(ctx context.Context, subject string, ttl time.Duration, metadata map[string]any) (string, error) {
	if strings.TrimSpace(subject) == "" {
		return "", errors.New("security: subject is required")
	}
	if ttl <= 0 {
		return "", errors.New("security: token TTL must be positive")
	}
	claims := Claims{
		Metadata: metadata,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(s.clock().UTC().Add(ttl)),
		},
	}
	return s.Sign(ctx, claims)
}

// Parse validates a token string and returns the claims when successful.
func (s *JWTService) Parse(_ context.Context, tokenString string) (*Claims, error) {
	if s == nil {
		return nil, errors.New("security: JWT service not initialised")
	}

	// Build all parser options including validation options
	parserOpts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{s.signingMethod.Alg()}),
		jwt.WithLeeway(s.leeway),
	}
	parserOpts = append(parserOpts, buildValidationOptions(s.issuer, s.audience)...)

	parser := jwt.NewParser(parserOpts...)

	var claims Claims
	token, err := parser.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	return &claims, nil
}

// buildValidationOptions constructs parser options based on issuer and audience configuration.
func buildValidationOptions(issuer string, audience []string) []jwt.ParserOption {
	opts := []jwt.ParserOption{}
	if issuer != "" {
		opts = append(opts, jwt.WithIssuer(issuer))
	}
	// WithAudience expects a single string, so add one option per audience
	for _, aud := range audience {
		opts = append(opts, jwt.WithAudience(aud))
	}
	return opts
}
