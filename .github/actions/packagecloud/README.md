# PackageCloud docker action

This action will push packages to PackageCloud via the package_cloud gem.

## Inputs

### `repo`

**Required** The repo to push to.

### `dir`

**Required** Directory where the packages are. All rpms and debs found here will be pushed.

## Outputs

### `stdout`
Stdout from the command execution.

## Example usage

```yaml
uses: ./.github/actions/packagecloud
env:
  PACKAGECLOUD_TOKEN: ${{ secrets.PACKAGECLOUD_TOKEN }}
with:
  repo: 'tyk/tyk-sync-unstable'
  dir: 'dist'
```
