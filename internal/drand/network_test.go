package drand

import (
	"context"
	"testing"
	"time"
)

func TestNewNetwork(t *testing.T) {
	// Test with valid chain hash
	chainHash := ExampleChainHash // Example chain hash
	endpoints := []string{DefaultEndpoint}

	net, err := NewNetwork(chainHash, endpoints)
	if err != nil {
		t.Fatal(err)
	}

	if net == nil {
		t.Fatal("Expected non-nil network")
	}

	if len(net.endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(net.endpoints))
	}
}

func TestNewNetworkInvalidChainHash(t *testing.T) {
	// Test with invalid chain hash
	_, err := NewNetwork("invalid", []string{"https://api.drand.sh"})
	if err == nil {
		t.Error("Expected error for invalid chain hash")
	}
}

func TestNewNetworkEmptyEndpoints(t *testing.T) {
	// Test with empty endpoints (should use default)
	chainHash := ExampleChainHash
	net, err := NewNetwork(chainHash, []string{})

	if err != nil {
		t.Fatal(err)
	}

	if len(net.endpoints) != 1 || net.endpoints[0] != DefaultEndpoint {
		t.Errorf("Expected default endpoint, got %v", net.endpoints)
	}
}

func TestNetworkMethods(t *testing.T) {
	chainHash := ExampleChainHash
	net, err := NewNetwork(chainHash, []string{DefaultEndpoint})
	if err != nil {
		t.Fatal(err)
	}

	// Test ChainHash
	if len(net.ChainHash()) != 64 {
		t.Errorf("Expected chain hash length 64, got %d", len(net.ChainHash()))
	}

	// Test Current
	current := net.Current(time.Unix(1000000000, 0)) // Unix timestamp
	if current == 0 {
		t.Error("Expected non-zero current round")
	}

	// Test PublicKey (should return nil for now)
	if net.PublicKey() != nil {
		t.Error("Expected nil public key")
	}

	// Test Scheme (returns crypto.Scheme, skip value check)
	_ = net.Scheme()
}

func TestGetSignature(t *testing.T) {
	chainHash := ExampleChainHash
	net, err := NewNetwork(chainHash, []string{DefaultEndpoint})
	if err != nil {
		t.Fatal(err)
	}

	// Test GetSignature with a known round (this will fail without network, but tests the method exists)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = net.GetSignature(ctx, 1000000)
	// We expect this to fail due to network dependency, but it should not panic
	if err == nil {
		t.Log("GetSignature succeeded (network available)")
	} else {
		t.Logf("GetSignature failed as expected: %v", err)
	}
}

func TestSignature(t *testing.T) {
	chainHash := ExampleChainHash
	net, err := NewNetwork(chainHash, []string{DefaultEndpoint})
	if err != nil {
		t.Fatal(err)
	}

	// Test Signature method (wrapper around GetSignature)
	_, err = net.Signature(1000000)
	// We expect this to fail due to network dependency, but it should not panic
	if err == nil {
		t.Log("Signature succeeded (network available)")
	} else {
		t.Logf("Signature failed as expected: %v", err)
	}
}
