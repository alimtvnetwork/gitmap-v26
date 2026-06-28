// Package cmd — `gitmap orphans`: find local clones whose remote no
// longer exists (HTTP 404 on the origin URL) and offer bulk delete.
// v6.68.0.
package cmd

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type orphanRepo struct {
	path   string
	remote string
	status int
}

// runOrphans executes `gitmap orphans`.
func runOrphans(args []string) {
	checkHelp("orphans", args)
	fs := flag.NewFlagSet("orphans", flag.ContinueOnError)
	root := fs.String("root", ".", "scan root directory")
	yes := fs.Bool("y", false, "delete without prompting")
	dryRun := fs.Bool("dry-run", false, "list only; do not delete")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	repos := scanForRepos(*root)
	var orphans []orphanRepo
	for _, r := range repos {
		remote, ok := originURL(r)
		if !ok {
			continue
		}
		status := remoteStatus(remote)
		if status == http.StatusNotFound || status == http.StatusGone {
			orphans = append(orphans, orphanRepo{path: r, remote: remote, status: status})
		}
	}
	if len(orphans) == 0 {
		fmt.Fprintln(os.Stdout, "\n  no orphaned clones found\n")

		return
	}
	fmt.Fprintf(os.Stdout, "\n  \033[36m%d orphan(s)\033[0m (remote returns 404/410)\n\n", len(orphans))
	for _, o := range orphans {
		fmt.Fprintf(os.Stdout, "  \033[31m%d\033[0m  %s  -> %s\n", o.status, o.path, o.remote)
	}
	fmt.Fprintln(os.Stdout, "")
	if *dryRun {
		return
	}
	if !*yes && !confirmYesNo(fmt.Sprintf("delete %d orphan(s)?", len(orphans))) {
		return
	}
	for _, o := range orphans {
		if err := os.RemoveAll(o.path); err != nil {
			fmt.Fprintf(os.Stderr, "  \033[31mfailed\033[0m %s: %v\n", o.path, err)

			continue
		}
		fmt.Fprintf(os.Stdout, "  \033[32mdeleted\033[0m %s\n", o.path)
	}
}

// originURL returns the `origin` remote URL for repo at dir.
func originURL(dir string) (string, bool) {
	out, err := exec.Command("git", "-C", dir, "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "", false
	}
	u := strings.TrimSpace(string(out))

	return u, u != ""
}

// remoteStatus probes the remote URL via HTTP HEAD.
func remoteStatus(remote string) int {
	web := gitURLToHTTPS(remote)
	if web == "" {
		return 0
	}
	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest(http.MethodHead, web, nil)
	if err != nil {
		return 0
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	return resp.StatusCode
}

// gitURLToHTTPS converts ssh/https git URLs into a browseable HTTPS URL.
func gitURLToHTTPS(u string) string {
	u = strings.TrimSuffix(u, ".git")
	if strings.HasPrefix(u, "git@") {
		// git@github.com:owner/repo -> https://github.com/owner/repo
		parts := strings.SplitN(strings.TrimPrefix(u, "git@"), ":", 2)
		if len(parts) != 2 {
			return ""
		}

		return "https://" + parts[0] + "/" + parts[1]
	}
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return u
	}

	return ""
}

// confirmYesNo prompts on stderr and reads a y/N answer from stdin.
func confirmYesNo(prompt string) bool {
	fmt.Fprintf(os.Stderr, "%s [y/N] ", prompt)
	var resp string
	if _, err := fmt.Fscanln(os.Stdin, &resp); err != nil {
		return false
	}
	resp = strings.ToLower(strings.TrimSpace(resp))

	return resp == "y" || resp == "yes"
}

// helper to silence unused import when filepath isn't referenced.
var _ = filepath.Join
