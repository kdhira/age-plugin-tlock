package tlock

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"filippo.io/age"
	"github.com/drand/tlock"
	"github.com/kdhira/age-plugin-tlock/internal/codec"
	"github.com/kdhira/age-plugin-tlock/internal/drand"
	timemap "github.com/kdhira/age-plugin-tlock/internal/time"
)

// IdentityAdapter implements age.Identity for time-lock decryption.
//
// It provides the cryptographic identity needed to decrypt time-locked
// files using drand's tlock scheme. The adapter verifies that the target
// drand round has been published before attempting decryption.
type IdentityAdapter struct {
	payload codec.IdentityPayload // Decoded identity information
	network *drand.Network        // Drand network client
}

// NewIdentityAdapter creates a new identity adapter from binary payload data.
//
// It parses the identity payload and initializes a drand network client
// for the specified chain. The adapter verifies chain information during
// initialization to ensure the identity is valid.
//
// Parameters:
//   - data: Binary identity payload from bech32 decoding
//
// Returns an age.Identity or an error if initialization fails.
//
// Example:
//
//	data := // ... decoded from bech32
//	identity, err := NewIdentityAdapter(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewIdentityAdapter(data []byte) (age.Identity, error) {
	payload, err := codec.ParseIdentityPayload(data)
	if err != nil {
		return nil, err
	}

	chainHashHex := fmt.Sprintf("%x", payload.ChainHash)
	net, err := drand.NewNetwork(chainHashHex, []string{drand.DefaultEndpoint}) // Default endpoint
	if err != nil {
		return nil, err
	}

	// Fetch chain info for proper initialization
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = net.GetChainInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to verify chain: %w", err)
	}

	return &IdentityAdapter{
		payload: payload,
		network: net,
	}, nil
}

// Unwrap implements age.Identity by decrypting time-locked file keys.
//
// It locates the tlock stanza in the provided stanzas, verifies that the
// target drand round has been published, and decrypts the file key using
// drand's tlock decryption. Returns an error if the round is not yet
// available or if decryption fails.
//
// Parameters:
//   - stanzas: Age stanzas containing the encrypted file key
//
// Returns the decrypted file key or an error if decryption fails.
//
// Example:
//
//	fileKey, err := identity.Unwrap(stanzas)
//	if err != nil {
//	    if strings.Contains(err.Error(), "not yet published") {
//	        log.Printf("File will be decryptable: %v", err)
//	        return
//	    }
//	    log.Fatal(err)
//	}
//	// Use fileKey for further decryption
func (i *IdentityAdapter) Unwrap(stanzas []*age.Stanza) ([]byte, error) {
	// Find the tlock stanza
	var tlockStanza *age.Stanza
	for _, stanza := range stanzas {
		if stanza.Type == "tlock" {
			tlockStanza = stanza
			break
		}
	}
	if tlockStanza == nil {
		return nil, fmt.Errorf("no tlock stanza found")
	}

	// Check if round is in future, return "too early" error
	round, err := parseRound(tlockStanza.Args[0])
	if err != nil {
		return nil, err
	}
	currentTime := time.Now()
	roundTime := timemap.RoundToTime(round, fmt.Sprintf("%x", i.payload.ChainHash))
	if roundTime.After(currentTime) {
		eta := timemap.GetRoundETA(round, fmt.Sprintf("%x", i.payload.ChainHash))
		return nil, fmt.Errorf("round %d not yet published: %s", round, eta)
	}

	// Create tlock instance for decryption
	tlockInstance := tlock.New(i.network)

	// Decrypt the file key using streams
	reader := bytes.NewReader(tlockStanza.Body)
	var buf bytes.Buffer
	err = tlockInstance.Decrypt(&buf, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file key: %w", err)
	}
	fileKey := buf.Bytes()

	return fileKey, nil
}

// Helper functions
func parseRound(s string) (uint64, error) {
	// Parse uint64 from string
	var round uint64
	_, err := fmt.Sscanf(s, "%d", &round)
	if err != nil {
		return 0, fmt.Errorf("invalid round format: %w", err)
	}
	return round, nil
}
