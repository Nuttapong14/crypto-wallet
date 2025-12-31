package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateWalletRequest models the payload for wallet creation.
type CreateWalletRequest struct {
	Chain string `json:"chain"`
	Label string `json:"label,omitempty"`
}

// Wallet represents a wallet summary returned to clients.
type Wallet struct {
	ID               uuid.UUID  `json:"id"`
	Chain            string     `json:"chain"`
	Address          string     `json:"address"`
	Label            string     `json:"label"`
	Balance          string     `json:"balance"`
	BalanceUSD       string     `json:"balance_usd,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	BalanceUpdatedAt *time.Time `json:"balance_updated_at,omitempty"`
}

// WalletDetail extends Wallet with additional metadata.
type WalletDetail struct {
	Wallet
	DerivationPath   string `json:"derivation_path,omitempty"`
	TransactionCount int    `json:"transaction_count,omitempty"`
}

// WalletList groups a collection of wallets with paging metadata.
type WalletList struct {
	Wallets []Wallet `json:"wallets"`
	Total   int      `json:"total"`
}

// WalletBalance summarises balance information for a wallet.
type WalletBalance struct {
	WalletID      uuid.UUID `json:"wallet_id"`
	Chain         string    `json:"chain"`
	Address       string    `json:"address"`
	Balance       string    `json:"balance"`
	BalanceUSD    string    `json:"balance_usd,omitempty"`
	Confirmations int       `json:"confirmations"`
	LastUpdated   time.Time `json:"last_updated"`
}
