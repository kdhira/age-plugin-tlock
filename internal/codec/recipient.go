package codec

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/kdhira/age-plugin-tlock/internal/drand"
	timemap "github.com/kdhira/age-plugin-tlock/internal/time"
)

// RecipientPayload represents the decoded recipient payload.
//
// A recipient contains the information needed to encrypt files for
// time-locked decryption, including the target drand round, chain hash,
// and public key for verification.
type RecipientPayload struct {
	Version     uint8    // Protocol version
	Round       uint64   // Target drand round for decryption
	ChainHash   [32]byte // Drand chain hash
	SchemeID    uint8    // Cryptographic scheme identifier
	DrandPubKey [48]byte // Drand public key for verification
}

// EncodeRecipient encodes a recipient payload into a bech32-encoded string.
//
// The resulting string follows the age plugin protocol format:
// age1tlock<base64-encoded-payload>
//
// This function converts the binary payload to 5-bit words and encodes
// it using the bech32 algorithm with the "age1tlock" human-readable part.
//
// Example:
//
//	payload := RecipientPayload{Round: 1000000, ...}
//	recipient, err := EncodeRecipient(payload)
//	if err != nil {
//	    return err
//	}
//	fmt.Println(recipient) // age1tlock...
func EncodeRecipient(payload RecipientPayload) (string, error) {
	data := make([]byte, RecipientPayloadLength)
	data[RecipientVersionOffset] = payload.Version
	binary.BigEndian.PutUint64(data[RoundOffset:RoundOffset+8], payload.Round)
	copy(data[RecipientChainHashOffset:RecipientChainHashOffset+ChainHashLength], payload.ChainHash[:])
	data[RecipientSchemeIDOffset] = payload.SchemeID
	copy(data[RecipientPublicKeyOffset:RecipientPublicKeyOffset+PublicKeyLength], payload.DrandPubKey[:])

	// Convert binary data to 5-bit words for bech32
	fiveBitData := convertTo5Bit(data)

	encoded, err := bech32.Encode("age1tlock", fiveBitData)
	if err != nil {
		return "", err
	}
	return encoded, nil
}

// convertTo5Bit converts 8-bit data to 5-bit words
func convertTo5Bit(data []byte) []byte {
	result := make([]byte, 0, (len(data)*BitsPerByte+4)/BitsPer5BitWord)
	bitBuffer := uint64(0)
	bitCount := 0

	for _, b := range data {
		bitBuffer = (bitBuffer << BitsPerByte) | uint64(b)
		bitCount += BitsPerByte

		for bitCount >= BitsPer5BitWord {
			bitCount -= BitsPer5BitWord
			//nolint:gosec
			word := byte((bitBuffer >> uint(bitCount)) & FiveBitMask)
			result = append(result, word)
		}
	}

	// Handle remaining bits
	if bitCount > 0 {
		//nolint:gosec
		word := byte((bitBuffer << uint(BitsPer5BitWord-bitCount)) & FiveBitMask)
		result = append(result, word)
	}

	return result
}

// convertFrom5Bit converts 5-bit words back to 8-bit data
func convertFrom5Bit(fiveBitData []byte) []byte {
	result := make([]byte, 0, (len(fiveBitData)*BitsPer5BitWord+7)/BitsPerByte)
	bitBuffer := uint64(0)
	bitCount := 0

	for _, word := range fiveBitData {
		bitBuffer = (bitBuffer << BitsPer5BitWord) | uint64(word)
		bitCount += BitsPer5BitWord

		for bitCount >= BitsPerByte {
			bitCount -= BitsPerByte
			//nolint:gosec
			b := byte((bitBuffer >> uint(bitCount)) & ByteMask)
			result = append(result, b)
		}
	}

	return result
}

// ParseRecipientPayload parses binary data into RecipientPayload
func ParseRecipientPayload(data []byte) (RecipientPayload, error) {
	if len(data) != RecipientPayloadLength {
		return RecipientPayload{}, fmt.Errorf("invalid payload length: expected %d, got %d", RecipientPayloadLength, len(data))
	}

	var payload RecipientPayload
	payload.Version = data[RecipientVersionOffset]
	payload.Round = binary.BigEndian.Uint64(data[RoundOffset : RoundOffset+8])
	copy(payload.ChainHash[:], data[RecipientChainHashOffset:RecipientChainHashOffset+ChainHashLength])
	payload.SchemeID = data[RecipientSchemeIDOffset]
	copy(payload.DrandPubKey[:], data[RecipientPublicKeyOffset:RecipientPublicKeyOffset+PublicKeyLength])

	return payload, nil
}

// GenerateRecipient generates a recipient string for the specified drand round and chain.
//
// It fetches the chain information from the drand network, constructs a
// recipient payload with the target round and public key, and encodes it
// into a bech32 string suitable for use with the age CLI.
//
// Parameters:
//   - round: Target drand round number for decryption
//   - chainHash: Hex-encoded drand chain hash (64 characters)
//   - endpoint: Drand HTTP API endpoint URL
//
// Returns an age-compatible recipient string or an error if the chain
// cannot be accessed or parameters are invalid.
//
// Example:
//
//	recipient, err := GenerateRecipient(1000000, "52db9ba70e0cc0f6...", "https://api.drand.sh")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(recipient) // age1tlock...
func GenerateRecipient(round uint64, chainHash string, endpoint string) (string, error) {
	// Decode chainHash from hex
	chainHashBytes := make([]byte, ChainHashLength)
	if len(chainHash) != HexChainHashLength {
		return "", ValidationError{
			Field:   "chain hash",
			Value:   chainHash,
			Message: fmt.Sprintf("must be %d hex characters", HexChainHashLength),
		}
	}
	for i := 0; i < ChainHashLength; i++ {
		if b, err := fmt.Sscanf(chainHash[i*2:i*2+2], "%02x", &chainHashBytes[i]); err != nil || b != 1 {
			return "", ValidationError{
				Field:   "chain hash",
				Value:   chainHash,
				Message: "contains invalid hex characters",
			}
		}
	}

	// Get real public key from drand client
	net, err := drand.NewNetwork(chainHash, []string{endpoint})
	if err != nil {
		return "", NetworkError{
			Endpoint:  endpoint,
			Operation: "create network client",
			Cause:     err,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var info *drand.ChainInfo
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		info, err = net.GetChainInfo(ctx)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	if err != nil {
		return "", fmt.Errorf("failed to get chain info after retries: %w", err)
	}

	var pubKey [PublicKeyLength]byte
	pubKeyBytes, err := info.PublicKey.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}
	copy(pubKey[:], pubKeyBytes[:PublicKeyLength])

	// Cache chain metadata for time calculations
	genesis := time.Unix(info.Genesis, 0)
	period := time.Duration(info.Period) * time.Second
	timemap.SetChainMetadata(chainHash, genesis, period)

	payload := RecipientPayload{
		Version:     CurrentVersion,
		Round:       round,
		ChainHash:   [ChainHashLength]byte(chainHashBytes),
		SchemeID:    CurrentSchemeID, // BLS
		DrandPubKey: pubKey,
	}

	return EncodeRecipient(payload)
}
