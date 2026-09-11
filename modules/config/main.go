package config

import (
	"fmt"
	"os"

	"github.com/ubgo/cfgkit"
)

// EnvFile is the optional local override file, read from the process's working
// directory.
//
// It is OPTIONAL by design: a missing .env is a normal state, not an error,
// because the compiled-in defaults already produce a valid configuration. The
// file is not searched for up the directory tree — the previous loader walked
// up to three levels looking for a config directory, which meant tmp/, uploads/
// and the log file all moved depending on where the process was launched from.
// Anchor paths with VOLT_ROOT instead; it is one variable and it is explicit.
const EnvFile = ".env"

// sources is the layer order, lowest precedence first.
//
// Precedence is positional: the last source to claim a key wins. The
// environment comes last so an operator can always override the file without
// editing it, which is what makes a container's -e flag behave the way anyone
// would expect.
func sources() []any {
	return []any{
		cfgkit.FromFiles(EnvFile),
		cfgkit.FromEnviron(),
	}
}

// Options is the load configuration, exported so that everything which reads
// this application's configuration reads it the same way.
//
// Its consumers are Get (the application) and the CLI's config command, which
// comes from cfgkit/contrib/cli-cobra and is handed these same options. A
// command that checked a different source list would be checking something the
// application never runs — worse than not checking, because it reports
// confidence it has not earned.
//
// There are deliberately no Check/Explain/Document wrappers here: those verbs
// live in the contrib command, so there is one implementation rather than two
// that can disagree.
func Options() []cfgkit.Option {
	return []cfgkit.Option{
		// volt-api names the environment CONFIG_ENV; cfgkit's own default is
		// APP_ENV. Telling it explicitly is what makes the production
		// validation rules fire on a real deployment.
		cfgkit.WithModeKey("CONFIG_ENV"),
		cfgkit.WithSources(sources()...),
	}
}

// Get loads and validates the configuration.
//
// It returns an error rather than panicking. The previous loader panicked on a
// missing directory, which turned a configuration mistake into a stack trace at
// container start — after the rollout, and before the logger existed to record
// it. Callers that genuinely cannot continue should use GetOrDie and say so.
func Get() (*Root, error) {
	cfg, _, err := cfgkit.Load[Root](Options()...)
	if err != nil {
		return nil, fmt.Errorf("loading configuration: %w", err)
	}
	return cfg, nil
}

// GetOrDie loads the configuration or exits.
//
// It exists so a main function can stay a one-liner while the panic remains an
// EXPLICIT choice at the call site rather than a hidden property of Get. The
// message goes to stderr and the exit code is 1, so a supervisor sees a failed
// start rather than a crash.
func GetOrDie() *Root {
	cfg, err := Get()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return cfg
}
