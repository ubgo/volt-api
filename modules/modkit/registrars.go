package modkit

import "github.com/spf13/cobra"

// Registrar interfaces are optional — modules implement only the ones that
// apply. Host inspects each Module via type assertion and calls the matching
// Register method during boot.
//
// These are intentionally typed as `any` for scheduler/router/resolver
// to keep modkit free of heavy deps (gocron, gin, gqlgen). Each module
// type-asserts back to the concrete type at registration time.

// CronRegistrar — module schedules background jobs.
// Host passes its gocron.Scheduler.
type CronRegistrar interface {
	RegisterCron(scheduler any) error
}

// CLIRegistrar — module adds subcommands to the host CLI.
type CLIRegistrar interface {
	RegisterCLI(root *cobra.Command)
}

// HTTPRegistrar — module mounts HTTP routes.
// Host passes its gin.IRouter (or compatible). Module type-asserts.
type HTTPRegistrar interface {
	RegisterHTTP(router any) error
}

// GQLResolverRegistrar — module contributes GraphQL resolver implementations.
// reg is a host-defined registry the module type-asserts.
type GQLResolverRegistrar interface {
	RegisterResolvers(reg any) error
}
