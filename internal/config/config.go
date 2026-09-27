package config

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port                        string
	CORSAllowedOrigins          []string
	Env                         string
	LogLevel                    string
	DatabaseURL                 string
	ReplicaDatabaseURL          string
	RedisURL                    string
	RedisSentinelMasterName     string
	RedisSentinelAddrs          []string
	RedisSentinelPassword       string
	StellarNetwork              string
	StellarHorizonURL           string
	StellarUSDCIssuer           string
	StellarEURCIssuer           string
	MasterEncryptionKey         []byte
	TreasurySecretKey           string
	PlatformFeeWalletPublicKey  string
	ColdStorageAddress          string
	MigrationsPath              string
	AlertWebhookURL             string
	PlatformWalletID            string
	FlutterwaveSecretKey        string
	FlutterwaveWebhookHash      string
	BalanceDiscrepancyThreshold string
	JWTSecret                   string
	FXSpreadBps                 int
	SorobanRPCURL               string
	ContractWalletWasmHash      string
	ContractWalletSpendingLimit string
	ContractWalletWindowSeconds int
	ContractWalletRecoveryQuota int
	YellowCardAPIKey            string
	YellowCardWebhookKey        string
	YellowCardSandbox           bool
	ComplianceEnabled           bool
	OFACSDNURL                  string
	ComplianceStructuringUnit   string
	ComplianceVelocityMax       int
	ComplianceVelocityWindowMin int
	ComplianceRoundTripMin      int
	ComplianceFuzzyThreshold    int
	ComplianceReloadMinutes     int
	WorkerEnabled               bool
	WebhookAllowPrivateNetworks bool
	// ClaimableBalanceSourceWalletID funds claimable balances whose request did
	// not name a source wallet.
	ClaimableBalanceSourceWalletID string

	// Indexer configuration
	IndexerPaymentsPageLimit int
	IndexerStreamMinBackoff  string
	IndexerStreamMaxBackoff  string
	IndexerSyncPageSize      int
}

func splitCSV(value string) []string {
	parts := []string{}
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func Load() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "3000")
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "localhost:*")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("STELLAR_NETWORK", "testnet")
	viper.SetDefault("STELLAR_HORIZON_URL", "https://horizon-testnet.stellar.org")
	viper.SetDefault("MIGRATIONS_PATH", "db/migrations")
	viper.SetDefault("FX_SPREAD_BPS", "50")
	viper.SetDefault("JWT_SECRET", "fluxa-default-jwt-secret-key-change-in-production")
	viper.SetDefault("SOROBAN_RPC_URL", "https://soroban-testnet.stellar.org")
	viper.SetDefault("CONTRACT_WALLET_SPENDING_LIMIT", "1000")
	viper.SetDefault("CONTRACT_WALLET_WINDOW_SECONDS", "86400")
	viper.SetDefault("CONTRACT_WALLET_RECOVERY_THRESHOLD", "2")
	viper.SetDefault("YELLOW_CARD_SANDBOX", "true")
	viper.SetDefault("COMPLIANCE_ENABLED", "true")
	viper.SetDefault("OFAC_SDN_URL", "https://sanctionslistservice.ofac.treas.gov/api/download/sdn.xml")
	viper.SetDefault("COMPLIANCE_STRUCTURING_UNIT", "1000")
	viper.SetDefault("COMPLIANCE_VELOCITY_MAX_TRANSFERS", "10")
	viper.SetDefault("COMPLIANCE_VELOCITY_WINDOW_MINUTES", "10")
	viper.SetDefault("COMPLIANCE_ROUND_TRIP_WINDOW_MINUTES", "60")
	viper.SetDefault("COMPLIANCE_FUZZY_THRESHOLD", "2")
	viper.SetDefault("COMPLIANCE_RELOAD_MINUTES", "15")
	viper.SetDefault("WORKER_ENABLED", "true")
	viper.SetDefault("WEBHOOK_ALLOW_PRIVATE_NETWORKS", "false")
	viper.SetDefault("INDEXER_PAYMENTS_PAGE_LIMIT", "50")
	viper.SetDefault("INDEXER_STREAM_MIN_BACKOFF", "1s")
	viper.SetDefault("INDEXER_STREAM_MAX_BACKOFF", "30s")
	viper.SetDefault("INDEXER_SYNC_PAGE_SIZE", "100")

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig()

	required := []string{"DATABASE_URL", "REDIS_URL", "MASTER_ENCRYPTION_KEY"}
	for _, key := range required {
		if viper.GetString(key) == "" {
			return nil, fmt.Errorf("required env var %s is not set", key)
		}
	}

	keyHex := viper.GetString("MASTER_ENCRYPTION_KEY")
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("MASTER_ENCRYPTION_KEY must be a valid hex string: %w", err)
	}
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("MASTER_ENCRYPTION_KEY must be 32 bytes (64 hex chars), got %d bytes", len(keyBytes))
	}
	if err := validateKeyEntropy(keyBytes); err != nil {
		return nil, fmt.Errorf("MASTER_ENCRYPTION_KEY entropy check failed: %w", err)
	}

	env := viper.GetString("ENV")
	jwtSecret := viper.GetString("JWT_SECRET")
	if env == "production" {
		if jwtSecret == "fluxa-default-jwt-secret-key-change-in-production" || len(jwtSecret) < 32 {
			return nil, fmt.Errorf("a secure, high-entropy JWT_SECRET (min 32 bytes) must be explicitly configured in production")
		}
	}

	ycSandbox, _ := strconv.ParseBool(viper.GetString("YELLOW_CARD_SANDBOX"))
	complianceEnabled, _ := strconv.ParseBool(viper.GetString("COMPLIANCE_ENABLED"))
	workerEnabled, _ := strconv.ParseBool(viper.GetString("WORKER_ENABLED"))
	webhookAllowPrivateNetworks, _ := strconv.ParseBool(viper.GetString("WEBHOOK_ALLOW_PRIVATE_NETWORKS"))

	indexerPaymentsPageLimit := viper.GetInt("INDEXER_PAYMENTS_PAGE_LIMIT")
	indexerStreamMinBackoff := viper.GetString("INDEXER_STREAM_MIN_BACKOFF")
	indexerStreamMaxBackoff := viper.GetString("INDEXER_STREAM_MAX_BACKOFF")
	indexerSyncPageSize := viper.GetInt("INDEXER_SYNC_PAGE_SIZE")

	if webhookAllowPrivateNetworks && env != "development" {
		return nil, fmt.Errorf("WEBHOOK_ALLOW_PRIVATE_NETWORKS can only be enabled in development environment")
	}

	return &Config{
		Port:                        viper.GetString("PORT"),
		CORSAllowedOrigins:          splitCSV(viper.GetString("CORS_ALLOWED_ORIGINS")),
		Env:                         env,
		LogLevel:                    viper.GetString("LOG_LEVEL"),
		DatabaseURL:                 viper.GetString("DATABASE_URL"),
		ReplicaDatabaseURL:          viper.GetString("REPLICA_DATABASE_URL"),
		RedisURL:                    viper.GetString("REDIS_URL"),
		RedisSentinelMasterName:     viper.GetString("REDIS_SENTINEL_MASTER_NAME"),
		RedisSentinelAddrs:          splitCSV(viper.GetString("REDIS_SENTINEL_ADDRS")),
		RedisSentinelPassword:       viper.GetString("REDIS_SENTINEL_PASSWORD"),
		StellarNetwork:              viper.GetString("STELLAR_NETWORK"),
		StellarHorizonURL:           viper.GetString("STELLAR_HORIZON_URL"),
		StellarUSDCIssuer:           viper.GetString("STELLAR_USDC_ISSUER"),
		StellarEURCIssuer:           viper.GetString("STELLAR_EURC_ISSUER"),
		MasterEncryptionKey:         keyBytes,
		TreasurySecretKey:           viper.GetString("TREASURY_SECRET_KEY"),
		PlatformFeeWalletPublicKey:  viper.GetString("PLATFORM_FEE_WALLET_PUBLIC_KEY"),
		ColdStorageAddress:          viper.GetString("COLD_STORAGE_ADDRESS"),
		MigrationsPath:              viper.GetString("MIGRATIONS_PATH"),
		AlertWebhookURL:             viper.GetString("ALERT_WEBHOOK_URL"),
		PlatformWalletID:            viper.GetString("PLATFORM_WALLET_ID"),
		FlutterwaveSecretKey:        viper.GetString("FLUTTERWAVE_SECRET_KEY"),
		FlutterwaveWebhookHash:      viper.GetString("FLUTTERWAVE_WEBHOOK_HASH"),
		BalanceDiscrepancyThreshold: viper.GetString("BALANCE_DISCREPANCY_THRESHOLD"),
		JWTSecret:                   viper.GetString("JWT_SECRET"),
		FXSpreadBps:                 viper.GetInt("FX_SPREAD_BPS"),
		SorobanRPCURL:               viper.GetString("SOROBAN_RPC_URL"),
		ContractWalletWasmHash:      viper.GetString("CONTRACT_WALLET_WASM_HASH"),
		ContractWalletSpendingLimit: viper.GetString("CONTRACT_WALLET_SPENDING_LIMIT"),
		ContractWalletWindowSeconds: viper.GetInt("CONTRACT_WALLET_WINDOW_SECONDS"),
		ContractWalletRecoveryQuota: viper.GetInt("CONTRACT_WALLET_RECOVERY_THRESHOLD"),
		YellowCardAPIKey:            viper.GetString("YELLOW_CARD_API_KEY"),
		YellowCardWebhookKey:        viper.GetString("YELLOW_CARD_WEBHOOK_KEY"),
		YellowCardSandbox:           ycSandbox,
		ComplianceEnabled:           complianceEnabled,
		OFACSDNURL:                  viper.GetString("OFAC_SDN_URL"),
		ComplianceStructuringUnit:   viper.GetString("COMPLIANCE_STRUCTURING_UNIT"),
		ComplianceVelocityMax:       viper.GetInt("COMPLIANCE_VELOCITY_MAX_TRANSFERS"),
		ComplianceVelocityWindowMin: viper.GetInt("COMPLIANCE_VELOCITY_WINDOW_MINUTES"),
		ComplianceRoundTripMin:      viper.GetInt("COMPLIANCE_ROUND_TRIP_WINDOW_MINUTES"),
		ComplianceFuzzyThreshold:    viper.GetInt("COMPLIANCE_FUZZY_THRESHOLD"),
		ComplianceReloadMinutes:     viper.GetInt("COMPLIANCE_RELOAD_MINUTES"),
		WorkerEnabled:               workerEnabled,
		WebhookAllowPrivateNetworks: webhookAllowPrivateNetworks,

		ClaimableBalanceSourceWalletID: viper.GetString("CLAIMABLE_BALANCE_SOURCE_WALLET_ID"),

		IndexerPaymentsPageLimit: indexerPaymentsPageLimit,
		IndexerStreamMinBackoff:  indexerStreamMinBackoff,
		IndexerStreamMaxBackoff:  indexerStreamMaxBackoff,
		IndexerSyncPageSize:      indexerSyncPageSize,
	}, nil
}

// validateKeyEntropy checks that the encryption key has sufficient entropy.
// It uses a simple Shannon entropy estimation to reject obviously weak keys
// (e.g., all zeros, repeated patterns, or low-entropy inputs).
func validateKeyEntropy(key []byte) error {
	if len(key) == 0 {
		return fmt.Errorf("key is empty")
	}

	// Check for all zeros
	allZero := true
	for _, b := range key {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return fmt.Errorf("key cannot be all zeros")
	}

	// Check for all same byte
	allSame := true
	first := key[0]
	for _, b := range key {
		if b != first {
			allSame = false
			break
		}
	}
	if allSame {
		return fmt.Errorf("key cannot be all identical bytes")
	}

	// Calculate Shannon entropy (bits per byte)
	// For a 32-byte key, we expect entropy close to 8 bits/byte
	freq := make(map[byte]int)
	for _, b := range key {
		freq[b]++
	}

	entropy := 0.0
	for _, count := range freq {
		p := float64(count) / float64(len(key))
		entropy -= p * log2(p)
	}

	// Require at least 7.5 bits/byte entropy (out of 8 max)
	// This catches keys with obvious patterns while allowing natural randomness
	if entropy < 7.5 {
		return fmt.Errorf("key entropy too low: %.2f bits/byte (minimum 7.5)", entropy)
	}

	return nil
}

func log2(x float64) float64 {
	return math.Log2(x)
}
