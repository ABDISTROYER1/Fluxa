package indexer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fluxa/fluxa/internal/domain"
	"github.com/fluxa/fluxa/internal/stellar"
	"github.com/fluxa/fluxa/internal/transfer"
	"github.com/fluxa/fluxa/internal/wallet"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/protocols/horizon/operations"
)

type Config struct {
	PaymentsPageLimit int
	StreamMinBackoff  time.Duration
	StreamMaxBackoff  time.Duration
	SyncPageSize      int
}

func DefaultConfig() Config {
	return Config{
		PaymentsPageLimit: 50,
		StreamMinBackoff:  1 * time.Second,
		StreamMaxBackoff:  30 * time.Second,
		SyncPageSize:      100,
	}
}

type Indexer struct {
	walletRepo wallet.Repository
	txRepo     transfer.Repository
	stellar    stellar.Client
	config     Config
}

func New(walletRepo wallet.Repository, txRepo transfer.Repository, stellarClient stellar.Client) *Indexer {
	return NewWithConfig(walletRepo, txRepo, stellarClient, DefaultConfig())
}

func NewWithConfig(walletRepo wallet.Repository, txRepo transfer.Repository, stellarClient stellar.Client, config Config) *Indexer {
	if config.PaymentsPageLimit <= 0 {
		config.PaymentsPageLimit = DefaultConfig().PaymentsPageLimit
	}
	if config.StreamMinBackoff <= 0 {
		config.StreamMinBackoff = DefaultConfig().StreamMinBackoff
	}
	if config.StreamMaxBackoff <= 0 {
		config.StreamMaxBackoff = DefaultConfig().StreamMaxBackoff
	}
	if config.SyncPageSize <= 0 {
		config.SyncPageSize = DefaultConfig().SyncPageSize
	}
	return &Indexer{
		walletRepo: walletRepo,
		txRepo:     txRepo,
		stellar:    stellarClient,
		config:     config,
	}
}

// SyncAll iterates over all wallets and syncs their recent payments from Horizon.
// It pages through all wallets using limit/offset until no more wallets are returned.
func (idx *Indexer) SyncAll(ctx context.Context) error {
	limit := idx.config.SyncPageSize
	offset := 0

	for {
		wallets, err := idx.walletRepo.List(ctx, limit, offset)
		if err != nil {
			return fmt.Errorf("list wallets: %w", err)
		}

		if len(wallets) == 0 {
			break
		}

		for _, w := range wallets {
			if err := idx.SyncWallet(ctx, w); err != nil {
				log.Error().Err(err).Str("wallet_id", w.ID).Msg("failed to sync wallet")
			}
		}

		if len(wallets) < limit {
			break
		}
		offset += limit
	}
	return nil
}

// SyncWallet syncs Horizon payment history for a single wallet into the local DB.
// It persists the account's current balances and processes every payment
// operation since the wallet's stored cursor, advancing the cursor as it goes
// so a subsequent call resumes rather than reprocessing history.
// TenantID is set from the indexed wallet on every created transaction.
func (idx *Indexer) SyncWallet(ctx context.Context, w *domain.Wallet) error {
	acct, err := stellar.LoadAccountWithContext(ctx, idx.stellar, w.PublicKey)
	if err != nil {
		if stellar.IsNotFound(err) {
			return nil // account not yet funded — nothing to sync
		}
		return fmt.Errorf("load account %s: %w", w.PublicKey, err)
	}

	if err := idx.persistBalances(ctx, w.ID, acct); err != nil {
		return fmt.Errorf("persist balances: %w", err)
	}

	cursor := w.SyncCursor
	for {
		ops, err := stellar.PaymentsWithContext(ctx, idx.stellar, w.PublicKey, cursor, uint(idx.config.PaymentsPageLimit))
		if err != nil {
			return fmt.Errorf("fetch payments since cursor %q: %w", cursor, err)
		}
		if len(ops) == 0 {
			break
		}

		for _, op := range ops {
			if err := idx.processPayment(ctx, w, op); err != nil {
				log.Error().Err(err).Str("wallet_id", w.ID).Str("op_id", op.GetID()).
					Msg("indexer: process payment failed")
			}
			cursor = op.PagingToken()
		}

		if err := idx.walletRepo.UpdateSyncCursor(ctx, w.ID, cursor); err != nil {
			return fmt.Errorf("update sync cursor: %w", err)
		}
		w.SyncCursor = cursor

		if len(ops) < idx.config.PaymentsPageLimit {
			break
		}
	}

	return nil
}

// persistBalances upserts every asset balance reported by Horizon for the wallet.
func (idx *Indexer) persistBalances(ctx context.Context, walletID string, acct horizon.Account) error {
	for _, b := range acct.Balances {
		code := b.Code
		if code == "" {
			code = "XLM"
		}
		amt, err := decimal.NewFromString(b.Balance)
		if err != nil {
			return fmt.Errorf("parse balance %q for asset %s: %w", b.Balance, code, err)
		}
		if err := idx.walletRepo.UpsertBalance(ctx, walletID, code, b.Issuer, amt); err != nil {
			return fmt.Errorf("upsert balance for asset %s: %w", code, err)
		}
	}
	return nil
}

// StreamAll starts a real-time Horizon payment stream for every wallet and
// blocks until ctx is canceled. It pages through all wallets using limit/offset.
// Each wallet streams on its own goroutine so a reconnect loop on one wallet
// never blocks or is affected by another.
func (idx *Indexer) StreamAll(ctx context.Context) error {
	limit := idx.config.SyncPageSize
	offset := 0

	var wg sync.WaitGroup
	for {
		wallets, err := idx.walletRepo.List(ctx, limit, offset)
		if err != nil {
			return fmt.Errorf("list wallets: %w", err)
		}

		if len(wallets) == 0 {
			break
		}

		for _, w := range wallets {
			wg.Add(1)
			go func(w *domain.Wallet) {
				defer wg.Done()
				idx.StreamWallet(ctx, w)
			}(w)
		}

		if len(wallets) < limit {
			break
		}
		offset += limit
	}

	wg.Wait()
	return nil
}

// StreamWallet streams new payment operations for a single wallet via Horizon
// SSE, persisting each one and advancing the stored cursor as events arrive.
// On stream failure it reconnects with exponential backoff, resuming from the
// last processed cursor, until ctx is canceled.
func (idx *Indexer) StreamWallet(ctx context.Context, w *domain.Wallet) {
	cursor := w.SyncCursor
	backoff := idx.config.StreamMinBackoff

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := stellar.StreamPaymentsWithContext(ctx, idx.stellar, w.PublicKey, cursor, func(op operations.Operation) error {
			if procErr := idx.processPayment(ctx, w, op); procErr != nil {
				log.Error().Err(procErr).Str("wallet_id", w.ID).Str("op_id", op.GetID()).
					Msg("indexer: process streamed payment failed")
			}

			cursor = op.PagingToken()
			if updErr := idx.walletRepo.UpdateSyncCursor(ctx, w.ID, cursor); updErr != nil {
				log.Error().Err(updErr).Str("wallet_id", w.ID).Msg("indexer: update sync cursor failed")
			}

			backoff = idx.config.StreamMinBackoff // connection is healthy; reset for the next disconnect
			return nil
		})

		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Error().Err(err).Str("wallet_id", w.ID).Dur("retry_in", backoff).
				Msg("indexer: payment stream disconnected, reconnecting")
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		backoff *= 2
		if backoff > idx.config.StreamMaxBackoff {
			backoff = idx.config.StreamMaxBackoff
		}
	}
}

// processPayment records an inbound payment operation as a transaction, if it
// isn't already known. Outgoing payments are skipped here since Fluxa records
// its own outbound transfers at submission time. Uses UpsertByTxHash to avoid
// TOCTOU race between ExistsByTxHash check and Create.
func (idx *Indexer) processPayment(ctx context.Context, w *domain.Wallet, op operations.Operation) error {
	if !op.IsTransactionSuccessful() {
		return nil
	}

	asset, amount, to, ok := paymentDetails(op)
	if !ok || to != w.PublicKey {
		return nil
	}

	hash := op.GetTransactionHash()

	var reference string
	horizonTx, txErr := stellar.TransactionDetailWithContext(ctx, idx.stellar, hash)
	if txErr == nil {
		if horizonTx.MemoType == "text" {
			reference = horizonTx.Memo
		} else if horizonTx.MemoType == "hash" {
			reference = horizonTx.Memo
		}
	}

	tx, err := newInboundTransaction(w.ID, w.PublicKey, hash, asset, amount, w.TenantID)
	if err != nil {
		return fmt.Errorf("build inbound transaction %s: %w", hash, err)
	}
	tx.Reference = reference

	// Use UpsertByTxHash to atomically insert only if tx_hash doesn't exist.
	// This avoids the TOCTOU race in the previous check-then-act pattern.
	if err := idx.txRepo.UpsertByTxHash(ctx, tx); err != nil {
		return fmt.Errorf("upsert transaction %s: %w", hash, err)
	}

	log.Info().Str("wallet_id", w.ID).Str("tx_hash", hash).Str("asset", asset).Str("amount", amount).
		Msg("indexer: recorded inbound payment")
	return nil
}

// paymentDetails extracts the asset code, amount, and destination account
// from a payment or path-payment operation. ok is false for any other
// operation type (e.g. trustline changes, offers).
func paymentDetails(op operations.Operation) (asset, amount, to string, ok bool) {
	switch p := op.(type) {
	case operations.PathPayment:
		return assetCode(p.Asset.Type, p.Asset.Code), p.Amount, p.To, true
	case operations.Payment:
		return assetCode(p.Asset.Type, p.Asset.Code), p.Amount, p.To, true
	default:
		return "", "", "", false
	}
}

func assetCode(assetType, code string) string {
	if assetType == "native" {
		return "XLM"
	}
	return code
}

// newInboundTransaction creates a confirmed inbound transaction with the wallet's tenant identity.
func newInboundTransaction(walletID, publicKey, txHash, asset, amount string, tenantID *string) (*domain.Transaction, error) {
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, err
	}
	return &domain.Transaction{
		ID:        uuid.New().String(),
		TxHash:    txHash,
		Type:      domain.TypeTransfer,
		Status:    domain.StatusConfirmed,
		ToWallet:  walletID,
		Asset:     asset,
		Amount:    amt,
		TenantID:  tenantID,
		CreatedAt: time.Now().UTC(),
	}, nil
}
