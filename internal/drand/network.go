// Package drand provides client functionality for interacting with drand networks.
//
// This package implements a client for the drand distributed randomness beacon,
// providing methods to fetch chain information, signatures, and implement
// the tlock.Network interface for time-lock encryption.
//
// The Network type manages connections to drand HTTP endpoints and caches
// chain metadata for performance.
package drand

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/drand/drand/v2/crypto"
	"github.com/drand/go-clients/client"
	"github.com/drand/go-clients/client/http"
	"github.com/drand/go-clients/drand"
	"github.com/drand/kyber"
)

// ChainInfo represents information about a drand chain.
//
// It contains the cryptographic parameters and metadata needed for
// time-lock encryption, including the public key, beacon period,
// genesis time, and chain hash.
type ChainInfo struct {
	PublicKey kyber.Point   // Public key for signature verification
	Period    int           // Beacon period in seconds
	Genesis   int64         // Genesis timestamp (Unix seconds)
	Hash      string        // Chain hash as hex string
	Scheme    crypto.Scheme // Cryptographic scheme information
}

// Network represents a client for interacting with a drand network.
//
// It manages HTTP connections to drand endpoints, caches chain information,
// and implements the tlock.Network interface for time-lock encryption.
// Thread-safe for concurrent use.
type Network struct {
	endpoints []string     // List of drand HTTP endpoints
	chainHash []byte       // Chain hash as raw bytes
	client    drand.Client // Underlying drand client
	chainInfo *ChainInfo   // Cached chain information
	mu        sync.RWMutex // Protects chainInfo
}

// NewNetwork creates a new client for the specified drand chain.
//
// It establishes connections to the provided endpoints and validates
// the chain hash format. If no endpoints are provided, it uses the
// default drand mainnet endpoint.
//
// Parameters:
//   - chainHash: Hex-encoded drand chain hash (64 characters)
//   - endpoints: List of drand HTTP API endpoints (optional)
//
// Returns a configured Network client or an error if initialization fails.
//
// Example:
//
//	network, err := NewNetwork("52db9ba70e0cc0f6...", []string{"https://api.drand.sh"})
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewNetwork(chainHash string, endpoints []string) (*Network, error) {
	if len(endpoints) == 0 {
		endpoints = []string{DefaultEndpoint}
	}

	chainHashBytes, err := hex.DecodeString(chainHash)
	if err != nil {
		return nil, fmt.Errorf("invalid chain hash: %w", err)
	}

	c, err := client.New(
		client.From(http.ForURLs(context.Background(), nil, endpoints, chainHashBytes)...),
		client.WithChainHash(chainHashBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return &Network{
		endpoints: endpoints,
		chainHash: chainHashBytes,
		client:    c,
	}, nil
}

// GetChainInfo fetches and caches information about the drand chain.
//
// It retrieves chain metadata including the public key, beacon period,
// genesis time, and cryptographic scheme. Results are cached to avoid
// redundant network requests.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//
// Returns chain information or an error if the request fails.
//
// Example:
//
//	info, err := network.GetChainInfo(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Chain period: %d seconds\n", info.Period)
func (n *Network) GetChainInfo(ctx context.Context) (*ChainInfo, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.chainInfo != nil {
		return n.chainInfo, nil
	}

	// Get chain information from the drand client
	chainInfo, err := n.client.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain info: %w", err)
	}

	// Use the public key from drand
	publicKey := chainInfo.PublicKey

	// Get the cryptographic scheme
	scheme, err := crypto.SchemeFromName(chainInfo.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheme from name %q: %w", chainInfo.Scheme, err)
	}

	// Extract chain information
	info := &ChainInfo{
		Hash:      hex.EncodeToString(n.chainHash),
		PublicKey: publicKey,
		Period:    int(chainInfo.Period.Seconds()),
		Genesis:   chainInfo.GenesisTime,
		Scheme:    *scheme,
	}

	n.chainInfo = info
	return info, nil
}

// GetSignature fetches the BLS signature for the specified drand round.
//
// It retrieves the cryptographic signature that can be used for
// time-lock encryption verification. The signature corresponds to
// the randomness generated at the given round.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - round: Drand round number
//
// Returns the signature bytes or an error if the request fails.
//
// Example:
//
//	sig, err := network.GetSignature(context.Background(), 1000000)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Signature length: %d bytes\n", len(sig))
func (n *Network) GetSignature(ctx context.Context, round uint64) ([]byte, error) {
	result, err := n.client.Get(ctx, round)
	if err != nil {
		return nil, fmt.Errorf("failed to get signature: %w", err)
	}

	return result.GetSignature(), nil
}

// Implement tlock.Network interface

// ChainHash returns the chain hash as hex string
func (n *Network) ChainHash() string {
	return hex.EncodeToString(n.chainHash)
}

// Current returns the current round number for the given time
func (n *Network) Current(t time.Time) uint64 {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Use cached chain info if available
	if n.chainInfo != nil {
		genesis := time.Unix(n.chainInfo.Genesis, 0)
		period := time.Duration(n.chainInfo.Period) * time.Second
		if t.After(genesis) {
			duration := t.Sub(genesis)
			return uint64(duration / period)
		}
		return 0
	}

	// Fallback to simplified calculation
	unix := t.Unix()
	if unix < 0 {
		return 0
	}
	//nolint:gosec
	return uint64(unix / SimplifiedPeriodSeconds) // Assuming 30 second periods
}

// PublicKey returns the public key for the chain
func (n *Network) PublicKey() kyber.Point {
	if n.chainInfo == nil {
		return nil
	}
	return n.chainInfo.PublicKey
}

// Scheme returns the cryptographic scheme
func (n *Network) Scheme() crypto.Scheme {
	if n.chainInfo == nil {
		return crypto.Scheme{}
	}
	return n.chainInfo.Scheme
}

// Signature returns the signature for the given round
func (n *Network) Signature(roundNumber uint64) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return n.GetSignature(ctx, roundNumber)
}

// SwitchChainHash switches to a new chain hash
func (n *Network) SwitchChainHash(chainHash string) error {
	chainHashBytes, err := hex.DecodeString(chainHash)
	if err != nil {
		return fmt.Errorf("invalid chain hash: %w", err)
	}
	n.chainHash = chainHashBytes
	// Reset chain info to force re-fetch
	n.chainInfo = nil
	return nil
}
