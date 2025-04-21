# Release Process

This document outlines the process for creating and publishing a new release of the Bridge project.

## 1. Update Version Information

1. Update the version number in `internal/version/version.go`:

```go
package version

const (
    Version = "x.y.z"  // Update with the new version
    BuildDate = ""     // Will be set during build
    GitCommit = ""     // Will be set during build
)
```

## 2. Create and Update Changelog

1. Update the `CHANGELOG.md` file with the changes since the last release:

```markdown
# Changelog

## v.x.y.z (YYYY-MM-DD)

### Added
- List new features

### Changed
- List changes to existing functionality

### Fixed
- List bug fixes

### Removed
- List removed features
```

## 3. Create a Release Build

1. Run the build script to create release binaries:

```bash
./scripts/build-release.sh
```

This script should:
- Set the build date
- Set the git commit hash
- Build for multiple platforms (Windows, Linux, macOS)
- Place binaries in `./dist/`

## 4. Tag the Release in Git

```bash
git add internal/version/version.go CHANGELOG.md
git commit -m "Bump version to x.y.z"
git tag -a vx.y.z -m "Release version x.y.z"
git push origin master vx.y.z
```

## 5. Create GitHub Release

1. Go to GitHub's Releases page for the repository
2. Click "Draft a new release"
3. Select the tag you just pushed
4. Title the release "Bridge vx.y.z"
5. Copy the changelog entries for this version into the description
6. Attach the binaries from the `./dist/` folder
7. Publish the release

## 6. Announce the Release

Announce the new release on:
- Project website/blog
- Social media
- Relevant forums or communities

## 7. Post-Release

1. Update development version in `internal/version/version.go` to next planned version with `-dev` suffix
2. Create a new section in `CHANGELOG.md` for the upcoming release

```go
// internal/version/version.go
const (
    Version = "x.y.(z+1)-dev"
    ...
)
```

3. Commit these changes:

```bash
git add internal/version/version.go CHANGELOG.md
git commit -m "Begin development on x.y.(z+1)"
git push origin master
```
