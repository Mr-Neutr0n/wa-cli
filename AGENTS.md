# Working in this fork

This repository is `Mr-Neutr0n/wa-cli`, the maintained fork. `codebyoketch/wa-cli`
is reference we read; it stays untouched. Work ships by pushing to `fork`.

- Push commits to `fork main`. `origin` is upstream and stays read-only.
- Review happens in this fork. A change that only helps upstream is out of scope.
- Import work from upstream or sibling forks (`INTERDICTOR1/wa-cli`) by
  cherry-pick: keep the original author, drop agent co-author trailers, then
  run `make verify`.
- `make verify` before every push. Code style and package layout are in
  CONTRIBUTING.md.
- The `wa` on Hari's Mac is a local build of this `main`, not a Homebrew cask.
  After a push that changes the binary: `make build && install -m 755 bin/wa
  /opt/homebrew/bin/wa`.
