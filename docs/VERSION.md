# Version Management

This project uses **automatic version injection** from Git tags. You don't need to manually update version numbers - they're automatically extracted from your Git tags when building.

## How It Works

The build system uses Go's `-ldflags` to inject version information at compile time:

- **Version**: Automatically extracted from the latest Git tag (e.g., `v0.3.0`)
- **Git Commit**: Short SHA of the current commit
- **Build Date**: UTC timestamp of when the binary was built

## Creating a Release

### 1. Tag a New Version

```bash
# Create a new version tag
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push the tag to GitHub
git push origin v1.0.0
```

### 2. Build the Release Binary

```bash
make build
```

The binary will automatically include the version from the tag:
```
Build complete: twingate-tray (version v1.0.0)
```

### 3. Check the Version

```bash
./twingate-tray version
```

Output:
```
Twingate Tray v1.0.0
Version: v1.0.0
Commit: abc1234
Built: 2026-02-10T21:00:00Z
```

## Version Format

Git tags should follow semantic versioning with a `v` prefix:

- `v1.0.0` - Major release
- `v1.2.0` - Minor release (new features)
- `v1.2.3` - Patch release (bug fixes)

### Development Builds

If you build without tagging or with uncommitted changes:

```bash
make build
# Output: twingate-tray (version v0.3.0-4-g98abff0-dirty)
```

The version format `v0.3.0-4-g98abff0-dirty` means:
- `v0.3.0` - Latest tag
- `4` - Number of commits since that tag
- `g98abff0` - Current commit SHA
- `dirty` - You have uncommitted changes

## Version in the About Dialog

The About dialog automatically displays the current version:

```
Twingate Tray

Version: v1.0.0

System Tray Indicator for Twingate VPN on Linux
...
```

No manual updates needed!

## Manual Version Override

If you need to override the version (e.g., for testing):

```bash
go build -ldflags "-X github.com/bisand/twingate-tray/internal/app.Version=1.2.3-custom" \
         -o twingate-tray ./cmd/twingate-tray
```

## GitHub Releases

When creating a GitHub release:

1. **Tag the release** on GitHub (or locally and push)
2. **Build the binary** with `make build`
3. **Upload the binary** to the GitHub release
4. The binary will show the correct version from the tag

Example workflow:
```bash
# Tag and push
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Build release binary
make clean
make build

# Create GitHub release and upload ./twingate-tray
gh release create v1.0.0 ./twingate-tray \
    --title "Twingate Tray v1.0.0" \
    --notes "Release notes here"
```

## Checking Version Information

Three ways to check version:

1. **CLI**: `./twingate-tray version`
2. **About Dialog**: Click tray icon → About
3. **Build Time**: `make version` shows what will be injected

## Summary

- ✅ **No manual version updates needed**
- ✅ **Version automatically from Git tags**
- ✅ **Works with GitHub releases**
- ✅ **Shows commit and build date**
- ✅ **Development builds show dirty status**
