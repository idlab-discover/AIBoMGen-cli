# Bubbletea UI enhancement ideas

## Interactive Hugging Face browser

```go
package ui

import (
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type ModelSelectorModel struct {
    textInput textinput.Model
    list      list.Model
    selected  []string
    searching bool
}

// Features:
// - Search-as-you-type for Hugging Face models
// - Multi-select with space bar
// - Fuzzy filtering
// - Show model metadata (downloads, likes) in list items
```

## Progress Spinner/Bar (generate command)

```go
package ui

import (
    "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/bubbles/spinner"
    tea "github.com/charmbracelet/bubbletea"
)

type GenerateProgressModel struct {
    spinner  spinner.Model
    progress progress.Model
    status   string
    steps    []Step
    current  int
}

type Step struct {
    Name     string
    Status   StepStatus // pending, running, done, failed
}

// Displays:
// ⣾ Scanning directory for AI imports...
// ✓ Found 3 model references
// ⣾ Fetching metadata for gpt2...
// ████████████░░░░░░░░ 60% (2/3 models)
```

## Interactive Enrichment Form

```go
package ui

import (
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/textarea"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/huh" // Form library from Charm
)

type EnrichFormModel struct {
    form        *huh.Form
    fields      []FieldSpec
    currentIdx  int
    values      map[string]string
}

// Features:
// - Tab navigation between fields
// - Field validation with inline errors
// - Dropdown selection for enum fields (license types, etc.)
// - Multi-line textarea for descriptions
// - Help text shown for each field
// - Required fields highlighted
```

## Completeness Visualization

```go
package ui

import (
    "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/lipgloss"
    tea "github.com/charmbracelet/bubbletea"
)

type CompletenessViewModel struct {
    modelScore    float64
    datasetScores map[string]float64
    missingFields []FieldGroup
    expanded      map[string]bool // expandable sections
}

// Visual output:
// ╭─────────────────────────────────────────╮
// │  AIBOM Completeness Report              │
// ├─────────────────────────────────────────┤
// │  Model: gpt2                            │
// │  Score: ████████████░░░░░░░░ 73.5%      │
// │                                         │
// │  ▼ Required Fields (3 missing)          │
// │    ✗ license                            │
// │    ✗ author                             │
// │    ✗ description                        │
// │                                         │
// │  ▶ Optional Fields (5 missing)          │
// │                                         │
// │  Datasets:                              │
// │    wikitext: ██████████████░░░ 85.0%    │
// ╰─────────────────────────────────────────╯
```

## Validation Results Table

```go
package ui

import (
    "github.com/charmbracelet/bubbles/table"
    tea "github.com/charmbracelet/bubbletea"
)

type ValidationTableModel struct {
    table       table.Model
    errors      []ValidationIssue
    warnings    []ValidationIssue
    activeTab   int // 0=summary, 1=errors, 2=warnings
}

// Features:
// - Tabbed view: Summary | Errors | Warnings
// - Sortable columns
// - Navigate to see details of each issue
// - Filter by severity

```

## Interactive Command Launcher

```go
package ui

import (
    "github.com/charmbracelet/bubbles/list"
    tea "github.com/charmbracelet/bubbletea"
)

type LauncherModel struct {
    list     list.Model
    commands []CommandOption
}

type CommandOption struct {
    Name        string
    Description string
    Action      func() tea.Cmd
}

// When user runs `aibomgen-cli` with no args:
// 
// ╭─ AIBoMGen-cli ─────────────────────────╮
// │                                        │
// │  What would you like to do?            │
// │                                        │
// │  > Generate AIBOM                      │
// │    Validate existing AIBOM             │
// │    Enrich AIBOM with metadata          │
// │    Check completeness score            │
// │                                        │
// │  ↑/↓ navigate • enter select • q quit │
// ╰────────────────────────────────────────╯
```
## File Browser for --input and --ouput

```go
package ui

import (
    "github.com/charmbracelet/bubbles/filepicker"
    tea "github.com/charmbracelet/bubbletea"
)

type FilePickerModel struct {
    picker   filepicker.Model
    selected string
    filter   []string // e.g., []string{".json", ".xml"} for validate
}
```