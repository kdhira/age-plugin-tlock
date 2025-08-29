package tlock

import (
	"testing"

	"github.com/kdhira/age-plugin-tlock/internal/codec"
)

func TestNewIdentityAdapter(t *testing.T) {
	// Create test payload data (binary format)
	data := make([]byte, codec.IdentityPayloadLength)
	data[codec.VersionOffset] = codec.CurrentVersion // Version
	data[codec.TrustChainHashOffset] = 1             // TrustChainHash = true
	// ChainHash - use valid 32-byte binary
	validChainHash := []byte{0x52, 0xdb, 0x9b, 0xa7, 0x0e, 0x0c, 0xc0, 0xf6, 0xea, 0xf7, 0x80, 0x3d, 0xd0, 0x74, 0x47, 0xa1, 0xf5, 0x47, 0x77, 0x35, 0xfd, 0x3f, 0x66, 0x17, 0x92, 0xba, 0x94, 0x60, 0x0c, 0x84, 0xe9, 0x71}
	copy(data[codec.ChainHashOffset:codec.ChainHashOffset+codec.ChainHashLength], validChainHash)
	// SchemeID
	data[codec.SchemeIDOffset] = codec.CurrentSchemeID
	// DrandPubKey
	copy(data[codec.PublicKeyOffset:codec.PublicKeyOffset+codec.PublicKeyLength], []byte{4, 5, 6})

	adapter, err := NewIdentityAdapter(data)
	if err != nil {
		t.Fatal(err)
	}

	if adapter == nil {
		t.Error("Expected non-nil adapter")
	}
}

func TestIdentityAdapterUnwrap(t *testing.T) {
	t.Skip("Skipping test due to network dependency and tlock library integration issues")
}

func TestParseRound(t *testing.T) {
	round, err := parseRound("12345")
	if err != nil {
		t.Fatal(err)
	}
	if round != 12345 {
		t.Errorf("Expected 12345, got %d", round)
	}

	round, err = parseRound("0")
	if err != nil {
		t.Fatal(err)
	}
	if round != 0 {
		t.Errorf("Expected 0, got %d", round)
	}
}
