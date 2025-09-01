package time

import (
	"testing"
	"time"
)

// BenchmarkToRound benchmarks the time to round conversion
func BenchmarkToRound(b *testing.B) {
	testTime := time.Now()
	chainHash := "52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971"

	// Set up test metadata to avoid network calls
	SetChainMetadata(chainHash, PlaceholderGenesisTime, time.Duration(DefaultPeriodSeconds)*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ToRound(testTime, chainHash)
	}
}

// BenchmarkRoundToTime benchmarks the round to time conversion
func BenchmarkRoundToTime(b *testing.B) {
	round := uint64(1000000)
	chainHash := "52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971"

	// Set up test metadata to avoid network calls
	SetChainMetadata(chainHash, PlaceholderGenesisTime, time.Duration(DefaultPeriodSeconds)*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RoundToTime(round, chainHash)
	}
}

// BenchmarkGetRoundETA benchmarks the ETA calculation
func BenchmarkGetRoundETA(b *testing.B) {
	round := uint64(1000000)
	chainHash := "52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971"

	// Set up test metadata to avoid network calls
	SetChainMetadata(chainHash, PlaceholderGenesisTime, time.Duration(DefaultPeriodSeconds)*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetRoundETA(round, chainHash)
	}
}
