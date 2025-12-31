package blockchain

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// BitcoinConfig captures connection parameters for the Bitcoin RPC client.
type BitcoinConfig struct {
	RPCURL                string
	RPCUser               string
	RPCPassword           string
	Network               string
	ConfirmationThreshold int
}

// BitcoinAdapter provides Bitcoin blockchain integration (stub implementation).
type BitcoinAdapter struct {
	BaseAdapter
	config BitcoinConfig
}

// NewBitcoinAdapter constructs a BitcoinAdapter stub.
func NewBitcoinAdapter(cfg BitcoinConfig, logger *slog.Logger) *BitcoinAdapter {
	threshold := cfg.ConfirmationThreshold
	if threshold <= 0 {
		threshold = 6
	}
	return &BitcoinAdapter{
		BaseAdapter: newBaseAdapter(ChainBTC, threshold, logger),
		config:      cfg,
	}
}

func (b *BitcoinAdapter) GenerateWallet(ctx context.Context) (*Wallet, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	privateKeyBytes, err := randomBytes(32)
	if err != nil {
		return nil, fmt.Errorf("bitcoin: generate private key: %w", err)
	}
	publicKey, err := randomPublicKeyString()
	if err != nil {
		return nil, fmt.Errorf("bitcoin: generate public key: %w", err)
	}

	addressSeed, err := randomBytes(20)
	if err != nil {
		return nil, fmt.Errorf("bitcoin: generate address seed: %w", err)
	}

	address := "bc1" + encodeBase32Lower(addressSeed)
	if len(address) > 42 {
		address = address[:42]
	}

	privateKey := "K" + encodeBase58(privateKeyBytes)

	return &Wallet{
		Address:        address,
		PublicKey:      publicKey,
		PrivateKey:     privateKey,
		DerivationPath: "m/84'/0'/0'/0/0",
		Chain:          ChainBTC,
	}, nil
}

func (b *BitcoinAdapter) ImportWallet(ctx context.Context, privateKey string) (*Wallet, error) {
	return nil, b.notImplemented("ImportWallet")
}

func (b *BitcoinAdapter) ValidateAddress(ctx context.Context, address string) (bool, error) {
	return false, b.notImplemented("ValidateAddress")
}

func (b *BitcoinAdapter) GetBalance(ctx context.Context, address string) (*Balance, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(address) == "" {
		return nil, fmt.Errorf("bitcoin: address is required")
	}
	return synthBalance(address, b.GetConfirmationThreshold()), nil
}

func (b *BitcoinAdapter) EstimateFee(ctx context.Context, req *FeeEstimateRequest) (*FeeEstimate, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &FeeEstimate{
		Slow: Fee{Amount: "0.0001", EstimatedTime: 3 * time.Minute},
		Standard: Fee{Amount: "0.0002", EstimatedTime: 90 * time.Second},
		Fast: Fee{Amount: "0.0003", EstimatedTime: 30 * time.Second},
	}, nil
}

func (b *BitcoinAdapter) CreateTransaction(ctx context.Context, req *TransactionRequest) (*UnsignedTransaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.New("bitcoin: request is required")
	}
	if strings.TrimSpace(req.FromAddress) == "" || strings.TrimSpace(req.ToAddress) == "" {
		return nil, ErrInvalidAddress
	}
	metadata := mergeMetadata(map[string]any{"memo": req.Memo}, cloneMetadata(req.Metadata))
	unsigned := &UnsignedTransaction{
		TxHash:   stubTxHash(b.GetChain()),
		RawTx:    []byte(time.Now().UTC().Format(time.RFC3339Nano)),
		Metadata: metadata,
	}
	return unsigned, nil
}

func (b *BitcoinAdapter) SignTransaction(ctx context.Context, tx *UnsignedTransaction, privateKey string) (*SignedTransaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, errors.New("bitcoin: unsigned transaction required")
	}
	_ = privateKey
	signed := &SignedTransaction{
		TxHash:   tx.TxHash,
		RawTx:    append([]byte{}, tx.RawTx...),
		Metadata: mergeMetadata(tx.Metadata, map[string]any{"signed_at": time.Now().UTC().Format(time.RFC3339Nano)}),
	}
	return signed, nil
}

func (b *BitcoinAdapter) BroadcastTransaction(ctx context.Context, tx *SignedTransaction) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if tx == nil {
		return "", errors.New("bitcoin: signed transaction required")
	}
	return tx.TxHash, nil
}

func (b *BitcoinAdapter) GetTransaction(ctx context.Context, txHash string) (*Transaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(txHash) == "" {
		return nil, errors.New("bitcoin: transaction hash required")
	}
	status, err := b.GetTransactionStatus(ctx, txHash)
	if err != nil {
		return nil, err
	}
	return &Transaction{
		Hash:          txHash,
		Status:        status.Status,
		Confirmations: status.Confirmations,
		BlockNumber:   status.BlockNumber,
		Metadata:      map[string]any{"stub": true},
	}, nil
}

func (b *BitcoinAdapter) GetTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(txHash) == "" {
		return nil, errors.New("bitcoin: transaction hash required")
	}
	return &TransactionStatus{
		TxHash:        txHash,
		Status:        TxStatusPending,
		Confirmations: 0,
		BlockNumber:   0,
	}, nil
}

func (b *BitcoinAdapter) GetBlockNumber(ctx context.Context) (uint64, error) {
	return 0, b.notImplemented("GetBlockNumber")
}

func (b *BitcoinAdapter) GetNetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	return nil, b.notImplemented("GetNetworkInfo")
}
