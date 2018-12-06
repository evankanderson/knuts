package pkg

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey"
)

// DryRun is a global flag indicating that no write actions should be taken.
var DryRun = true

// MultiSelect is a flags.Value implementing interface which selects one or
// more items from a list of Options, and will prompt if no items are
// selected by the flag.
type MultiSelect struct {
	// Options provides a mapping from shortname to a description and a selected object.
	Options  map[string]Option
	selected map[string]bool
	// Description is a string describing what the MultiSelect covers.
	Description string
}

// Option describes a single selection from a set of enumerated options.
type Option struct {
	Description string
	Data        interface{}
}

// Prompt is a flags.Value implementing interface for a string which will
// interactively prompt if the value is not provided.
type Prompt struct {
	Description string
	data        string
}

func (m *MultiSelect) init() {
	if m.selected == nil {
		m.selected = make(map[string]bool)
	}
}

// String implements the flag.Value interface.
func (m *MultiSelect) String() string {
	m.init()
	out := []string{}
	for k := range m.selected {
		out = append(out, k)
	}
	return strings.Join(out, ",")
}

// Set implements the flag.Value interface.
func (m *MultiSelect) Set(value string) error {
	m.init()
	for _, i := range strings.Split(value, ",") {
		if _, ok := m.Options[i]; !ok {
			return fmt.Errorf("Unable to recognize %q", i)
		}
		m.selected[i] = true
	}
	return nil
}

// Type implements the pflag.Value interface.
func (m *MultiSelect) Type() string {
	return "multiSelect"
}

// Get implements the flag.Get interface.
func (m *MultiSelect) Get() interface{} {
	m.init()
	if len(m.selected) == 0 {
		if err := m.prompt(); err != nil {
			fmt.Printf("Error prompting for %s: %v", m.Description, err)
			return []Option{}
		}
	}
	out := []Option{}
	for k := range m.selected {
		out = append(out, m.Options[k])
	}
	return out
}

func (m *MultiSelect) prompt() error {
	choices := make([]string, len(m.Options))
	i := 0
	for k, v := range m.Options {
		choices[i] = fmt.Sprintf("%s: %s", k, v.Description)
		i++
	}
	sort.Strings(choices)
	question := &survey.MultiSelect{
		Message: m.Description,
		Options: choices,
	}
	answers := []string{}
	err := survey.AskOne(question, &answers, nil)
	if err != nil {
		return err
	}
	for _, v := range answers {
		m.selected[strings.SplitN(v, ":", 2)[0]] = true
	}
	return nil
}

// String implements the flag.Value interface.
func (p *Prompt) String() string {
	return p.data
}

// Set implements the flag.Value interface.
func (p *Prompt) Set(in string) error {
	p.data = in
	return nil
}

// Type implements the pflag.Value interface.
func (p *Prompt) Type() string {
	return "string"
}

// Get implements the flag.Getter interface.
func (p *Prompt) Get() interface{} {
	if p.data == "" {
		prompt := &survey.Input{Message: p.Description}
		survey.AskOne(prompt, &p.data, nil)
	}
	return p.data
}
