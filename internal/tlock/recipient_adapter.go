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
	"bytes"
	"context"
	"fmt"
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
	payload codec.RecipientPayload // Decoded recipient information
	network *drand.Network         // Drand network client
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

	return &RecipientAdapter{
		payload: payload,
		network: net,
	}, nil
}

// Wrap implements age.Recipient by encrypting the file key with time-lock encryption.
//
// It uses drand's tlock to encrypt the file key such that it can only be
// decrypted after the target drand round is published. The encrypted data
// is wrapped in an age stanza with the round and chain information.
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
	// Create tlock instance
	tlockInstance := tlock.New(r.network)

	// Encrypt the file key using streams
	reader := bytes.NewReader(fileKey)
	var buf bytes.Buffer
	err := tlockInstance.Encrypt(&buf, reader, r.payload.Round)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt file key: %w", err)
	}
	encryptedData := buf.Bytes()

	// Create age stanza with tlock format
	ageStanza := &age.Stanza{
		Type: "tlock",
		Args: []string{fmt.Sprintf("%d", r.payload.Round), fmt.Sprintf("%x", r.payload.ChainHash)},
		Body: encryptedData,
	}

	return []*age.Stanza{ageStanza}, nil
}
