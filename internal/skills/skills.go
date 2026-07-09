package skills

import (
	"embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

//go:embed embedded/*.md
var embeddedFS embed.FS

var (
	readDir   = os.ReadDir
	readFile  = os.ReadFile
	mkdirAll  = os.MkdirAll
	writeFile = os.WriteFile
)

// Skill describes one available skill document.
type Skill struct {
	// Name is the extensionless skill identifier.
	Name string
	// Source is "embedded" or "custom".
	Source string
	// Body is the Markdown skill document.
	Body string
}

// TemplateClass is a skill-new/audit template category.
type TemplateClass string

const (
	// FullWorkflow describes broad multi-step skills with phases or tracks.
	FullWorkflow TemplateClass = "full_workflow"
	// MicroSkill describes narrow methods or rules for one concern.
	MicroSkill TemplateClass = "micro_skill"
	// ToolBridge describes skills centered on external tools or MCP bridges.
	ToolBridge TemplateClass = "tool_bridge"
	// NoneClass means the body is too small or lacks enough signals.
	NoneClass TemplateClass = "none"
)

// Classification describes the scored template choice for a skill body.
type Classification struct {
	// Skill is the skill name being classified.
	Skill string
	// Class is the selected template class.
	Class TemplateClass
	// Confidence is a normalized score in the range [0,1].
	Confidence float64
	// Scores records the raw score by template class.
	Scores map[TemplateClass]int
	// Reasoning explains the evidence behind the classification.
	Reasoning []string
}

// SkillNewDraft describes a fetched skill before it is written.
type SkillNewDraft struct {
	// Name is the normalized proposed skill name.
	Name string
	// URL is the source URL supplied by the user.
	URL string
	// Body is the fetched skill Markdown.
	Body string
	// Classification is the template classification for Body.
	Classification Classification
	// FetchedPath is the documented cache path for fetched source.
	FetchedPath string
	// TargetPath is the documented custom skill output path.
	TargetPath string
	// Instructions explain the manual approval and write workflow.
	Instructions []string
}

var contentSignals = []*regexp.Regexp{
	regexp.MustCompile(`(?im)^#{1,3}\s+(workflow|procedure|rules|phases?|tracks|setup|installation|mental\s+model)\b`),
	regexp.MustCompile(`(?im)\b(use\s+this\s+skill|when\s+to\s+use|do\s+not\s+use|stop\s+condition)\b`),
	regexp.MustCompile(`(?im)\b(mcp|connect|debug|tool|operation)\b`),
	regexp.MustCompile(`(?im)\b(recomp|decomp|build|test|cmake|make|ghidra|objdiff)\b`),
}

// Embedded returns all compiled skill documents.
func Embedded() []Skill {
	entries, _ := embeddedFS.ReadDir("embedded")
	var out []Skill
	for _, entry := range entries {
		body, _ := embeddedFS.ReadFile("embedded/" + entry.Name())
		out = append(out, Skill{Name: strings.TrimSuffix(entry.Name(), ".md"), Source: "embedded", Body: string(body)})
	}
	sortSkills(out)
	return out
}

// Names returns sorted embedded and custom skill names with custom names merged.
func Names(customDir string) ([]string, error) {
	all := Embedded()
	custom, err := LoadCustom(customDir)
	if err != nil {
		return nil, err
	}
	seen := map[string]string{}
	for _, skill := range all {
		seen[strings.ToLower(skill.Name)] = skill.Name
	}
	for _, skill := range custom {
		seen[strings.ToLower(skill.Name)] = skill.Name
	}
	names := make([]string, 0, len(seen))
	for _, name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// LoadCustom reads custom skill documents from dir.
func LoadCustom(dir string) ([]Skill, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, nil
	}
	entries, err := readDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Skill
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		body, err := readFile(path)
		if err != nil {
			return nil, err
		}
		out = append(out, Skill{Name: strings.TrimSuffix(entry.Name(), ".md"), Source: "custom", Body: string(body)})
	}
	sortSkills(out)
	return out, nil
}

// Resolve finds a skill by exact, case-insensitive, or .md-suffixed name.
func Resolve(query string, customDir string) (Skill, error) {
	query = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(query)), ".md")
	if query == "" {
		return Skill{}, fmt.Errorf("skill name is empty")
	}
	all := Embedded()
	custom, err := LoadCustom(customDir)
	if err != nil {
		return Skill{}, err
	}
	all = append(custom, all...)
	for _, skill := range all {
		if strings.ToLower(skill.Name) == query {
			return skill, nil
		}
	}
	return Skill{}, fmt.Errorf("unknown skill %q", query)
}

// Get returns the resolved skill document.
func Get(query string, customDir string) (Skill, error) {
	return Resolve(query, customDir)
}

// IsEmbedded reports whether name belongs to the embedded bundle.
func IsEmbedded(name string) bool {
	query := normalizeName(name)
	if query == "" {
		return false
	}
	for _, skill := range Embedded() {
		if strings.ToLower(skill.Name) == query {
			return true
		}
	}
	return false
}

// ListMarkdown renders a skill list with active and custom markers.
func ListMarkdown(active []string, customDir string) (string, error) {
	all := Embedded()
	custom, err := LoadCustom(customDir)
	if err != nil {
		return "", err
	}
	byName := map[string]Skill{}
	for _, skill := range all {
		byName[strings.ToLower(skill.Name)] = skill
	}
	for _, skill := range custom {
		byName[strings.ToLower(skill.Name)] = skill
	}
	names := make([]string, 0, len(byName))
	for _, skill := range byName {
		names = append(names, skill.Name)
	}
	sort.Strings(names)
	activeSet := map[string]bool{}
	for _, name := range active {
		activeSet[normalizeName(name)] = true
	}
	var b strings.Builder
	b.WriteString("Built-in RE skills:\n")
	for _, name := range names {
		skill := byName[strings.ToLower(name)]
		mark := " "
		if activeSet[strings.ToLower(name)] {
			mark = "*"
		}
		label := name
		if skill.Source == "custom" {
			label += " (custom)"
		}
		fmt.Fprintf(&b, "%s %s\n", mark, label)
	}
	b.WriteString("\nLoad one with /skill <name>.\n")
	return b.String(), nil
}

// Audit classifies a skill name into a broad template category.
func Audit(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "mcp"), strings.Contains(lower, "debug"):
		return "runtime-integration"
	case strings.Contains(lower, "recomp"), strings.Contains(lower, "decomp"):
		return "reverse-engineering-workflow"
	default:
		return "methodology"
	}
}

// Classify scores a skill body into a Phase 9 template class.
func Classify(name string, body string) Classification {
	scores := map[TemplateClass]int{FullWorkflow: 0, MicroSkill: 0, ToolBridge: 0}
	var reasons []string
	lower := strings.ToLower(body)
	if len(strings.TrimSpace(body)) < 60 || !hasContentSignals(body) {
		return Classification{Skill: name, Class: NoneClass, Confidence: 1, Scores: scores, Reasoning: []string{"body too short or missing skill content signals"}}
	}
	if strings.Contains(lower, "phase") || strings.Contains(lower, "track") || strings.Contains(lower, "workflow") || len(body) > 5000 {
		scores[FullWorkflow] += 3
		reasons = append(reasons, "multi-step workflow signals")
	}
	if strings.Contains(lower, "rule") || strings.Contains(lower, "when to use") || strings.Contains(lower, "stop condition") {
		scores[MicroSkill] += 3
		reasons = append(reasons, "single-skill rules or trigger signals")
	}
	if strings.Contains(lower, "mcp") || strings.Contains(lower, "connect") || strings.Contains(lower, "tool") || strings.Contains(lower, "debug") {
		scores[ToolBridge] += 3
		reasons = append(reasons, "external tool or bridge signals")
	}
	if strings.Contains(lower, "build") || strings.Contains(lower, "test") || strings.Contains(lower, "cmake") {
		scores[FullWorkflow]++
	}
	class, high, second := winningClass(scores)
	confidence := 0.0
	if high > 0 {
		confidence = float64(high-second) / float64(high)
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "content signals present but weakly separated")
	}
	return Classification{Skill: name, Class: class, Confidence: confidence, Scores: scores, Reasoning: reasons}
}

// NewDraft validates fetched skill content and builds a skill-new draft.
func NewDraft(rawURL string, body string) (SkillNewDraft, error) {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return SkillNewDraft{}, fmt.Errorf("invalid skill URL %q", rawURL)
	}
	if len(strings.TrimSpace(body)) < 60 {
		return SkillNewDraft{}, fmt.Errorf("fetched content too short to classify")
	}
	name := skillNameFromURL(parsed)
	classification := Classify(name, body)
	draft := SkillNewDraft{
		Name:           name,
		URL:            rawURL,
		Body:           body,
		Classification: classification,
		FetchedPath:    filepath.ToSlash(filepath.Join(".rehamr", "fetched", name+".md")),
		TargetPath:     filepath.ToSlash(filepath.Join(".rehamr", "skills", name+".md")),
		Instructions: []string{
			"ask the user to confirm the classification before writing",
			"cache fetched content under " + filepath.ToSlash(filepath.Join(".rehamr", "fetched", name+".md")),
			"write the approved skill under " + filepath.ToSlash(filepath.Join(".rehamr", "skills", name+".md")),
			"reload with /skills and activate with /skill " + name,
		},
	}
	return draft, nil
}

// ScaffoldCustomSkill writes an approved custom skill file.
func ScaffoldCustomSkill(customDir string, name string, body string) (string, error) {
	if strings.TrimSpace(customDir) == "" {
		return "", fmt.Errorf("custom skills directory is empty")
	}
	clean := skillNameFromURL(&url.URL{Path: normalizeName(name)})
	if clean == "" || clean == "new-skill" {
		return "", fmt.Errorf("skill name is empty")
	}
	if len(strings.TrimSpace(body)) < 60 {
		return "", fmt.Errorf("skill body is too short")
	}
	if err := mkdirAll(customDir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(customDir, clean+".md")
	if err := writeFile(path, []byte(body), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func sortSkills(skills []Skill) {
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })
}

func normalizeName(name string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(name)), ".md")
}

func hasContentSignals(body string) bool {
	for _, signal := range contentSignals {
		if signal.MatchString(body) {
			return true
		}
	}
	return false
}

func winningClass(scores map[TemplateClass]int) (TemplateClass, int, int) {
	order := []TemplateClass{FullWorkflow, MicroSkill, ToolBridge}
	winner := order[0]
	high, second := scores[winner], 0
	for _, class := range order[1:] {
		score := scores[class]
		if score > high {
			second = high
			winner, high = class, score
			continue
		}
		if score > second {
			second = score
		}
	}
	if high <= 0 {
		return NoneClass, high, second
	}
	return winner, high, second
}

func skillNameFromURL(u *url.URL) string {
	segment := strings.TrimRight(u.Path, "/")
	fromHost := false
	if idx := strings.LastIndexByte(segment, '/'); idx >= 0 {
		segment = segment[idx+1:]
	}
	if segment == "" {
		segment = u.Host
		fromHost = true
	}
	if ext := filepath.Ext(segment); ext != "" && !fromHost {
		segment = strings.TrimSuffix(segment, ext)
	}
	segment = strings.ToLower(segment)
	var b strings.Builder
	lastHyphen := false
	for _, r := range segment {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if valid {
			b.WriteRune(r)
			lastHyphen = false
			continue
		}
		if !lastHyphen && b.Len() > 0 {
			b.WriteByte('-')
			lastHyphen = true
		}
	}
	name := strings.Trim(b.String(), "-")
	if len(name) < 2 {
		return "new-skill"
	}
	return name
}
