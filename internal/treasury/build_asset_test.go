package treasury

import (
	"testing"

	"github.com/shopspring/decimal"
)

// TestBuildAsset_RejectsUnknownCode is the regression test for the sweep path
// silently building an invalid on-chain transaction for an unrecognised asset.
func TestBuildAsset_RejectsUnknownCode(t *testing.T) {
	svc := &service{usdcIssuer: "GUSDCISSUER", eurcIssuer: "GEURCISSUER"}

	if _, err := svc.buildAsset("DOGE"); err == nil {
		t.Fatal("expected an error for an unknown asset code")
	}
	if _, err := svc.buildAsset(""); err == nil {
		t.Fatal("expected an error for an empty asset code")
	}
	if _, err := svc.buildAsset("USDC"); err != nil {
		t.Fatalf("USDC should build: %v", err)
	}
	if _, err := svc.buildAsset("XLM"); err != nil {
		t.Fatalf("XLM should build: %v", err)
	}
}

// TestBuildAsset_RejectsUnconfiguredIssuer verifies a known asset code with no
// configured issuer errors instead of producing a credit asset with an empty
// issuer.
func TestBuildAsset_RejectsUnconfiguredIssuer(t *testing.T) {
	svc := &service{usdcIssuer: "GUSDCISSUER"}

	if _, err := svc.buildAsset("EURC"); err == nil {
		t.Fatal("expected an error when the EURC issuer is not configured")
	}
}

// TestWithBaseReserve_IgnoresNonPositive keeps a misconfigured zero base
// reserve from silently zeroing every reserve figure.
func TestWithBaseReserve_IgnoresNonPositive(t *testing.T) {
	svc := &service{baseReserve: decimal.RequireFromString("0.5")}
	WithBaseReserve(decimal.Zero)(svc)
	if !svc.baseReserve.Equal(decimal.RequireFromString("0.5")) {
		t.Fatalf("base reserve = %s, want 0.5", svc.baseReserve)
	}
	WithBaseReserve(decimal.NewFromInt(2))(svc)
	if !svc.baseReserve.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("base reserve = %s, want 2", svc.baseReserve)
	}
}
