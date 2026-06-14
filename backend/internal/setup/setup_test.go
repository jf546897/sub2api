package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecideAdminBootstrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		totalUsers int64
		adminUsers int64
		should     bool
		reason     string
	}{
		{
			name:       "empty database should create admin",
			totalUsers: 0,
			adminUsers: 0,
			should:     true,
			reason:     adminBootstrapReasonEmptyDatabase,
		},
		{
			name:       "admin exists should skip",
			totalUsers: 10,
			adminUsers: 1,
			should:     false,
			reason:     adminBootstrapReasonAdminExists,
		},
		{
			name:       "users exist without admin should skip",
			totalUsers: 5,
			adminUsers: 0,
			should:     false,
			reason:     adminBootstrapReasonUsersExistWithoutAdmin,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := decideAdminBootstrap(tc.totalUsers, tc.adminUsers)
			if got.shouldCreate != tc.should {
				t.Fatalf("shouldCreate=%v, want %v", got.shouldCreate, tc.should)
			}
			if got.reason != tc.reason {
				t.Fatalf("reason=%q, want %q", got.reason, tc.reason)
			}
		})
	}
}

func TestSetupDefaultAdminConcurrency(t *testing.T) {
	t.Run("simple mode admin uses higher concurrency", func(t *testing.T) {
		t.Setenv("RUN_MODE", "simple")
		if got := setupDefaultAdminConcurrency(); got != simpleModeAdminConcurrency {
			t.Fatalf("setupDefaultAdminConcurrency()=%d, want %d", got, simpleModeAdminConcurrency)
		}
	})

	t.Run("standard mode keeps existing default", func(t *testing.T) {
		t.Setenv("RUN_MODE", "standard")
		if got := setupDefaultAdminConcurrency(); got != defaultUserConcurrency {
			t.Fatalf("setupDefaultAdminConcurrency()=%d, want %d", got, defaultUserConcurrency)
		}
	})
}

func TestWriteConfigFileKeepsDefaultUserConcurrency(t *testing.T) {
	t.Setenv("RUN_MODE", "simple")
	t.Setenv("DATA_DIR", t.TempDir())

	if err := writeConfigFile(&SetupConfig{}); err != nil {
		t.Fatalf("writeConfigFile() error = %v", err)
	}

	data, err := os.ReadFile(GetConfigFilePath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if !strings.Contains(string(data), "user_concurrency: 5") {
		t.Fatalf("config missing default user concurrency, got:\n%s", string(data))
	}
}

func TestEnsureDataDirCreatesMissingDirectory(t *testing.T) {
	base := t.TempDir()
	target := filepath.Join(base, "nested", "data")
	t.Setenv("DATA_DIR", target)

	if err := ensureDataDir(); err != nil {
		t.Fatalf("ensureDataDir() error = %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("ensureDataDir() did not create directory %q", target)
	}
}

func TestBuildDatabaseConnectionDSNsUsesPostgresForBootstrap(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:     "db",
		Port:     5432,
		User:     "sub2api",
		Password: "secret",
		DBName:   "sub2api",
		SSLMode:  "disable",
	}

	bootstrapDSN, targetDSN := buildDatabaseConnectionDSNs(cfg)

	if !strings.Contains(bootstrapDSN, "dbname=postgres") {
		t.Fatalf("bootstrap DSN = %q, want default postgres database", bootstrapDSN)
	}
	if strings.Contains(bootstrapDSN, "dbname=sub2api") {
		t.Fatalf("bootstrap DSN = %q, should not connect to target database before checking/creating it", bootstrapDSN)
	}
	if !strings.Contains(targetDSN, "dbname=sub2api") {
		t.Fatalf("target DSN = %q, want configured database", targetDSN)
	}
}

func TestBuildPostgresDSNEscapesSpecialCharacters(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:     "db.internal",
		Port:     5432,
		User:     "sub2 api",
		Password: `pa ss'\word`,
		DBName:   "sub2 api",
		SSLMode:  "prefer",
	}

	_, targetDSN := buildDatabaseConnectionDSNs(cfg)

	for _, want := range []string{
		`user='sub2 api'`,
		`password='pa ss\'\\word'`,
		`dbname='sub2 api'`,
		"sslmode=prefer",
	} {
		if !strings.Contains(targetDSN, want) {
			t.Fatalf("target DSN = %q, want %q", targetDSN, want)
		}
	}
}

func TestValidateDatabaseConfigForConnection(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *DatabaseConfig
		wantErr bool
	}{
		{
			name: "allows quoted identifier names and prefer sslmode",
			cfg: &DatabaseConfig{
				Host:    "db.internal",
				Port:    5432,
				User:    "sub2 api",
				DBName:  "sub2 api",
				SSLMode: "prefer",
			},
		},
		{
			name:    "rejects nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "rejects invalid port",
			cfg: &DatabaseConfig{
				Host:    "db.internal",
				Port:    0,
				User:    "sub2api",
				DBName:  "sub2api",
				SSLMode: "disable",
			},
			wantErr: true,
		},
		{
			name: "rejects empty database name",
			cfg: &DatabaseConfig{
				Host:    "db.internal",
				Port:    5432,
				User:    "sub2api",
				DBName:  " ",
				SSLMode: "disable",
			},
			wantErr: true,
		},
		{
			name: "rejects invalid sslmode",
			cfg: &DatabaseConfig{
				Host:    "db.internal",
				Port:    5432,
				User:    "sub2api",
				DBName:  "sub2api",
				SSLMode: "unsafe",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateDatabaseConfigForConnection(tc.cfg)
			if tc.wantErr && err == nil {
				t.Fatalf("validateDatabaseConfigForConnection() error = nil, want error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("validateDatabaseConfigForConnection() error = %v", err)
			}
		})
	}
}
