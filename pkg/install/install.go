package install

import (
	"fmt"

	"github.com/evankanderson/knuts/pkg"
)

// Component describes a Knative software component to be installed.
type Component struct {
	Name        string
	Description string
	Yaml        string
	hidden      bool
	provides    string
	preferred   bool
	deps        []string
}

// ComponentsAsFlag returns the public versions of the components list as a
// MultiSelect flag.
func ComponentsAsFlag() pkg.MultiSelect {
	o := map[string]pkg.Option{}
	for _, c := range components {
		if !c.hidden {
			o[c.Name] = pkg.Option{
				Description: c.Description,
				Data:        c,
			}
		}
	}
	return pkg.MultiSelect{
		Description: "Which components to install",
		Options:     o,
	}
}

// Expand returns the transitive set of dependencies for the Component
// (prompting if needed), given the currently selected set of Components
// to install.
func (c Component) Expand(selected []Component) []Component {
	for _, x := range selected {
		if x.Name == c.Name {
			return selected
		}
	}
	ret := selected[:]

	unresolved := c.deps
DEP:
	for len(unresolved) > 0 {
		dep := unresolved[0]
		unresolved = unresolved[1:]

		if opts, ok := providers[dep]; ok {
			// Is a "provides", resolve it.
			// Select an opt via menu
			menu := map[string]pkg.Option{}
			for _, opt := range opts {
				for _, c := range ret {
					if c.Name == opt {
						fmt.Printf("Resolved %q to %q\n", dep, opt)
						continue DEP
					}
				}
				choice := componentMap[opt]
				menu[opt] = pkg.Option{Description: choice.Description, Data: choice}
				if choice.preferred {
					unresolved = append([]string{opt}, unresolved...)
					continue DEP
				}
			}
			// TODO: this should be a single-select
			prompt := &pkg.MultiSelect{
				Description: fmt.Sprintf("Select an implementation for %q", dep),
				Options:     menu,
			}
			dep = prompt.Get().([]pkg.Option)[0].Data.(Component).Name
		}
		if resolved, ok := componentMap[dep]; ok {
			// Is a real dependency
			ret = resolved.Expand(ret)
		}
	}
	return append(ret, c)
}

var (
	components = []Component{
		{
			Name:        "build",
			Description: "Knative build: cluster-hosted container build",
			Yaml:        "https://github.com/knative/serving/releases/download/v0.2.2/build.yaml",
		},
		{
			Name:        "serving",
			Description: "Knative serving: scale from zero stateless web services",
			Yaml:        "https://github.com/knative/serving/releases/download/v0.2.2/serving.yaml",
			deps:        []string{"istio"},
		},
		{
			Name:        "eventing",
			Description: "Knative eventing: Channels and orchestration",
			Yaml:        "https://github.com/knative/eventing/releases/download/v0.2.1/release.yaml",
			deps:        []string{"istio-sidecar"},
		},
		{
			Name:        "eventing-sources",
			Description: "Knative event sources",
			Yaml:        "https://github.com/knative/eventing-sources/releases/download/v0.2.1/release.yaml",
			deps:        []string{"serving", "istio-sidecar"},
		},
		{
			Name:        "monitoring",
			Description: "Monitoring and instrumentation for Knative",
			Yaml:        "https://github.com/knative/serving/releases/download/v0.2.2/monitoring.yaml",
		},
		{
			Name:        "istio-sidecar",
			Description: "Knative tested version of Istio with sidecar",
			Yaml:        "https://github.com/knative/serving/releases/download/v0.2.2/istio.yaml",
			hidden:      true,
			provides:    "istio",
			deps:        []string{"istio-crd"},
			preferred:   true,
		},
		{
			Name:        "istio-lean",
			Description: "Knative tested version of Istio without sidecar",
			Yaml:        "https://github.com/knative/serving/releases/download/v0.2.2/istio-lean.yaml",
			hidden:      true,
			provides:    "istio",
			deps:        []string{"istio-crd"},
		},
		{
			Name:        "istio-crd",
			Description: "Istio CRDs",
			Yaml:        "https://github.com/knative/serving/releases/download/v0.2.2/istio-crds.yaml",
			hidden:      true,
		},
	}

	// providers tracks which components provide a dependency.
	providers    = map[string][]string{}
	componentMap = map[string]Component{}
)

func init() {
	for _, c := range components {
		if c.provides != "" {
			providers[c.provides] = append(providers[c.provides], c.Name)
		}
		componentMap[c.Name] = c
	}
}
