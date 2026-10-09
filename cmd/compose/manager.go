package compose

import (
	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/containers"
	"github.com/eng618/eng/internal/paths"
)

// newManagerFromConfig adapts containers config into a pure-data Manager.
// It never passes Viper into the domain layer: registered entries are
// expanded here before being handed to containers as DTOs.
func newManagerFromConfig(cfg config.ContainersConfig) *containers.Manager {
	registered := make([]containers.RegisteredStack, 0, len(cfg.Stacks))
	for _, s := range cfg.Stacks {
		registered = append(registered, containers.RegisteredStack{
			Name: s.Name,
			Path: paths.Expand(s.Path),
		})
	}
	base := paths.Expand(cfg.Path)
	return containers.NewManagerWithStacks(base, registered)
}
