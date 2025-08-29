package time

import "time"

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

// GetChainMetadata returns cached or default metadata for a chain
func GetChainMetadata(chainHash string) *ChainMetadata {
	if meta, ok := chainMetadataCache[chainHash]; ok {
		return meta
	}
	// Default to LoE mainnet
	return &ChainMetadata{
		Genesis: PlaceholderGenesisTime,
		Period:  time.Duration(DefaultPeriodSeconds) * time.Second,
	}
}

// SetChainMetadata sets metadata for a chain
func SetChainMetadata(chainHash string, genesis time.Time, period time.Duration) {
	chainMetadataCache[chainHash] = &ChainMetadata{
		Genesis: genesis,
		Period:  period,
	}
}
