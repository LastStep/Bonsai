package nonint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
)

// remove_test.go exercises the headless remove cores (RunRemoveAgent /
// RunRemoveItem) directly: typed options in, (*Result, exitCode, error) out.
// These are the Plan 41 Phase 3 negative controls + security guards.

// initWithAgents initialises a project (tech-lead) then adds each named extra
// agent through RunAdd. Returns the project root and its config path. Every
// agent installs its catalog defaults, so backend + security both end up with
// the (non-required) coding-standards skill — the multi-owner fixture.
func initWithAgents(t *testing.T, extras ...string) (string, string) {
	t.Helper()
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".bonsai.yaml")

	cfg := minimalInitCfg(t, tmp)
	if _, code, err := RunInit(tmp, configPath, cfg, cat, "test"); err != nil || code != ExitOK {
		t.Fatalf("RunInit setup: code=%d err=%v", code, err)
	}
	for _, a := range extras {
		overlay, err := LoadConfig(writeYAML(t, tmp, a+"-overlay.yaml", "agents:\n  "+a+": {}\n"), tmp, cat)
		if err != nil {
			t.Fatalf("LoadConfig %s overlay: %v", a, err)
		}
		if _, code, err := RunAdd(tmp, configPath, overlay, cat, "test"); err != nil || code != ExitOK {
			t.Fatalf("RunAdd %s setup: code=%d err=%v", a, code, err)
		}
	}
	return tmp, configPath
}

func reloadConfig(t *testing.T, configPath string) *config.ProjectConfig {
	t.Helper()
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	return cfg
}

// ─── Agent removal: tech-lead guard ─────────────────────────────────────

// TestRunRemoveAgent_TechLeadInUse_Exit2: with backend installed alongside
// tech-lead, removing tech-lead must exit 2 and the error must contain
// "tech-lead". Removing backend first, then tech-lead, must succeed (exit 0).
func TestRunRemoveAgent_TechLeadInUse_Exit2(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp, configPath := initWithAgents(t, "backend")

	cfg := reloadConfig(t, configPath)
	lock, _ := config.LoadLockFile(tmp)
	_, code, err := RunRemoveAgent(tmp, cfg, cat, lock, "test", "tech-lead", false)
	if code != ExitInvalidConfig {
		t.Fatalf("remove tech-lead in use: want exit %d, got %d (err=%v)", ExitInvalidConfig, code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "tech-lead") {
		t.Errorf("error must contain tech-lead; got %v", err)
	}

	// Remove backend first.
	cfg = reloadConfig(t, configPath)
	lock, _ = config.LoadLockFile(tmp)
	if _, code, err := RunRemoveAgent(tmp, cfg, cat, lock, "test", "backend", false); err != nil || code != ExitOK {
		t.Fatalf("remove backend: code=%d err=%v", code, err)
	}

	// Now tech-lead is the sole agent — removal must succeed.
	cfg = reloadConfig(t, configPath)
	lock, _ = config.LoadLockFile(tmp)
	if _, code, err := RunRemoveAgent(tmp, cfg, cat, lock, "test", "tech-lead", false); err != nil || code != ExitOK {
		t.Fatalf("remove sole tech-lead: want exit 0, got %d (err=%v)", code, err)
	}
	post := reloadConfig(t, configPath)
	if _, ok := post.Agents["tech-lead"]; ok {
		t.Errorf("tech-lead still registered after removal")
	}
}

// TestRunRemoveAgent_NotInstalled_Exit2: removing an agent that isn't in the
// project → exit 2.
func TestRunRemoveAgent_NotInstalled_Exit2(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp, configPath := initWithAgents(t)
	cfg := reloadConfig(t, configPath)
	lock, _ := config.LoadLockFile(tmp)

	_, code, err := RunRemoveAgent(tmp, cfg, cat, lock, "test", "backend", false)
	if code != ExitInvalidConfig {
		t.Errorf("not-installed: want exit %d, got %d (err=%v)", ExitInvalidConfig, code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("error must mention not installed; got %v", err)
	}
}

// ─── Safety: empty / wildcard targets ───────────────────────────────────

// TestRunRemoveAgent_EmptyAndWildcard_Exit2: empty and "*" targets are
// rejected with exit 2 and zero filesystem mutation (config unchanged).
func TestRunRemoveAgent_EmptyAndWildcard_Exit2(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp, configPath := initWithAgents(t, "backend")
	before := mustReadFile(t, configPath)

	for _, bad := range []string{"", "   ", "*"} {
		cfg := reloadConfig(t, configPath)
		lock, _ := config.LoadLockFile(tmp)
		_, code, err := RunRemoveAgent(tmp, cfg, cat, lock, "test", bad, true)
		if code != ExitInvalidConfig {
			t.Errorf("target %q: want exit %d, got %d (err=%v)", bad, ExitInvalidConfig, code, err)
		}
		if err == nil {
			t.Errorf("target %q: expected error", bad)
		}
	}
	if after := mustReadFile(t, configPath); after != before {
		t.Errorf("config mutated by a rejected unsafe target.\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestRunRemoveItem_EmptyAndWildcard_Exit2(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp, configPath := initWithAgents(t, "backend")
	before := mustReadFile(t, configPath)

	for _, bad := range []string{"", "   ", "*"} {
		cfg := reloadConfig(t, configPath)
		lock, _ := config.LoadLockFile(tmp)
		_, code, err := RunRemoveItem(tmp, cfg, cat, lock, "test", "skill", bad, "")
		if code != ExitInvalidConfig {
			t.Errorf("item target %q: want exit %d, got %d (err=%v)", bad, ExitInvalidConfig, code, err)
		}
	}
	if after := mustReadFile(t, configPath); after != before {
		t.Errorf("config mutated by a rejected unsafe item target")
	}
}

// ─── Symlink refusal under --delete-files ───────────────────────────────

// TestRunRemoveAgent_SymlinkDeleteTarget_Refused: replace EACH of the three
// delete targets (agentDir / CLAUDE.md / .claude) with a symlink in turn and
// assert --delete-files refuses with exit 2 and ZERO deletion. The symlink
// (and its target) must survive.
func TestRunRemoveAgent_SymlinkDeleteTarget_Refused(t *testing.T) {
	targets := []struct {
		name string
		rel  string
	}{
		{"agentDir", filepath.Join("agent")},
		{"CLAUDE.md", "CLAUDE.md"},
		{".claude", ".claude"},
	}
	for _, tc := range targets {
		t.Run(tc.name, func(t *testing.T) {
			cat := loadTestCatalog(t)
			tmp, configPath := initWithAgents(t, "backend")
			cfg := reloadConfig(t, configPath)
			ws := cfg.Agents["backend"].Workspace

			// Build the symlink: point it at a sentinel outside the delete
			// target so we can prove the link's TARGET wasn't followed/deleted.
			sentinelDir := t.TempDir()
			sentinelFile := filepath.Join(sentinelDir, "DO-NOT-DELETE")
			if err := os.WriteFile(sentinelFile, []byte("keep me"), 0o644); err != nil {
				t.Fatalf("write sentinel: %v", err)
			}

			linkPath := filepath.Join(tmp, ws, tc.rel)
			// Remove whatever the real target is, then symlink it to the sentinel.
			_ = os.RemoveAll(linkPath)
			linkDest := sentinelDir
			if tc.rel == "CLAUDE.md" {
				linkDest = sentinelFile
			}
			if err := os.Symlink(linkDest, linkPath); err != nil {
				t.Fatalf("symlink %s: %v", linkPath, err)
			}

			lock, _ := config.LoadLockFile(tmp)
			_, code, err := RunRemoveAgent(tmp, cfg, cat, lock, "test", "backend", true)
			if code != ExitInvalidConfig {
				t.Fatalf("symlinked %s: want exit %d, got %d (err=%v)", tc.name, ExitInvalidConfig, code, err)
			}
			if err == nil || !strings.Contains(err.Error(), "symlink") {
				t.Errorf("error must mention symlink; got %v", err)
			}
			// The symlink itself must still exist (zero deletion).
			if _, lerr := os.Lstat(linkPath); lerr != nil {
				t.Errorf("symlink was deleted despite refusal: %v", lerr)
			}
			// The sentinel (the link's target) must be untouched.
			if _, serr := os.Stat(sentinelFile); serr != nil {
				t.Errorf("symlink target was followed and deleted: %v", serr)
			}
			// The agent must remain registered (config not mutated on the
			// reject path — the symlink check fires before any mutation).
			post := reloadConfig(t, configPath)
			if _, ok := post.Agents["backend"]; !ok {
				t.Errorf("backend de-registered despite the symlink refusal (mutation leaked before the guard)")
			}
		})
	}
}

// ─── Item removal: multi-owner disambiguation ───────────────────────────

// TestRunRemoveItem_MultiOwner_NeedsFrom: coding-standards is a default skill
// of both backend and security (and required for neither). Removing it with
// no --from must exit 2 and the message must NAME both owners. With
// --from backend it must succeed (exit 0) and remove it from backend only,
// leaving it on security.
func TestRunRemoveItem_MultiOwner_NeedsFrom(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp, configPath := initWithAgents(t, "backend", "security")

	// Sanity: both own coding-standards.
	pre := reloadConfig(t, configPath)
	if !hasSkill(pre, "backend", "coding-standards") || !hasSkill(pre, "security", "coding-standards") {
		t.Fatalf("fixture invalid: backend=%v security=%v",
			pre.Agents["backend"].Skills, pre.Agents["security"].Skills)
	}

	// No --from → exit 2, message names both owners.
	cfg := reloadConfig(t, configPath)
	lock, _ := config.LoadLockFile(tmp)
	_, code, err := RunRemoveItem(tmp, cfg, cat, lock, "test", "skill", "coding-standards", "")
	if code != ExitInvalidConfig {
		t.Fatalf("multi-owner no --from: want exit %d, got %d (err=%v)", ExitInvalidConfig, code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "backend") || !strings.Contains(err.Error(), "security") {
		t.Errorf("error must name both owners (backend, security); got %v", err)
	}

	// --from backend → exit 0, removed from backend only.
	cfg = reloadConfig(t, configPath)
	lock, _ = config.LoadLockFile(tmp)
	if _, code, err := RunRemoveItem(tmp, cfg, cat, lock, "test", "skill", "coding-standards", "backend"); err != nil || code != ExitOK {
		t.Fatalf("--from backend: want exit 0, got %d (err=%v)", code, err)
	}
	post := reloadConfig(t, configPath)
	if hasSkill(post, "backend", "coding-standards") {
		t.Errorf("coding-standards still on backend after --from backend removal")
	}
	if !hasSkill(post, "security", "coding-standards") {
		t.Errorf("coding-standards wrongly removed from security (should be scoped to backend)")
	}
}

// TestRunRemoveItem_FromAgentNotOwner_Exit2: --from naming an agent that does
// not own the item → exit 2, zero mutation.
func TestRunRemoveItem_FromAgentNotOwner_Exit2(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp, configPath := initWithAgents(t, "backend", "security")
	before := mustReadFile(t, configPath)

	// auth-patterns is a security default but NOT a backend default — scoping
	// to backend must fail.
	cfg := reloadConfig(t, configPath)
	lock, _ := config.LoadLockFile(tmp)
	_, code, err := RunRemoveItem(tmp, cfg, cat, lock, "test", "skill", "auth-patterns", "backend")
	if code != ExitInvalidConfig {
		t.Fatalf("--from non-owner: want exit %d, got %d (err=%v)", ExitInvalidConfig, code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "backend") {
		t.Errorf("error must name the requested agent; got %v", err)
	}
	if after := mustReadFile(t, configPath); after != before {
		t.Errorf("config mutated by a rejected --from non-owner")
	}
}

// ─── Required-item protection (filterRequired on the --from branch) ──────

// requiredProtocol is an installed, required-for-all protocol (session-start /
// security / scope-boundaries are all required:all and present in every
// agent's defaults). Used by the required-item controls.
const requiredProtocol = "security"

// TestRunRemoveItem_RequiredItem_NoFrom_Exit2: the security protocol is
// required for all agents. Removing it with no --from → exit 2, zero mutation.
func TestRunRemoveItem_RequiredItem_NoFrom_Exit2(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp, configPath := initWithAgents(t)
	before := mustReadFile(t, configPath)

	cfg := reloadConfig(t, configPath)
	lock, _ := config.LoadLockFile(tmp)
	_, code, err := RunRemoveItem(tmp, cfg, cat, lock, "test", "protocol", requiredProtocol, "")
	if code != ExitInvalidConfig {
		t.Fatalf("required item no --from: want exit %d, got %d (err=%v)", ExitInvalidConfig, code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Errorf("error must mention required; got %v", err)
	}
	if after := mustReadFile(t, configPath); after != before {
		t.Errorf("config mutated removing a required item")
	}
}

// TestRunRemoveItem_RequiredItem_WithFrom_Exit2 is the H1 control: --from must
// NOT bypass required-protection. The security protocol is required for
// tech-lead — removing it with --from tech-lead → exit 2, ZERO filesystem
// mutation (config + the generated protocol file both unchanged).
func TestRunRemoveItem_RequiredItem_WithFrom_Exit2(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp, configPath := initWithAgents(t)
	before := mustReadFile(t, configPath)

	cfg := reloadConfig(t, configPath)
	ws := cfg.Agents["tech-lead"].Workspace
	protoFile := filepath.Join(tmp, ws, "agent", "Protocols", requiredProtocol+".md")
	if _, err := os.Stat(protoFile); err != nil {
		t.Fatalf("fixture invalid: %s protocol file missing: %v", requiredProtocol, err)
	}

	lock, _ := config.LoadLockFile(tmp)
	_, code, err := RunRemoveItem(tmp, cfg, cat, lock, "test", "protocol", requiredProtocol, "tech-lead")
	if code != ExitInvalidConfig {
		t.Fatalf("required item --from: want exit %d, got %d (err=%v)", ExitInvalidConfig, code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "required") {
		t.Errorf("error must mention required (filterRequired not bypassed by --from); got %v", err)
	}
	// Zero FS mutation: config + the protocol file both unchanged.
	if after := mustReadFile(t, configPath); after != before {
		t.Errorf("config mutated removing a required item via --from")
	}
	if _, err := os.Stat(protoFile); err != nil {
		t.Errorf("generated protocol file deleted despite required refusal: %v", err)
	}
}

// ─── routine-check auto-managed sensor block ────────────────────────────

// TestRunRemoveItem_RoutineCheck_Blocked: routine-check is auto-managed and
// cannot be removed directly → exit 2.
func TestRunRemoveItem_RoutineCheck_Blocked(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp, configPath := initWithAgents(t)
	cfg := reloadConfig(t, configPath)
	lock, _ := config.LoadLockFile(tmp)

	_, code, err := RunRemoveItem(tmp, cfg, cat, lock, "test", "sensor", "routine-check", "")
	if code != ExitInvalidConfig {
		t.Errorf("routine-check: want exit %d, got %d (err=%v)", ExitInvalidConfig, code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "auto-managed") {
		t.Errorf("error must mention auto-managed; got %v", err)
	}
}

// TestRunRemoveItem_NotInstalled_Exit2: a phantom item → exit 2.
func TestRunRemoveItem_NotInstalled_Exit2(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp, configPath := initWithAgents(t)
	cfg := reloadConfig(t, configPath)
	lock, _ := config.LoadLockFile(tmp)

	_, code, err := RunRemoveItem(tmp, cfg, cat, lock, "test", "skill", "does-not-exist", "")
	if code != ExitInvalidConfig {
		t.Errorf("phantom item: want exit %d, got %d (err=%v)", ExitInvalidConfig, code, err)
	}
	if err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("error must mention not installed; got %v", err)
	}
}

// TestRunRemoveItem_SingleOwner_Success: a non-required skill owned by exactly
// one agent removes cleanly without --from.
func TestRunRemoveItem_SingleOwner_Success(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp, configPath := initWithAgents(t, "backend")

	// database-conventions is a backend default, not required, and tech-lead
	// does NOT default-install it — single owner.
	pre := reloadConfig(t, configPath)
	if !hasSkill(pre, "backend", "database-conventions") {
		t.Skip("fixture: backend lacks database-conventions default")
	}
	if hasSkill(pre, "tech-lead", "database-conventions") {
		t.Skip("fixture: database-conventions unexpectedly shared with tech-lead")
	}

	cfg := reloadConfig(t, configPath)
	lock, _ := config.LoadLockFile(tmp)
	if _, code, err := RunRemoveItem(tmp, cfg, cat, lock, "test", "skill", "database-conventions", ""); err != nil || code != ExitOK {
		t.Fatalf("single-owner removal: want exit 0, got %d (err=%v)", code, err)
	}
	post := reloadConfig(t, configPath)
	if hasSkill(post, "backend", "database-conventions") {
		t.Errorf("database-conventions still on backend after removal")
	}
}

// ─── small test helpers ─────────────────────────────────────────────────

func hasSkill(cfg *config.ProjectConfig, agent, skill string) bool {
	a := cfg.Agents[agent]
	if a == nil {
		return false
	}
	for _, s := range a.Skills {
		if s == skill {
			return true
		}
	}
	return false
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}
