# Contributing to Bridge

First off, thank you for considering contributing to Bridge! It's people like you that make Bridge such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

* **Use a clear and descriptive title**
* **Describe the exact steps to reproduce the problem**
* **Provide specific examples**
* **Describe the behavior you observed and what you expected**
* **Include screenshots if relevant**
* **Include your environment details:**
  - Bridge version (`bridge version`)
  - Go version (`go version`)
  - Docker version (`docker version`)
  - Operating system and version

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

* **Use a clear and descriptive title**
* **Provide a step-by-step description of the suggested enhancement**
* **Provide specific examples to demonstrate the steps**
* **Describe the current behavior and expected behavior**
* **Explain why this enhancement would be useful**

### Pull Requests

* Fill in the required template
* Follow the Go coding style
* Include appropriate test cases
* Update documentation as needed
* End all files with a newline

## Development Setup

### Prerequisites

* Go 1.23 or later
* Docker
* Make (optional but recommended)
* Git

### Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/bridge.git
   cd bridge
   ```

3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/mdxabu/bridge.git
   ```

4. Create a branch:
   ```bash
   git checkout -b feature/my-new-feature
   ```

5. Install dependencies:
   ```bash
   make deps
   ```

6. Make your changes and test:
   ```bash
   make test
   make build
   ```

## Development Workflow

### Making Changes

1. **Write tests first** (TDD approach recommended)
2. **Implement your feature**
3. **Ensure tests pass:**
   ```bash
   make test
   make test-race
   ```
4. **Format your code:**
   ```bash
   make fmt
   ```
5. **Run linters:**
   ```bash
   make lint
   ```
6. **Check for issues:**
   ```bash
   make vet
   ```

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(translator): add ICMPv6 translation support

Implement translation for ICMPv6 echo requests and replies to
enable ping functionality between IPv6 and IPv4 networks.

Closes #123
```

```
fix(nat): correct port allocation race condition

The port allocator had a race condition when multiple goroutines
attempted to allocate ports simultaneously. Added proper locking.
```

### Testing

#### Unit Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/translator/

# Run with coverage
make test-coverage

# Run with race detector
make test-race
```

#### Integration Tests

```bash
# Setup test environment
bridge setup

# Run integration test script
./scripts/integration-test.sh
```

#### Manual Testing

```bash
# Build and run
make build
sudo ./bridge start

# Test in another terminal
curl http://localhost:8080/api/health
```

### Code Style

* Follow standard Go conventions
* Use `gofmt` for formatting
* Keep functions focused and small
* Add comments for exported functions
* Use meaningful variable names
* Avoid global variables when possible

**Example:**

```go
// TranslatePacket translates an IPv6 packet to IPv4
// and returns the translated packet or an error.
func TranslatePacket(pkt *Packet) ([]byte, error) {
    if pkt == nil {
        return nil, fmt.Errorf("packet cannot be nil")
    }
    
    // Translation logic here
    return translatedPacket, nil
}
```

### Documentation

* Update README.md for user-facing changes
* Update QUICKREF.md for command changes
* Add examples in examples/ directory
* Update docs/ for technical details
* Include docstrings for all exported functions

### Project Structure

```
bridge/
├── cmd/              # CLI commands
├── internal/         # Internal packages
│   ├── translator/   # Packet translation
│   ├── nat/          # NAT state management
│   ├── tun/          # TUN interface handling
│   ├── api/          # REST API
│   └── ...
├── docs/             # Documentation
├── examples/         # Usage examples
└── scripts/          # Build and utility scripts
```

## Submitting Changes

### Before Submitting

- [ ] Tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] No linting errors (`make lint`)
- [ ] Documentation is updated
- [ ] Commit messages follow conventions
- [ ] Changes are rebased on latest main

### Submit Pull Request

1. Push your changes:
   ```bash
   git push origin feature/my-new-feature
   ```

2. Create a Pull Request on GitHub

3. Fill in the PR template completely

4. Link any related issues

5. Request review from maintainers

### PR Review Process

1. Automated checks must pass (CI/CD)
2. Code review by at least one maintainer
3. Address any requested changes
4. Once approved, PR will be merged

## Additional Notes

### Issue Labels

* `bug` - Something isn't working
* `enhancement` - New feature or request
* `documentation` - Documentation improvements
* `good first issue` - Good for newcomers
* `help wanted` - Extra attention needed
* `question` - Further information requested

### Branches

* `main` - Stable release branch
* `develop` - Development branch
* `feature/*` - Feature branches
* `fix/*` - Bug fix branches
* `docs/*` - Documentation branches

### Release Process

1. Update version in code
2. Update CHANGELOG.md
3. Tag release
4. Build binaries for all platforms
5. Create GitHub release
6. Update documentation

## Getting Help

* **Questions?** Open a GitHub Discussion
* **Bug?** Open a GitHub Issue
* **Feature idea?** Open a GitHub Issue
* **Chat?** Join our community (if available)

## Recognition

Contributors will be recognized in:
* README.md contributors section
* Release notes
* Project documentation

Thank you for contributing to Bridge! 

---

**Happy Coding!** 
