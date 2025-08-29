package codec

import (
	"testing"
)

// BenchmarkEncodeIdentity benchmarks the identity encoding performance
func BenchmarkEncodeIdentity(b *testing.B) {
	payload := IdentityPayload{
		Version:        CurrentVersion,
		TrustChainHash: true,
		ChainHash:      [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
		SchemeID:       CurrentSchemeID,
		DrandPubKey:    [48]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EncodeIdentity(payload)
	}
}

// BenchmarkParseIdentityPayload benchmarks the identity parsing performance
func BenchmarkParseIdentityPayload(b *testing.B) {
	data := make([]byte, IdentityPayloadLength)
	data[VersionOffset] = CurrentVersion
	data[TrustChainHashOffset] = 1
	copy(data[ChainHashOffset:ChainHashOffset+ChainHashLength], []byte{1, 2, 3})
	data[SchemeIDOffset] = CurrentSchemeID
	copy(data[PublicKeyOffset:PublicKeyOffset+PublicKeyLength], []byte{4, 5, 6})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseIdentityPayload(data)
	}
}
