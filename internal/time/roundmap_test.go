package time

import (
	"testing"
	"time"
)

func TestToRound(t *testing.T) {
	genesis := PlaceholderGenesisTime

	// Test time before genesis
	pastTime := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	round := ToRound(pastTime, "dummy")
	if round != 0 {
		t.Errorf("Expected round 0 for time before genesis, got %d", round)
	}

	// Test time at genesis
	round = ToRound(genesis, "dummy")
	if round != 0 {
		t.Errorf("Expected round 0 for genesis time, got %d", round)
	}

	// Test time after genesis
	futureTime := genesis.Add(2 * time.Duration(DefaultPeriodSeconds) * time.Second) // 2 periods
	round = ToRound(futureTime, "dummy")
	if round != 2 {
		t.Errorf("Expected round 2, got %d", round)
	}
}

func TestRoundToTime(t *testing.T) {
	genesis := PlaceholderGenesisTime
	period := time.Duration(DefaultPeriodSeconds) * time.Second

	// Test round 0
	roundTime := RoundToTime(0, "dummy")
	if !roundTime.Equal(genesis) {
		t.Errorf("Expected genesis time, got %v", roundTime)
	}

	// Test round 5
	expectedTime := genesis.Add(5 * period)
	roundTime = RoundToTime(5, "dummy")
	if !roundTime.Equal(expectedTime) {
		t.Errorf("Expected %v, got %v", expectedTime, roundTime)
	}
}

func TestGetRoundETA(t *testing.T) {
	// Test with future round
	futureRound := uint64(1000000) // Some large round
	eta := GetRoundETA(futureRound, "dummy")

	if eta == "" {
		t.Error("Expected non-empty ETA string")
	}

	// Test with past round (this might be tricky since it depends on current time)
	// For now, just ensure it doesn't panic
	pastRound := uint64(1)
	eta = GetRoundETA(pastRound, "dummy")
	if eta == "" {
		t.Error("Expected non-empty ETA string for past round")
	}
}
