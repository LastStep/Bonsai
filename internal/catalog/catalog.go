package catalog

import (
	"io/fs"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// DisplayNameFrom derives a human-readable display name from a kebab-case machine name.
// "scope-guard-files" → "Scope Guard Files"
func DisplayNameFrom(name string) string {
	var result strings.Builder
	capitalize := true
	for _, r := range name {
		if r == '-' {
			result.WriteRune(' ')
			capitalize = true
			continue
		}
		if capitalize {
			result.WriteRune(unicode.ToUpper(r))
			capitalize = false
		} else {
			result.WriteRune(r)
		}
		if r == ' ' {
			capitalize = true
		}
	}
	return result.String()
}

// AgentCompat handles the YAML "agents" field which can be "all" or a list of strings.
type AgentCompat struct {
	All   bool
	Names []string
}

func (a *AgentCompat) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode && value.Value == "all" {
		a.All = true
		return nil
	}
	return value.Decode(&a.Names)
}

func (a AgentCompat) CompatibleWith(agentType string) bool {
	if a.All {
		return true
	}
	for _, n := range a.Names {
		if n == agentType {
			return true
		}
	}
	return false
}

func (a AgentCompat) String() string {
	if a.All {
		return "all"
	}
	return strings.Join(a.Names, ", ")
}

// CatalogItem represents a skill, workflow, or protocol.
type CatalogItem struct {
	Name        string      `yaml:"name"`
	DisplayName string      `yaml:"display_name"`
	Description string      `yaml:"description"`
	Agents      AgentCompat `yaml:"agents"`
	Required    AgentCompat `yaml:"required"`
	ContentPath string      `yaml:"-"`
}

// SensorItem represents a sensor (hook).
type SensorItem struct {
	Name        string      `yaml:"name"`
	DisplayName string      `yaml:"display_name"`
	Description string      `yaml:"description"`
	Agents      AgentCompat `yaml:"agents"`
	Required    AgentCompat `yaml:"required"`
	Event       string      `yaml:"event"`
	Matcher     string      `yaml:"matcher,omitempty"`
	ContentPath string      `yaml:"-"`
}

// RoutineItem represents a periodic self-maintenance routine.
type RoutineItem struct {
	Name        string      `yaml:"name"`
	DisplayName string      `yaml:"display_name"`
	Description string      `yaml:"description"`
	Agents      AgentCompat `yaml:"agents"`
	Required    AgentCompat `yaml:"required"`
	Frequency   string      `yaml:"frequency"`
	ContentPath string      `yaml:"-"`
}

// ScaffoldingItem represents a project scaffolding group (e.g. logs, playbook, reports).
type ScaffoldingItem struct {
	Name        string   `yaml:"name"`
	DisplayName string   `yaml:"display_name"`
	Description string   `yaml:"description"`
	Required    bool     `yaml:"required"`
	Affects     string   `yaml:"affects"`
	Files       []string `yaml:"files"`
}

// SharedCoreDir is the path within the embedded FS to shared core files
// (memory, self-awareness). Agent-specific core/ dirs can override any shared file.
const SharedCoreDir = "core"

// AgentDef represents an agent type definition from the catalog.
type AgentDef struct {
	Name             string
	DisplayName      string
	Description      string
	DefaultSkills    []string
	DefaultWorkflows []string
	DefaultProtocols []string
	DefaultSensors   []string
	DefaultRoutines  []string
	CoreDir          string // path within FS to core/ directory
}

type agentYAML struct {
	Name        string `yaml:"name"`
	DisplayName string `yaml:"display_name"`
	Description string `yaml:"description"`
	Defaults    struct {
		Skills    []string `yaml:"skills"`
		Workflows []string `yaml:"workflows"`
		Protocols []string `yaml:"protocols"`
		Sensors   []string `yaml:"sensors"`
		Routines  []string `yaml:"routines"`
	} `yaml:"defaults"`
}

// Catalog holds all loaded catalog data with lookup helpers.
type Catalog struct {
	Agents      []AgentDef
	Skills      []CatalogItem
	Workflows   []CatalogItem
	Protocols   []CatalogItem
	Sensors     []SensorItem
	Routines    []RoutineItem
	Scaffolding []ScaffoldingItem

	fsys fs.FS

	skillsByName      map[string]*CatalogItem
	workflowsByName   map[string]*CatalogItem
	protocolsByName   map[string]*CatalogItem
	sensorsByName     map[string]*SensorItem
	routinesByName    map[string]*RoutineItem
	scaffoldingByName map[string]*ScaffoldingItem
}

// New loads the full catalog from an embedded filesystem.
func New(fsys fs.FS) (*Catalog, error) {
	c := &Catalog{fsys: fsys}

	c.Agents = loadAgents(fsys)
	c.Skills = loadItems(fsys, "skills")
	c.Workflows = loadItems(fsys, "workflows")
	c.Protocols = loadItems(fsys, "protocols")
	c.Sensors = loadSensors(fsys)
	c.Routines = loadRoutines(fsys)
	c.Scaffolding = loadScaffolding(fsys)

	c.skillsByName = make(map[string]*CatalogItem)
	for i := range c.Skills {
		c.skillsByName[c.Skills[i].Name] = &c.Skills[i]
	}
	c.workflowsByName = make(map[string]*CatalogItem)
	for i := range c.Workflows {
		c.workflowsByName[c.Workflows[i].Name] = &c.Workflows[i]
	}
	c.protocolsByName = make(map[string]*CatalogItem)
	for i := range c.Protocols {
		c.protocolsByName[c.Protocols[i].Name] = &c.Protocols[i]
	}
	c.sensorsByName = make(map[string]*SensorItem)
	for i := range c.Sensors {
		c.sensorsByName[c.Sensors[i].Name] = &c.Sensors[i]
	}
	c.routinesByName = make(map[string]*RoutineItem)
	for i := range c.Routines {
		c.routinesByName[c.Routines[i].Name] = &c.Routines[i]
	}
	c.scaffoldingByName = make(map[string]*ScaffoldingItem)
	for i := range c.Scaffolding {
		c.scaffoldingByName[c.Scaffolding[i].Name] = &c.Scaffolding[i]
	}

	return c, nil
}

func (c *Catalog) FS() fs.FS                                   { return c.fsys }
func (c *Catalog) GetSkill(name string) *CatalogItem           { return c.skillsByName[name] }
func (c *Catalog) GetWorkflow(name string) *CatalogItem        { return c.workflowsByName[name] }
func (c *Catalog) GetProtocol(name string) *CatalogItem        { return c.protocolsByName[name] }
func (c *Catalog) GetSensor(name string) *SensorItem           { return c.sensorsByName[name] }
func (c *Catalog) GetRoutine(name string) *RoutineItem         { return c.routinesByName[name] }
func (c *Catalog) GetScaffolding(name string) *ScaffoldingItem { return c.scaffoldingByName[name] }

func (c *Catalog) GetAgent(name string) *AgentDef {
	for i := range c.Agents {
		if c.Agents[i].Name == name {
			return &c.Agents[i]
		}
	}
	return nil
}

func (c *Catalog) GetItem(name string) *CatalogItem {
	if s := c.GetSkill(name); s != nil {
		return s
	}
	if w := c.GetWorkflow(name); w != nil {
		return w
	}
	return c.GetProtocol(name)
}

func (c *Catalog) SkillsFor(agentType string) []CatalogItem {
	return filterItems(c.Skills, agentType)
}
func (c *Catalog) WorkflowsFor(agentType string) []CatalogItem {
	return filterItems(c.Workflows, agentType)
}
func (c *Catalog) ProtocolsFor(agentType string) []CatalogItem {
	return filterItems(c.Protocols, agentType)
}
func (c *Catalog) SensorsFor(agentType string) []SensorItem {
	var result []SensorItem
	for _, s := range c.Sensors {
		if s.Agents.CompatibleWith(agentType) {
			result = append(result, s)
		}
	}
	return result
}
func (c *Catalog) RoutinesFor(agentType string) []RoutineItem {
	var result []RoutineItem
	for _, r := range c.Routines {
		if r.Agents.CompatibleWith(agentType) {
			result = append(result, r)
		}
	}
	return result
}

func filterItems(items []CatalogItem, agentType string) []CatalogItem {
	var result []CatalogItem
	for _, item := range items {
		if item.Agents.CompatibleWith(agentType) {
			result = append(result, item)
		}
	}
	return result
}

func loadItems(fsys fs.FS, category string) []CatalogItem {
	entries, err := fs.ReadDir(fsys, category)
	if err != nil {
		return nil
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var items []CatalogItem
	for _, name := range names {
		metaPath := category + "/" + name + "/meta.yaml"
		data, err := fs.ReadFile(fsys, metaPath)
		if err != nil {
			continue
		}

		var item CatalogItem
		if err := yaml.Unmarshal(data, &item); err != nil || item.Name == "" {
			continue
		}
		if item.DisplayName == "" {
			item.DisplayName = DisplayNameFrom(item.Name)
		}

		// Find content .md file
		itemDir := category + "/" + name
		dirEntries, err := fs.ReadDir(fsys, itemDir)
		if err != nil {
			continue
		}
		for _, f := range dirEntries {
			if !f.IsDir() && (strings.HasSuffix(f.Name(), ".md") || strings.HasSuffix(f.Name(), ".md.tmpl")) {
				item.ContentPath = itemDir + "/" + f.Name()
				break
			}
		}
		if item.ContentPath == "" {
			continue
		}

		items = append(items, item)
	}
	return items
}

func loadSensors(fsys fs.FS) []SensorItem {
	entries, err := fs.ReadDir(fsys, "sensors")
	if err != nil {
		return nil
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var sensors []SensorItem
	for _, name := range names {
		metaPath := "sensors/" + name + "/meta.yaml"
		data, err := fs.ReadFile(fsys, metaPath)
		if err != nil {
			continue
		}

		var sensor SensorItem
		if err := yaml.Unmarshal(data, &sensor); err != nil || sensor.Name == "" || sensor.Event == "" {
			continue
		}
		if sensor.DisplayName == "" {
			sensor.DisplayName = DisplayNameFrom(sensor.Name)
		}

		// Find script file
		itemDir := "sensors/" + name
		dirEntries, err := fs.ReadDir(fsys, itemDir)
		if err != nil {
			continue
		}
		for _, f := range dirEntries {
			if !f.IsDir() && f.Name() != "meta.yaml" {
				sensor.ContentPath = itemDir + "/" + f.Name()
				break
			}
		}
		if sensor.ContentPath == "" {
			continue
		}

		sensors = append(sensors, sensor)
	}
	return sensors
}

func loadRoutines(fsys fs.FS) []RoutineItem {
	entries, err := fs.ReadDir(fsys, "routines")
	if err != nil {
		return nil
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var routines []RoutineItem
	for _, name := range names {
		metaPath := "routines/" + name + "/meta.yaml"
		data, err := fs.ReadFile(fsys, metaPath)
		if err != nil {
			continue
		}

		var routine RoutineItem
		if err := yaml.Unmarshal(data, &routine); err != nil || routine.Name == "" || routine.Frequency == "" {
			continue
		}
		if routine.DisplayName == "" {
			routine.DisplayName = DisplayNameFrom(routine.Name)
		}

		// Find content file (.md or .md.tmpl)
		itemDir := "routines/" + name
		dirEntries, err := fs.ReadDir(fsys, itemDir)
		if err != nil {
			continue
		}
		for _, f := range dirEntries {
			if !f.IsDir() && f.Name() != "meta.yaml" {
				routine.ContentPath = itemDir + "/" + f.Name()
				break
			}
		}
		if routine.ContentPath == "" {
			continue
		}

		routines = append(routines, routine)
	}
	return routines
}

func loadScaffolding(fsys fs.FS) []ScaffoldingItem {
	data, err := fs.ReadFile(fsys, "scaffolding/manifest.yaml")
	if err != nil {
		return nil
	}
	var items []ScaffoldingItem
	if err := yaml.Unmarshal(data, &items); err != nil {
		return nil
	}
	for i := range items {
		if items[i].DisplayName == "" {
			items[i].DisplayName = DisplayNameFrom(items[i].Name)
		}
	}
	return items
}

func loadAgents(fsys fs.FS) []AgentDef {
	entries, err := fs.ReadDir(fsys, "agents")
	if err != nil {
		return nil
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var agents []AgentDef
	for _, name := range names {
		yamlPath := "agents/" + name + "/agent.yaml"
		data, err := fs.ReadFile(fsys, yamlPath)
		if err != nil {
			continue
		}

		var raw agentYAML
		if err := yaml.Unmarshal(data, &raw); err != nil || raw.Name == "" {
			continue
		}

		displayName := raw.DisplayName
		if displayName == "" {
			displayName = DisplayNameFrom(raw.Name)
		}

		agents = append(agents, AgentDef{
			Name:             raw.Name,
			DisplayName:      displayName,
			Description:      raw.Description,
			DefaultSkills:    raw.Defaults.Skills,
			DefaultWorkflows: raw.Defaults.Workflows,
			DefaultProtocols: raw.Defaults.Protocols,
			DefaultSensors:   raw.Defaults.Sensors,
			DefaultRoutines:  raw.Defaults.Routines,
			CoreDir:          "agents/" + name + "/core",
		})
	}
	return agents
}
