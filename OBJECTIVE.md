# age-plugin-tlock — Implementation Plan

## Goal
Build `age-plugin-tlock`, a Go plugin that adds **time-lock recipients/identities** to `age`, using drand’s **tlock** scheme. Conform to the official age plugin protocol and file format; interoperate with the stock `age` CLI.

## Non-Goals
- Not a general drand client; only what’s required for tlock encrypt/decrypt.
- No long-polling “wait until time” UX in v1 (return “too early” and exit).
- No GUI; CLI only.

## High-Level Design
- **Plugin binary**: `age-plugin-tlock` on `PATH`. Age invokes it for `age1tlock…` recipients and `AGE-PLUGIN-TLOCK-…` identities. Use `filippo.io/age/plugin` to satisfy the plugin protocol.
- **Crypto**: Delegate to `github.com/drand/tlock`:
  - Encrypt: produce a `-> tlock <round> <chain-hash>` stanza wrapping the file key.
  - Decrypt: fetch drand **signature** for `round`; use it as the IBE key to unwrap the file key (fails with “too early” before the round).
- **Network**: default to LoE mainnet quicknet (3-second beacons); use drand HTTP v2 endpoints (e.g., `https://api.drand.sh`). Make endpoint configurable.

## User Stories
1. **Encrypt to time**: Encrypt to a future round so no one can decrypt earlier.
2. **Decrypt after time**: Decrypt **after** the target round is published.
3. **Offline encrypt**: Generate recipients and encrypt offline (recipient encodes round + chain + pubkey).
4. **Identity & chain policy**: Generate an identity bound to a chain, with **strict** mode to forbid chain switching.

## CLI / UX
- Generate identity:
  ```bash
  age-plugin-tlock --generate-identity --chain <chain-hash> [--strict] > tlock.txt
  ```
- Produce recipient:
  ```bash
  age-plugin-tlock --recipient-round <uint64> --chain <chain-hash> > recip.txt
  # or
  age-plugin-tlock --recipient-time "2025-09-01T12:00:00Z" > recip.txt
  ```
- Encrypt:
  ```bash
  age -r "$(cat recip.txt)" -o secret.age secret.txt
  ```
- Decrypt:
  ```bash
  age -d -i tlock.txt -o secret.txt secret.age
  ```

**Notes**
- Drand endpoints configurable via env/flags; default to LoE mainnet.
- Identity string: `AGE-PLUGIN-TLOCK-…` (encodes chain hash, scheme, pubkey, trust policy). Recipient: `age1tlock…` (encodes round, chain hash, scheme, pubkey).

## Architecture & Modules
```
cmd/age-plugin-tlock/main.go
internal/cli/flags.go                // flag parsing, subcommands
internal/codec/recipient.go          // encode/decode bech32 recipient payload
internal/codec/identity.go           // encode/decode identity payload
internal/drand/network.go            // http client constructor, caching
internal/tlock/recipient_adapter.go  // implements age.Recipient via tlock
internal/tlock/identity_adapter.go   // implements age.Identity via tlock
internal/time/roundmap.go            // time->round mapping (period, genesis)
```
- Use `filippo.io/age/plugin` for protocol plumbing, handlers, prompts.
- Use `github.com/drand/tlock` for `Wrap/Unwrap` semantics (`tlock_age.go`).
- Use drand HTTP v2 for round beacons/signatures with verification.

## Recipient / Identity Encoding
- **Recipient payload** (bech32 inside `age1tlock…`):  
  `version | round(u64) | chain_hash(32) | scheme_id(1) | drand_pubkey(48)`  
  Self-contained for **offline encrypt**; verify chain hash ↔ pubkey if network reachable.
- **Identity payload** (`AGE-PLUGIN-TLOCK-…`):  
  `version | trust_chainhash(bool) | chain_hash(32) | scheme_id(1) | drand_pubkey(48)`

## Time → Round Mapping
- Quicknet (mainnet): **period = 3s**. Compute `round = floor((t - genesis)/period)`. Derive genesis from drand chain info or maintain a compiled default (with override). Document approximation caveats.

## Errors & Messages
- Before target round: return clear error (“round X not yet published”), suggest ETA.
- Network errors: include endpoint in message; recommend retry/failover.
- Chain mismatch in identity:
  - default: **warn + switch** (convenience).
  - `--strict`: **fail** on mismatch.

## Security Considerations
- **Threat model**: time-lock relies on drand’s threshold BLS; early decryption requires breaking threshold or collusion.
- **Don’t mix recipients** if you need enforced timing (any non-tlock recipient would subvert the time-lock).
- Verify chain hash/public key; default to LoE mainnet; make overrides explicit.
- Age format: use proper stanzas; plugin protocol greasing prevents ossification.

## Acceptance Criteria
- **Encrypt**: `age -r "$(age-plugin-tlock --recipient-round N)"` produces header containing exactly one `-> tlock N <chain>` stanza; decryption fails with “too early” before N and succeeds after N (using a mocked network in tests for determinism).
- **Decrypt**: with identity file, plugin fetches drand round signature and unwraps.
- **Offline encrypt**: no network required to produce recipient and stanza.
- **Protocol conformance**: Works with `age` (Go) and `rage` (Rust) CLIs; unknown stanza tolerance preserved.
- **Flags**: `--generate-identity`, `--recipient-round`, `--recipient-time`, `--chain`, `--strict`, `--endpoint`, `--version` behave as documented.

## Tasks
1. **Skeleton & wiring**
   - `go mod init …`
   - Integrate `filippo.io/age/plugin` and register handlers for recipient/identity.
2. **Codec**
   - Implement recipient/identity (de)serialization (use `age/plugin` helpers where available).
   - Fuzz tests for parsing.
3. **Network**
   - `http.NewNetwork(endpoint, chainHash)` using drand HTTP v2; cache pubkey; verify chain hash. Provide endpoint failover list.
4. **Recipient adapter**
   - Parse payload → `tlock.NewRecipient(network, round)`; implement `Wrap`.
5. **Identity adapter**
   - Parse payload → `tlock.NewIdentity(network, trustChainhash)`; implement `Unwrap`. Handle “too early” vs. other errors distinctly.
6. **CLI utilities**
   - `--generate-identity`, `--recipient-round`, `--recipient-time` (time→round).
   - Output formats mirroring other plugins (YubiKey/TPM) for familiarity.
7. **Tests**
   - Unit: codec, round mapping, error paths.
   - Integration: encrypt/decrypt (mock drand server; golden header tests).
   - Interop: ensure `rage` tolerates tlock stanza.
8. **Docs**
   - README: install, examples, chain hashes, endpoints, caveats.
   - Security section with drand references.
9. **Release**
   - GitHub Actions: build matrix (macOS/Linux/Windows), static binaries.
   - Homebrew formula (follow age-plugin-yubikey pattern).

## Example Pseudocode
```go
// cmd/age-plugin-tlock/main.go
func main() {
    p, err := plugin.New("tlock")
    if err != nil { log.Fatal(err) }

    registerFlags(p) // utility mode for identity/recipient generation

    p.HandleRecipient(func(data []byte) (age.Recipient, error) {
        meta, err := codec.ParseRecipientPayload(data)
        if err != nil { return nil, err }
        net := drandnet.Must(meta.ChainHash, meta.Endpoint)
        return tlock.NewRecipient(net, meta.Round)
    })

    p.HandleIdentity(func(data []byte) (age.Identity, error) {
        meta, err := codec.ParseIdentityPayload(data)
        if err != nil { return nil, err }
        net := drandnet.Must(meta.ChainHash, meta.Endpoint)
        return tlock.NewIdentity(net, meta.TrustChain)
    })

    os.Exit(p.Main())
}
```

## Test Matrix
- **Too-early decrypt**: expect error string mentioning target round.
- **After round**: success; filekey recovered; plaintext matches.
- **Strict identity + mismatched chain**: fails with chain mismatch.
- **Endpoint failover**: primary down → try alternate endpoints.
