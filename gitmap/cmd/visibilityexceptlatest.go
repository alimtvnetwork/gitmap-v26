// Package cmd — visibilityexceptlatest.go: filters a matched repo set
// so the newest `-vN` sibling per base group is preserved (i.e. removed
// from the apply set). Repos that don't carry a `-vN` suffix are left
// untouched.
//
// Spec: spec/01-app/116-bulk-visibility-mapub-mapri.md §except-latest.
package cmd

import (
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/alimtvnetwork/gitmap-v26/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v26/gitmap/visibility"
)

// versionSuffixRE captures the trailing `-v<digits>` segment. Anchored
// at the end so multi-segment names like `proj-v1-fix-v2` match only
// the final `-v2`.
var versionSuffixRE = regexp.MustCompile(`(?i)^(.*)-v(\d+)$`)

// filterExceptLatest removes the highest -vN entry per base from
// `in`. Bases with only one versioned entry collapse to "no peers"
// and are left intact (nothing to preserve over). Unversioned repos
// are passed through untouched. Each drop is logged via `w`.
func filterExceptLatest(in []visibility.MatchedRepo, w io.Writer) []visibility.MatchedRepo {
	type peak struct {
		idx int
		ver int
	}
	peaks := map[string]peak{}
	counts := map[string]int{}
	for i, m := range in {
		base, ver, ok := parseVersionedName(m.RepoName)
		if !ok {
			continue
		}
		counts[base]++
		if cur, seen := peaks[base]; !seen || ver > cur.ver {
			peaks[base] = peak{idx: i, ver: ver}
		}
	}

	drop := map[int]int{}
	for base, p := range peaks {
		if counts[base] < 2 {
			continue
		}
		drop[p.idx] = p.ver
	}

	out := make([]visibility.MatchedRepo, 0, len(in))
	for i, m := range in {
		if v, ok := drop[i]; ok {
			fmt.Fprintf(w, constants.MsgBulkExceptDropFmt, m.RepoName, v)

			continue
		}
		out = append(out, m)
	}

	return out
}

// parseVersionedName returns (base, version, true) when `name` ends
// in `-v<digits>`, or ("", 0, false) otherwise.
func parseVersionedName(name string) (string, int, bool) {
	m := versionSuffixRE.FindStringSubmatch(name)
	if m == nil {
		return "", 0, false
	}
	v, err := strconv.Atoi(m[2])
	if err != nil {
		return "", 0, false
	}

	return m[1], v, true
}
