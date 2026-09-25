package config

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// TestMain silences the package's slog output; CheckURLs logs on every
// successful dial and drowns the actual test results otherwise.
func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

// Every environment variable the package binds. Tests clear all of them so an
// ambient value in the developer's shell cannot change a result.
var configEnvKeys = []string{
	"BASE_URL", "BASE_PORT", "LOG_LEVEL",
	"MONGODB_HOST", "MONGODB_PORT", "MONGODB_USERNAME", "MONGODB_PASSWORD", "MONGODB_DATABASE",
	"RABBITMQ_HOST", "RABBITMQ_PORT", "RABBITMQ_USERNAME", "RABBITMQ_PASSWORD",
	"EMAIL_SMTP_HOST", "EMAIL_SMTP_PORT", "EMAIL_SMTP_USERNAME", "EMAIL_SMTP_PASSWORD", "EMAIL_FROM_ADDRESS",
	"OIDC_CLIENT_ID", "OIDC_CLIENT_SECRET", "OIDC_URL", "OIDC_REDIRECT_URL", "OIDC_SWAGGER_CLIENT_ID",
	"ADMIN_GROUPS",
}

const testAdminGroup = "admins"

// The smallest environment that satisfies validateConfiguration.
func requiredEnv() map[string]string {
	return map[string]string{
		"BASE_URL":               "localhost",
		"BASE_PORT":              "8080",
		"MONGODB_HOST":           "mongodb://localhost",
		"MONGODB_PORT":           "27017",
		"MONGODB_USERNAME":       "mongo-user",
		"MONGODB_PASSWORD":       "mongo-pass",
		"MONGODB_DATABASE":       "unimq",
		"RABBITMQ_HOST":          "http://localhost",
		"RABBITMQ_PORT":          "15672",
		"RABBITMQ_USERNAME":      "rmq-user",
		"RABBITMQ_PASSWORD":      "rmq-pass",
		"OIDC_CLIENT_ID":         "unimq-dashboard",
		"OIDC_CLIENT_SECRET":     "oidc-secret",
		"OIDC_URL":               "http://localhost:5556/dex",
		"OIDC_REDIRECT_URL":      "http://localhost:8080/api/v1/login/callback",
		"OIDC_SWAGGER_CLIENT_ID": "unimq-swagger",
		"ADMIN_GROUPS":           testAdminGroup,
	}
}

// newLoadTest isolates the three pieces of global state Load touches: the
// process environment, viper's singleton, and the working directory that .env
// is resolved against. Tests using it must not call t.Parallel.
func newLoadTest(t *testing.T) {
	t.Helper()

	saved := make(map[string]*string, len(configEnvKeys))
	for _, k := range configEnvKeys {
		if v, ok := os.LookupEnv(k); ok {
			val := v
			saved[k] = &val
		}
		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("unset %s: %v", k, err)
		}
	}
	t.Cleanup(func() {
		for _, k := range configEnvKeys {
			var err error
			if v, ok := saved[k]; ok {
				err = os.Setenv(k, *v)
			} else {
				err = os.Unsetenv(k)
			}
			if err != nil {
				t.Errorf("restore %s: %v", k, err)
			}
		}
	})

	viper.Reset()
	t.Cleanup(viper.Reset)

	t.Chdir(t.TempDir())
}

func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func writeEnvFile(t *testing.T, env map[string]string) {
	t.Helper()
	var b strings.Builder
	for k, v := range env {
		fmt.Fprintf(&b, "%s=%s\n", k, v)
	}
	if err := os.WriteFile(filepath.Join(".", ".env"), []byte(b.String()), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
}

// listenLocal returns a listening socket and its port. The listener stays open
// for the test so CheckURLs finds something to connect to.
func listenLocal(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() {
		if err := l.Close(); err != nil {
			t.Logf("close listener: %v", err)
		}
	})
	return l.Addr().(*net.TCPAddr).Port
}

// closedPort returns a port with nothing listening on it.
func closedPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err := l.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}
	return port
}

func TestNewConfigDefaults(t *testing.T) {
	c := NewConfig()

	if c.BaseURL != "localhost" {
		t.Errorf("BaseURL = %q, want localhost", c.BaseURL)
	}
	if c.BasePort != 8080 {
		t.Errorf("BasePort = %d, want 8080", c.BasePort)
	}
	if c.LogLevel != 0 {
		t.Errorf("LogLevel = %d, want 0 (slog.LevelInfo)", c.LogLevel)
	}
	if c.MongoDBPort != 27017 {
		t.Errorf("MongoDBPort = %d, want 27017", c.MongoDBPort)
	}
	if c.RabbitMQPort != 15672 {
		t.Errorf("RabbitMQPort = %d, want 15672", c.RabbitMQPort)
	}
	if c.Email == nil || c.OIDC == nil {
		t.Fatal("Email and OIDC must be non-nil; validateConfiguration dereferences OIDC")
	}
	if c.Email.EmailSMTPPort != 587 {
		t.Errorf("EmailSMTPPort = %d, want 587", c.Email.EmailSMTPPort)
	}
	if c.OIDC.OIDCSwaggerClientID != "unimq-swagger" {
		t.Errorf("OIDCSwaggerClientID = %q, want unimq-swagger", c.OIDC.OIDCSwaggerClientID)
	}
	if c.AdminGroups == nil {
		t.Error("AdminGroups should be an empty slice, not nil")
	}
}

func TestIsPresent(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"non-empty string", "value", true},
		{"empty string", "", false},
		{"positive int", 8080, true},
		{"zero int", 0, false},
		{"negative int", -4, true},
		{"populated slice", []string{"admins"}, true},
		{"empty slice", []string{}, false},
		{"nil slice", []string(nil), false},
		{"unsupported type", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPresent(tt.value); got != tt.want {
				t.Errorf("isPresent(%#v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// Zero is a legitimate value for a slog.Level, so any int config whose zero
// value is meaningful must not be routed through isPresent. LOG_LEVEL was
// checked here once and made the shipped LOG_LEVEL=0 unstartable.
func TestIsPresentRejectsZeroInt(t *testing.T) {
	if isPresent(0) {
		t.Fatal("expected isPresent(0) to be false")
	}
	if _, checked := requiredParameters(t)["LOG_LEVEL"]; checked {
		t.Error("LOG_LEVEL must not be a required parameter: slog.LevelInfo is 0, " +
			"which isPresent reports as missing, so the shipped LOG_LEVEL=0 cannot start")
	}
}

// requiredParameters rebuilds the parameter map validateConfiguration checks,
// by validating a config where every field is zero and reading back the names.
func requiredParameters(t *testing.T) map[string]bool {
	t.Helper()
	empty := &Config{OIDC: &OIDCConfig{}, Email: &EmailConfig{}}
	err := empty.validateConfiguration()
	if err == nil {
		t.Fatal("expected an all-zero config to fail validation")
	}
	out := map[string]bool{}
	for _, part := range strings.Split(err.Error(), ", ") {
		out[strings.TrimSuffix(part, " missing")] = true
	}
	return out
}

func TestValidateConfiguration(t *testing.T) {
	t.Run("all required present", func(t *testing.T) {
		c := NewConfig()
		for k, v := range requiredEnv() {
			applyToConfig(t, c, k, v)
		}
		if err := c.validateConfiguration(); err != nil {
			t.Fatalf("expected valid config, got %v", err)
		}
	})

	t.Run("reports each missing parameter by name", func(t *testing.T) {
		c := NewConfig()
		for k, v := range requiredEnv() {
			applyToConfig(t, c, k, v)
		}
		c.RabbitMQUsername = ""
		c.OIDC.OIDCClientID = ""

		err := c.validateConfiguration()
		if err == nil {
			t.Fatal("expected an error")
		}
		for _, want := range []string{"RABBITMQ_USERNAME missing", "OIDC_CLIENT_ID missing"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q does not mention %q", err, want)
			}
		}
	})

	t.Run("empty admin groups is rejected", func(t *testing.T) {
		c := NewConfig()
		for k, v := range requiredEnv() {
			applyToConfig(t, c, k, v)
		}
		c.AdminGroups = nil

		err := c.validateConfiguration()
		if err == nil || !strings.Contains(err.Error(), "ADMIN_GROUPS missing") {
			t.Fatalf("expected ADMIN_GROUPS to be required, got %v", err)
		}
	})
}

func applyToConfig(t *testing.T, c *Config, key, value string) {
	t.Helper()
	switch key {
	case "BASE_URL":
		c.BaseURL = value
	case "MONGODB_HOST":
		c.MongoDBHost = value
	case "MONGODB_USERNAME":
		c.MongoDBUsername = value
	case "MONGODB_PASSWORD":
		c.MongoDBPassword = value
	case "MONGODB_DATABASE":
		c.MongoDBDatabase = value
	case "RABBITMQ_HOST":
		c.RabbitMQHost = value
	case "RABBITMQ_USERNAME":
		c.RabbitMQUsername = value
	case "RABBITMQ_PASSWORD":
		c.RabbitMQPassword = value
	case "OIDC_CLIENT_ID":
		c.OIDC.OIDCClientID = value
	case "OIDC_CLIENT_SECRET":
		c.OIDC.OIDCClientSecret = value
	case "OIDC_URL":
		c.OIDC.OIDCURL = value
	case "OIDC_REDIRECT_URL":
		c.OIDC.OIDCRedirectURL = value
	case "OIDC_SWAGGER_CLIENT_ID":
		c.OIDC.OIDCSwaggerClientID = value
	case "ADMIN_GROUPS":
		c.AdminGroups = strings.Split(value, ",")
	case "BASE_PORT", "MONGODB_PORT", "RABBITMQ_PORT":
		// ports keep their non-zero defaults
	default:
		t.Fatalf("applyToConfig: unhandled key %q", key)
	}
}

func TestCheckParameters(t *testing.T) {
	t.Run("all present returns empty string", func(t *testing.T) {
		if got := checkParameters(map[string]bool{"A": true, "B": true}); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("single missing", func(t *testing.T) {
		if got := checkParameters(map[string]bool{"A": false}); got != "A missing" {
			t.Errorf("got %q, want %q", got, "A missing")
		}
	})

	t.Run("multiple missing are comma separated", func(t *testing.T) {
		got := checkParameters(map[string]bool{"A": false, "B": false, "C": true})
		// map iteration order is random, so assert on content rather than order
		if !strings.Contains(got, "A missing") || !strings.Contains(got, "B missing") {
			t.Errorf("got %q, want both A and B reported", got)
		}
		if strings.Contains(got, "C") {
			t.Errorf("got %q, should not mention the present parameter C", got)
		}
		if strings.Count(got, ",") != 1 {
			t.Errorf("got %q, want exactly one separator", got)
		}
	})
}

func TestOIDCConfigIsValid(t *testing.T) {
	full := func() *OIDCConfig {
		return &OIDCConfig{
			OIDCClientID:     "id",
			OIDCClientSecret: "secret",
			OIDCURL:          "http://dex",
			OIDCRedirectURL:  "http://callback",
		}
	}

	if !full().IsValid() {
		t.Fatal("a fully populated OIDC config should be valid")
	}

	tests := []struct {
		name  string
		blank func(*OIDCConfig)
	}{
		{"missing client id", func(o *OIDCConfig) { o.OIDCClientID = "" }},
		{"missing client secret", func(o *OIDCConfig) { o.OIDCClientSecret = "" }},
		{"missing url", func(o *OIDCConfig) { o.OIDCURL = "" }},
		{"missing redirect url", func(o *OIDCConfig) { o.OIDCRedirectURL = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := full()
			tt.blank(o)
			if o.IsValid() {
				t.Error("expected invalid")
			}
		})
	}

	// The swagger client id is deliberately not part of IsValid.
	t.Run("swagger client id is not required", func(t *testing.T) {
		o := full()
		o.OIDCSwaggerClientID = ""
		if !o.IsValid() {
			t.Error("swagger client id should not affect validity")
		}
	})
}

func TestEmailConfigIsValid(t *testing.T) {
	tests := []struct {
		name string
		cfg  EmailConfig
		want bool
	}{
		{"host and from address", EmailConfig{EmailSMTPHost: "smtp.example.com", EmailFromAddress: "a@b.c"}, true},
		{"missing host disables email", EmailConfig{EmailFromAddress: "a@b.c"}, false},
		{"missing from address", EmailConfig{EmailSMTPHost: "smtp.example.com"}, false},
		{"empty", EmailConfig{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadFromEnvFile(t *testing.T) {
	newLoadTest(t)
	writeEnvFile(t, requiredEnv())

	c := NewConfig()
	if err := c.Load(); err != nil {
		t.Fatalf("Load() = %v, want nil", err)
	}

	if c.MongoDBUsername != "mongo-user" {
		t.Errorf("MongoDBUsername = %q, want mongo-user", c.MongoDBUsername)
	}
	if c.RabbitMQPort != 15672 {
		t.Errorf("RabbitMQPort = %d, want 15672", c.RabbitMQPort)
	}
	if c.OIDC.OIDCClientID != "unimq-dashboard" {
		t.Errorf("OIDCClientID = %q, want unimq-dashboard", c.OIDC.OIDCClientID)
	}
}

// The deployed configuration: env vars only, no .env anywhere. Dockerfile.backend
// copies just the binary into the final stage, and the Helm chart injects config
// through a ConfigMap, so no container ever has a .env file.
func TestLoadFromEnvironmentWithoutEnvFile(t *testing.T) {
	newLoadTest(t)
	setEnv(t, requiredEnv())

	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		t.Fatalf("precondition: .env should not exist, got %v", err)
	}

	c := NewConfig()
	if err := c.Load(); err != nil {
		t.Fatalf("Load() = %v, want nil.\n"+
			"A missing .env must not be fatal — this is the container and Helm path.\n"+
			"loadConfigurationFile guards with errors.Is(err, new(fs.PathError)), but errors.Is\n"+
			"compares pointers, so a freshly allocated *fs.PathError never matches.\n"+
			"Use errors.As(err, &pathErr) with a declared var, or os.Stat(\".env\") before ReadInConfig.", err)
	}

	if c.MongoDBUsername != "mongo-user" {
		t.Errorf("MongoDBUsername = %q, want mongo-user", c.MongoDBUsername)
	}
	if c.OIDC.OIDCRedirectURL != requiredEnv()["OIDC_REDIRECT_URL"] {
		t.Errorf("OIDCRedirectURL = %q, want it read from the environment", c.OIDC.OIDCRedirectURL)
	}
}

func TestLoadEnvironmentOverridesEnvFile(t *testing.T) {
	newLoadTest(t)

	fileEnv := requiredEnv()
	fileEnv["MONGODB_USERNAME"] = "from-file"
	writeEnvFile(t, fileEnv)

	t.Setenv("MONGODB_USERNAME", "from-environment")

	c := NewConfig()
	if err := c.Load(); err != nil {
		t.Fatalf("Load() = %v", err)
	}

	if c.MongoDBUsername != "from-environment" {
		t.Errorf("MongoDBUsername = %q, want from-environment — an explicit environment "+
			"variable must win over .env so deployments can override the baked-in file",
			c.MongoDBUsername)
	}
}

// LOG_LEVEL=0 is slog.LevelInfo and is what .env.example and the Helm chart ship.
func TestLoadAcceptsZeroLogLevel(t *testing.T) {
	newLoadTest(t)
	env := requiredEnv()
	env["LOG_LEVEL"] = "0"
	writeEnvFile(t, env)

	c := NewConfig()
	if err := c.Load(); err != nil {
		t.Fatalf("Load() = %v, want nil: LOG_LEVEL=0 is the shipped default (info)", err)
	}
	if c.LogLevel != 0 {
		t.Errorf("LogLevel = %d, want 0", c.LogLevel)
	}
}

func TestLoadLogLevel(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value string
		want  int
	}{
		{"trace", "-8", -8},
		{"debug", "-4", -4},
		{"info", "0", 0},
		{"warn", "4", 4},
		{"error", "8", 8},
	} {
		t.Run(tt.name, func(t *testing.T) {
			newLoadTest(t)
			env := requiredEnv()
			env["LOG_LEVEL"] = tt.value
			writeEnvFile(t, env)

			c := NewConfig()
			if err := c.Load(); err != nil {
				t.Fatalf("Load() = %v", err)
			}
			if c.LogLevel != tt.want {
				t.Errorf("LogLevel = %d, want %d", c.LogLevel, tt.want)
			}
		})
	}
}

func TestLoadAdminGroups(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{"single group", testAdminGroup, []string{testAdminGroup}},
		{"comma separated", testAdminGroup + ",ops", []string{testAdminGroup, "ops"}},
		{"realistic values", "group.name1@local,group.name2@local", []string{"group.name1@local", "group.name2@local"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newLoadTest(t)
			env := requiredEnv()
			env["ADMIN_GROUPS"] = tt.value
			writeEnvFile(t, env)

			c := NewConfig()
			if err := c.Load(); err != nil {
				t.Fatalf("Load() = %v", err)
			}
			if len(c.AdminGroups) != len(tt.want) {
				t.Fatalf("AdminGroups = %#v, want %#v", c.AdminGroups, tt.want)
			}
			for i := range tt.want {
				if c.AdminGroups[i] != tt.want[i] {
					t.Errorf("AdminGroups[%d] = %q, want %q", i, c.AdminGroups[i], tt.want[i])
				}
			}
		})
	}
}

// Documents current behaviour: the comma split does not trim, so a space after
// a comma becomes part of the group name and will never match a groups claim.
// Tighten this assertion to "ops" when the values are trimmed.
func TestLoadAdminGroupsDoesNotTrimWhitespace(t *testing.T) {
	newLoadTest(t)
	env := requiredEnv()
	env["ADMIN_GROUPS"] = testAdminGroup + ", ops"
	writeEnvFile(t, env)

	c := NewConfig()
	if err := c.Load(); err != nil {
		t.Fatalf("Load() = %v", err)
	}

	if len(c.AdminGroups) != 2 {
		t.Fatalf("AdminGroups = %#v, want 2 entries", c.AdminGroups)
	}
	if c.AdminGroups[1] != " ops" {
		t.Logf("AdminGroups[1] = %q — whitespace now trimmed, tighten this test", c.AdminGroups[1])
	}
}

func TestLoadMissingRequiredParameter(t *testing.T) {
	tests := []struct {
		name    string
		omit    string
		wantErr string
	}{
		{"rabbitmq username", "RABBITMQ_USERNAME", "RABBITMQ_USERNAME missing"},
		{"rabbitmq password", "RABBITMQ_PASSWORD", "RABBITMQ_PASSWORD missing"},
		{"mongodb username", "MONGODB_USERNAME", "MONGODB_USERNAME missing"},
		{"mongodb password", "MONGODB_PASSWORD", "MONGODB_PASSWORD missing"},
		{"oidc client id", "OIDC_CLIENT_ID", "OIDC_CLIENT_ID missing"},
		{"oidc client secret", "OIDC_CLIENT_SECRET", "OIDC_CLIENT_SECRET missing"},
		{"oidc url", "OIDC_URL", "OIDC_URL missing"},
		{"oidc redirect url", "OIDC_REDIRECT_URL", "OIDC_REDIRECT_URL missing"},
		{"admin groups", "ADMIN_GROUPS", "ADMIN_GROUPS missing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newLoadTest(t)
			env := requiredEnv()
			delete(env, tt.omit)
			writeEnvFile(t, env)

			c := NewConfig()
			err := c.Load()
			if err == nil {
				t.Fatalf("Load() = nil, want an error naming %s", tt.omit)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Load() = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadRejectsNonNumericPort(t *testing.T) {
	newLoadTest(t)
	env := requiredEnv()
	env["RABBITMQ_PORT"] = "not-a-number"
	writeEnvFile(t, env)

	c := NewConfig()
	err := c.Load()
	if err == nil {
		t.Fatal("Load() = nil, want a decoding error for a non-numeric port")
	}
	if !strings.Contains(err.Error(), "RABBITMQ_PORT") {
		t.Errorf("Load() = %q, want it to name the offending key", err)
	}
}

func TestCheckURLs(t *testing.T) {
	// A config whose RabbitMQ and MongoDB endpoints both point at live listeners.
	reachable := func(t *testing.T) *Config {
		t.Helper()
		c := NewConfig()
		c.RabbitMQHost = "http://127.0.0.1"
		c.RabbitMQPort = listenLocal(t)
		c.MongoDBHost = "mongodb://127.0.0.1"
		c.MongoDBPort = listenLocal(t)
		c.OIDC = &OIDCConfig{
			OIDCClientID:     "id",
			OIDCClientSecret: "secret",
			OIDCURL:          "http://dex",
			OIDCRedirectURL:  "http://callback",
		}
		c.Email = &EmailConfig{EmailSMTPHost: "smtp.example.com", EmailFromAddress: "a@b.c"}
		return c
	}

	t.Run("both reachable and config valid", func(t *testing.T) {
		if err := reachable(t).CheckURLs(); err != nil {
			t.Fatalf("CheckURLs() = %v, want nil", err)
		}
	})

	t.Run("strips http, https and mongodb schemes", func(t *testing.T) {
		for _, scheme := range []string{"http://", "https://", ""} {
			c := reachable(t)
			c.RabbitMQHost = scheme + "127.0.0.1"
			if err := c.CheckURLs(); err != nil {
				t.Errorf("RabbitMQHost=%q: CheckURLs() = %v, want nil", c.RabbitMQHost, err)
			}
		}

		for _, scheme := range []string{"mongodb://", ""} {
			c := reachable(t)
			c.MongoDBHost = scheme + "127.0.0.1"
			if err := c.CheckURLs(); err != nil {
				t.Errorf("MongoDBHost=%q: CheckURLs() = %v, want nil", c.MongoDBHost, err)
			}
		}
	})

	t.Run("rabbitmq unreachable", func(t *testing.T) {
		c := reachable(t)
		c.RabbitMQPort = closedPort(t)

		err := c.CheckURLs()
		if err == nil {
			t.Fatal("CheckURLs() = nil, want a connection error")
		}
		if !strings.Contains(err.Error(), "RabbitMQ") {
			t.Errorf("CheckURLs() = %q, want it to name RabbitMQ", err)
		}
	})

	t.Run("mongodb unreachable", func(t *testing.T) {
		c := reachable(t)
		c.MongoDBPort = closedPort(t)

		err := c.CheckURLs()
		if err == nil {
			t.Fatal("CheckURLs() = nil, want a connection error")
		}
		if !strings.Contains(err.Error(), "MongoDB") {
			t.Errorf("CheckURLs() = %q, want it to name MongoDB", err)
		}
	})

	t.Run("invalid oidc is fatal", func(t *testing.T) {
		c := reachable(t)
		c.OIDC.OIDCClientSecret = ""

		err := c.CheckURLs()
		if err == nil {
			t.Fatal("CheckURLs() = nil, want an OIDC error")
		}
		if !strings.Contains(err.Error(), "OIDC") {
			t.Errorf("CheckURLs() = %q, want it to name OIDC", err)
		}
	})

	// Email is optional: an invalid block warns and disables notifications
	// rather than preventing startup.
	t.Run("invalid email is not fatal", func(t *testing.T) {
		c := reachable(t)
		c.Email = &EmailConfig{}

		if err := c.CheckURLs(); err != nil {
			t.Fatalf("CheckURLs() = %v, want nil: email is optional", err)
		}
	})
}
