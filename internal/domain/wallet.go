package domain

import (
	"time"
)

// CustodyType distinguishes custodial wallets (Fluxa holds the encrypted
// secret) from contract wallets (a Soroban contract holds the funds and
// enforces spending policy on-chain).
type CustodyType string

const (
	CustodyCustodial CustodyType = "custodial"
	CustodyContract  CustodyType = "contract"
)

type Wallet struct {
	ID              string      `json:"id"`
	TenantID        *string     `json:"tenant_id,omitempty"`
	PublicKey       string      `json:"public_key"`
	EncryptedSecret string      `json:"-"`
	SyncCursor      string      `json:"-"`
	CustodyType     CustodyType `json:"custody_type"`
	ContractID      string      `json:"contract_id,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
}

// BalanceRecord is one cached on-chain balance for a wallet/asset pair.
type BalanceRecord struct {
	WalletID  string    `json:"wallet_id"`
	AssetCode string    `json:"asset_code"`
	Issuer    string    `json:"issuer"`
	Balance   string    `json:"balance"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Balance struct {
	AssetCode   string `json:"asset_code"`
	AssetIssuer string `json:"asset_issuer,omitempty"`
	Balance     string `json:"balance"`
	Limit       string `json:"limit,omitempty"`
}

type WalletBalance struct {
	WalletID  string    `json:"wallet_id"`
	Balances  []Balance `json:"balances"`
	Stale     bool      `json:"stale"`
	UpdatedAt time.Time `json:"updated_at"`
}
