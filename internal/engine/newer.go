package engine

// Newer releases of a locked pack (contract §12: status --full's checks list "newer pack versions"): the source's tags,
// read with git ls-remote (the network; no fetch, nothing written), and those that name a version newer than the
// locked one. A release tag names the pack's version (spec §5: vX.Y.Z, or a prefix and vX.Y.Z such as base-v1.0.0 for
// a repository of several packs); only tags with the prefix bonsai.yaml's ref has (or, for a ref that is a commit, no
// prefix or the pack's id and a dash) count, so another pack's releases in the same repository are not taken for this
// one's.

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"
)

// lsRemoteTimeout is the longest git ls-remote may take.
const lsRemoteTimeout = 30 * time.Second

// RemoteTags lists a source's tag names with git ls-remote, prompts off.
func RemoteTags(source string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), lsRemoteTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-c", "protocol.ext.allow=never", "ls-remote", "--tags", "--refs", "--", source)
	cmd.Env = gitEnv()
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return nil, errors.New("git ls-remote took longer than " + lsRemoteTimeout.String())
	}
	if err != nil {
		msg := "git ls-remote failed"
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			msg += ": " + firstLine(string(ee.Stderr))
		}
		return nil, errors.New(msg)
	}
	var tags []string
	for _, line := range strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n") {
		if i := strings.Index(line, "\trefs/tags/"); i >= 0 {
			tags = append(tags, line[i+len("\trefs/tags/"):])
		}
	}
	sort.Strings(tags)
	return tags, nil
}

var releaseTag = regexp.MustCompile(`^(.*?)v?([0-9]+\.[0-9]+\.[0-9]+)$`)

// NewerTags gives the tags that name a version newer than the locked one, oldest first: tags with ref's prefix when
// ref is a release tag, else with none or "<id>-".
func NewerTags(tags []string, ref, id, version string) []string {
	locked, _, ok := ParseVersion(version)
	if !ok {
		return []string{}
	}
	prefixes := map[string]bool{"": true, id + "-": true, id + "/": true}
	if m := releaseTag.FindStringSubmatch(ref); m != nil {
		prefixes = map[string]bool{m[1]: true}
	}
	type tagged struct {
		tag string
		v   [3]int
	}
	var newer []tagged
	for _, t := range tags {
		m := releaseTag.FindStringSubmatch(t)
		if m == nil || !prefixes[m[1]] {
			continue
		}
		v, _, ok := ParseVersion(m[2])
		if ok && older(locked, v) {
			newer = append(newer, tagged{t, v})
		}
	}
	sort.SliceStable(newer, func(i, j int) bool { return older(newer[i].v, newer[j].v) })
	out := []string{}
	for _, n := range newer {
		out = append(out, n.tag)
	}
	return out
}
