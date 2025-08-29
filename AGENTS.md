# AGENTS.md - AI Agent Guidelines for age-plugin-tlock

This document provides guidelines for AI agents and automated tools working on the `age-plugin-tlock` codebase. It covers project structure, development workflows, testing protocols, and best practices for maintaining code quality.

## 📋 Project Overview

**age-plugin-tlock** is a Go plugin for the [age](https://github.com/FiloSottile/age) encryption tool that adds time-lock encryption using [drand](https://drand.love/)'s tlock scheme. The plugin enables encrypting data that can only be decrypted after a specific future time.

### Key Features
- Time-lock encryption/decryption using drand beacons
- Offline recipient generation
- Age plugin protocol v1 compliance
- Multiple drand network support
- Strict mode for chain validation

## 🏗️ Project Structure

```
age-plugin-tlock/
├── cmd/age-plugin-tlock/main.go      # Plugin entry point
├── internal/
│   ├── cli/flags.go                  # CLI utility flags
│   ├── codec/
│   │   ├── identity.go              # Identity encoding/decoding
│   │   └── recipient.go             # Recipient encoding/decoding
│   ├── drand/network.go             # Drand HTTP client
│   ├── time/roundmap.go             # Time ↔ round conversion
│   └── tlock/
│       ├── recipient_adapter.go     # Age recipient interface
│       └── identity_adapter.go      # Age identity interface
├── testdata/                        # Test data and scripts
├── go.mod                           # Go module definition
├── go.sum                           # Dependency checksums
├── README.md                        # User documentation
├── AGENTS.md                        # This file
└── .gitignore                       # Git ignore rules
```

## 🤖 Agent Interaction Guidelines

### Communication Protocol

**Always use clear, technical language** in responses. Avoid conversational fillers like "Great!", "Sure!", or "Okay!". Be direct and focus on technical details.

**Example:**
- ✅ "Updated the recipient adapter to handle offline encryption"
- ❌ "Great! I've updated the recipient adapter to handle offline encryption"

### Task Management

1. **Read existing code** before making changes
2. **Understand the context** from OBJECTIVE.md and PLAN_1.md in the `.agent` directory
3. **Test changes** using provided test scripts
4. **Update documentation** when making significant changes
5. **Follow Go best practices** and project conventions

### Code Review Process

**Before submitting changes:**
1. Run `go build ./cmd/age-plugin-tlock` to ensure compilation
2. Run `go test ./...` to ensure tests pass
3. Check for linting issues with `go vet ./...`
4. Ensure all imports are properly organized
5. Test with both `age` and `rage` CLIs if possible

## 🔧 Development Workflows

### Adding New Features

1. **Understand Requirements**: Read OBJECTIVE.md and relevant issues
2. **Design Integration**: Ensure compatibility with age plugin protocol
3. **Implement Incrementally**: Make small, testable changes
4. **Test Thoroughly**: Include unit tests and integration tests
5. **Update Documentation**: Modify README.md and this file as needed

### Bug Fixes

1. **Reproduce Issue**: Use provided test scripts to reproduce
2. **Identify Root Cause**: Trace through the code execution path
3. **Implement Fix**: Make minimal changes to address the issue
4. **Verify Fix**: Ensure existing tests still pass
5. **Add Regression Test**: Prevent future occurrences

### Testing Protocols

#### Unit Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/codec

# Run with coverage
go test -cover ./...
```

#### Integration Testing
```bash
# End-to-end smoke test
./smoke-test.sh

# Manual plugin testing
./test_plugin.sh | ~/bin/age-plugin-tlock
```

#### Compatibility Testing
```bash
# Test with age CLI
echo "test" | age -r "$(./age-plugin-tlock --recipient-round 500000 --chain 52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971)" -a

# Test decryption
age -d -i identity.txt -o output.txt encrypted.age
```

#### Skipped Tests
Some tests are currently skipped due to external dependencies:

- **`TestIdentityAdapterUnwrap`** in `internal/tlock/identity_adapter_test.go`
- **`TestRecipientAdapterWrap`** in `internal/tlock/recipient_adapter_test.go`

**Skip Reason:** These tests require network connectivity to drand endpoints and proper integration with the tlock cryptographic library. They are skipped to avoid test failures in environments without network access or when the drand API is unavailable.

**Future Implementation:** These tests should be converted to integration tests with proper mocking of the drand network client and tlock library dependencies.

## 📝 Code Style Guidelines

### Go Best Practices

1. **Use `gofmt`**: All code should be formatted with `gofmt`
2. **Meaningful Names**: Use descriptive variable and function names
3. **Error Handling**: Always handle errors appropriately
4. **Documentation**: Add comments for complex logic
5. **Package Organization**: Keep related functionality together

### Project Conventions

1. **Error Messages**: Use lowercase, no trailing punctuation
2. **Logging**: Use `log` package for errors, avoid `fmt.Printf`
3. **Imports**: Group standard library, third-party, and local imports
4. **Constants**: Define magic numbers as named constants
5. **Tests**: Use descriptive test names with `TestXxx` format

## 🔒 Security Considerations

### For Agents Working on Security Features

1. **Never commit secrets**: API keys, private keys, or sensitive data
2. **Validate inputs**: Always validate user inputs and network responses
3. **Use secure defaults**: Prefer secure configurations by default
4. **Document assumptions**: Clearly document security assumptions
5. **Test edge cases**: Include tests for malformed inputs and error conditions

### Drand Integration Security

1. **Chain validation**: Always verify chain hashes match expected values
2. **Network security**: Use HTTPS for all drand API calls
3. **Time validation**: Validate round numbers and time calculations
4. **Error handling**: Don't leak sensitive information in error messages

## 🚀 Deployment and Release

### Build Process
```bash
# Build for current platform
go build -o age-plugin-tlock ./cmd/age-plugin-tlock

# Cross-compilation
GOOS=linux GOARCH=amd64 go build -o age-plugin-tlock-linux ./cmd/age-plugin-tlock
GOOS=darwin GOARCH=amd64 go build -o age-plugin-tlock-macos ./cmd/age-plugin-tlock
GOOS=windows GOARCH=amd64 go build -o age-plugin-tlock-windows.exe ./cmd/age-plugin-tlock
```

### Release Checklist
- [ ] All tests pass
- [ ] Code builds on all target platforms
- [ ] Documentation is up-to-date
- [ ] Security review completed
- [ ] Version number updated
- [ ] Changelog updated
- [ ] Git tag created

## 📚 Documentation Updates

### When to Update Documentation

1. **New Features**: Add usage examples and configuration options
2. **API Changes**: Update function signatures and parameters
3. **Security Changes**: Document new security considerations
4. **Breaking Changes**: Clearly mark backward-incompatible changes
5. **Bug Fixes**: Document workarounds or important fixes

### Documentation Files

- **README.md**: User-facing documentation
- **AGENTS.md**: This file - agent interaction guidelines
- **Code Comments**: Inline documentation for complex logic
- **Commit Messages**: Clear, descriptive commit messages

## 🔄 Continuous Integration

### GitHub Actions (Planned)

When CI is implemented, agents should:
1. Ensure all CI checks pass before merging
2. Update CI configuration for new dependencies
3. Test on all supported platforms
4. Monitor for security vulnerabilities

### Pre-commit Hooks

```bash
# Recommended pre-commit checks
go fmt ./...
go vet ./...
go test ./...
gofmt -d .  # Check formatting
```

## 🆘 Troubleshooting

### Common Issues

1. **Build Failures**: Check Go version (1.21+) and dependencies
2. **Plugin Not Found**: Ensure plugin is in PATH or use full path
3. **Network Errors**: Check drand endpoint availability
4. **Encoding Errors**: Verify bech32 format using age's internal functions

### Debug Mode
```bash
# Enable verbose logging
export AGEDEBUG=plugin

# Test with debug output
age -r "$(./age-plugin-tlock --recipient-round 500000 --chain ...)" -a < input.txt
```

## 📞 Contact and Support

For questions about agent interactions or development guidelines:
1. Check existing documentation (README.md, this file)
2. Review recent commits for similar changes
3. Test changes thoroughly before proposing
4. Include context and rationale for changes

## 🔗 Related Resources

- [Age Plugin Protocol](https://github.com/FiloSottile/age/blob/main/age-plugin.md)
- [Drand Documentation](https://drand.love/docs/)
- [Go Best Practices](https://golang.org/doc/effective_go.html)
- [Age Source Code](https://github.com/FiloSottile/age)

---

*This document should be updated whenever development workflows or project structure changes significantly.*
