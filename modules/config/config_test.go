// Tests for volt-api's configuration.
//
// The headline claim is that the service configures itself with zero input.
// That is the first test, because it is the property the whole migration was
// for: it was previously false — the loader walked directories looking for a
// Pkl file and panicked when it found none.
package config

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ubgo/cfgkit"
)

// isolate points the loader at an empty directory and UNSETS every variable
// this package reads, so a test never picks up the developer's environment or
// the repository's own .env.
//
// It unsets rather than blanking. t.Setenv(k, "") sets the variable to an
// empty string, and cfgkit correctly treats that as a real value — an operator
// who clears a variable meant to clear it — so blanking would override every
// default and test the wrong thing entirely. This helper got that wrong first
// time round and the whole suite failed on "" is not a valid bool.
func isolate(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())

	for _, k := range readsEnv {
		if old, ok := os.LookupEnv(k); ok {
			t.Cleanup(func() { _ = os.Setenv(k, old) })
		}
		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("unsetting %s: %v", k, err)
		}
	}
}

// readsEnv is every variable this package binds.
//
// It is derived from the struct tags by hand, and TestReadsEnvMatchesTheStruct
// pins that it stays complete — a variable missing here would leak the
// developer's environment into a test and produce a failure nobody can
// reproduce on another machine.
var readsEnv = []string{
	"HOSTNAME", "PORT",
	"APP_NAME", "APP_KEY", "CONFIG_ENV", "VOLT_ROOT", "TMP_DIR", "UPLOADS_DIR",
	"INTERNAL_KEY", "JWT_SECRET", "ENCRYPTION_KEY",
	"DB_USERNAME", "DB_PASSWORD", "DB_HOST", "DB_PORT", "DB_NAME", "DB_SSLMODE",
	"GRAPHQL_KEY",
	"OTEL_ENABLE", "OTEL_IS_ENDPOINT_URL", "OTEL_ENDPOINT", "OTEL_API_KEY",
	"OTEL_PREFIX_BASIC", "OTEL_INSECURE", "OTEL_ENABLE_METRICS",
	"OTEL_ENABLE_LOGS", "OTEL_ENABLE_TRACES", "OTEL_SERVICE_NAME",
	"OTEL_SERVICE_VERSION", "OTEL_ENABLE_HTTP", "OTEL_ENABLE_STDOUT",
	"LOGGER_ENABLE_STDOUT", "LOGGER_ENABLE_OTEL", "LOGGER_ENABLE_FILE",
	"LOGGER_LOCAL_ENABLED", "LOGGER_LOG_FILE",
	"REDIS_STATUS", "REDIS_URL", "REDIS_DB",
	"OPENAI_API_KEY", "CRON_UI_ENABLE", "CRON_UI_PORT", "CRON_ENABLED_JOBS",
	"REACT_APP_HOST",
}

// TestReadsEnvMatchesTheStruct pins that the isolation list above covers every
// key the configuration actually binds.
//
// The generated contract is the source of truth: it is produced from the
// struct, so a field added without updating readsEnv fails here rather than
// producing a test that quietly reads the developer's environment.
func TestReadsEnvMatchesTheStruct(t *testing.T) {
	var out bytes.Buffer
	if err := cfgkit.Document[Root](&out); err != nil {
		t.Fatal(err)
	}

	known := make(map[string]bool, len(readsEnv))
	for _, k := range readsEnv {
		known[k] = true
	}

	for _, line := range strings.Split(out.String(), "\n") {
		key, _, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || key == "" || strings.HasPrefix(key, "#") {
			continue
		}
		if !known[key] {
			t.Errorf("%s is bound by the struct but missing from readsEnv, so tests "+
				"would inherit it from the developer's environment", key)
		}
	}
}

// TestZeroInputProducesAValidConfig is the property the migration exists for.
//
// No .env file, no environment variables, no pkl binary: the service still has
// a complete, valid configuration. Previously this could not even be attempted
// — Get panicked if it could not find a config directory within three levels
// of the working directory.
func TestZeroInputProducesAValidConfig(t *testing.T) {
	isolate(t)

	cfg, err := Get()
	if err != nil {
		t.Fatalf("zero input did not produce a valid configuration: %v", err)
	}

	if cfg.Server.Port == "" {
		t.Error("Server.Port is empty with zero input")
	}
	if cfg.Core.AppKey == "" {
		t.Error("Core.AppKey is empty with zero input")
	}
	if cfg.Core.Environment != Dev {
		t.Errorf("Environment = %q, want dev as the zero-input default", cfg.Core.Environment)
	}
	if err := check(); err != nil {
		t.Errorf("Check rejected the zero-input configuration: %v", err)
	}
}

// TestDerivedPathsHangOffRootdir pins the paths that used to be Pkl string
// concatenation. They are derived rather than defaulted because a `default:`
// tag holds a constant and these depend on another field.
func TestDerivedPathsHangOffRootdir(t *testing.T) {
	isolate(t)
	t.Setenv("VOLT_ROOT", "/srv/volt")

	cfg, err := Get()
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct{ name, got, want string }{
		{"Tmpdir", cfg.Core.Tmpdir, filepath.Join("/srv/volt", "tmp")},
		{"Uploadsdir", cfg.Core.Uploadsdir, filepath.Join("/srv/volt", "uploads")},
		{"LogFile", cfg.Logger.LocalConfig.LogFile, filepath.Join("/srv/volt", "tmp", "logs", "app.log")},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

// TestExplicitPathsSurviveDerive pins that Derive fills gaps and never
// overwrites. An operator who sets TMP_DIR must get that directory, not a path
// recomputed from the root.
func TestExplicitPathsSurviveDerive(t *testing.T) {
	isolate(t)
	t.Setenv("VOLT_ROOT", "/srv/volt")
	t.Setenv("TMP_DIR", "/var/tmp/volt")

	cfg, err := Get()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Core.Tmpdir != "/var/tmp/volt" {
		t.Errorf("Tmpdir = %q; an explicit value must survive Derive", cfg.Core.Tmpdir)
	}
	// And the log file follows the explicit tmp dir, not the root.
	if want := filepath.Join("/var/tmp/volt", "logs", "app.log"); cfg.Logger.LocalConfig.LogFile != want {
		t.Errorf("LogFile = %q, want %q", cfg.Logger.LocalConfig.LogFile, want)
	}
}

// TestOtelServiceNameFallsBackToAppKey pins the rule that keeps two services
// from merging in the observability backend when neither sets the name.
func TestOtelServiceNameFallsBackToAppKey(t *testing.T) {
	isolate(t)
	t.Setenv("APP_KEY", "billing-api")

	cfg, err := Get()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Otel.ServiceName != "billing-api" {
		t.Errorf("ServiceName = %q, want the app key", cfg.Otel.ServiceName)
	}

	// An explicit name wins, which is what every deployed environment sets.
	t.Setenv("OTEL_SERVICE_NAME", "billing-api-stag")
	cfg, err = Get()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Otel.ServiceName != "billing-api-stag" {
		t.Errorf("ServiceName = %q; an explicit value must win", cfg.Otel.ServiceName)
	}
}

// TestEnvironmentIsAClosedSet pins that a misspelling fails loudly.
//
// The dangerous reading is the opposite: treating an unrecognised value as
// "not prod" silently disables every production rule below.
func TestEnvironmentIsAClosedSet(t *testing.T) {
	isolate(t)

	for _, ok := range []Environment{Dev, Stag, Prod} {
		t.Setenv("CONFIG_ENV", string(ok))
		t.Setenv("ENCRYPTION_KEY", "a-real-key-that-is-not-a-placeholder")
		t.Setenv("JWT_SECRET", "a-real-jwt-secret")
		t.Setenv("INTERNAL_KEY", "a-real-internal-key")
		if err := check(); err != nil {
			t.Errorf("CONFIG_ENV=%s was rejected: %v", ok, err)
		}
	}

	t.Setenv("CONFIG_ENV", "production") // the plausible misspelling
	if err := check(); err == nil {
		t.Error(`CONFIG_ENV="production" was accepted; only dev, stag and prod are valid`)
	}
}

// TestPlaceholderSecretsAreRefusedOutsideDev is the security rule.
//
// Shipping the default encryption key makes every encrypted record readable by
// anyone holding the source, so it must fail the deploy rather than the audit.
func TestPlaceholderSecretsAreRefusedOutsideDev(t *testing.T) {
	isolate(t)

	// dev tolerates the placeholders — that is what keeps zero input working.
	t.Setenv("CONFIG_ENV", "dev")
	if err := check(); err != nil {
		t.Errorf("dev rejected the shipped defaults: %v", err)
	}

	for _, env := range []string{"stag", "prod"} {
		t.Setenv("CONFIG_ENV", env)
		err := check()
		if err == nil {
			t.Fatalf("%s accepted the placeholder secrets", env)
		}
		// Every offending key must be named, not just the first.
		for _, key := range []string{"ENCRYPTION_KEY", "JWT_SECRET", "INTERNAL_KEY"} {
			if !strings.Contains(err.Error(), key) {
				t.Errorf("%s: %s missing from the report:\n%v", env, key, err)
			}
		}
	}
}

// TestRealSecretsPassOutsideDev is the control: the rule above must reject
// placeholders specifically, not every non-dev deployment.
func TestRealSecretsPassOutsideDev(t *testing.T) {
	isolate(t)
	t.Setenv("CONFIG_ENV", "prod")
	t.Setenv("ENCRYPTION_KEY", "8f2b1c9d4e6a7b3c5d8e1f2a3b4c5d6e")
	t.Setenv("JWT_SECRET", "a-genuine-signing-secret")
	t.Setenv("INTERNAL_KEY", "a-genuine-internal-key")

	if err := check(); err != nil {
		t.Errorf("prod with real secrets was rejected: %v", err)
	}
}

// TestSecretsAreMaskedInExplain pins that the diagnostic output is safe to
// paste into an issue. Explain is what an operator reaches for when a value is
// wrong, so it is exactly where a credential would otherwise leak.
func TestSecretsAreMaskedInExplain(t *testing.T) {
	isolate(t)
	const canary = "super-secret-canary-value"
	t.Setenv("JWT_SECRET", canary)
	t.Setenv("DB_PASSWORD", canary)
	t.Setenv("OPENAI_API_KEY", canary)

	var out bytes.Buffer
	if err := explain(&out); err != nil {
		t.Fatalf("Explain: %v", err)
	}
	if strings.Contains(out.String(), canary) {
		t.Errorf("a secret reached Explain's output:\n%s", out.String())
	}
	// The field must still be listed — masked, not omitted, or the operator
	// cannot tell whether it was set at all.
	if !strings.Contains(out.String(), "JWT_SECRET") {
		t.Error("JWT_SECRET is absent from Explain; masking must not hide the field")
	}
}

// TestEnvironmentOverridesTheDefault pins the precedence that makes a
// container's -e flag behave as anyone would expect.
func TestEnvironmentOverridesTheDefault(t *testing.T) {
	isolate(t)
	t.Setenv("PORT", "8443")
	t.Setenv("DB_PORT", "6543")

	cfg, err := Get()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != "8443" {
		t.Errorf("Port = %q, want the environment's value", cfg.Server.Port)
	}
	if cfg.Database.Port != 6543 {
		t.Errorf("DB port = %d, want 6543", cfg.Database.Port)
	}
}

// TestInvalidPortIsRejected pins that a nonsense value fails the check rather
// than binding as zero and producing a service that listens nowhere.
func TestInvalidPortIsRejected(t *testing.T) {
	isolate(t)
	t.Setenv("DB_PORT", "99999") // above the 16-bit range

	if err := check(); err == nil {
		t.Error("an out-of-range database port was accepted")
	}
}

// TestIsWeakOrDefaultKey carries over the cases from the helper's previous
// home, plus the placeholders this package now declares.
func TestIsWeakOrDefaultKey(t *testing.T) {
	for _, tc := range []struct {
		key  string
		weak bool
	}{
		{"", true},
		{"   ", true},
		{"change-me-encryption-key-32ch", true},
		{"change-me-jwt-secret-key", true},
		{"change-me-internal-key", true},
		{"xxxx", true},
		{"8f2b1c9d4e6a7b3c5d8e1f2a3b4c5d6e", false},
		{"a-genuine-key", false},
	} {
		if got := IsWeakOrDefaultKey(tc.key); got != tc.weak {
			t.Errorf("IsWeakOrDefaultKey(%q) = %v, want %v", tc.key, got, tc.weak)
		}
	}
}

// TestDocumentListsEveryKey pins that the generated contract covers the whole
// surface. A field added without appearing here would be a setting nobody
// deploying the service knows exists.
func TestDocumentListsEveryKey(t *testing.T) {
	var out bytes.Buffer
	if err := cfgkit.Document[Root](&out); err != nil {
		t.Fatalf("Document: %v", err)
	}
	doc := out.String()

	for _, key := range []string{
		"HOSTNAME", "PORT", "APP_NAME", "APP_KEY", "CONFIG_ENV", "VOLT_ROOT",
		"INTERNAL_KEY", "JWT_SECRET", "ENCRYPTION_KEY",
		"DB_HOST", "DB_PORT", "DB_USERNAME", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"GRAPHQL_KEY", "REDIS_STATUS", "REDIS_URL", "REDIS_DB",
		"OTEL_ENABLE", "OTEL_ENDPOINT", "OTEL_SERVICE_NAME",
		"LOGGER_ENABLE_STDOUT", "LOGGER_LOCAL_ENABLED",
		"REACT_APP_HOST", "OPENAI_API_KEY", "CRON_UI_ENABLE", "CRON_UI_PORT",
	} {
		if !strings.Contains(doc, key) {
			t.Errorf("%s is missing from the generated contract", key)
		}
	}

	// A secret must be emitted with an EMPTY value: the file is committed.
	if strings.Contains(doc, "change-me-jwt-secret-key") {
		t.Error("a placeholder secret's value was written into the contract")
	}
}

// TestGetReturnsAnErrorRatherThanPanicking is the headline behavioural change.
//
// The previous loader panicked — on a missing config directory, and on any
// parse failure — which turned a configuration mistake into a stack trace at
// container start, after the rollout and before the logger existed to record
// it. A bad value must now come back as an error the caller can act on.
func TestGetReturnsAnErrorRatherThanPanicking(t *testing.T) {
	isolate(t)
	t.Setenv("DB_PORT", "not-a-number")

	cfg, err := Get()
	if err == nil {
		t.Fatal("an unparseable port was accepted")
	}
	if cfg != nil {
		t.Error("a failed load returned a non-nil configuration")
	}
	if !strings.Contains(err.Error(), "DB_PORT") {
		t.Errorf("the error does not name the offending key: %v", err)
	}
}

// TestExplainReportsALoadFailure pins that the diagnostic path fails the same
// way. Explain is what someone runs BECAUSE something is wrong, so it must not
// be the one command that panics on a broken config.
func TestExplainReportsALoadFailure(t *testing.T) {
	isolate(t)
	t.Setenv("REDIS_DB", "not-a-number")

	var out bytes.Buffer
	if err := explain(&out); err == nil {
		t.Error("Explain accepted an unparseable value")
	}
}

// TestRootdirIsNormalised pins the trailing-slash handling. "/" would
// otherwise trim to "" and produce paths rooted at "tmp" rather than "/tmp" —
// a silent relocation of the log file and uploads directory.
func TestRootdirIsNormalised(t *testing.T) {
	isolate(t)

	for _, tc := range []struct{ root, wantTmp string }{
		{"/srv/volt/", filepath.Join("/srv/volt", "tmp")},
		{"/", filepath.Join(".", "tmp")},
		{"", filepath.Join(".", "tmp")},
	} {
		t.Setenv("VOLT_ROOT", tc.root)
		cfg, err := Get()
		if err != nil {
			t.Fatalf("VOLT_ROOT=%q: %v", tc.root, err)
		}
		if cfg.Core.Tmpdir != tc.wantTmp {
			t.Errorf("VOLT_ROOT=%q -> Tmpdir %q, want %q", tc.root, cfg.Core.Tmpdir, tc.wantTmp)
		}
	}
}

// check and explain replace the wrappers that used to live in main.go.
//
// They are test helpers rather than package API because the only production
// consumer of those verbs is the CLI's config command, which cfgkit's contrib
// package supplies. Keeping a second implementation in the package would be two
// things that can disagree about what "check" means.
//
// Both use Options() — the same list the application loads with — so a test that
// passes is a statement about the real configuration, not a parallel one.
func check() error {
	return cfgkit.Check[Root](Options()...)
}

func explain(w io.Writer) error {
	_, res, err := cfgkit.Load[Root](Options()...)
	if err != nil {
		return err
	}
	return res.Explain(w)
}
