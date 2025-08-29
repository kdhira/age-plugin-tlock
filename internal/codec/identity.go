// Package codec provides encoding and decoding functionality for age-plugin-tlock.
//
// It handles the conversion between binary payloads and bech32-encoded strings
// for both identities and recipients, following the age plugin protocol.
// The package also includes utilities for generating identities and recipients
// from drand chain information.
package codec

import (
	"context"
	"fmt"
	"time"

	"filippo.io/age/plugin"
	"github.com/kdhira/age-plugin-tlock/internal/drand"
	timemap "github.com/kdhira/age-plugin-tlock/internal/time"
)

// IdentityPayload represents the decoded identity payload.
//
// An identity contains the cryptographic information needed to decrypt
// time-locked files, including the drand chain hash, public key, and
// trust settings.
type IdentityPayload struct {
	Version        uint8    // Protocol version
	TrustChainHash bool     // Whether to trust chain hash changes
	ChainHash      [32]byte // Drand chain hash
	SchemeID       uint8    // Cryptographic scheme identifier
	DrandPubKey    [48]byte // Drand public key for verification
}

// EncodeIdentity encodes an identity payload into a bech32-encoded string.
//
// The resulting string follows the age plugin protocol format:
// AGE-PLUGIN-TLOCK-<base64-encoded-payload>
//
// Example:
//
//	payload := IdentityPayload{...}
//	identity, err := EncodeIdentity(payload)
//	if err != nil {
//	    return err
//	}
//	fmt.Println(identity) // AGE-PLUGIN-TLOCK-...
func EncodeIdentity(payload IdentityPayload) (string, error) {
	data := make([]byte, IdentityPayloadLength)
	data[VersionOffset] = payload.Version
	if payload.TrustChainHash {
		data[TrustChainHashOffset] = 1
	} else {
		data[TrustChainHashOffset] = 0
	}
	copy(data[ChainHashOffset:ChainHashOffset+ChainHashLength], payload.ChainHash[:])
	data[SchemeIDOffset] = payload.SchemeID
	copy(data[PublicKeyOffset:PublicKeyOffset+PublicKeyLength], payload.DrandPubKey[:])

	return plugin.EncodeIdentity("tlock", data), nil
}

// ParseIdentityPayload parses binary data into an IdentityPayload.
//
// It validates the payload length and extracts the individual fields
// according to the protocol specification. Returns an error if the
// data is malformed or has incorrect length.
func ParseIdentityPayload(data []byte) (IdentityPayload, error) {
	if len(data) != IdentityPayloadLength {
		return IdentityPayload{}, fmt.Errorf("invalid payload length: expected %d, got %d", IdentityPayloadLength, len(data))
	}

	var payload IdentityPayload
	payload.Version = data[VersionOffset]
	payload.TrustChainHash = data[TrustChainHashOffset] == 1
	copy(payload.ChainHash[:], data[ChainHashOffset:ChainHashOffset+ChainHashLength])
	payload.SchemeID = data[SchemeIDOffset]
	copy(payload.DrandPubKey[:], data[PublicKeyOffset:PublicKeyOffset+PublicKeyLength])

	return payload, nil
}

// GenerateIdentity generates an identity string for the given drand chain.
//
// It fetches the chain information from the drand network, constructs an
// identity payload with the public key and chain details, and encodes it
// into a bech32 string suitable for use with the age CLI.
//
// Parameters:
//   - chainHash: Hex-encoded drand chain hash (64 characters)
//   - strict: If true, disables chain hash trust for enhanced security
//   - endpoint: Drand HTTP API endpoint URL
//
// Returns an age-compatible identity string or an error if the chain
// cannot be accessed or the hash is invalid.
//
// Example:
//
//	identity, err := GenerateIdentity("52db9ba70e0cc0f6...", false, "https://api.drand.sh")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(identity) // AGE-PLUGIN-TLOCK-...
func GenerateIdentity(chainHash string, strict bool, endpoint string) (string, error) {
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

	payload := IdentityPayload{
		Version:        CurrentVersion,
		TrustChainHash: !strict,
		ChainHash:      [ChainHashLength]byte(chainHashBytes),
		SchemeID:       CurrentSchemeID,
		DrandPubKey:    pubKey,
	}

	return EncodeIdentity(payload)
}
