package status

// needs, --full's checks and --active (step 5.1.6; contract §12, §13; spec §5, §7).

import (
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// need is one needs entry: all five fields, null where it does not apply (bonsai.status/1).
func need(kind string, id, name, source, version any) schema.Object {
	return schema.Object{{Key: "kind", Value: kind}, {Key: "id", Value: id}, {Key: "name", Value: name},
		{Key: "source", Value: source}, {Key: "version", Value: version}}
}

// needsOf lists what the workspace needs from a machine: the Claude Code floor (kind tool, never a problem), then each
// locked pack, kind plugin when Claude Code reported its plugin not installed for this checkout (installed, from
// --full; nil when not asked), else kind pack. lock may be nil (no pack locked).
func needsOf(lock *workspace.Lock, installed map[string]bool) []any {
	floor, _ := engine.Floor(lock)
	out := []any{need("tool", nil, "claude-code", nil, ">="+floor)}
	if lock == nil {
		return out
	}
	for _, lp := range lock.Packs {
		kind := "pack"
		if have, asked := installed[lp.ID]; asked && !have {
			kind = "plugin"
		}
		out = append(out, need(kind, lp.ID, nil, lp.Source, lp.Version))
	}
	return out
}

// fullChecks are --full's checks (contract §12: "newer pack versions, the agent's version, MCP servers reachable"):
//
//	packs        each locked pack: id, version (locked), ref (bonsai.yaml's), newer (its source's release tags past the
//	             locked version, oldest first; [] for none), why (null, or why the tags could not be read)
//	plugins      each locked pack's plugin as Claude Code reports it for this checkout: id, plugin, installed (true,
//	             false, or null when Claude Code could not be asked), why
//	claude_code  version (read, or null), floor, from (bonsai or a pack's id), state (ok, old or unknown), why
//	mcp          the MCP servers the packs' needs name, each with whether it is reachable as far as Bonsai can tell
//	             without starting a session, unknown where it cannot: none in formats set 4 (bonsai.pack/1's needs holds
//	             only claude_code), so []
//
// It gives the packs' plugins installed, by pack id, for needs (nil when Claude Code could not be asked).
func fullChecks(r *engine.CheckResult, cfg *workspace.Config, lock *workspace.Lock, o Options) (schema.Object, map[string]bool) {
	tags := o.Tags
	if tags == nil {
		tags = engine.RemoteTags
	}
	refs := map[string]string{}
	for _, p := range cfg.Packs {
		refs[p.ID] = p.Ref
	}
	packs, plugins := []any{}, []any{}
	installed, perr := engine.PluginsInstalled(r, o.Plugins)
	if perr != nil {
		installed = nil
	}
	market := ""
	if lock != nil && len(lock.Packs) > 0 {
		market = engine.LockMarket(cfg, lock)
	}
	if lock != nil {
		for _, lp := range lock.Packs {
			newer, why := []string{}, any(nil)
			if list, err := tags(lp.Source); err != nil {
				why = ascii(err.Error())
			} else {
				newer = engine.NewerTags(list, refs[lp.ID], lp.ID, lp.Version)
			}
			nv := make([]any, len(newer))
			for i, t := range newer {
				nv[i] = t
			}
			packs = append(packs, schema.Object{{Key: "id", Value: lp.ID}, {Key: "version", Value: lp.Version},
				{Key: "ref", Value: refs[lp.ID]}, {Key: "newer", Value: nv}, {Key: "why", Value: why}})
			var have, pwhy any
			if installed != nil {
				have = installed[lp.ID]
			} else {
				pwhy = "Claude Code could not be asked: " + ascii(perr.Error())
			}
			plugins = append(plugins, schema.Object{{Key: "id", Value: lp.ID}, {Key: "plugin", Value: engine.PluginID(lp.ID, market)},
				{Key: "installed", Value: have}, {Key: "why", Value: pwhy}})
		}
	}
	ask := o.Claude
	if ask == nil {
		ask = func() (string, error) { return "", engine.ErrNoClaude }
	}
	st := engine.JudgeClaude(lock, ask)
	var version, why any
	if st.Version != "" {
		version = st.Version
	}
	if st.Why != "" {
		why = st.Why
	}
	claude := schema.Object{{Key: "version", Value: version}, {Key: "floor", Value: st.Floor}, {Key: "from", Value: st.From},
		{Key: "state", Value: st.State}, {Key: "why", Value: why}}
	return schema.Object{{Key: "packs", Value: packs}, {Key: "plugins", Value: plugins}, {Key: "claude_code", Value: claude},
		{Key: "mcp", Value: []any{}}}, installed
}

// Active is status --active (contract §13's read-only command; spec §4): the active task for the checkout holding
// dir and this process's environment, as status --json's active_task holds it. It reads only what the answer needs:
// the checkout, bonsai.yaml and the task files, never the lock, the network or Claude Code. Its error is Bonsai not
// reading the workspace at all (exit 3, as status's).
func Active(dir string) (workspace.ActiveTask, *engine.Error) {
	co, err := workspace.Find(dir)
	if err != nil {
		return workspace.ActiveTask{}, engine.FindError(err)
	}
	cfg, err := workspace.LoadConfigFull(co.Root)
	if err != nil {
		return workspace.ActiveTask{}, engine.ConfigError(err)
	}
	return activeTask(co, cfg), nil
}

// ActiveSchema is bonsai.status/1's active_task: what status --active --json prints is held to it.
func ActiveSchema() schema.Object {
	props, _ := Schema().Get("properties")
	at, _ := props.(schema.Object).Get("active_task")
	return at.(schema.Object)
}
