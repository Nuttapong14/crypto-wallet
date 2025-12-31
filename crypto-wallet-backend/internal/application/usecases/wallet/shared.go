package wallet

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/crypto-wallet/backend/internal/application/dto"
	"github.com/crypto-wallet/backend/internal/domain/entities"
	"github.com/crypto-wallet/backend/internal/domain/repositories"
	"github.com/crypto-wallet/backend/internal/domain/services"
	"github.com/crypto-wallet/backend/internal/infrastructure/blockchain"
)

// Service defines the contract required from the domain wallet service.
type Service interface {
	CreateWallet(ctx context.Context, params services.CreateWalletParams) (entities.Wallet, error)
	ListWallets(ctx context.Context, userID uuid.UUID, filter repositories.WalletFilter, opts repositories.ListOptions) ([]entities.Wallet, error)
	GetWalletByID(ctx context.Context, id uuid.UUID) (entities.Wallet, error)
	RefreshWalletBalance(ctx context.Context, walletID uuid.UUID) (entities.Wallet, *blockchain.Balance, error)
}

func mapWalletEntity(entity entities.Wallet) dto.Wallet {
	if entity == nil {
		return dto.Wallet{}
	}

	balanceUpdated := entity.GetBalanceUpdatedAt()
	var copiedBalanceUpdated *time.Time
	if balanceUpdated != nil {
		value := balanceUpdated.UTC()
		copiedBalanceUpdated = &value
	}

	return dto.Wallet{
		ID:               entity.GetID(),
		Chain:            string(entity.GetChain()),
		Address:          entity.GetAddress(),
		Label:            entity.GetLabel(),
		Balance:          entity.GetBalance().String(),
		Status:           string(entity.GetStatus()),
		CreatedAt:        entity.GetCreatedAt().UTC(),
		UpdatedAt:        entity.GetUpdatedAt().UTC(),
		BalanceUpdatedAt: copiedBalanceUpdated,
	}
}

func mapWallets(entitiesList []entities.Wallet) []dto.Wallet {
	results := make([]dto.Wallet, 0, len(entitiesList))
	for _, entity := range entitiesList {
		results = append(results, mapWalletEntity(entity))
	}
	return results
}

func mapBalance(wallet entities.Wallet, balance *blockchain.Balance, balanceUSD string) dto.WalletBalance {
	result := dto.WalletBalance{
		WalletID: wallet.GetID(),
		Chain:    string(wallet.GetChain()),
		Address:  wallet.GetAddress(),
		Balance:  wallet.GetBalance().String(),
	}

	if balance != nil {
		if strings.TrimSpace(balance.Balance) != "" {
			result.Balance = balance.Balance
		}
		result.Confirmations = balance.Confirmations
		if !balance.LastUpdated.IsZero() {
			result.LastUpdated = balance.LastUpdated.UTC()
		}
	}

	if result.LastUpdated.IsZero() {
		if ts := wallet.GetBalanceUpdatedAt(); ts != nil {
			result.LastUpdated = ts.UTC()
		} else {
			result.LastUpdated = wallet.GetUpdatedAt().UTC()
		}
	}

	result.BalanceUSD = balanceUSD

	return result
}
