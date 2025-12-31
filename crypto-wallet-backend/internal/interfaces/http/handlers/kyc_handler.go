package handlers

import (
	"io"
	"log/slog"
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/crypto-wallet/backend/internal/application/dto"
	kycusecase "github.com/crypto-wallet/backend/internal/application/usecases/kyc"
	"github.com/crypto-wallet/backend/internal/interfaces/http/middleware"
	"github.com/crypto-wallet/backend/pkg/utils"
)

// KYCHandler wires KYC-related use cases to HTTP endpoints.
type KYCHandler struct {
	submitUC *kycusecase.SubmitKYCUseCase
	uploadUC *kycusecase.UploadDocumentUseCase
	statusUC *kycusecase.GetKYCStatusUseCase
	logger   *slog.Logger
}

// KYCHandlerConfig configures handler dependencies.
type KYCHandlerConfig struct {
	SubmitUseCase *kycusecase.SubmitKYCUseCase
	UploadUseCase *kycusecase.UploadDocumentUseCase
	StatusUseCase *kycusecase.GetKYCStatusUseCase
	Logger        *slog.Logger
}

// NewKYCHandler constructs a KYCHandler.
func NewKYCHandler(cfg KYCHandlerConfig) *KYCHandler {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &KYCHandler{
		submitUC: cfg.SubmitUseCase,
		uploadUC: cfg.UploadUseCase,
		statusUC: cfg.StatusUseCase,
		logger:   logger,
	}
}

// Register attaches routes to the router.
func (h *KYCHandler) Register(router fiber.Router) {
	if router == nil {
		return
	}

	router.Post("/submit", h.handleSubmit)
	router.Post("/documents", h.handleUploadDocument)
	router.Get("/status", h.handleStatus)
}

func (h *KYCHandler) handleSubmit(c *fiber.Ctx) error {
	if h.submitUC == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "kyc submission not configured")
	}

	userID, err := extractUserID(c)
	if err != nil {
		return err
	}

	var payload dto.KYCSubmitRequest
	if err := c.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request payload")
	}

	result, err := h.submitUC.Execute(c.UserContext(), kycusecase.SubmitKYCInput{
		UserID:  userID,
		Payload: payload,
		Email:   extractUserEmail(c),
	})
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *KYCHandler) handleUploadDocument(c *fiber.Ctx) error {
	if h.uploadUC == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "kyc document upload not configured")
	}

	userID, err := extractUserID(c)
	if err != nil {
		return err
	}

	documentType := c.FormValue("document_type")
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "document file is required")
	}

	if fileHeader.Size > 10*1024*1024 {
		return fiber.NewError(fiber.StatusBadRequest, "document file exceeds 10MB limit")
	}

	content, err := readFileContent(fileHeader)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	result, err := h.uploadUC.Execute(c.UserContext(), kycusecase.UploadDocumentInput{
		UserID:       userID,
		DocumentType: documentType,
		FileName:     fileHeader.Filename,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		Content:      content,
	})
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *KYCHandler) handleStatus(c *fiber.Ctx) error {
	if h.statusUC == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "kyc status not configured")
	}

	userID, err := extractUserID(c)
	if err != nil {
		return err
	}

	result, err := h.statusUC.Execute(c.UserContext(), kycusecase.GetKYCStatusInput{
		UserID: userID,
	})
	if err != nil {
		return respondError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func readFileContent(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, 10*1024*1024))
	if err != nil {
		return nil, err
	}
	return content, nil
}

func extractUserEmail(c *fiber.Ctx) string {
	claims := c.Locals(middleware.AuthContextKey)
	switch value := claims.(type) {
	case map[string]any:
		if v, ok := value["email"].(string); ok {
			return v
		}
	}
	return ""
}
