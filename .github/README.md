# GitHub Actions for minty

This directory contains CI/CD workflows for the minty project.

## Workflows

### CI (`ci.yml`)

Runs on every push to `main`/`develop` and on pull requests:

- **Test**: Runs unit tests with race detection and coverage
- **Lint**: Runs golangci-lint
- **Vet**: Runs `go vet` on all modules
- **Build Matrix**: Verifies cross-compilation for key platforms
- **Security**: Runs govulncheck for vulnerability scanning

### Release (`release.yml`)

Triggered by version tags (`v*`) or manual dispatch:

- Builds binaries for 11 OS/architecture combinations
- Creates GitHub Release with all artifacts
- Generates SHA256 checksums

**Supported platforms:**
| OS | Architectures |
|----|---------------|
| Linux | amd64, arm64, 386, arm (v7) |
| macOS | amd64 (Intel), arm64 (Apple Silicon) |
| Windows | amd64, 386, arm64 |
| FreeBSD | amd64, arm64 |

### Docker (`docker.yml`)

Triggered by version tags or manual dispatch:

- Builds multi-arch Docker images (amd64, arm64)
- Pushes to GitHub Container Registry (ghcr.io)
- Tags: version, major.minor, major, latest

## Creating a Release

1. **Tag the release:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **Or use manual dispatch:**
   - Go to Actions → Release → Run workflow
   - Enter version (e.g., `v1.0.0`)

## Local Development

Use the provided Makefile:

```bash
# Build examples for current platform
make build

# Run tests
make test

# Build for all platforms
make release-local

# Build Docker images
make docker

# Run examples
make run-assettrack
make run-insurequote
```

## Configuration

### Required Secrets

None required for public repositories. GitHub Actions provides `GITHUB_TOKEN` automatically.

### Optional Configuration

Edit `.github/dependabot.yml` to adjust dependency update frequency.

## File Structure

```
.github/
├── workflows/
│   ├── ci.yml          # Continuous integration
│   ├── release.yml     # Binary releases
│   └── docker.yml      # Docker images
└── dependabot.yml      # Dependency updates

examples/
├── Dockerfile.assettrack
└── insurance-quote/
    └── Dockerfile

Makefile                # Local build commands
```

## Docker Images

After release, images are available at:

```bash
# AssetTrack
docker pull ghcr.io/ha1tch/minty/assettrack:latest
docker pull ghcr.io/ha1tch/minty/assettrack:v1.0.0

# InsureQuote  
docker pull ghcr.io/ha1tch/minty/insurequote:latest
docker pull ghcr.io/ha1tch/minty/insurequote:v1.0.0

# Run
docker run -p 8080:8080 ghcr.io/ha1tch/minty/assettrack:latest
```

## Customisation

### Adding More Platforms

Edit the `matrix.include` section in `release.yml`:

```yaml
- goos: openbsd
  goarch: amd64
  suffix: ''
```

### Changing Go Version

Update `GO_VERSION` in workflow env sections and Makefile.

### Adding Examples

1. Add build job following existing patterns
2. Add Dockerfile if container image needed
3. Update Makefile targets
