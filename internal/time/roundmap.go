// Package time provides utilities for converting between times and drand round numbers.
//
// This package handles the mapping between wall-clock time and drand beacon rounds,
// taking into account chain-specific genesis times and beacon periods. It provides
// functions for both directions of conversion and ETA calculations.
package time

import (
	"fmt"
	"time"
)

// ToRound converts a wall-clock time to the corresponding drand round number.
//
// It calculates which drand round would be published at or after the given time,
// based on the chain's genesis time and beacon period. Times before genesis
// return round 0.
//
// Parameters:
//   - t: The time to convert
//   - chainHash: Hex-encoded drand chain hash for metadata lookup
//
// Returns the drand round number that covers the given time.
//
// Example:
//
//	round := ToRound(time.Now(), "52db9ba70e0cc0f6...")
//	fmt.Printf("Current round: %d\n", round)
func ToRound(t time.Time, chainHash string) uint64 {
	meta := GetChainMetadata(chainHash)
	genesis := meta.Genesis
	period := meta.Period

	if t.Before(genesis) {
		return 0
	}

	duration := t.Sub(genesis)
	//nolint:gosec
	round := uint64(duration / period)
	return round
}

// RoundToTime converts a drand round number to the approximate publication time.
//
// It calculates when the specified drand round was or will be published,
// based on the chain's genesis time and beacon period.
//
// Parameters:
//   - round: Drand round number
//   - chainHash: Hex-encoded drand chain hash for metadata lookup
//
// Returns the estimated publication time for the round.
//
// Example:
//
//	pubTime := RoundToTime(1000000, "52db9ba70e0cc0f6...")
//	fmt.Printf("Round published at: %s\n", pubTime.Format(time.RFC3339))
func RoundToTime(round uint64, chainHash string) time.Time {
	meta := GetChainMetadata(chainHash)
	genesis := meta.Genesis
	period := meta.Period

	//nolint:gosec
	return genesis.Add(time.Duration(round) * period)
}

// GetRoundETA returns a human-readable estimate of when a drand round will be published.
//
// It calculates the time remaining until the specified round becomes available,
// or indicates when it was published if it's in the past.
//
// Parameters:
//   - round: Drand round number
//   - chainHash: Hex-encoded drand chain hash for metadata lookup
//
// Returns a formatted string describing the round's status and timing.
//
// Example:
//
//	eta := GetRoundETA(1000000, "52db9ba70e0cc0f6...")
//	fmt.Println(eta) // "round 1000000 expected at 2024-01-01T12:00:00Z (in 2h30m45s)"
func GetRoundETA(round uint64, chainHash string) string {
	now := time.Now()
	roundTime := RoundToTime(round, chainHash)
	if roundTime.After(now) {
		return fmt.Sprintf("round %d expected at %s (in %v)", round, roundTime.Format(time.RFC3339), roundTime.Sub(now))
	}
	return fmt.Sprintf("round %d was published at %s", round, roundTime.Format(time.RFC3339))
}
