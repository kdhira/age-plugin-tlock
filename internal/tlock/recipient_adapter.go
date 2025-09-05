// Package tlock provides age plugin adapters for time-lock encryption.
//
// This package implements the age.Recipient and age.Identity interfaces
// using drand's tlock cryptographic scheme. It enables encryption of data
// that can only be decrypted after a specific future drand round.
//
// The adapters handle the conversion between age's protocol and tlock's
// stream-based encryption/decryption operations.
package tlock

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"filippo.io/age"
	"github.com/drand/tlock"
	"github.com/kdhira/age-plugin-tlock/internal/codec"
	"github.com/kdhira/age-plugin-tlock/internal/drand"
)

// RecipientAdapter implements age.Recipient for time-lock encryption.
//
// It wraps drand's tlock encryption functionality to provide age-compatible
// recipient operations. The adapter encrypts file keys using the specified
// drand round and chain, creating stanzas that can only be decrypted
// after the target round is published.
type RecipientAdapter struct {
	payload         codec.RecipientPayload // Decoded recipient information
	roundNumber     uint64
	isDynamic       bool           // Whether to calculate round dynamically
	durationSeconds uint64         // Duration in seconds for dynamic mode
	network         *drand.Network // Drand network client
}

// NewRecipientAdapter creates a new recipient adapter from binary payload data.
//
// It parses the recipient payload and initializes a drand network client
// for the specified chain. The adapter is ready to encrypt file keys
// for time-locked decryption.
//
// Parameters:
//   - data: Binary recipient payload from bech32 decoding
//
// Returns an age.Recipient or an error if initialization fails.
//
// Example:
//
//	data := // ... decoded from bech32
//	recipient, err := NewRecipientAdapter(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewRecipientAdapter(data []byte) (age.Recipient, error) {
	payload, err := codec.ParseRecipientPayload(data)
	if err != nil {
		return nil, err
	}

	// For offline encryption, network is not required for Wrap
	// But create network for potential verification
	chainHashHex := fmt.Sprintf("%x", payload.ChainHash)
	net, err := drand.NewNetwork(chainHashHex, []string{drand.DefaultEndpoint}) // Default endpoint
	if err != nil {
		return nil, err
	}

	// Optional: Verify chain hash by fetching info
	// For strict verification, uncomment below
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = net.GetChainInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to verify chain: %w", err)
	}

	// Check if this is a dynamic recipient (high bit of Round is set)
	isDynamic := (payload.Round & (1 << 63)) != 0
	var roundNumber, durationSeconds uint64
	if isDynamic {
		durationSeconds = payload.Round & 0x7FFFFFFFFFFFFFFF // Clear high bit
		roundNumber = 0                                      // Will be calculated dynamically
	} else {
		roundNumber = payload.Round
		durationSeconds = 0
	}

	return &RecipientAdapter{
		network:         net,
		roundNumber:     roundNumber,
		isDynamic:       isDynamic,
		durationSeconds: durationSeconds,
		payload:         payload,
	}, nil
}

// Wrap implements age.Recipient by encrypting the file key with time-lock encryption.
//
// It uses drand's tlock to encrypt the file key such that it can only be
// decrypted after the target drand round is published. The encrypted data
// is wrapped in an age stanza with the round and chain information.
//
// For dynamic recipients, the target round is calculated at encryption time
// based on the current round plus the specified duration.
//
// Parameters:
//   - fileKey: The symmetric file key to encrypt (32 bytes)
//
// Returns age stanzas containing the encrypted file key, or an error
// if encryption fails.
//
// Example:
//
//	stanzas, err := recipient.Wrap(fileKey)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// stanzas[0] contains the tlock-encrypted file key
func (r *RecipientAdapter) Wrap(fileKey []byte) ([]*age.Stanza, error) {
	targetRound := r.roundNumber

	if r.isDynamic {
		// Calculate target round dynamically
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Ensure chain info is loaded
		chainInfo, err := r.network.GetChainInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get chain info for dynamic round: %w", err)
		}

		// Get current round
		currentRound := r.network.Current(time.Now())

		// Calculate duration in rounds
		periodSeconds := uint64(chainInfo.Period)
		durationRounds := r.durationSeconds / periodSeconds
		if r.durationSeconds%periodSeconds != 0 {
			durationRounds++ // Round up
		}

		targetRound = currentRound + durationRounds
	}

	ciphertext, err := tlock.TimeLock(r.network.Scheme(), r.network.PublicKey(), targetRound, fileKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt dek: %w", err)
	}

	body, err := tlock.CiphertextToBytes(r.network.Scheme(), ciphertext)
	if err != nil {
		return nil, fmt.Errorf("bytes: %w", err)
	}

	stanza := age.Stanza{
		Type: "tlock",
		Args: []string{strconv.FormatUint(targetRound, 10), r.network.ChainHash()},
		Body: body,
	}

	return []*age.Stanza{&stanza}, nil
}
