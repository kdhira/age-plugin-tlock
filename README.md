# age-plugin-tlock

[![Go Report Card](https://goreportcard.com/badge/github.com/kdhira/age-plugin-tlock)](https://goreportcard.com/report/github.com/kdhira/age-plugin-tlock)

An [age](https://github.com/FiloSottile/age) plugin that adds **time-lock encryption** using [drand](https://drand.love/)'s tlock scheme. Encrypt files that can only be decrypted after a specific future time, leveraging drand's threshold BLS signatures for cryptographic time proofs.

## Features

- **Time-lock encryption**: Encrypt data that becomes decryptable only after a target round
- **Offline recipient generation**: Create recipients without network access
- **Multiple drand networks**: Support for different drand chains (mainnet, testnet, etc.)
- **Strict mode**: Optional strict chain validation to prevent chain switching
- **Age compatibility**: Works seamlessly with `age` and `rage` CLIs
- **Plugin protocol**: Implements official age plugin protocol with full framework integration

## Installation

### From Source

```bash
git clone https://github.com/kdhira/age-plugin-tlock.git
cd age-plugin-tlock

# Build with version information
go build -ldflags "-X 'main.Version=v1.0.0' -X 'main.Commit=$(git rev-parse --short HEAD)' -X 'main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" -o age-plugin-tlock ./cmd/age-plugin-tlock

# Or for development builds
go build -o age-plugin-tlock ./cmd/age-plugin-tlock
```

### Configuration

The plugin can be configured using environment variables:

- `AGE_PLUGIN_TLOCK_ENDPOINT`: Default drand HTTP endpoint (default: "https://api.drand.sh")
- `AGE_PLUGIN_TLOCK_TIMEOUT`: Request timeout duration (default: "10s")
- `AGE_PLUGIN_TLOCK_MAX_RETRIES`: Maximum retry attempts for network requests (default: "3")
- `AGE_PLUGIN_TLOCK_RETRY_DELAY`: Base delay between retries (default: "1s")
- `AGE_PLUGIN_TLOCK_STRICT_MODE`: Enable strict chain validation (default: "false")

Example:
```bash
export AGE_PLUGIN_TLOCK_ENDPOINT="https://custom-drand-endpoint.com"
export AGE_PLUGIN_TLOCK_TIMEOUT="30s"
./age-plugin-tlock --generate-identity --chain 52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971
```

**Requirements:**
- Go 1.21+ (uses `filippo.io/age/plugin` framework)
- Network access to drand endpoints for identity/recipient generation
- Age v1.2.1+ for plugin protocol compatibility

### Pre-built Binaries

*Coming soon - check releases for pre-built binaries*

## Usage

### CLI Utility Mode

The plugin provides CLI utilities for generating identities and recipients:

#### Generate Identity

```bash
# Generate identity for LoE mainnet with strict mode
./age-plugin-tlock --generate-identity --chain 52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971 --strict > identity.txt

# Generate identity for custom chain
./age-plugin-tlock --generate-identity --chain <chain-hash> --endpoint https://custom-drand-endpoint.com > identity.txt
```

#### Generate Recipient

```bash
# Generate recipient for specific round
./age-plugin-tlock --recipient-round 5000000 --chain 52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971 > recipient.txt

# Generate recipient for future time (RFC3339 format)
./age-plugin-tlock --recipient-time "2025-01-01T00:00:00Z" --chain 52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971 > recipient.txt

# Generate recipient for duration from now (e.g., 1 hour 30 minutes)
./age-plugin-tlock --recipient-duration 1h30m --chain 52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971 > recipient.txt
```

### Age Plugin Mode

Once installed, the plugin works seamlessly with age:

#### Encrypt

```bash
# Encrypt file to be decryptable after round 5000000
age -r "$(cat recipient.txt)" -o secret.enc secret.txt

# Encrypt to multiple recipients (mixing with other age recipients)
age -r "$(cat recipient.txt)" -r "age1..." -o secret.enc secret.txt
```

#### Decrypt

```bash
# Decrypt using identity
age -d -i identity.txt -o secret.txt secret.enc
```

## Plugin Protocol Integration

This plugin implements the official [age plugin protocol](https://github.com/FiloSottile/age/blob/main/age-plugin.md) using the `filippo.io/age/plugin` framework. The integration includes:

### Framework Components

- **`plugin.New("tlock")`**: Registers the plugin with the age framework
- **`p.HandleRecipient()`**: Processes `age1tlock*` recipients during encryption
- **`p.HandleIdentity()`**: Processes `AGE-PLUGIN-TLOCK-*` identities during decryption
- **`p.RegisterFlags()`**: Integrates CLI utility flags with the plugin framework
- **`p.Main()`**: Runs the plugin event loop handling age's protocol messages

### Protocol Flow

#### Recipient Processing (Encryption)
1. Age sends: `add-recipient <recipient-string>`
2. Plugin parses bech32-encoded recipient using `plugin.ParseRecipient()`
3. Plugin validates recipient format and extracts payload
4. Plugin creates tlock recipient adapter
5. Plugin responds with: `-> recipient-stanza <round> <chain-hash> <encrypted-file-key>`

#### Identity Processing (Decryption)
1. Age sends: `add-identity <identity-string>`
2. Plugin parses bech32-encoded identity using `plugin.ParseIdentity()`
3. Plugin validates identity format and extracts payload
4. Plugin creates tlock identity adapter
5. Plugin processes: `recipient-stanza <encrypted-file-key>`
6. Plugin fetches drand signature for target round
7. Plugin decrypts file key using tlock scheme
8. Plugin responds with: `-> file-key <decrypted-file-key>`

### Bech32 Encoding

The plugin uses age's built-in bech32 encoding functions:
- **`plugin.EncodeIdentity("tlock", data)`**: Creates `AGE-PLUGIN-TLOCK-*` identities
- **`plugin.ParseIdentity(encoded)`**: Parses identities back to binary data
- **`plugin.EncodeRecipient("tlock", data)`**: Creates `age1tlock*` recipients
- **`plugin.ParseRecipient(encoded)`**: Parses recipients back to binary data

**Important**: The plugin relies on age's internal bech32 implementation which automatically converts 8-bit binary data to 5-bit bech32 format. Custom bech32 implementations may fail due to byte value restrictions.

### Error Handling

The plugin provides structured error messages with context:
- **Validation errors**: Include field names and invalid values
- **Network errors**: Include endpoint and operation details
- **Decoding errors**: Specify input format and parsing issues
- **Time errors**: Provide ETA information for future rounds
- **"round X not yet published"**: Target round is in the future with ETA
- **"chain hash mismatch"**: Identity/recipient chain mismatch (strict mode)
- **"invalid identity encoding"**: Malformed bech32 identity string

### Compatibility

- **Age versions**: Compatible with age v1.2.1+ (uses plugin framework)
- **Rage support**: Compatible with rage CLI (accepts tlock stanzas)
- **Protocol version**: Implements current age plugin protocol v1

## Drand Networks

### LoE Mainnet (Recommended)
- **Chain Hash**: `52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971`
- **Endpoint**: `https://api.drand.sh`
- **Period**: 3 seconds
- **Beacon ID**: quicknet

### Testnet
- **Chain Hash**: `7672797f548f3f4748ac4bf3352fc6c6b6468c9ad40ad456a397545c6e2df5bf`
- **Endpoint**: `https://pl-us.testnet.drand.sh`
- **Period**: 3 seconds

### Custom Networks

You can use any drand network by specifying the chain hash and endpoint:

```bash
./age-plugin-tlock --generate-identity --chain <custom-chain-hash> --endpoint <custom-endpoint>
```

## Security Considerations

### Threat Model

This plugin relies on drand's threshold BLS signatures for time proofs. The security assumptions are:

- **Trusted Setup**: drand's initial key ceremony was properly executed
- **Threshold Security**: At least 2/3 of drand nodes remain honest
- **Network Security**: The drand HTTP endpoints are trustworthy
- **Time Oracle**: drand provides accurate time proofs

### Limitations

- **Early Decryption**: If the threshold is compromised, files could be decrypted early
- **Network Dependency**: Decryption requires network access to fetch drand signatures
- **Clock Synchronization**: Relies on drand's time, not local clocks
- **Chain Switching**: Default mode allows chain switching (use `--strict` to prevent)

### Best Practices

1. **Use strict mode** for high-security applications
2. **Verify chain hashes** before use
3. **Keep identities secure** - they allow decryption once time conditions are met
4. **Don't mix recipients** if you need guaranteed timing (any non-tlock recipient can subvert the time-lock)
5. **Test with small rounds** first to verify functionality

## Caveats and Limitations

### Time Approximation

- Round-to-time conversion uses approximate genesis time
- Actual drand genesis time may vary slightly
- Use `--recipient-round` for precise control

### Network Issues

- **Network failures**: If drand endpoints are unreachable, decryption will fail
- **Chain mismatches**: Wrong chain hash will result in decryption failures
- **Rate limiting**: Drand endpoints may have rate limits

### Compatibility

- **Age versions**: Tested with age v1.2.1+ (NB special build: `v1.2.1-0.20240926110859-2214a556f604`)
- **Rage support**: Compatible with rage (Rust implementation)
- **Plugin protocol**: Follows official age plugin protocol

### Performance

- **Encryption**: Fast (no network required for recipient generation)
- **Decryption**: Requires network round-trip to fetch signatures
- **Large files**: No size limits, but decryption requires full file in memory

## Examples

### Basic Time-Lock

```bash
# 1. Generate identity
./age-plugin-tlock --generate-identity --chain 52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971 > identity.txt

# 2. Generate recipient for 1 hour from now
./age-plugin-tlock --recipient-duration 1h --chain 52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971 > recipient.txt

# 3. Encrypt
echo "Secret message" | age -r "$(cat recipient.txt)" -o secret.enc

# 4. Try to decrypt immediately (will fail)
age -d -i identity.txt -o secret.txt secret.enc  # Error: round not yet published

# 5. Wait ~1 hour, then decrypt
age -d -i identity.txt -o secret.txt secret.enc  # Success!
```

### Strict Mode

```bash
# Generate identity with strict chain validation
./age-plugin-tlock --generate-identity --chain 52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971 --strict > identity.txt

# This identity will reject decryption if chain hash changes
```

## References

- [Age Encryption](https://github.com/FiloSottile/age)
- [Drand Documentation](https://drand.love/docs/)
- [Tlock Paper](https://eprint.iacr.org/2021/573)
- [Age Plugin Protocol](https://github.com/FiloSottile/age/blob/main/age-plugin.md)
- [Age Plugin Framework](https://pkg.go.dev/filippo.io/age/plugin) - Go package documentation
- [Age v1.2.1 Plugin Framework](https://github.com/FiloSottile/age/releases/tag/v1.2.1) - Version used for implementation

## Contributing

Contributions welcome! Please:

1. Test your changes with both `age` and `rage`
2. Add tests for new functionality
3. Update documentation
4. Follow Go best practices

## License

This project is licensed under the BSD 3-Clause License - see the LICENSE file for details.

## Acknowledgments

- [Filippo Valsorda](https://github.com/FiloSottile) for the age encryption tool
- [Drand team](https://drand.love/) for the time-lock encryption scheme
- [Age plugin ecosystem](https://github.com/FiloSottile/age#plugins) for inspiration
