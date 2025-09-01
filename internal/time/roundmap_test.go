package time

import (
	"testing"
	"time"
)

func TestToRound(t *testing.T) {
	genesis := PlaceholderGenesisTime
	chainHash := "dummy"

	// Set up test metadata
	SetChainMetadata(chainHash, genesis, time.Duration(DefaultPeriodSeconds)*time.Second)

	// Test time before genesis
	pastTime := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	round, err := ToRound(pastTime, chainHash)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if round != 0 {
		t.Errorf("Expected round 0 for time before genesis, got %d", round)
	}

	// Test time at genesis
	round, err = ToRound(genesis, chainHash)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if round != 0 {
		t.Errorf("Expected round 0 for genesis time, got %d", round)
	}

	// Test time after genesis
	futureTime := genesis.Add(2 * time.Duration(DefaultPeriodSeconds) * time.Second) // 2 periods
	round, err = ToRound(futureTime, chainHash)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if round != 2 {
		t.Errorf("Expected round 2, got %d", round)
	}
}

func TestRoundToTime(t *testing.T) {
	genesis := PlaceholderGenesisTime
	period := time.Duration(DefaultPeriodSeconds) * time.Second
	chainHash := "dummy"

	// Set up test metadata
	SetChainMetadata(chainHash, genesis, period)

	// Test round 0
	roundTime, err := RoundToTime(0, chainHash)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !roundTime.Equal(genesis) {
		t.Errorf("Expected genesis time, got %v", roundTime)
	}

	// Test round 5
	expectedTime := genesis.Add(5 * period)
	roundTime, err = RoundToTime(5, chainHash)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !roundTime.Equal(expectedTime) {
		t.Errorf("Expected %v, got %v", expectedTime, roundTime)
	}
}

func TestGetRoundETA(t *testing.T) {
	genesis := PlaceholderGenesisTime
	chainHash := "dummy"

	// Set up test metadata
	SetChainMetadata(chainHash, genesis, time.Duration(DefaultPeriodSeconds)*time.Second)

	// Test with future round
	futureRound := uint64(1000000) // Some large round
	eta, err := GetRoundETA(futureRound, chainHash)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if eta == "" {
		t.Error("Expected non-empty ETA string")
	}

	// Test with past round (this might be tricky since it depends on current time)
	// For now, just ensure it doesn't panic
	pastRound := uint64(1)
	eta, err = GetRoundETA(pastRound, chainHash)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if eta == "" {
		t.Error("Expected non-empty ETA string for past round")
	}
}
