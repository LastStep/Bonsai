package engine

// What bonsai pack reads of a pack beyond the lock (T-6009's plan, step 7): a new pack's id from its pack.yaml, for
// pack add (bonsai.yaml names a pack by its id, which only the pack itself says), and the skills a locked pack brings,
// for pack outdated --json (bonsai.packs/1's skills: the lock does not hold them), read from the home's clone.

import (
	"path"
	"sort"
	"strings"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// PackIDAt fetches source at ref into the home's cache and gives the id its bonsai/pack.yaml names there, the pack
// kept in its folder p ("" for the repository's top). Its refusals are fetch's and the pack reader's.
func PackIDAt(home, source, p, ref string) (string, error) {
	c := cache{home: home}
	commit, err := c.fetch(source, ref)
	if err != nil {
		return "", err
	}
	pd, err := c.packAt(workspace.PackRef{Source: source, Path: p, Ref: ref}, commit)
	if err != nil {
		return "", err
	}
	return pd.Manifest.ID, nil
}

// PackSkills gives the names of the skills a pack brings at a commit (each folder under its skills/ holding a
// SKILL.md), sorted, from the home's clone alone, never the network; ok false when the clone does not hold the commit.
func PackSkills(home, source, p, commit string) (skills []string, ok bool) {
	c := cache{home: home}
	if !commitPattern().MatchString(commit) || !c.has(source, commit) {
		return nil, false
	}
	prefix := "skills/"
	if p != "" {
		prefix = path.Clean(p) + "/skills/"
	}
	out, err := git(c.dir(source), nil, "ls-tree", "-r", "-z", "--name-only", "--full-tree", commit, "--", prefix)
	if err != nil {
		return nil, false
	}
	seen := map[string]bool{}
	skills = []string{}
	for _, name := range strings.Split(string(out), "\x00") {
		rest, found := strings.CutPrefix(name, prefix)
		if !found {
			continue
		}
		if dir, file, cut := strings.Cut(rest, "/"); cut && file == "SKILL.md" && !seen[dir] {
			seen[dir] = true
			skills = append(skills, dir)
		}
	}
	sort.Strings(skills)
	return skills, true
}
