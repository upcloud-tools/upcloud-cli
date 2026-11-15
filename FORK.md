# Community fork

This repository is an independent community fork of [UpCloudLtd/upcloud-cli](https://github.com/UpCloudLtd/upcloud-cli).
It tracks upstream releases and carries fixes that have not (yet) been merged upstream.
Changes are kept small and self-contained so that upstream can merge them at any time.

## Versioning

- Fork releases are always suffixed with `-community`, e.g. `3.37.1-community` (git tag `v3.37.1-community`).
- The base version is the upstream release the fork is synced to; fork-only changes bump the patch (or minor) component on top of it.
- Like upstream, the binary version is derived from the git tag at build time (goreleaser injects it into `internal/config.Version`); release preparation only updates `CHANGELOG.md`.
- Compare links for `-community` versions point to this fork; links for upstream versions point to UpCloudLtd/upcloud-cli.

## Feature parity

**Base: upstream v3.37.0 (2026-10-13).** The fork contains all upstream features; no upstream feature is removed or disabled.

### Fork-only changes

| Change | Introduced in | Related issue | Upstream status |
| ------ | ------------- | ------------- | --------------- |
| `kubernetes node-group scale --count 0` sends `{"count": 0}` instead of silently doing nothing. The command now uses the v9 SDK; JSON/YAML output shape is preserved. | 3.37.1-community | [#1](https://github.com/upcloud-tools/upcloud-cli/issues/1) | Not fixed as of upstream v3.37.0 |
| Command help no longer shows `Default:` for required flags (e.g. `scale --count` showed a misleading `Default: 0`). | 3.37.1-community | [#1](https://github.com/upcloud-tools/upcloud-cli/issues/1) | Not fixed as of upstream v3.37.0 |

### Notes for upstream merging

- The scale fix works around a bug in the `upcloud-go-api` v8 SDK: `request.ModifyKubernetesNodeGroup.Count` is tagged `json:"count,omitempty"`, so a zero count is dropped from the PATCH body and the API no-ops while the CLI reports success. The v9 SDK (`KubernetesNodeGroupPatch`) models `count` without `omitempty` and is not affected. Fixing the v8 SDK would additionally benefit all other v8 consumers.
- The help renderer change is in `internal/ui/usage.go` (`formatFlags`) and applies to every command with required flags.

## Syncing with upstream

```sh
git remote add upstream git@github.com:UpCloudLtd/upcloud-cli.git
git fetch upstream
git merge upstream/main
```

After syncing: update the base version above, drop fork-only changes that upstream merged from the table, and base the next `-community` version on the new upstream release.