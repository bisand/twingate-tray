# Version System - Simple Explanation

## How It Works

**The version is embedded in the binary when you build it. The binary NEVER calls git or GitHub at runtime.**

### Build Time (Only Once)
When you run `make build`, the Makefile:
1. Reads the git tag (e.g., `v1.0.0`)
2. Embeds it into the binary using Go's `-ldflags`
3. The version is now **permanently inside** the binary

### Runtime (Every Time)
When you run `./twingate-tray version`:
- It just prints the embedded version
- No git calls
- No GitHub calls
- Instant

## For GitHub Releases

### Step 1: Create Release on GitHub
```bash
# Via GitHub CLI
gh release create v1.0.0 --title "Release v1.0.0" --notes "Release notes"

# Or via GitHub web UI
# Go to Releases → New Release → Tag: v1.0.0
```

### Step 2: GitHub Actions Builds Automatically
Your `.github/workflows/release.yml` now:
1. Gets the tag from GitHub (`v1.0.0`)
2. Embeds it: `-X internal/app.Version=v1.0.0`
3. Builds the binary with that version baked in
4. Uploads to the release

### Step 3: Users Download
When someone downloads and runs the binary:
```bash
./twingate-tray version
# Output: Twingate Tray v1.0.0
```

The version `v1.0.0` is **inside the file** - no external calls needed.

## Proof

```bash
# Build with version v1.0.0
make build

# The version is in the binary file itself
strings twingate-tray | grep "v1.0.0"
# Output: v1.0.0

# Running it doesn't call git
./twingate-tray version
# Output: Twingate Tray v1.0.0 (instant, no git)
```

## That's It!

- ✅ Version comes from git tag **at build time**
- ✅ Version is embedded in the binary **permanently**  
- ✅ No runtime calls to git or GitHub
- ✅ Works perfectly with GitHub releases
- ✅ Already configured in your workflow

When you create a GitHub release with tag `v1.0.0`, the workflow builds a binary with `v1.0.0` inside it. Simple!
