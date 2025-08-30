package tlock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"filippo.io/age"
	chain "github.com/drand/drand/v2/common"
	"github.com/drand/tlock"
	"github.com/kdhira/age-plugin-tlock/internal/codec"
	"github.com/kdhira/age-plugin-tlock/internal/drand"
)

// IdentityAdapter implements age.Identity for time-lock decryption.
//
// It provides the cryptographic identity needed to decrypt time-locked
// files using drand's tlock scheme. The adapter verifies that the target
// drand round has been published before attempting decryption.
type IdentityAdapter struct {
	payload        codec.IdentityPayload // Decoded identity information
	network        *drand.Network        // Drand network client
	trustChainhash bool
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
		payload:        payload,
		network:        net,
		trustChainhash: payload.TrustChainHash,
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
	if len(stanzas) < 1 {
		return nil, errors.New("check stanzas length: should be at least one")
	}

	invalid := ""
	for _, stanza := range stanzas {
		if stanza.Type != "tlock" {
			continue
		}

		if len(stanza.Args) != 2 {
			continue
		}

		roundNumber, err := strconv.ParseUint(stanza.Args[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse block round: %w", err)
		}

		if i.network.ChainHash() != stanza.Args[1] {
			invalid = stanza.Args[1]
			if i.trustChainhash {
				fmt.Fprintf(os.Stderr, "WARN: stanza using different chainhash '%s', trying to use it instead.\n", invalid)
				err = i.network.SwitchChainHash(invalid)
				if err != nil {
					continue
				}
			} else {
				continue
			}
		}

		ciphertext, err := tlock.BytesToCiphertext(i.network.Scheme(), stanza.Body)
		if err != nil {
			return nil, fmt.Errorf("parse cipher dek: %w", err)
		}

		signature, err := i.network.Signature(roundNumber)
		if err != nil {
			// return nil, fmt.Errorf(
			// 	"%w: expected round %d > %d current round",
			// 	tlock.ErrTooEarly,
			// 	roundNumber,
			// 	i.network.Current(time.Now()))
			return nil, age.ErrIncorrectIdentity
		}

		beacon := chain.Beacon{
			Round:     roundNumber,
			Signature: signature,
		}

		fileKey, err := tlock.TimeUnlock(i.network.Scheme(), i.network.PublicKey(), beacon, ciphertext)
		if err != nil {
			return nil, fmt.Errorf("decrypt dek: %w", err)
		}

		return fileKey, nil
	}

	if len(invalid) > 0 {
		return nil, fmt.Errorf("%w: current network uses %s != %s the ciphertext requires.\n"+
			"Note that is might have been encrypted using our testnet instead", tlock.ErrWrongChainhash, i.network.ChainHash(), invalid)
	}

	return nil, fmt.Errorf("check stanza type: wrong type: %w", age.ErrIncorrectIdentity)
}

func parseRound(s string) (uint64, error) {
	// Parse uint64 from string
	var round uint64
	_, err := fmt.Sscanf(s, "%d", &round)
	if err != nil {
		return 0, fmt.Errorf("invalid round format: %w", err)
	}
	return round, nil
}
