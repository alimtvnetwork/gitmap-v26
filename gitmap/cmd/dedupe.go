// Package cmd — `gitmap dedupe`: detect identical repos cloned under
// different folders by hashing each repo's HEAD tree SHA. v6.68.0.
package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// runDedupe executes `gitmap dedupe`.
func runDedupe(args []string) {
	checkHelp("dedupe", args)
	fs := flag.NewFlagSet("dedupe", flag.ContinueOnError)
	root := fs.String("root", ".", "scan root directory")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	repos := scanForRepos(*root)
	groups := groupByHeadTree(repos)
	dupes := filterDuplicateGroups(groups)
	printDedupeReport(dupes)
}

// groupByHeadTree maps `HEAD^{tree}` SHA -> list of repo paths.
func groupByHeadTree(repos []string) map[string][]string {
	out := map[string][]string{}
	for _, r := range repos {
		sha, ok := headTreeSHA(r)
		if !ok {
			continue
		}
		out[sha] = append(out[sha], r)
	}

	return out
}

// headTreeSHA returns the tree SHA pointed to by HEAD for repo at dir.
func headTreeSHA(dir string) (string, bool) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD^{tree}").Output()
	if err != nil {
		return "", false
	}
	s := strings.TrimSpace(string(out))

	return s, s != ""
}

// filterDuplicateGroups keeps only groups with 2+ entries.
func filterDuplicateGroups(groups map[string][]string) map[string][]string {
	out := map[string][]string{}
	for k, v := range groups {
		if len(v) > 1 {
			out[k] = v
		}
	}

	return out
}

// printDedupeReport renders the duplicate-group table.
func printDedupeReport(dupes map[string][]string) {
	if len(dupes) == 0 {
		fmt.Fprintln(os.Stdout, "\n  no duplicate repos found\n")

		return
	}
	keys := make([]string, 0, len(dupes))
	for k := range dupes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Fprintf(os.Stdout, "\n  \033[36m%d duplicate group(s)\033[0m (identical HEAD tree)\n\n", len(dupes))
	for _, k := range keys {
		fmt.Fprintf(os.Stdout, "  \033[1mtree %s\033[0m\n", k[:12])
		for _, p := range dupes[k] {
			fmt.Fprintf(os.Stdout, "    • %s\n", p)
		}
	}
	fmt.Fprintln(os.Stdout, "")
}
