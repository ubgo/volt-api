// Package config is volt-api's configuration, defined in Go and loaded with
// cfgkit.
//
// It replaces a Pkl schema plus generated bindings. What that buys:
//
//   - Zero input boots. Every field has a compiled-in default, so the service
//     starts with no config file, no environment variable and no pkl binary
//     installed. Previously the loader walked up to three directories looking
//     for modules/config/pkl and PANICKED if it did not find one, which made
//     the working directory part of the runtime contract.
//   - Errors instead of panics. Get returns an error; GetOrDie is the explicit
//     opt-in for main functions that genuinely cannot continue.
//   - Provenance. Every field can say which source set it, so "why is this
//     value what it is in staging" is answerable without a bisect.
//   - Secrets are structurally masked, so a config dump cannot leak the JWT
//     secret or the database password.
//
// The environment keys are NOT new. They are exactly the ones the deployed
// .env.prod already used, which the prod Pkl file read one by one — so a
// deployment that works today keeps working, and this file is the first place
// the full key list has existed in one piece.
package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ubgo/cfgkit"
)

// Environment is the deployment environment.
//
// It stays a named type rather than a bare string so a comparison against a
// misspelled literal is a compile error, and it keeps the same three values
// the Pkl typealias had.
type Environment string

// The deployment environments. Anything else fails validation rather than
// being treated as "not prod", which is the reading that quietly disables
// production hardening.
const (
	Dev  Environment = "dev"
	Stag Environment = "stag"
	Prod Environment = "prod"
)

// Environments is the canonical iteration order, so validation and any UI that
// lists environments read from one place.
var Environments = []Environment{Dev, Stag, Prod}

// Server is where the HTTP server binds.
type Server struct {
	Hostname string `env:"HOSTNAME" default:"localhost" doc:"interface the HTTP server binds"`
	Port     string `env:"PORT" default:"9001" doc:"port the HTTP server listens on"`
}

// Core is the identity and filesystem layout every other component reads.
type Core struct {
	AppName string `env:"APP_NAME" default:"boilerplate" doc:"human-readable application name"`

	// AppKey identifies THIS microservice. It is the fallback for the
	// telemetry service name, so two services sharing it would merge in the
	// observability backend.
	AppKey string `env:"APP_KEY" default:"boilerplate_local" doc:"unique key for this microservice"`

	Environment Environment `env:"CONFIG_ENV" default:"dev" doc:"deployment environment: dev, stag or prod"`

	// Rootdir anchors every derived path. It defaults to "." rather than a
	// discovered directory: the previous loader walked upwards looking for a
	// config directory, which made tmp/, uploads/ and the log file move
	// depending on which directory the process was launched from.
	Rootdir string `env:"VOLT_ROOT" default:"." doc:"root directory; anchors tmp, uploads and logs"`

	// Tmpdir and Uploadsdir are DERIVED from Rootdir unless set explicitly.
	// See Derive.
	Tmpdir     string `env:"TMP_DIR" doc:"defaults to <root>/tmp"`
	Uploadsdir string `env:"UPLOADS_DIR" doc:"defaults to <root>/uploads"`

	InternalKey   string `env:"INTERNAL_KEY" default:"change-me-internal-key" secret:"true" doc:"service-to-service auth key"`
	Jwtsecret     string `env:"JWT_SECRET" default:"change-me-jwt-secret-key" secret:"true" doc:"JWT signing secret"`
	EncryptionKey string `env:"ENCRYPTION_KEY" default:"change-me-encryption-key-32ch" secret:"true" doc:"data encryption key"`
}

// Database is the Postgres connection.
type Database struct {
	Username string `env:"DB_USERNAME" default:"postgres"`
	Password string `env:"DB_PASSWORD" default:"root" secret:"true"`
	Host     string `env:"DB_HOST" default:"localhost"`
	Port     uint16 `env:"DB_PORT" default:"5432"`
	DbName   string `env:"DB_NAME" default:"boilerplate_db"`
	Sslmode  string `env:"DB_SSLMODE" default:"disable" doc:"disable, require, verify-ca or verify-full"`
}

// Graphql holds the GraphQL surface's settings.
type Graphql struct {
	Key string `env:"GRAPHQL_KEY" default:"1234" secret:"true"`
}

// Otel is the OpenTelemetry exporter configuration.
type Otel struct {
	Enable        bool   `env:"OTEL_ENABLE" default:"false" doc:"master switch for the OTEL middleware"`
	IsEndpoingUrl bool   `env:"OTEL_IS_ENDPOINT_URL" default:"false"`
	Endpoint      string `env:"OTEL_ENDPOINT" default:"localhost:4317"`
	ApiKey        string `env:"OTEL_API_KEY" secret:"true"`
	PrefixBasic   bool   `env:"OTEL_PREFIX_BASIC" default:"false"`
	Insecure      bool   `env:"OTEL_INSECURE" default:"true"`
	EnableMetrics bool   `env:"OTEL_ENABLE_METRICS" default:"false"`
	EnableLogs    bool   `env:"OTEL_ENABLE_LOGS" default:"true"`
	EnableTraces  bool   `env:"OTEL_ENABLE_TRACES" default:"true"`

	// ServiceName is what appears in the observability backend's service
	// dropdown. Every deployed environment sets it so environments never mix
	// in a backend without tenant separation; locally it derives from AppKey,
	// which is already distinct per service. See Derive.
	ServiceName string `env:"OTEL_SERVICE_NAME" doc:"defaults to the app key"`

	ServiceVersion string `env:"OTEL_SERVICE_VERSION" default:"1.0.0"`
	EnableHttp     bool   `env:"OTEL_ENABLE_HTTP" default:"true"`
	EnableStdout   bool   `env:"OTEL_ENABLE_STDOUT" default:"true"`
}

// LocalConfig is the on-disk log file.
type LocalConfig struct {
	Enabled bool   `env:"LOGGER_LOCAL_ENABLED" default:"true"`
	LogFile string `env:"LOGGER_LOG_FILE" doc:"defaults to <tmp>/logs/app.log"`
}

// Logger selects where log records go.
type Logger struct {
	EnableStdout bool `env:"LOGGER_ENABLE_STDOUT" default:"true"`
	EnableOtel   bool `env:"LOGGER_ENABLE_OTEL" default:"false"`
	EnableFile   bool `env:"LOGGER_ENABLE_FILE" default:"true"`

	LocalConfig LocalConfig
}

// Redis is the cache connection.
//
// Status defaults to TRUE, matching the previous configuration exactly. This
// migration deliberately changes no behaviour — a migration that also changes
// defaults makes any resulting incident impossible to attribute.
//
// It is worth knowing what that costs: with Status true and no Redis running,
// boot still fails, so "starts with zero input" is currently true of the
// CONFIG and not yet of the SERVICE. Making it true of the service means
// tolerating a nil client — today `rediscache.New(nil)` stores the nil and
// panics on first use, and there is no in-memory or no-op Cache to fall back
// to. That is a separate piece of work, tracked, and not bundled here.
type Redis struct {
	Status bool   `env:"REDIS_STATUS" default:"true" doc:"enable the Redis cache"`
	Url    string `env:"REDIS_URL" default:"redis://default:@127.0.0.1:6379"`
	Db     int    `env:"REDIS_DB" default:"1"`
}

// ApiKeys holds third-party credentials.
type ApiKeys struct {
	Openai string `env:"OPENAI_API_KEY" secret:"true"`
}

// Gocron configures the scheduled-job runner.
type Gocron struct {
	UiEnable bool   `env:"CRON_UI_ENABLE" default:"true"`
	UiPort   string `env:"CRON_UI_PORT" default:"8089"`

	// EnabledJobs names the jobs to register. Empty means register all of
	// them, which is what the Pkl default expressed as an empty listing.
	EnabledJobs []string `env:"CRON_ENABLED_JOBS" delim:"," doc:"job names to register; empty registers all"`
}

// App holds the settings specific to this product rather than the platform.
type App struct {
	ReactAppHost string `env:"REACT_APP_HOST" default:"http://localhost:2312"`
	ApiKeys      ApiKeys
	Gocron       Gocron
}

// Root is the whole configuration surface.
//
// The field names and nesting match what the generated Pkl bindings produced,
// so consumers reading cfg.Core.AppKey or cfg.Otel.Endpoint are unchanged.
type Root struct {
	Server   Server
	Core     Core
	Database Database
	Graphql  Graphql
	Otel     Otel
	Logger   Logger
	Redis    Redis
	App      App
}

// Derive fills the fields whose defaults depend on other fields.
//
// A `default:` tag holds a constant, and these are not constants: three paths
// hang off Rootdir and the telemetry service name falls back to AppKey. Derive
// runs AFTER every source, so an explicitly configured value is never
// overwritten — which is why each assignment is guarded on the zero value.
func (r *Root) Derive() error {
	root := strings.TrimRight(r.Core.Rootdir, "/")
	if root == "" {
		root = "."
	}
	r.Core.Rootdir = root

	if r.Core.Tmpdir == "" {
		r.Core.Tmpdir = filepath.Join(root, "tmp")
	}
	if r.Core.Uploadsdir == "" {
		r.Core.Uploadsdir = filepath.Join(root, "uploads")
	}
	if r.Logger.LocalConfig.LogFile == "" {
		r.Logger.LocalConfig.LogFile = filepath.Join(r.Core.Tmpdir, "logs", "app.log")
	}
	if r.Otel.ServiceName == "" {
		r.Otel.ServiceName = r.Core.AppKey
	}
	return nil
}

// knownPlaceholderSecrets are the credentials shipped as defaults in this
// file.
//
// They live HERE, beside the `default:` tags that declare them, because the
// previous list lived in another package and carried a comment telling the
// reader to keep it in sync with the Pkl templates by hand. A list that must
// be manually synchronised with values in a different file is a list that goes
// stale — and a placeholder that stops being recognised is a placeholder that
// ships to production.
//
// Adding a default credential above means adding it here.
var knownPlaceholderSecrets = []string{
	"change-me-internal-key",
	"change-me-jwt-secret-key",
	"change-me-encryption-key-32ch",
	"xxxx", // the historical sample-config placeholder
}

// IsWeakOrDefaultKey reports whether a secret is unusable for protecting real
// data: blank, or one of the shipped placeholders above.
//
// Exported because the boot path checks the encryption key before the rest of
// the application starts, and the rule must not be written twice.
func IsWeakOrDefaultKey(v string) bool {
	if strings.TrimSpace(v) == "" {
		return true
	}
	return slices.Contains(knownPlaceholderSecrets, v)
}

// Validate reports every problem at once.
//
// Joined rather than returning on the first failure: a misconfigured deploy
// should be one fix, not one fix per restart, and each restart is another
// failed rollout.
//
// The production rules are scoped to non-dev environments on purpose. That is
// what keeps the zero-input promise honest — a laptop with nothing set still
// boots, while a production deploy carrying a placeholder secret does not.
func (r *Root) Validate() error {
	prod := r.Core.Environment != Dev

	return errors.Join(
		cfgkit.OneOf("CONFIG_ENV", r.Core.Environment, Environments...),
		cfgkit.NotEmpty("APP_KEY", r.Core.AppKey),
		cfgkit.NotEmpty("HOSTNAME", r.Server.Hostname),
		cfgkit.NotEmpty("PORT", r.Server.Port),
		cfgkit.NotEmpty("DB_HOST", r.Database.Host),
		cfgkit.NotEmpty("DB_NAME", r.Database.DbName),
		cfgkit.Range("DB_PORT", int(r.Database.Port), 1, 65535),
		cfgkit.OneOf("DB_SSLMODE", r.Database.Sslmode,
			"disable", "allow", "prefer", "require", "verify-ca", "verify-full"),

		// Placeholders are fatal outside dev. Shipping the default encryption
		// key is the failure that makes every encrypted record readable by
		// anyone holding the source.
		requiredStrongInProd(prod, "ENCRYPTION_KEY", r.Core.EncryptionKey),
		requiredStrongInProd(prod, "JWT_SECRET", r.Core.Jwtsecret),
		requiredStrongInProd(prod, "INTERNAL_KEY", r.Core.InternalKey),
	)
}

// requiredStrongInProd rejects a placeholder credential outside dev.
//
// It names the variable rather than showing the value, because a validation
// message is one of the easiest routes for a credential to reach a log
// aggregator.
func requiredStrongInProd(prod bool, key, value string) error {
	if !prod || !IsWeakOrDefaultKey(value) {
		return nil
	}
	return fmt.Errorf("%s: is a default or placeholder value, which is refused outside dev", key)
}
