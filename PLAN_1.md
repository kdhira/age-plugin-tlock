

# Tasks for age-plugin-tlock

## Setup
- [ ] Initialize Go module: `go mod init github.com/kdhira/age-plugin-tlock`
- [ ] Add dependencies:
  - `filippo.io/age@latest`
  - `filippo.io/age/plugin@latest`
  - `github.com/drand/tlock@latest`
- [ ] Create directory structure:
  ```
  cmd/age-plugin-tlock/main.go
  internal/cli/flags.go
  internal/codec/recipient.go
  internal/codec/identity.go
  internal/drand/network.go
  internal/tlock/recipient_adapter.go
  internal/tlock/identity_adapter.go
  internal/time/roundmap.go
  ```

## Implementation
- [ ] Implement flag parsing in `internal/cli/flags.go` (`--generate-identity`, `--recipient-round`, `--recipient-time`, `--chain`, `--strict`, `--endpoint`).
- [ ] Implement codec helpers for recipient/identity encoding & decoding.
- [ ] Implement drand network client wrapper (http client, pubkey caching, chain verification).
- [ ] Implement recipient adapter: parse payload → `tlock.NewRecipient`; support offline encryption.
- [ ] Implement identity adapter: parse payload → `tlock.NewIdentity`; handle `trustChainhash` and `Strict` mode.
- [ ] Wire everything in `cmd/age-plugin-tlock/main.go` using `filippo.io/age/plugin` handlers.
- [ ] Implement time→round conversion utility in `internal/time/roundmap.go`.

## Testing
- [ ] Unit tests for codec parsing (fuzz inputs, malformed payloads).
- [ ] Unit tests for round mapping (time → round).
- [ ] Integration test: encrypt + decrypt flow with mocked drand network.
- [ ] Integration test: decrypt too early (expect error).
- [ ] Integration test: chain mismatch with strict identity (expect failure).
- [ ] Ensure compatibility with `rage` CLI (file with tlock stanza is accepted, though not decrypted).
- [ ] Golden header tests: verify stanza output format.

## Release
- [ ] Write README with usage examples and security considerations.
- [ ] Add CI via GitHub Actions (build matrix: Linux, macOS, Windows).
- [ ] Add cross-compile scripts for static binaries.
- [ ] Create Homebrew formula (follow age-plugin-yubikey pattern).
- [ ] Tag v0.1.0 release once basic functionality is complete.
