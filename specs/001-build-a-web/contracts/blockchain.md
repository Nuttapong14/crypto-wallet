# Blockchain Adapter Interface Contracts

**Feature**: Multi-Chain Blockchain Integration
**Version**: 1.0.0
**Date**: 2025-10-14

---

## Overview

Unified interface for interacting with Bitcoin, Ethereum, Solana, and Stellar blockchains. Provides abstraction layer following Hexagonal Architecture principles.

---

## BlockchainAdapter Interface

All blockchain implementations must implement this interface:

```go
package blockchain

import (
    "context"
    "time"
)

// BlockchainAdapter provides unified interface for blockchain operations
type BlockchainAdapter interface {
    // Wallet Operations
    GenerateWallet(ctx context.Context) (*Wallet, error)
    ImportWallet(ctx context.Context, privateKey string) (*Wallet, error)
    ValidateAddress(ctx context.Context, address string) (bool, error)

    // Balance Operations
    GetBalance(ctx context.Context, address string) (*Balance, error)

    // Transaction Operations
    EstimateFee(ctx context.Context, req *FeeEstimateRequest) (*FeeEstimate, error)
    CreateTransaction(ctx context.Context, req *TransactionRequest) (*UnsignedTransaction, error)
    SignTransaction(ctx context.Context, tx *UnsignedTransaction, privateKey string) (*SignedTransaction, error)
    BroadcastTransaction(ctx context.Context, tx *SignedTransaction) (string, error)
    GetTransaction(ctx context.Context, txHash string) (*Transaction, error)
    GetTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error)

    // Network Operations
    GetBlockNumber(ctx context.Context) (uint64, error)
    GetNetworkInfo(ctx context.Context) (*NetworkInfo, error)

    // Utility
    GetChain() Chain
    GetConfirmationThreshold() int
}
```

---

## Common Data Structures

### Wallet
```go
type Wallet struct {
    Address        string
    PublicKey      string
    PrivateKey     string // Encrypted before storage
    DerivationPath string // For HD wallets
    Chain          Chain
}
```

### Balance
```go
type Balance struct {
    Address      string
    Balance      string // Decimal string representation
    Confirmations int
    LastUpdated  time.Time
}
```

### FeeEstimateRequest
```go
type FeeEstimateRequest struct {
    FromAddress string
    ToAddress   string
    Amount      string
    Priority    FeePriority // slow, standard, fast
}
```

### FeeEstimate
```go
type FeeEstimate struct {
    Slow     Fee
    Standard Fee
    Fast     Fee
}

type Fee struct {
    Amount        string
    EstimatedTime time.Duration
}
```

### TransactionRequest
```go
type TransactionRequest struct {
    FromAddress string
    ToAddress   string
    Amount      string
    Fee         string
    Memo        string // For blockchains that support memos (Stellar)
    Metadata    map[string]interface{} // Chain-specific data
}
```

### UnsignedTransaction
```go
type UnsignedTransaction struct {
    RawTx    []byte
    TxHash   string
    Metadata map[string]interface{}
}
```

### SignedTransaction
```go
type SignedTransaction struct {
    RawTx      []byte
    TxHash     string
    Signature  []byte
    Metadata   map[string]interface{}
}
```

### Transaction
```go
type Transaction struct {
    TxHash        string
    BlockNumber   uint64
    FromAddress   string
    ToAddress     string
    Amount        string
    Fee           string
    Status        TxStatus
    Confirmations int
    Timestamp     time.Time
    Metadata      map[string]interface{}
}
```

### TransactionStatus
```go
type TransactionStatus struct {
    TxHash        string
    Status        TxStatus // pending, confirmed, failed
    Confirmations int
    BlockNumber   uint64
    ErrorMessage  string
}

type TxStatus string
const (
    TxStatusPending   TxStatus = "pending"
    TxStatusConfirmed TxStatus = "confirmed"
    TxStatusFailed    TxStatus = "failed"
)
```

### NetworkInfo
```go
type NetworkInfo struct {
    Chain             Chain
    NetworkType       string // mainnet, testnet
    CurrentBlockNumber uint64
    AverageBlockTime  time.Duration
    PeerCount         int
    IsHealthy         bool
}
```

### Chain
```go
type Chain string
const (
    ChainBTC Chain = "BTC"
    ChainETH Chain = "ETH"
    ChainSOL Chain = "SOL"
    ChainXLM Chain = "XLM"
)
```

### FeePriority
```go
type FeePriority string
const (
    FeePrioritySlow     FeePriority = "slow"
    FeePriorityStandard FeePriority = "standard"
    FeePriorityFast     FeePriority = "fast"
)
```

---

## Bitcoin Adapter Implementation

### Library
```go
import (
    "github.com/btcsuite/btcd/chaincfg"
    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/rpcclient"
    "github.com/btcsuite/btcutil"
)
```

### Specific Considerations
- Uses UTXO model (must select inputs and create change outputs)
- Fee calculation based on transaction size (bytes) and sat/byte rate
- Confirmation threshold: 6 blocks (~60 minutes)
- Support for P2PKH, P2SH, and Bech32 addresses
- Average block time: ~10 minutes

### Example Fee Estimate
```go
func (b *BitcoinAdapter) EstimateFee(ctx context.Context, req *FeeEstimateRequest) (*FeeEstimate, error) {
    // Estimate transaction size
    txSize := estimateTransactionSize(req)

    // Get current fee rates (sat/byte)
    slowRate := b.getFeeRate(ctx, 10)    // Confirm in ~100 minutes
    standardRate := b.getFeeRate(ctx, 3) // Confirm in ~30 minutes
    fastRate := b.getFeeRate(ctx, 1)     // Confirm in ~10 minutes

    return &FeeEstimate{
        Slow:     Fee{Amount: toDecimal(txSize * slowRate), EstimatedTime: 100 * time.Minute},
        Standard: Fee{Amount: toDecimal(txSize * standardRate), EstimatedTime: 30 * time.Minute},
        Fast:     Fee{Amount: toDecimal(txSize * fastRate), EstimatedTime: 10 * time.Minute},
    }, nil
}
```

---

## Ethereum Adapter Implementation

### Library
```go
import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/crypto"
)
```

### Specific Considerations
- Account model (not UTXO)
- Fee = gasPrice × gasUsed (typically 21,000 gas for simple transfer)
- Confirmation threshold: 12 blocks (~3 minutes)
- Supports EIP-1559 (base fee + priority fee)
- Average block time: ~12 seconds
- Nonce management required

### Example Transaction Creation
```go
func (e *EthereumAdapter) CreateTransaction(ctx context.Context, req *TransactionRequest) (*UnsignedTransaction, error) {
    nonce, _ := e.client.PendingNonceAt(ctx, common.HexToAddress(req.FromAddress))
    gasPrice, _ := e.client.SuggestGasPrice(ctx)
    gasLimit := uint64(21000)

    toAddress := common.HexToAddress(req.ToAddress)
    amount := toBigInt(req.Amount)

    tx := types.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, nil)

    return &UnsignedTransaction{
        RawTx: serializeTransaction(tx),
        TxHash: tx.Hash().Hex(),
        Metadata: map[string]interface{}{
            "nonce": nonce,
            "gasPrice": gasPrice.String(),
            "gasLimit": gasLimit,
        },
    }, nil
}
```

---

## Solana Adapter Implementation

### Library
```go
import (
    "github.com/portto/solana-go-sdk/client"
    "github.com/portto/solana-go-sdk/common"
    "github.com/portto/solana-go-sdk/types"
)
```

### Specific Considerations
- Account model with parallel transaction processing
- Fee structure: base fee + priority fee (measured in lamports)
- Confirmation threshold: 32 confirmations (~15 seconds)
- Very fast block time (~400ms)
- Recent blockhash required for transactions (valid for ~90 seconds)
- Compute units for transaction complexity

### Example Balance Query
```go
func (s *SolanaAdapter) GetBalance(ctx context.Context, address string) (*Balance, error) {
    pubkey := common.PublicKeyFromString(address)
    balance, err := s.client.GetBalance(ctx, pubkey.ToBase58())
    if err != nil {
        return nil, err
    }

    // Convert lamports to SOL (1 SOL = 1e9 lamports)
    balanceSOL := toDecimal(balance / 1e9)

    return &Balance{
        Address: address,
        Balance: balanceSOL,
        LastUpdated: time.Now(),
    }, nil
}
```

---

## Stellar Adapter Implementation

### Library
```go
import (
    "github.com/stellar/go/clients/horizonclient"
    "github.com/stellar/go/keypair"
    "github.com/stellar/go/network"
    "github.com/stellar/go/txnbuild"
)
```

### Specific Considerations
- Account model with minimum balance requirement (1 XLM base reserve)
- Fixed fee: 100 stroops (0.00001 XLM) per operation
- Confirmation threshold: 1 confirmation (~5 seconds)
- Supports memo field for payment identification
- Transaction requires sequence number
- Average ledger close time: ~5 seconds

### Example Transaction with Memo
```go
func (s *StellarAdapter) CreateTransaction(ctx context.Context, req *TransactionRequest) (*UnsignedTransaction, error) {
    sourceAccount, _ := s.client.AccountDetail(horizonclient.AccountRequest{
        AccountID: req.FromAddress,
    })

    tx, _ := txnbuild.NewTransaction(
        txnbuild.TransactionParams{
            SourceAccount:        &sourceAccount,
            IncrementSequenceNum: true,
            Operations: []txnbuild.Operation{
                &txnbuild.Payment{
                    Destination: req.ToAddress,
                    Amount:      req.Amount,
                    Asset:       txnbuild.NativeAsset{},
                },
            },
            BaseFee:       txnbuild.MinBaseFee,
            Memo:          txnbuild.MemoText(req.Memo),
            Preconditions: txnbuild.Preconditions{TimeBounds: txnbuild.NewTimeout(300)},
        },
    )

    return &UnsignedTransaction{
        RawTx: []byte(tx.Base64()),
        TxHash: tx.Hash(network.PublicNetworkPassphrase).Hex(),
    }, nil
}
```

---

## Error Handling

### Standard Error Codes
```go
type BlockchainError struct {
    Code    ErrorCode
    Message string
    Details map[string]interface{}
}

type ErrorCode string
const (
    ErrInvalidAddress      ErrorCode = "INVALID_ADDRESS"
    ErrInsufficientFunds   ErrorCode = "INSUFFICIENT_FUNDS"
    ErrInvalidAmount       ErrorCode = "INVALID_AMOUNT"
    ErrNetworkError        ErrorCode = "NETWORK_ERROR"
    ErrTransactionFailed   ErrorCode = "TRANSACTION_FAILED"
    ErrInvalidPrivateKey   ErrorCode = "INVALID_PRIVATE_KEY"
    ErrInvalidSignature    ErrorCode = "INVALID_SIGNATURE"
    ErrNonceTooLow         ErrorCode = "NONCE_TOO_LOW"
    ErrGasPriceTooLow      ErrorCode = "GAS_PRICE_TOO_LOW"
    ErrBlockNotFound       ErrorCode = "BLOCK_NOT_FOUND"
    ErrTransactionNotFound ErrorCode = "TRANSACTION_NOT_FOUND"
)
```

---

## Configuration

### Blockchain Connection Config
```yaml
blockchains:
  btc:
    rpc_url: "https://bitcoin-rpc.example.com"
    rpc_user: "user"
    rpc_password: "pass"
    network: "mainnet"
    confirmation_threshold: 6

  eth:
    rpc_url: "https://ethereum-rpc.example.com"
    network: "mainnet"
    chain_id: 1
    confirmation_threshold: 12

  sol:
    rpc_url: "https://solana-rpc.example.com"
    network: "mainnet-beta"
    confirmation_threshold: 32

  xlm:
    horizon_url: "https://horizon.stellar.org"
    network: "public"
    confirmation_threshold: 1
```

---

## Testing Strategy

### Unit Tests
- Mock blockchain responses
- Test address validation
- Test transaction creation/signing
- Test error handling

### Integration Tests
- Use testnets for real blockchain interaction
- Test full transaction lifecycle
- Test network failure scenarios
- Test concurrent operations

### Test Fixtures
```go
// Example test fixture
func TestBitcoinAdapter_GetBalance(t *testing.T) {
    mockClient := &MockBitcoinClient{}
    adapter := NewBitcoinAdapter(mockClient)

    mockClient.On("GetBalance", "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa").
        Return(btcutil.Amount(100000000), nil) // 1 BTC

    balance, err := adapter.GetBalance(context.Background(), "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa")

    assert.NoError(t, err)
    assert.Equal(t, "1.00000000", balance.Balance)
}
```

---

## Performance Requirements

- Balance query: <2s (SC-007)
- Transaction submission: <5s
- Fee estimation: <1s
- Address validation: <100ms
- Block number query: <500ms

---

## Security Considerations

1. **Private Key Handling**:
   - Never log private keys
   - Encrypt before storage
   - Use secure memory for operations
   - Wipe memory after use

2. **Transaction Validation**:
   - Validate all addresses before transactions
   - Check sufficient balance including fees
   - Verify transaction data before signing

3. **Network Security**:
   - Use TLS for all RPC connections
   - Verify SSL certificates
   - Implement request timeouts
   - Rate limit RPC calls

4. **Error Handling**:
   - Don't expose internal errors to users
   - Log all blockchain errors for debugging
   - Implement retry logic for transient failures
