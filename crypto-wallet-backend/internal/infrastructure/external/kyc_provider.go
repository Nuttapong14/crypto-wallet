package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	defaultKYCProviderTimeout   = 10 * time.Second
	defaultKYCProviderUserAgent = "atlas-wallet-kyc-client/1.0"
)

var (
	// ErrKYCProviderUnavailable indicates the upstream KYC provider is unavailable.
	ErrKYCProviderUnavailable = errors.New("kyc provider: service unavailable")
	// ErrKYCProviderUnauthorized indicates the provided API credentials are invalid.
	ErrKYCProviderUnauthorized = errors.New("kyc provider: invalid credentials")
	// ErrKYCProviderRequest indicates the upstream service rejected the request payload.
	ErrKYCProviderRequest = errors.New("kyc provider: request rejected")
)

// KYCSubmissionPayload captures identity information required by the external provider.
type KYCSubmissionPayload struct {
	ExternalUserID string            `json:"externalUserId"`
	Email          string            `json:"email,omitempty"`
	FirstName      string            `json:"firstName"`
	LastName       string            `json:"lastName"`
	DateOfBirth    string            `json:"dateOfBirth"`
	Nationality    string            `json:"nationality"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// KYCSubmissionResult represents the provider response when initiating verification.
type KYCSubmissionResult struct {
	ApplicationID string `json:"applicationId"`
	Status        string `json:"status"`
	ReviewURL     string `json:"reviewUrl,omitempty"`
}

// KYCDocumentUploadPayload describes a document upload request.
type KYCDocumentUploadPayload struct {
	ApplicationID string
	DocumentType  string
	FileName      string
	MimeType      string
	Content       []byte
}

// KYCDocumentUploadResult represents the provider response when uploading a document.
type KYCDocumentUploadResult struct {
	DocumentID string `json:"documentId"`
	Status     string `json:"status"`
}

// KYCStatusResult describes the external verification state.
type KYCStatusResult struct {
	Status        string `json:"status"`
	ReviewResult  string `json:"reviewResult,omitempty"`
	ApprovedAt    string `json:"approvedAt,omitempty"`
	RejectionCode string `json:"rejectionCode,omitempty"`
}

// KYCProviderClient defines the operations required from a third-party KYC provider.
type KYCProviderClient interface {
	SubmitApplication(ctx context.Context, payload KYCSubmissionPayload) (*KYCSubmissionResult, error)
	UploadDocument(ctx context.Context, payload KYCDocumentUploadPayload) (*KYCDocumentUploadResult, error)
	GetStatus(ctx context.Context, applicationID string) (*KYCStatusResult, error)
}

// KYCProviderConfig configures the SumSub-compatible client.
type KYCProviderConfig struct {
	BaseURL   string
	APIKey    string
	Secret    string
	Timeout   time.Duration
	Logger    *slog.Logger
	UserAgent string
	HTTPClient *http.Client
}

type kycProviderClient struct {
	baseURL    *url.URL
	apiKey     string
	secret     string
	httpClient *http.Client
	logger     *slog.Logger
	userAgent  string
}

// NewKYCProviderClient constructs a SumSub-compatible HTTP client.
func NewKYCProviderClient(cfg KYCProviderConfig) (KYCProviderClient, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, errors.New("kyc provider: baseURL is required")
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("kyc provider: api key is required")
	}

	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("kyc provider: parse baseURL: %w", err)
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultKYCProviderTimeout
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	userAgent := cfg.UserAgent
	if strings.TrimSpace(userAgent) == "" {
		userAgent = defaultKYCProviderUserAgent
	}

	return &kycProviderClient{
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		secret:     cfg.Secret,
		httpClient: httpClient,
		logger:     logger,
		userAgent:  userAgent,
	}, nil
}

func (c *kycProviderClient) SubmitApplication(ctx context.Context, payload KYCSubmissionPayload) (*KYCSubmissionResult, error) {
	endpoint := c.endpoint("/applicants")
	response := &KYCSubmissionResult{}

	if err := c.doRequest(ctx, http.MethodPost, endpoint, payload, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *kycProviderClient) UploadDocument(ctx context.Context, payload KYCDocumentUploadPayload) (*KYCDocumentUploadResult, error) {
	if strings.TrimSpace(payload.ApplicationID) == "" {
		return nil, errors.New("kyc provider: application id is required")
}

	endpoint := c.endpoint(path.Join("/applicants", payload.ApplicationID, "documents"))

	body := map[string]any{
		"documentType": payload.DocumentType,
		"fileName":     payload.FileName,
		"mimeType":     payload.MimeType,
		"content":      payload.Content,
	}

	response := &KYCDocumentUploadResult{}
	if err := c.doRequest(ctx, http.MethodPost, endpoint, body, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *kycProviderClient) GetStatus(ctx context.Context, applicationID string) (*KYCStatusResult, error) {
	if strings.TrimSpace(applicationID) == "" {
		return nil, errors.New("kyc provider: application id is required")
	}

	endpoint := c.endpoint(path.Join("/applicants", applicationID, "status"))
	response := &KYCStatusResult{}
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *kycProviderClient) doRequest(ctx context.Context, method string, endpoint *url.URL, payload any, dest any) error {
	var body io.Reader
	if payload != nil {
		buf, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("kyc provider: encode payload: %w", err)
		}
		body = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return fmt.Errorf("kyc provider: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", c.apiKey)
	if strings.TrimSpace(c.secret) != "" {
		req.Header.Set("X-API-SECRET", c.secret)
	}
	req.Header.Set("User-Agent", c.userAgent)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrKYCProviderUnavailable, err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return ErrKYCProviderUnauthorized
	}
	if res.StatusCode >= 400 && res.StatusCode < 500 {
		detail, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		c.logger.Warn("kyc provider request rejected", slog.Int("status", res.StatusCode), slog.String("response", string(detail)))
		return ErrKYCProviderRequest
	}
	if res.StatusCode >= 500 {
		return fmt.Errorf("%w: status=%d", ErrKYCProviderUnavailable, res.StatusCode)
	}

	if dest == nil {
		return nil
	}

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("kyc provider: decode response: %w", err)
	}
	return nil
}

func (c *kycProviderClient) endpoint(p string) *url.URL {
	clone := *c.baseURL
	clone.Path = path.Join(clone.Path, strings.TrimPrefix(p, "/"))
	return &clone
}
