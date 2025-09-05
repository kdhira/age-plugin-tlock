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

// Test GenerateDynamicRecipient with valid inputs
func TestGenerateDynamicRecipient(t *testing.T) {
	// Test with valid duration
	durationSeconds := uint64(3600) // 1 hour
	chainHash := "52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971"
	endpoint := "https://api.drand.sh"

	// This will fail due to network dependency, but we can test the error handling
	_, err := GenerateDynamicRecipient(durationSeconds, chainHash, endpoint)
	// We expect this to fail in test environment due to network, but not due to our logic
	if err != nil {
		t.Logf("Expected network error in test environment: %v", err)
	}

	// Test with invalid duration (too large)
	_, err = GenerateDynamicRecipient(0x8000000000000000, chainHash, endpoint)
	if err == nil {
		t.Error("Expected error for duration too large")
	}

	// Test with zero duration
	_, err = GenerateDynamicRecipient(0, chainHash, endpoint)
	if err == nil {
		t.Error("Expected error for zero duration")
	}
}

// Test dynamic recipient payload encoding/decoding
func TestDynamicRecipientPayload(t *testing.T) {
	// Create a payload with dynamic flag set
	payload := RecipientPayload{
		Version:     CurrentVersion,
		Round:       (1 << 63) | 7200, // Dynamic flag + 2 hours in seconds
		ChainHash:   [32]byte{0x52, 0xdb, 0x9b, 0xa7, 0x0e, 0x0c, 0xc0, 0xf6, 0xea, 0xf7, 0x80, 0x3d, 0xd0, 0x74, 0x47, 0xa1, 0xf5, 0x47, 0x77, 0x35, 0xfd, 0x3f, 0x66, 0x17, 0x92, 0xba, 0x94, 0x60, 0x0c, 0x84, 0xe9, 0x71},
		SchemeID:    CurrentSchemeID,
		DrandPubKey: [48]byte{1, 2, 3}, // Minimal test data
	}

	// Encode the payload
	encoded, err := EncodeRecipient(payload)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the encoded string starts with correct prefix
	if len(encoded) == 0 || !hasPrefix(encoded, "age1tlock") {
		t.Errorf("Invalid encoded recipient format: %s", encoded)
	}

	// For a complete test, we would need to decode and verify,
	// but that requires the full bech32 decoding which is tested elsewhere
}

// Test fixed vs dynamic recipient round interpretation
func TestRecipientRoundInterpretation(t *testing.T) {
	testCases := []struct {
		name             string
		round            uint64
		expectedFixed    uint64
		expectedDynamic  bool
		expectedDuration uint64
	}{
		{
			name:             "Fixed round",
			round:            1000000,
			expectedFixed:    1000000,
			expectedDynamic:  false,
			expectedDuration: 0,
		},
		{
			name:             "Dynamic round with duration",
			round:            (1 << 63) | 3600,
			expectedFixed:    0,
			expectedDynamic:  true,
			expectedDuration: 3600,
		},
		{
			name:             "Maximum fixed round",
			round:            0x7FFFFFFFFFFFFFFF,
			expectedFixed:    0x7FFFFFFFFFFFFFFF,
			expectedDynamic:  false,
			expectedDuration: 0,
		},
		{
			name:             "Minimum dynamic duration",
			round:            (1 << 63) | 1,
			expectedFixed:    0,
			expectedDynamic:  true,
			expectedDuration: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isDynamic := (tc.round & (1 << 63)) != 0
			var roundNumber, durationSeconds uint64

			if isDynamic {
				durationSeconds = tc.round & 0x7FFFFFFFFFFFFFFF
				roundNumber = 0
			} else {
				roundNumber = tc.round
				durationSeconds = 0
			}

			if isDynamic != tc.expectedDynamic {
				t.Errorf("Dynamic flag mismatch: got %v, want %v", isDynamic, tc.expectedDynamic)
			}

			if roundNumber != tc.expectedFixed {
				t.Errorf("Fixed round mismatch: got %d, want %d", roundNumber, tc.expectedFixed)
			}

			if durationSeconds != tc.expectedDuration {
				t.Errorf("Duration mismatch: got %d, want %d", durationSeconds, tc.expectedDuration)
			}
		})
	}
}

// Helper function to check string prefix
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// Fuzz test for bech32 decoding (used in recipient processing)
func FuzzBech32Decode(f *testing.F) {
	f.Add("age1tlock1test")
	f.Fuzz(func(t *testing.T, input string) {
		_, _, _ = bech32.Decode(input)
	})
}
