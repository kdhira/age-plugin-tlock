// Package main implements the age-plugin-tlock command-line tool.
//
// This plugin adds time-lock encryption capabilities to the age encryption tool
// using drand's tlock scheme. It can operate in two modes:
//
// 1. Plugin mode: Integrates with age CLI for encryption/decryption
// 2. Utility mode: Generates identities and recipients for offline use
//
// Usage examples:
//
//	# Generate identity
//	./age-plugin-tlock --generate-identity --chain <chain-hash>
//
//	# Generate recipient for specific round
//	./age-plugin-tlock --recipient-round 1000000 --chain <chain-hash>
//
//	# Generate recipient for duration from now
//	./age-plugin-tlock --recipient-duration 1h30m --chain <chain-hash>
//
//	# Use with age CLI
//	echo "secret" | age -r "$(./age-plugin-tlock --recipient-round 1000000 --chain <chain-hash>)" -o encrypted.txt
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"filippo.io/age"
	"filippo.io/age/plugin"
	"github.com/kdhira/age-plugin-tlock/internal/cli"
	"github.com/kdhira/age-plugin-tlock/internal/config"
	"github.com/kdhira/age-plugin-tlock/internal/tlock"
)

// Version information - set at build time using -ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	// Load configuration from environment
	cfg := config.LoadFromEnv()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// fmt.Fprintf(os.Stderr, "%v\n", os.Args)

	p, err := plugin.New("tlock")
	if err != nil {
		log.Fatal(err)
	}

	// Register utility flags with the plugin framework
	p.RegisterFlags(nil)

	// HandleRecipient: parse the plugin recipient string and return an age.Recipient
	p.HandleRecipient(func(data []byte) (age.Recipient, error) {
		return tlock.NewRecipientAdapter(data)
	})

	// HandleIdentity: parse the plugin identity string and return an age.Identity
	p.HandleIdentity(func(data []byte) (age.Identity, error) {
		return tlock.NewIdentityAdapter(data)
	})

	// Add our custom flags
	version := flag.Bool("version", false, "Print version information")
	generateIdentity := flag.Bool("generate-identity", false, "Generate a new identity")
	recipientRound := flag.Uint64("recipient-round", 0, "Generate recipient for specific round")
	recipientTime := flag.String("recipient-time", "", "Generate recipient for time (RFC3339 format)")
	recipientDuration := flag.String("recipient-duration", "", "Generate recipient for duration from now (e.g., 1h30m)")
	chain := flag.String("chain", "", "Drand chain hash (hex)")
	strict := flag.Bool("strict", false, "Strict mode for identity (no chain switching)")
	endpoint := flag.String("endpoint", cfg.DrandEndpoint, "Drand HTTP endpoint")

	// Parse flags (this will be called automatically by p.Main() if not called before)
	flag.Parse()

	if *version {
		printVersion()
		os.Exit(0)
	}

	// Handle utility mode flags
	if *generateIdentity || *recipientRound > 0 || *recipientTime != "" || *recipientDuration != "" {
		cli.HandleUtilityFlags(*generateIdentity, *recipientRound, *recipientTime, *recipientDuration, *chain, *strict, *endpoint)
		os.Exit(0)
	}
	os.Exit(p.Main())
}

// printVersion prints version information to stdout
func printVersion() {
	fmt.Printf("age-plugin-tlock %s\n", Version)
	fmt.Printf("Commit: %s\n", Commit)
	fmt.Printf("Built: %s\n", BuildTime)
}
