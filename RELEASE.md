# Release Process

## Creating a New Release

1. **Update CHANGELOG.md**
   - Move items from "Unreleased" to a new version section
   - Add the release date

2. **Commit changes**
   ```bash
   git add CHANGELOG.md
   git commit -m "Prepare for v0.x.x release"
   git push origin main
   ```

3. **Create and push tag**
   ```bash
   git tag v0.x.x
   git push origin v0.x.x
   ```

4. **Create GitHub Release** (optional but recommended)
   ```bash
   gh release create v0.x.x --notes "See CHANGELOG.md for details"
   ```

## Version Number Guidelines

- **Patch release (v0.1.0 -> v0.1.1)**: Bug fixes only
- **Minor release (v0.1.0 -> v0.2.0)**: New features, backward compatible
- **Major release (v0.x.x -> v1.0.0)**: Stable API
- **Major release (v1.0.0 -> v2.0.0)**: Breaking changes

## Checking Package on pkg.go.dev

After pushing a tag, the package will be available at:
https://pkg.go.dev/github.com/brads3290/claude-code-hooks-go@v0.x.x

It may take a few minutes for pkg.go.dev to index the new version.
EOF < /dev/null