package codec

// Protocol constants
const (
	// CurrentVersion is the current version of the tlock payload format
	CurrentVersion = 1

	// CurrentSchemeID is the cryptographic scheme ID (BLS)
	CurrentSchemeID = 1

	// IdentityPayloadLength is the fixed length of identity payload in bytes
	IdentityPayloadLength = 83

	// RecipientPayloadLength is the fixed length of recipient payload in bytes
	RecipientPayloadLength = 90
)

// Cryptographic size constants
const (
	// ChainHashLength is the length of drand chain hash in bytes
	ChainHashLength = 32

	// PublicKeyLength is the length of drand public key in bytes
	PublicKeyLength = 48

	// HexChainHashLength is the length of chain hash in hex string (32*2)
	HexChainHashLength = 64
)

// Bit manipulation constants
const (
	// BitsPer5BitWord is the number of bits in a 5-bit word for bech32
	BitsPer5BitWord = 5

	// BitsPerByte is the number of bits in a byte
	BitsPerByte = 8

	// FiveBitMask is the mask for 5-bit values (0x1F = 31)
	FiveBitMask = 0x1F

	// ByteMask is the mask for byte values (0xFF = 255)
	ByteMask = 0xFF
)

// Payload field offsets (for identity)
const (
	// VersionOffset is the offset of version byte
	VersionOffset = 0

	// TrustChainHashOffset is the offset of trust chain hash flag
	TrustChainHashOffset = 1

	// ChainHashOffset is the offset of chain hash
	ChainHashOffset = 2

	// SchemeIDOffset is the offset of scheme ID
	SchemeIDOffset = 34

	// PublicKeyOffset is the offset of public key
	PublicKeyOffset = 35
)

// Payload field offsets (for recipient)
const (
	// RecipientVersionOffset is the offset of version byte
	RecipientVersionOffset = 0

	// RoundOffset is the offset of round number
	RoundOffset = 1

	// RecipientChainHashOffset is the offset of chain hash
	RecipientChainHashOffset = 9

	// RecipientSchemeIDOffset is the offset of scheme ID
	RecipientSchemeIDOffset = 41

	// RecipientPublicKeyOffset is the offset of public key
	RecipientPublicKeyOffset = 42
)
