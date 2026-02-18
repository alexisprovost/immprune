# immprune

Safely free up iCloud Photos storage by listing files that already exist in Immich.

`immprune` is read-only for iCloud Photos: it never deletes files by itself.

## Install

No Go is required for end users.

Install latest release (one-liner):

```bash
curl -fsSL https://raw.githubusercontent.com/alexisprovost/immprune/main/scripts/install.sh | sh
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/alexisprovost/immprune/main/scripts/install.sh | sh -s -- v1.0.0
```

Update to latest:

```bash
curl -fsSL https://raw.githubusercontent.com/alexisprovost/immprune/main/scripts/install.sh | sh
```

Alternative (developer/source install with Go):

```bash
go install github.com/alexisprovost/immprune/cmd/immprune@latest
```

## Quick start

```bash
immprune compare --only-videos --after 2024-01-01 --limit 500
```

On first run, setup is interactive and writes config to:

```text
~/.config/immprune/config.yaml
```

## Build locally

```bash
make build
./bin/immprune compare --limit 100
```

## Developer commands

```bash
make tidy
make fmt
make vet
make test
make ci
```

## Cross-build all platforms

```bash
make dist
```

Artifacts are generated in `dist/` for:

- linux/amd64
- linux/arm64
- darwin/amd64
- darwin/arm64
- windows/amd64
- windows/arm64

## CI/CD and releases

GitHub Actions pipeline does the following:

1. Runs checks (`go test`, `go vet`) on pushes and pull requests.
2. Builds binaries for all supported OS/ARCH targets.
3. On tag push (`v*`), creates a GitHub Release and uploads all build artifacts.

Create a release with:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Disclaimer

- You are solely responsible for any manual deletion actions you take in iCloud Photos.
- Always review generated output carefully before deleting anything.
- Keep independent backups before running cleanup workflows.
- This software is provided "as is", without warranties, and the authors are not liable for data loss, corruption, or service disruption.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
