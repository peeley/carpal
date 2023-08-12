package config

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDeserializeConfigYaml(t *testing.T) {
	testYaml := `
driver: file
file:
  directory: /foo/bar
`

	wizard := configWizard {}

	t.Run("can deserialize config yaml", func(t *testing.T) {
		got, _ := wizard.deserializeConfigYaml([]byte(testYaml))

		want := &Configuration{
			Driver: "file",
			FileConfiguration: &FileConfiguration{
				Directory: "/foo/bar",
			},
			LDAPConfiguration: nil,
			DatabaseConfiguration: nil,
		}

		if !cmp.Equal(got, want) {
			t.Errorf("got: %+v, want: %+v",	got, want)
		}
	})
}

func TestConfigWizardGetConfiguration(t *testing.T) {
	t.Run("config wizard can read config file and parse configuration", func(t *testing.T) {
		wizard := NewConfigWizard("../../test/config.yml")

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
		wizard := NewConfigWizard("missingno")

		got, err := wizard.GetConfiguration()

		if err == nil {
			t.Fatalf("expected wizard to error on missing file, got %+v", got)
		}

		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("error should be NotExist: %+v", err)
		}
	})
}
