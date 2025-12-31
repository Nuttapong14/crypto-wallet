package blockchain

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/domain/entities"
)

var (
	// ErrNotImplemented is returned by adapter stubs for unimplemented features.
	ErrNotImplemented = errors.New("blockchain: operation not implemented")
	// ErrInvalidAddress indicates an address failed validation.
	ErrInvalidAddress = errors.New("blockchain: invalid address")
	// ErrInsufficientFunds indicates balance is insufficient for requested operation.
	ErrInsufficientFunds = errors.New("blockchain: insufficient funds")
)

// Chain aliases the domain chain enumeration for use within blockchain adapters.
type Chain = entities.Chain

const (
	ChainBTC = entities.ChainBTC
	ChainETH = entities.ChainETH
	ChainSOL = entities.ChainSOL
	ChainXLM = entities.ChainXLM
)

// FeePriority influences fee estimation behaviour.
type FeePriority string

const (
	FeePrioritySlow     FeePriority = "slow"
	FeePriorityStandard FeePriority = "standard"
	FeePriorityFast     FeePriority = "fast"
)

// BlockchainError represents a structured blockchain operation failure.
type BlockchainError struct {
	Code    string
	Message string
	Details map[string]any
}

// Error implements the error interface.
func (e *BlockchainError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Message == "" {
		return e.Code
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Wallet represents a generated or imported wallet.
type Wallet struct {
	Address        string
	PublicKey      string
	PrivateKey     string
	DerivationPath string
	Chain          Chain
}

// Balance captures the native token balance for an address.
type Balance struct {
	Address       string
	Balance       string
	Confirmations int
	LastUpdated   time.Time
}

// Fee captures fee estimation data.
type Fee struct {
	Amount        string
	EstimatedTime time.Duration
}

// FeeEstimateRequest describes a fee estimation query.
type FeeEstimateRequest struct {
	FromAddress string
	ToAddress   string
	Amount      string
	Priority    FeePriority
}

// FeeEstimate groups fee options for different priority levels.
type FeeEstimate struct {
	Slow     Fee
	Standard Fee
	Fast     Fee
}

// TransactionRequest describes a transaction creation request.
type TransactionRequest struct {
	FromAddress string
	ToAddress   string
	Amount      string
	Fee         string
	Memo        string
	Metadata    map[string]any
}

// UnsignedTransaction captures chain-specific unsigned transaction data.
type UnsignedTransaction struct {
	RawTx    []byte
	TxHash   string
	Metadata map[string]any
}

// SignedTransaction captures chain-specific signed transaction data.
type SignedTransaction struct {
	RawTx    []byte
	TxHash   string
	Metadata map[string]any
}

// TxStatus enumerates blockchain transaction states.
type TxStatus string

const (
	TxStatusPending   TxStatus = "pending"
	TxStatusConfirmed TxStatus = "confirmed"
	TxStatusFailed    TxStatus = "failed"
)

// Transaction provides details about a broadcast transaction.
type Transaction struct {
	Hash          string
	FromAddress   string
	ToAddress     string
	Amount        string
	Fee           string
	BlockNumber   uint64
	Confirmations int
	Status        TxStatus
	ErrorMessage  string
	Metadata      map[string]any
}

// TransactionStatus summarises the current state of a blockchain transaction.
type TransactionStatus struct {
	TxHash        string
	Status        TxStatus
	Confirmations int
	BlockNumber   uint64
	ErrorMessage  string
}

// NetworkInfo provides basic network statistics.
type NetworkInfo struct {
	Chain              Chain
	NetworkType        string
	CurrentBlockNumber uint64
	AverageBlockTime   time.Duration
	PeerCount          int
	IsHealthy          bool
}

// BlockchainAdapter defines the contract implemented by each blockchain integration.
type BlockchainAdapter interface {
	GenerateWallet(ctx context.Context) (*Wallet, error)
	ImportWallet(ctx context.Context, privateKey string) (*Wallet, error)
	ValidateAddress(ctx context.Context, address string) (bool, error)

	GetBalance(ctx context.Context, address string) (*Balance, error)

	EstimateFee(ctx context.Context, req *FeeEstimateRequest) (*FeeEstimate, error)
	CreateTransaction(ctx context.Context, req *TransactionRequest) (*UnsignedTransaction, error)
	SignTransaction(ctx context.Context, tx *UnsignedTransaction, privateKey string) (*SignedTransaction, error)
	BroadcastTransaction(ctx context.Context, tx *SignedTransaction) (string, error)
	GetTransaction(ctx context.Context, txHash string) (*Transaction, error)
	GetTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error)

	GetBlockNumber(ctx context.Context) (uint64, error)
	GetNetworkInfo(ctx context.Context) (*NetworkInfo, error)

	GetChain() Chain
	GetConfirmationThreshold() int
}

// BaseAdapter provides shared helpers for chain-specific adapters.
type BaseAdapter struct {
	chain                 Chain
	confirmationThreshold int
	logger                *slog.Logger
}

// newBaseAdapter constructs a BaseAdapter with sane defaults.
func newBaseAdapter(chain Chain, confirmationThreshold int, logger *slog.Logger) BaseAdapter {
	if confirmationThreshold <= 0 {
		confirmationThreshold = 1
	}
	if logger == nil {
		logger = slog.Default()
	}
	return BaseAdapter{
		chain:                 chain,
		confirmationThreshold: confirmationThreshold,
		logger:                logger,
	}
}

func (b BaseAdapter) GetChain() Chain {
	return b.chain
}

func (b BaseAdapter) GetConfirmationThreshold() int {
	return b.confirmationThreshold
}

func (b BaseAdapter) notImplemented(operation string) error {
	b.logger.Warn("blockchain operation not implemented",
		slog.String("chain", string(b.chain)),
		slog.String("operation", operation),
	)
	return ErrNotImplemented
}


func stubTxHash(chain Chain) string {
	hash := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))
	return fmt.Sprintf("%s-%s", chain, hash)
}

func mergeMetadata(values ...map[string]any) map[string]any {
	merged := map[string]any{}
	for _, value := range values {
		for k, v := range value {
			merged[k] = v
		}
	}
	return merged
}

func cloneMetadata(src map[string]any) map[string]any {
	if src == nil {
		return map[string]any{}
	}
	clone := make(map[string]any, len(src))
	for k, v := range src {
		clone[k] = v
	}
	return clone
}
