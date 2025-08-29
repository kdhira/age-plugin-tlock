package drand

// Network constants
const (
	// DefaultEndpoint is the default drand HTTP endpoint
	DefaultEndpoint = "https://api.drand.sh"

	// DefaultScheme is the default cryptographic scheme
	DefaultScheme = "pedersen-bls-chained"

	// SimplifiedPeriodSeconds is a simplified period for round calculation (30 seconds)
	// This is a placeholder and should be replaced with actual chain period
	SimplifiedPeriodSeconds = 30
)

// Cryptographic size constants
const (
	// ChainHashLength is the length of drand chain hash in bytes
	ChainHashLength = 32
)

// Example values for testing
const (
	// ExampleChainHash is an example drand chain hash for testing
	ExampleChainHash = "52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971"
)
