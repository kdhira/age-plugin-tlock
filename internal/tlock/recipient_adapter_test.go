package tlock

import (
	"testing"

	"github.com/kdhira/age-plugin-tlock/internal/codec"
)

func TestNewRecipientAdapter(t *testing.T) {
	// Create test payload data (binary format)
	data := make([]byte, codec.RecipientPayloadLength)
	data[codec.RecipientVersionOffset] = codec.CurrentVersion // Version
	// Round = 1000 (big endian)
	data[1] = 0
	data[2] = 0
	data[3] = 0
	data[4] = 0
	data[5] = 0
	data[6] = 0
	data[7] = 0
	data[8] = 0
	data[9] = 0
	data[10] = 0
	// ChainHash - use valid 32-byte binary
	validChainHash := []byte{0x52, 0xdb, 0x9b, 0xa7, 0x0e, 0x0c, 0xc0, 0xf6, 0xea, 0xf7, 0x80, 0x3d, 0xd0, 0x74, 0x47, 0xa1, 0xf5, 0x47, 0x77, 0x35, 0xfd, 0x3f, 0x66, 0x17, 0x92, 0xba, 0x94, 0x60, 0x0c, 0x84, 0xe9, 0x71}
	copy(data[codec.RecipientChainHashOffset:codec.RecipientChainHashOffset+codec.ChainHashLength], validChainHash)
	// SchemeID
	data[codec.RecipientSchemeIDOffset] = codec.CurrentSchemeID
	// DrandPubKey
	copy(data[codec.RecipientPublicKeyOffset:codec.RecipientPublicKeyOffset+codec.PublicKeyLength], []byte{4, 5, 6})

	adapter, err := NewRecipientAdapter(data)
	if err != nil {
		t.Fatal(err)
	}

	if adapter == nil {
		t.Error("Expected non-nil adapter")
	}
}

func TestRecipientAdapterWrap(t *testing.T) {
	t.Skip("Skipping test due to network dependency on drand API and tlock library integration issues. " +
		"This test requires a live drand network connection for chain verification and tlock encryption. " +
		"Consider using mocks for the drand client or integration tests in a CI environment with network access.")
}
