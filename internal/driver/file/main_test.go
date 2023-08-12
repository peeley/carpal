package file

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/peeley/carpal/internal/config"
	"github.com/peeley/carpal/internal/driver"
	"github.com/peeley/carpal/internal/resource"
)

func TestFileDriverGetResource(t *testing.T) {
	config := config.Configuration{
		Driver: "file",
		FileConfiguration: &config.FileConfiguration{
			Directory: "../../../test/",
		},
	}

	fileDriver := NewFileDriver(config)

	t.Run("can get resource from file", func(t *testing.T){

		got, err := fileDriver.GetResource("acct:bob@foobar.com")

		if err != nil {
			t.Fatal(err)
		}

		profilePage := "https://www.example.com/~bob/"
		businessCard := "https://www.example.com/~bob/bob.vcf"
		want := &resource.Resource{
			Subject: "acct:bob@foobar.com",
			Aliases: []string{"mailto:bob@foobar.com", "https://mastodon/bob"},
			Properties: map[string]any{"http://webfinger.example/ns/name": "Bob Smith"},
			Links: []resource.Link{
				{
					Rel: "http://webfinger.example/rel/profile-page",
					Href: &profilePage,
				},
				{
					Rel: "http://webfinger.example/rel/businesscard",
					Href: &businessCard,
				},
			},
		}

		if !cmp.Equal(got, want) {
			t.Errorf("\n got: %+v \n want: %+v", got, want)
		}
	})

	t.Run("missing resource files should throw error", func(t *testing.T) {
		resource, err := fileDriver.GetResource("missingno")

		if err == nil {
			t.Errorf("should have gotten error, instead got resource: %+v", resource)
		}

		if !errors.As(err, &driver.ResourceNotFound{}) {
			t.Errorf("error should be ResourceNotFound: %+v", err)
		}
	})
}
