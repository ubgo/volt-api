package modhost

import (
	"github.com/khanakia/boilerplate-golang/modules/modkit"
	"github.com/spf13/cobra"
)

// DispatchCron iterates modules and calls RegisterCron on those that opt in.
// scheduler is passed through as `any` per modkit's dep-light design;
// modules type-assert back to gocron.Scheduler.
func (l *Loader) DispatchCron(scheduler any) error {
	for _, m := range l.mods {
		if r, ok := m.(modkit.CronRegistrar); ok {
			if err := r.RegisterCron(scheduler); err != nil {
				return err
			}
		}
	}
	return nil
}

// DispatchCLI mounts each module's CLI commands on the host root cmd.
func (l *Loader) DispatchCLI(root *cobra.Command) {
	for _, m := range l.mods {
		if r, ok := m.(modkit.CLIRegistrar); ok {
			r.RegisterCLI(root)
		}
	}
}

// DispatchHTTP mounts module HTTP routes on the host router.
func (l *Loader) DispatchHTTP(router any) error {
	for _, m := range l.mods {
		if r, ok := m.(modkit.HTTPRegistrar); ok {
			if err := r.RegisterHTTP(router); err != nil {
				return err
			}
		}
	}
	return nil
}

// DispatchGQL invokes RegisterResolvers on opted-in modules.
func (l *Loader) DispatchGQL(reg any) error {
	for _, m := range l.mods {
		if r, ok := m.(modkit.GQLResolverRegistrar); ok {
			if err := r.RegisterResolvers(reg); err != nil {
				return err
			}
		}
	}
	return nil
}
