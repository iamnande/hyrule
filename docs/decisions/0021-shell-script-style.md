# 0021: shell script style - defer to Google's guide, add `shellcheck` to CI's `lint` job

didn't need a full decision doc, per its own stub - the actual rule lives
in [conventions.md#shell-scripts](../conventions.md#shell-scripts):
defer to the [Google Shell Style Guide](https://google.github.io/styleguide/shellguide.html),
same approach [style.md](../style.md) takes for Go. no repo-specific rule
harder than the guide - [local/new-service.sh](../../local/new-service.sh),
the only script that exists today, already follows it.

`shellcheck` runs in CI's existing `lint` job alongside `golangci-lint` -
not a new job, just another step in the one that already exists. pinned
in [mise.toml](../../mise.toml) like the repo's other tools.
