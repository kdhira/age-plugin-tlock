// Package cli provides command-line interface utilities for age-plugin-tlock.
//
// This package handles the utility mode operations of the plugin, including
// generating identities and recipients from command-line flags. It provides
// functions to process CLI arguments and output the results to stdout.
package cli

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kdhira/age-plugin-tlock/internal/codec"
	timemap "github.com/kdhira/age-plugin-tlock/internal/time"
)

// HandleUtilityFlags handles utility mode operations for the age-plugin-tlock CLI.
//
// It processes command-line flags to generate identities or recipients for offline use.
// When an operation is requested, it performs the generation and prints the result
// to stdout before exiting the program.
//
// Parameters:
//   - generateIdentity: If true, generate a new identity for the specified chain
//   - recipientRound: Specific drand round number for recipient generation
//   - recipientTime: RFC3339 timestamp for recipient generation (alternative to round)
//   - chain: Hex-encoded drand chain hash (required for all operations)
//   - strict: Enable strict mode for identity generation
//   - endpoint: Drand HTTP API endpoint URL
//
// The function exits the program with status 0 after successful operation,
// or calls log.Fatal for errors.
//
// Example usage (called from main):
//
//	HandleUtilityFlags(true, 0, "", "52db9ba70e0cc0f6...", false, "https://api.drand.sh")
//	// Prints: AGE-PLUGIN-TLOCK-... and exits
func HandleUtilityFlags(generateIdentity bool, recipientRound uint64, recipientTime, chain string, strict bool, endpoint string) {
	if generateIdentity {
		if chain == "" {
			log.Fatal("--chain is required for --generate-identity")
		}
		identity, err := codec.GenerateIdentity(chain, strict, endpoint)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(identity)
		os.Exit(0)
	}

	//nolint:nestif
	if recipientRound > 0 || recipientTime != "" {
		if chain == "" {
			log.Fatal("--chain is required for recipient generation")
		}
		var round uint64
		if recipientRound > 0 {
			round = recipientRound
		} else {
			t, err := time.Parse(time.RFC3339, recipientTime)
			if err != nil {
				log.Fatal("Invalid time format:", err)
			}
			round = timemap.ToRound(t, chain)
		}
		recipient, err := codec.GenerateRecipient(round, chain, endpoint)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(recipient)
		os.Exit(0)
	}
}
