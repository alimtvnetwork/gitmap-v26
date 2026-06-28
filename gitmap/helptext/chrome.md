# chrome

Umbrella for Chrome profile snapshot/diff utilities.

## Subcommands

- `backup` — snapshot all profiles to a tar.gz under `.gitmap/chrome/backup/`.
- `restore <tarball>` — restore a snapshot into the User Data dir.
- `diff <A> <B>` — list extensions/bookmarks only-in-A vs only-in-B.
- `export-bookmarks <profile> [--format md|html|json] [--out <file>]` — export bookmarks tree.
- `which` — print directory + display name of the currently-active profile.

## Examples

```bash
gitmap chrome backup
gitmap chrome backup --out ~/chrome-2026.tar.gz
gitmap chrome restore ~/chrome-2026.tar.gz
gitmap chrome diff Default "Profile 1"
gitmap chrome export-bookmarks Lovable --format md --out bm.md
gitmap chrome export-bookmarks "Profile 1" --format json
gitmap chrome which
```
