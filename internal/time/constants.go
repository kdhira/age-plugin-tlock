package time

import (
	"context"
	"fmt"
	"time"

	"github.com/kdhira/age-plugin-tlock/internal/drand"
)

// Time constants for drand round calculations
const (
	// DefaultPeriodSeconds is the default period between drand rounds in seconds
	DefaultPeriodSeconds = 30

	// PlaceholderGenesisYear is the placeholder genesis year
	PlaceholderGenesisYear = 2020

	// PlaceholderGenesisMonth is the placeholder genesis month
	PlaceholderGenesisMonth = 1

	// PlaceholderGenesisDay is the placeholder genesis day
	PlaceholderGenesisDay = 1
)

// PlaceholderGenesisTime is the placeholder genesis time for LoE mainnet
var PlaceholderGenesisTime = time.Unix(1677685200, 0) // LoE mainnet genesis: 2023-03-01T12:00:00Z

// ChainMetadata holds genesis and period for a chain
type ChainMetadata struct {
	Genesis time.Time
	Period  time.Duration
}

// chainMetadataCache caches metadata per chain hash
var chainMetadataCache = make(map[string]*ChainMetadata)

// GetChainMetadata returns cached or fetched metadata for a chain
func GetChainMetadata(chainHash string) (*ChainMetadata, error) {
	if meta, ok := chainMetadataCache[chainHash]; ok {
		return meta, nil
	}

	// Attempt to fetch chain info from drand network
	network, err := drand.NewNetwork(chainHash, nil) // Uses default endpoint
	if err != nil {
		return nil, fmt.Errorf("failed to create drand network client for chain %s: %w", chainHash, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	chainInfo, err := network.GetChainInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chain info: %w", err)
	}

	// Cache the fetched metadata
	SetChainMetadata(chainHash, time.Unix(chainInfo.Genesis, 0), time.Duration(chainInfo.Period)*time.Second)
	return chainMetadataCache[chainHash], nil
}

// SetChainMetadata sets metadata for a chain
func SetChainMetadata(chainHash string, genesis time.Time, period time.Duration) {
	chainMetadataCache[chainHash] = &ChainMetadata{
		Genesis: genesis,
		Period:  period,
	}
}
