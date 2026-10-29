package engine

// The packs' newer releases, kept (T-6009's plan, step 4): the home's cache/packs.json, one record per pack source,
// so plain bonsai status can say whether a pack has a newer release without touching the network.
//
//   - What is read: a source's tags with git ls-remote (RemoteTags, given PackTagsTimeout), by bonsai pack outdated
//     and bonsai status --full, which both read every locked pack's source live and then write the record
//     (RecordPackTags). Plain status, check and every other reader only load the file (LoadPackCache).
//   - What is kept: every tag the source holds (git ls-remote returns all of a repository's tags, and several packs
//     may share one repository), when that read worked, and the last read's failure after it. Never a verdict: which
//     tags count for a pack (NewerTags' prefix rules) and whether one is newer than the locked version are worked out
//     each time the record is read (PackNewerOf), so moving the lock never leaves a stale "newer" behind.
//   - A write that fails never fails its caller: the caller says so in one line.
//
// cache/packs.json, Bonsai's own file, read by Bonsai alone (formats/README.md: not given a schema; bonsai.status/1's
// packs[].newer is its public face). Its fields:
//
//	sources  one record per source read, sorted by source:
//	  source  the source as the lock names it, "https://github.com/LastStep/Bonsai.git"
//	  tags    every tag name the last read that worked found, sorted, ["base-v1.0.0", "base-v1.0.1"]; [] for none
//	  read    when that read worked, RFC 3339 in UTC, "2026-10-29T08:15:00Z", or null when none has
//	  failed  the last read, when it failed after the last that worked: {"at": "...", "why": "git ls-remote failed:
//	          ..."}, else null

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// PackCacheFile is the home's record of the packs' sources' tags, relative to the home.
const PackCacheFile = "cache/packs.json"

// PackStale is how old a source's last good read may be before its pack's newer state reads stale.
const PackStale = 7 * 24 * time.Hour

// The newer states, bonsai.status/1's packs[].newer.state (a closed list, its one home the schema).
const (
	PackNewer   = "newer"   // a release tag newer than the locked version
	PackCurrent = "current" // none newer, the record fresh
	PackStaleS  = "stale"   // the record older than PackStale, or its newest tag below the locked version
)

// PackSourceRecord is one source's record in cache/packs.json (this file's comment).
type PackSourceRecord struct {
	Source string          `json:"source"`
	Tags   []string        `json:"tags"`
	Read   *string         `json:"read"`
	Failed *ReleaseFailure `json:"failed"`
}

// PackCache is cache/packs.json.
type PackCache struct {
	Sources []PackSourceRecord `json:"sources"`
}

// PackCachePath is the record's full path in the home, forward slashes.
func PackCachePath(home string) string {
	return filepath.ToSlash(filepath.Join(home, filepath.FromSlash(PackCacheFile)))
}

// LoadPackCache reads the home's record: an empty cache when there is none, an error when it is not one Bonsai
// reads (a time that is not RFC 3339, an unknown field), so nothing in a broken file reaches status.
func LoadPackCache(home string) (PackCache, error) {
	raw, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(PackCacheFile)))
	if errors.Is(err, fs.ErrNotExist) {
		return PackCache{}, nil
	}
	if err != nil {
		return PackCache{}, err
	}
	var c PackCache
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return PackCache{}, errors.New(PackCachePath(home) + " is not a record Bonsai reads")
	}
	for _, s := range c.Sources {
		for _, t := range []*string{s.Read, failedAt(s.Failed)} {
			if t != nil {
				if _, err := time.Parse(time.RFC3339, *t); err != nil {
					return PackCache{}, errors.New(PackCachePath(home) + " holds a time that is not RFC 3339")
				}
			}
		}
	}
	return c, nil
}

// Record is the source's record, nil when the cache holds none.
func (c PackCache) Record(source string) *PackSourceRecord {
	for i := range c.Sources {
		if c.Sources[i].Source == source {
			return &c.Sources[i]
		}
	}
	return nil
}

// PackTagRead is one live read of a source's tags: the tags, or the error that stopped it.
type PackTagRead struct {
	Source string
	Tags   []string
	Err    error
}

// RecordPackTags writes reads into the home's record, one per source: a read that worked replaces the source's tags
// and its time; one that failed keeps the last that worked and records the failure. Sources not read are kept as
// they are. It gives the cache as it now stands and why it could not be written (never a failure of the caller).
func RecordPackTags(home string, reads []PackTagRead) (PackCache, error) {
	c, err := LoadPackCache(home)
	if err != nil {
		c = PackCache{} // a broken record is replaced by this run's reads
	}
	stamp := releaseNow().UTC().Format(time.RFC3339)
	for _, r := range reads {
		rec := c.Record(r.Source)
		if rec == nil {
			c.Sources = append(c.Sources, PackSourceRecord{Source: r.Source, Tags: []string{}})
			rec = &c.Sources[len(c.Sources)-1]
		}
		if r.Err != nil {
			rec.Failed = &ReleaseFailure{At: stamp, Why: ascii(oneLine(r.Err.Error()))}
			continue
		}
		tags := append([]string{}, r.Tags...)
		sort.Strings(tags)
		s := stamp
		rec.Tags, rec.Read, rec.Failed = tags, &s, nil
	}
	sort.SliceStable(c.Sources, func(i, j int) bool { return c.Sources[i].Source < c.Sources[j].Source })
	if c.Sources == nil {
		c.Sources = []PackSourceRecord{}
	}
	raw, err := json.MarshalIndent(c, "", "  ")
	if err == nil {
		err = workspace.WriteFileAtomic(filepath.Join(home, filepath.FromSlash(PackCacheFile)), append(raw, '\n'))
	}
	return c, err
}

// PackNewerState is one pack's newer record as status --json's packs[].newer holds it: the newest tag the record
// holds for the pack by NewerTags' rules (or the locked ref itself when none is newer), when the source was last read
// successfully, and the state.
type PackNewerState struct {
	Tag   string
	Read  string
	State string
}

// PackNewerOf is a locked pack's newer state from the cache, worked out now against the locked version: nil when the
// cache holds no good read of its source (none made, or only failed ones). ref and path are bonsai.yaml's.
func PackNewerOf(c PackCache, source, id, ref, path, version string) *PackNewerState {
	rec := c.Record(source)
	if rec == nil || rec.Read == nil {
		return nil
	}
	out := &PackNewerState{Tag: ref, Read: *rec.Read, State: PackCurrent}
	if newer := NewerTags(rec.Tags, ref, id, path, version); len(newer) > 0 {
		out.Tag, out.State = newer[len(newer)-1], PackNewer
		return out
	}
	read, _ := time.Parse(time.RFC3339, *rec.Read)
	if releaseNow().Sub(read) > PackStale {
		out.State = PackStaleS
		return out
	}
	// The newest release tag the record holds for this pack sorts below the locked version: the record was read
	// before the lock moved, so it cannot say what is newest now.
	if locked, _, ok := ParseVersion(version); ok {
		if tags := ReleaseTagsFor(rec.Tags, ref, id, path); len(tags) > 0 {
			if newest, _, ok := ParseVersion(releaseTag().FindStringSubmatch(tags[len(tags)-1])[2]); ok && older(newest, locked) {
				out.State = PackStaleS
			}
		}
	}
	return out
}
