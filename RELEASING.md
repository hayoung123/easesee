# Releasing

End-to-end release process for `easesee`. Aim: bump version → build cross-platform binaries → tag → upload to GitHub Release → publish to npm.

## Prerequisites (one time)

- `gh` CLI authenticated as the repo owner
- npm logged in (`npm whoami`) with 2FA-enabled account or a granular publish token
- `go` 1.22+

## Step-by-step

Replace `X.Y.Z` with the next version (e.g. `0.2.0`).

```bash
# 1. Bump npm package version (mirrors the git tag)
cd npm && npm version X.Y.Z --no-git-tag-version && cd ..

# 2. Update CHANGELOG.md — move [Unreleased] entries under a new [X.Y.Z] section
$EDITOR CHANGELOG.md

# 3. Commit + tag
git add CHANGELOG.md npm/package.json
git commit -m "release: vX.Y.Z"
git tag -a vX.Y.Z -m "vX.Y.Z"
git push && git push --tags

# 4. Build cross-platform binaries
make release-binaries VERSION=X.Y.Z

# 5. Create GitHub release with binaries attached
gh release create vX.Y.Z \
  --title "vX.Y.Z" \
  --notes-file <(awk '/^## \[X.Y.Z\]/{flag=1;next} /^## \[/{flag=0} flag' CHANGELOG.md) \
  dist/easesee-darwin-arm64 \
  dist/easesee-darwin-amd64 \
  dist/easesee-linux-amd64 \
  dist/easesee-linux-arm64

# 6. Publish to npm (use --otp= or rely on session passkey/token)
cd npm && npm publish --access public
```

## Version source of truth

- Git tag `vX.Y.Z` is authoritative
- `internal/cli.Version` injected at build time via `-ldflags "-X 'github.com/hayoung123/easesee/internal/cli.Version=X.Y.Z'"` (handled by `make release-binaries`)
- `npm/package.json` version mirrors the tag

## Yanking a bad release

```bash
# npm — deprecate (cannot fully unpublish after 72h)
npm deprecate easesee@X.Y.Z "broken release, use X.Y.Z+1"

# GitHub
gh release delete vX.Y.Z --yes
git push --delete origin vX.Y.Z
git tag -d vX.Y.Z
```
