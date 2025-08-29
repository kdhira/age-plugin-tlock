package codec

import (
	"testing"

	"github.com/btcsuite/btcd/btcutil/bech32"
)

func TestEncodeDecodeRecipient(t *testing.T) {
	// Create a small test payload that fits bech32 limits
	// Use only essential data for testing
	testData := []byte{1, 0, 0, 0, 0, 0, 0, 0, 123, 45} // version=1, round=12345 (big endian)

	encoded, err := bech32.Encode("age1tlock", convertTo5Bit(testData))
	if err != nil {
		t.Fatal(err)
	}

	hrp, decodedData, err := bech32.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}

	if hrp != "age1tlock" {
		t.Errorf("HRP mismatch: got %s, want age1tlock", hrp)
	}

	originalData := convertFrom5Bit(decodedData)
	if len(originalData) != len(testData) {
		t.Errorf("Data length mismatch: got %d, want %d", len(originalData), len(testData))
	}

	for i, v := range testData {
		if i < len(originalData) && originalData[i] != v {
			t.Errorf("Data mismatch at index %d: got %d, want %d", i, originalData[i], v)
		}
	}
}

func TestParseRecipientPayload(t *testing.T) {
	data := make([]byte, RecipientPayloadLength)
	data[RecipientVersionOffset] = CurrentVersion
	// Set round, chain, etc.

	payload, err := ParseRecipientPayload(data)
	if err != nil {
		t.Fatal(err)
	}

	if payload.Version != CurrentVersion {
		t.Errorf("Version mismatch")
	}
}

// Test convertTo5Bit and convertFrom5Bit
func TestConvertToFrom5Bit(t *testing.T) {
	testData := []byte{1, 2, 3, 4, 5}

	fiveBit := convertTo5Bit(testData)
	original := convertFrom5Bit(fiveBit)

	if len(original) != len(testData) {
		t.Errorf("Length mismatch: got %d, want %d", len(original), len(testData))
	}

	for i, v := range testData {
		if original[i] != v {
			t.Errorf("Data mismatch at index %d: got %d, want %d", i, original[i], v)
		}
	}
}

// Test GenerateRecipient with mock data
func TestGenerateRecipient(t *testing.T) {
	// This test would require mocking the drand client
	// For now, test error cases
	_, err := GenerateRecipient(1000, "", "https://api.drand.sh")
	if err == nil {
		t.Error("Expected error for empty chain hash")
	}

	_, err = GenerateRecipient(1000, "invalid", "https://api.drand.sh")
	if err == nil {
		t.Error("Expected error for invalid chain hash")
	}

	// Test with too short chain hash
	_, err = GenerateRecipient(1000, "short", "https://api.drand.sh")
	if err == nil {
		t.Error("Expected error for too short chain hash")
	}

	// Test with invalid hex characters
	_, err = GenerateRecipient(1000, "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg", "https://api.drand.sh")
	if err == nil {
		t.Error("Expected error for invalid hex characters")
	}
}

// Fuzz test for bech32 decoding (used in recipient processing)
func FuzzBech32Decode(f *testing.F) {
	f.Add("age1tlock1test")
	f.Fuzz(func(t *testing.T, input string) {
		_, _, _ = bech32.Decode(input)
	})
}
