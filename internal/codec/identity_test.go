package codec

import (
	"testing"
)

func TestParseIdentityPayload(t *testing.T) {
	data := make([]byte, IdentityPayloadLength)
	data[VersionOffset] = CurrentVersion
	data[TrustChainHashOffset] = 1                                               // TrustChainHash = true
	copy(data[ChainHashOffset:ChainHashOffset+ChainHashLength], []byte{1, 2, 3}) // ChainHash
	data[SchemeIDOffset] = CurrentSchemeID                                       // SchemeID
	copy(data[PublicKeyOffset:PublicKeyOffset+PublicKeyLength], []byte{4, 5, 6}) // DrandPubKey

	payload, err := ParseIdentityPayload(data)
	if err != nil {
		t.Fatal(err)
	}

	if payload.Version != CurrentVersion {
		t.Errorf("Version mismatch")
	}

	if !payload.TrustChainHash {
		t.Error("TrustChainHash should be true")
	}

	if payload.SchemeID != CurrentSchemeID {
		t.Errorf("SchemeID mismatch")
	}
}

func TestGenerateIdentity(t *testing.T) {
	// Test error cases
	_, err := GenerateIdentity("", false, "https://api.drand.sh")
	if err == nil {
		t.Error("Expected error for empty chain hash")
	}

	_, err = GenerateIdentity("invalid", false, "https://api.drand.sh")
	if err == nil {
		t.Error("Expected error for invalid chain hash")
	}

	// Test with too short chain hash
	_, err = GenerateIdentity("short", false, "https://api.drand.sh")
	if err == nil {
		t.Error("Expected error for too short chain hash")
	}

	// Test with invalid hex characters
	_, err = GenerateIdentity("gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg", false, "https://api.drand.sh")
	if err == nil {
		t.Error("Expected error for invalid hex characters")
	}
}
