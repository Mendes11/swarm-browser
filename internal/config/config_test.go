package config

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/mendes11/swarm-browser/internal/core/models"
	"github.com/spf13/viper"
)

func TestLoadClustersConfigFromViper(t *testing.T) {
	yamlContent := `clusters:
  prod:
    name: "Production"
    host: "manager-01.example.com"
    nodes:
      manager-01:
        host: "manager-01.example.com"
        hostname: "manager-01.example.com"
  staging:
    name: "Staging"
    host: "swarm-02.test.com"
    nodes:
      swarm-02:
        host: "swarm-02.test.com"
        hostname: "swarm-02.test.com"
      worker-01:
        host: "worker-01.test.com"
        hostname: "worker-01.test.com"`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clusters.yml")

	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	v := viper.New()
	v.SetConfigFile(tmpFile)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config with viper: %v", err)
	}

	cfg, err := LoadClustersConfigFromViper(v)
	if err != nil {
		t.Fatalf("Failed to load clusters config from viper: %v", err)
	}

	if len(cfg.Clusters) != 2 {
		t.Errorf("Expected 2 clusters, got %d", len(cfg.Clusters))
	}

	prod, exists := cfg.GetCluster("prod")
	if !exists {
		t.Fatal("Expected to find 'prod' cluster")
	}
	if prod.Name != "Production" {
		t.Errorf("Expected cluster name 'Production', got '%s'", prod.Name)
	}
	if prod.Host != "manager-01.example.com" {
		t.Errorf("Expected host 'manager-01.example.com', got '%s'", prod.Host)
	}

	staging, exists := cfg.GetCluster("staging")
	if !exists {
		t.Fatal("Expected to find 'staging' cluster")
	}
	if len(staging.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in staging cluster, got %d", len(staging.Nodes))
	}

	node, found := staging.GetNodeByHostname("worker-01.test.com")
	if !found {
		t.Error("Expected to find node with hostname 'worker-01.test.com'")
	}
	if node.Host != "worker-01.test.com" {
		t.Errorf("Expected node host 'worker-01.test.com', got '%s'", node.Host)
	}
}

func TestLoadClustersConfigWithCommands(t *testing.T) {
	yamlContent := `commands:
  rails-console:
    name: "Rails Console"
    cmd: "bundle exec rails console"
    match_services: ["*web*", "*api*"]
  postgres-shell:
    name: "PostgreSQL Shell"
    cmd: "psql -U postgres"
    match_services: ["*postgres*", "*db*"]
  view-logs:
    name: "View Logs"
    cmd: "tail -f /var/log/*.log"
clusters:
  prod:
    name: "Production"
    host: "manager-01"
    commands:
      rails-console: {}
      postgres-shell:
        match_services: ["prod-db*"]
      prod-backup:
        name: "DB Backup"
        cmd: "pg_dump -U admin production_db"
        match_services: ["*postgres*"]
    nodes:
      manager-01:
        host: "manager-01"
        hostname: "manager-01"
  staging:
    name: "Staging"
    host: "staging-01"
    nodes:
      staging-01:
        host: "staging-01"
        hostname: "staging-01"`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clusters.yml")
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	cfg, err := LoadClustersConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test global commands parsed correctly
	if len(cfg.Commands) != 3 {
		t.Fatalf("Expected 3 global commands, got %d", len(cfg.Commands))
	}
	railsCmd, ok := cfg.Commands["rails-console"]
	if !ok {
		t.Fatal("Expected to find 'rails-console' in global commands")
	}
	if railsCmd.Name != "Rails Console" {
		t.Errorf("Expected name 'Rails Console', got '%s'", railsCmd.Name)
	}
	if railsCmd.Cmd != "bundle exec rails console" {
		t.Errorf("Expected cmd 'bundle exec rails console', got '%s'", railsCmd.Cmd)
	}
	if len(railsCmd.MatchServices) != 2 {
		t.Errorf("Expected 2 match_services patterns, got %d", len(railsCmd.MatchServices))
	}

	// Test cluster commands parsed correctly
	prod, exists := cfg.GetCluster("prod")
	if !exists {
		t.Fatal("Expected to find 'prod' cluster")
	}
	if len(prod.Commands) != 3 {
		t.Fatalf("Expected 3 cluster commands for prod, got %d", len(prod.Commands))
	}

	// Test staging has no cluster commands
	staging, exists := cfg.GetCluster("staging")
	if !exists {
		t.Fatal("Expected to find 'staging' cluster")
	}
	if len(staging.Commands) != 0 {
		t.Errorf("Expected 0 cluster commands for staging, got %d", len(staging.Commands))
	}
}

func TestResolveCommandsForCluster_InheritedEmptyOverride(t *testing.T) {
	yamlContent := `commands:
  rails-console:
    name: "Rails Console"
    cmd: "bin/docker-entrypoint bin/rails c"
    match_services: ["*app"]
  postgres-console:
    name: "Postgres Console"
    cmd: "psql -U postgres"
    match_services: ["*pg"]
clusters:
  test:
    name: "test Cluster"
    host: "manager-01.example.com"
    commands:
      rails-console:
        match_services: ["tally*"]
      postgres-console: {}
    nodes:
      manager-01:
        host: "manager-01.example.com"
        hostname: "manager-01.example.com"`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clusters.yml")
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	cfg, err := LoadClustersConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cmds := cfg.ResolveCommandsForCluster("test")
	if len(cmds) != 2 {
		t.Fatalf("Expected 2 resolved commands, got %d: %+v", len(cmds), cmds)
	}

	// Find the postgres-console command and verify it inherited from global
	var pgCmd *models.Command
	for _, cmd := range cmds {
		if cmd.Name == "Postgres Console" {
			pgCmd = &cmd
			break
		}
	}
	if pgCmd == nil {
		t.Fatal("Expected to find 'Postgres Console' in resolved commands")
	}
	if pgCmd.Cmd != "psql -U postgres" {
		t.Errorf("Expected inherited cmd 'psql -U postgres', got '%s'", pgCmd.Cmd)
	}
	if len(pgCmd.MatchServices) != 1 || pgCmd.MatchServices[0] != "*pg" {
		t.Errorf("Expected inherited match_services [\"*pg\"], got %v", pgCmd.MatchServices)
	}

	// Verify that the pattern actually matches the service name "tally_pg"
	matched, _ := path.Match(pgCmd.MatchServices[0], "tally_pg")
	if !matched {
		t.Errorf("Expected pattern %q to match service name 'tally_pg'", pgCmd.MatchServices[0])
	}
}

func TestResolveCommandsForCluster_InheritedEmptyOverrideViper(t *testing.T) {
	yamlContent := `commands:
  rails-console:
    name: "Rails Console"
    cmd: "bin/docker-entrypoint bin/rails c"
    match_services: ["*app"]
  postgres-console:
    name: "Postgres Console"
    cmd: "psql -U postgres"
    match_services: ["*pg"]
clusters:
  test:
    name: "test Cluster"
    host: "manager-01.example.com"
    commands:
      rails-console:
        match_services: ["tally*"]
      postgres-console: {}
    nodes:
      manager-01:
        host: "manager-01.example.com"
        hostname: "manager-01.example.com"`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clusters.yml")
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	v := viper.New()
	v.SetConfigFile(tmpFile)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config with viper: %v", err)
	}

	cfg, err := LoadClustersConfigFromViper(v)
	if err != nil {
		t.Fatalf("Failed to load config from viper: %v", err)
	}

	cmds := cfg.ResolveCommandsForCluster("test")
	if len(cmds) != 2 {
		t.Fatalf("Expected 2 resolved commands, got %d: %+v", len(cmds), cmds)
	}

	var pgCmd *models.Command
	for _, cmd := range cmds {
		if cmd.Name == "Postgres Console" {
			pgCmd = &cmd
			break
		}
	}
	if pgCmd == nil {
		t.Fatal("Expected to find 'Postgres Console' in resolved commands")
	}
	if pgCmd.Cmd != "psql -U postgres" {
		t.Errorf("Expected inherited cmd 'psql -U postgres', got '%s'", pgCmd.Cmd)
	}
	if len(pgCmd.MatchServices) != 1 || pgCmd.MatchServices[0] != "*pg" {
		t.Errorf("Expected inherited match_services [\"*pg\"], got %v", pgCmd.MatchServices)
	}
}

func TestResolveCommandsForCluster(t *testing.T) {
	yamlContent := `commands:
  rails-console:
    name: "Rails Console"
    cmd: "bundle exec rails console"
    match_services: ["*web*"]
  postgres-shell:
    name: "PostgreSQL Shell"
    cmd: "psql -U postgres"
    match_services: ["*postgres*"]
  view-logs:
    name: "View Logs"
    cmd: "tail -f /var/log/*.log"
clusters:
  prod:
    name: "Production"
    host: "manager-01"
    commands:
      rails-console: {}
      prod-backup:
        name: "DB Backup"
        cmd: "pg_dump production_db"
    nodes:
      manager-01:
        host: "manager-01"
        hostname: "manager-01"
  staging:
    name: "Staging"
    host: "staging-01"
    nodes:
      staging-01:
        host: "staging-01"
        hostname: "staging-01"`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clusters.yml")
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	cfg, err := LoadClustersConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test prod: has commands section, so only listed commands
	prodCmds := cfg.ResolveCommandsForCluster("prod")
	if len(prodCmds) != 2 {
		t.Fatalf("Expected 2 resolved commands for prod, got %d", len(prodCmds))
	}
	// Check that rails-console inherited from global
	foundRails := false
	foundBackup := false
	for _, cmd := range prodCmds {
		if cmd.Name == "Rails Console" {
			foundRails = true
			if cmd.Cmd != "bundle exec rails console" {
				t.Errorf("Expected inherited cmd, got '%s'", cmd.Cmd)
			}
			if len(cmd.MatchServices) != 1 || cmd.MatchServices[0] != "*web*" {
				t.Errorf("Expected inherited match_services, got %v", cmd.MatchServices)
			}
		}
		if cmd.Name == "DB Backup" {
			foundBackup = true
			if cmd.Cmd != "pg_dump production_db" {
				t.Errorf("Expected cluster-specific cmd, got '%s'", cmd.Cmd)
			}
		}
	}
	if !foundRails {
		t.Error("Expected to find 'Rails Console' in resolved commands")
	}
	if !foundBackup {
		t.Error("Expected to find 'DB Backup' in resolved commands")
	}

	// Test staging: no commands section, so all global commands
	stagingCmds := cfg.ResolveCommandsForCluster("staging")
	if len(stagingCmds) != 3 {
		t.Fatalf("Expected 3 resolved commands for staging (all global), got %d", len(stagingCmds))
	}

	// Test nonexistent cluster
	nonexistent := cfg.ResolveCommandsForCluster("nonexistent")
	if nonexistent != nil {
		t.Errorf("Expected nil for nonexistent cluster, got %v", nonexistent)
	}
}
