package domain

// SourceWallet is a Fluxa-custodied Stellar account usable as the funding or
// sponsor account for a claimable balance.
type SourceWallet struct {
	ID string
	// PublicKey is the Stellar account ID (G...).
	PublicKey string
	// EncryptedSecret is the hex-encoded AES-GCM envelope of the account seed,
	// exactly as stored on Wallet.
	EncryptedSecret string
}
