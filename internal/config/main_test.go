package config

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDeserializeConfigYaml(t *testing.T) {
	testYaml := `
driver: file
file:
  directory: /foo/bar
`

	wizard := configWizard{}

	t.Run("can deserialize config yaml", func(t *testing.T) {
		got, _ := wizard.deserializeConfigYaml([]byte(testYaml))

		want := &Configuration{
			Driver: "file",
			FileConfiguration: &FileConfiguration{
				Directory: "/foo/bar",
			},
			LDAPConfiguration:     nil,
			DatabaseConfiguration: nil,
		}

		if !cmp.Equal(got, want) {
			t.Errorf("got: %+v, want: %+v", got, want)
		}
	})
}

func TestConfigWizardGetConfigurationWithLDAPBindPassFile(t *testing.T) {
	testYaml := `
driver: ldap
ldap:
  bind_pass_file: ../../test/secret_file
`
	passwordContent := "test_secret"
	passwordFile := "../../test/secret_file"

	wizard := configWizard{}

	t.Run("config wizard can read password from file", func(t *testing.T) {
		got, err := wizard.processConfigYaml([]byte(testYaml))
		if err != nil {
			t.Fatal(err)
		}

		if got.LDAPConfiguration.BindPass != passwordContent {
			t.Errorf("expected BindPass to be '%s', got '%s'", passwordContent, got.LDAPConfiguration.BindPass)
		}

		if got.LDAPConfiguration.BindPassFile != passwordFile {
			t.Errorf("expected BindPassFile to be '%s', got '%s'", passwordFile, got.LDAPConfiguration.BindPassFile)
		}
	})
}

func TestConfigWizardGetConfigurationWithBothLDAPPasswordAndFile(t *testing.T) {
	testYaml := `
driver: ldap
ldap:
  bind_pass: password
  bind_pass_file: ../../test/secret_file
`
	wizard := configWizard{}
	t.Run("config wizard errors when both bind_pass and bind_pass_file are specified", func(t *testing.T) {
		_, err := wizard.processConfigYaml([]byte(testYaml))

		if err == nil {
			t.Fatal("expected error when both bind_pass and bind_pass_file are specified")
		}

		if err.Error() != "must specify either bind_pass or bind_pass_file" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestConfigWizardGetConfigurationWithMissingLDAPPasswordFile(t *testing.T) {
	testYaml := `
driver: ldap
ldap:
  bind_pass_file: /non/existent/file
`
	wizard := configWizard{}
	t.Run("config wizard errors when LDAP password file is missing", func(t *testing.T) {
		_, err := wizard.processConfigYaml([]byte(testYaml))
		if err == nil {
			t.Fatal("expected error when LDAP password file is missing")
		}

		if !strings.Contains(err.Error(), "cannot read LDAP bind password file") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestConfigWizardGetConfigurationWithDatabaseURLFile(t *testing.T) {
	testYaml := `
driver: sql
database:
  url_file: ../../test/secret_file
`
	urlContent := "test_secret"
	urlFile := "../../test/secret_file"

	wizard := configWizard{}

	t.Run("config wizard can read database URL from file", func(t *testing.T) {
		got, err := wizard.processConfigYaml([]byte(testYaml))
		if err != nil {
			t.Fatal(err)
		}

		if got.DatabaseConfiguration.URL != urlContent {
			t.Errorf("expected URL to be '%s', got '%s'", urlContent, got.DatabaseConfiguration.URL)
		}

		if got.DatabaseConfiguration.URLFile != urlFile {
			t.Errorf("expected URLFile to be '%s', got '%s'", urlFile, got.DatabaseConfiguration.URLFile)
		}
	})
}

func TestConfigWizardGetConfigurationWithBothDatabaseURLAndURLFile(t *testing.T) {
	testYaml := `
driver: sql
database:
  url: "postgres://user:password@localhost:5432/dbname?sslmode=disable"
  url_file: ../../test/secret_file
`
	wizard := configWizard{}
	t.Run("config wizard errors when both url and url_file are specified", func(t *testing.T) {
		_, err := wizard.processConfigYaml([]byte(testYaml))
		if err == nil {
			t.Fatal("expected error when both url and url_file are specified")
		}

		if err.Error() != "must specify either url or url_file" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestConfigWizardGetConfigurationWithMissingDatabaseURLFile(t *testing.T) {
	testYaml := `
driver: sql
database:
  url_file: /non/existant/file
`
	wizard := configWizard{}
	t.Run("config wizard errors when database URL file is missing", func(t *testing.T) {
		_, err := wizard.processConfigYaml([]byte(testYaml))
		if err == nil {
			t.Fatal("expected error when database URL file is missing")
		}

		if !strings.Contains(err.Error(), "cannot read database URL file") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestConfigWizardGetConfiguration(t *testing.T) {
	t.Run("config wizard can read config file and parse configuration", func(t *testing.T) {
		wizard := NewConfigWizard("../../test/config.yml", false)

		got, err := wizard.GetConfiguration()
		if err != nil {
			t.Fatal(err)
		}

		want := Configuration{
			Driver: "file",
			FileConfiguration: &FileConfiguration{
				Directory: "./test/",
			},
		}

		if cmp.Equal(got, want) {
			t.Fatalf("got: %+v, want: %+v", got, want)
		}
	})

	t.Run("config wizard errors on nonexistent config file", func(t *testing.T) {
		wizard := NewConfigWizard("missingno", false)

		got, err := wizard.GetConfiguration()

		if err == nil {
			t.Fatalf("expected wizard to error on missing file, got %+v", got)
		}

		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("error should be NotExist: %+v", err)
		}
	})
}

func TestConfigWizardExpandEnvs(t *testing.T) {
	t.Run("config wizard can expand environment variables in configuration file", func(t *testing.T) {
		wizard := NewConfigWizard("../../test/config-with-envs.yml", true)

		databaseUrl := "test_dabase_url"
		t.Setenv("DATABASE_URL", databaseUrl)

		got, err := wizard.GetConfiguration()

		if err != nil {
			t.Fatal(err)
		}

		want := &Configuration{
			Driver: "sql",
			DatabaseConfiguration: &DatabaseConfiguration{
				Driver:      "postgres",
				URL:         databaseUrl,
				URLFile:     "",
				Table:       "users",
				KeyColumn:   "email",
				ColumnNames: []string{"email", "handle", "name"},
				Template:    "",
			},
		}

		if !cmp.Equal(got, want) {
			t.Errorf("got: %+v, want: %+v", got, want)
		}
	})
}
