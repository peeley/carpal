package config

import (
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
