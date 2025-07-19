package sql

import (
	"testing"
	"text/template"

	"github.com/peeley/carpal/internal/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	_ "github.com/lib/pq"
	"github.com/peeley/carpal/internal/config"
	"github.com/peeley/carpal/internal/resource"
)

func TestSQLDriverGetResource(t *testing.T) {
	conf := config.Configuration{
		Driver: "sql",
		DatabaseConfiguration: &config.DatabaseConfiguration{
			Driver:      "postgres",
			URL:         "postgres://user:password@localhost:5432/dbname?sslmode=disable",
			Table:       "users",
			KeyColumn:   "email",
			ColumnNames: []string{"email", "handle", "name"},
			Template:    "test/sql_template.gotempl",
		},
	}

	tmplContent := `aliases:
  - "mailto:{{ .email }}"
  - "https://mastodon/{{ .handle }}"
properties:
  'http://webfinger.example/ns/name': '{{ .name }}'
links:
  - rel: "http://webfinger.example/rel/profile-page"
    href: "https://www.example.com/~{{ .handle }}/"
`
	tmpl, err := template.New("test").Parse(tmplContent)
	if err != nil {
		t.Fatal(err)
	}

	sql, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("could not mock sql database: %v", err)
	}

	expectedRows := sqlmock.NewRows([]string{"email", "handle", "name"}).
		AddRow("bob@example.com", "bob", "Bob Smith")

	mock.ExpectQuery("SELECT email,handle,name FROM users WHERE email = (.+)").
		WithArgs("bob@example.com").
		WillReturnRows(expectedRows)

	driverInstance := &sqlDriver{
		Configuration: conf,
		Template:      tmpl,
		DB:            sql,
	}

	t.Run("can get resource from SQL", func(t *testing.T) {
		got, err := driverInstance.GetResource("acct:bob@example.com")
		if err != nil {
			t.Fatal(err)
		}

		linkHref := "https://www.example.com/~bob/"
		want := &resource.Resource{
			Subject:    "acct:bob@example.com",
			Aliases:    []string{"mailto:bob@example.com", "https://mastodon/bob"},
			Properties: map[string]any{"http://webfinger.example/ns/name": "Bob Smith"},
			Links: []resource.Link{
				{
					Rel:  "http://webfinger.example/rel/profile-page",
					Href: &linkHref,
				},
			},
		}

		if !cmp.Equal(got, want) {
			t.Errorf("got:  %+v,\n want: %+v", got, want)
		}
	})

	mock.ExpectQuery("SELECT email,handle,name FROM users WHERE email = (.+)").
		WithArgs("bob@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"email", "handle", "name"}))

	t.Run("handles missing resource in SQL", func(t *testing.T) {
		got, err := driverInstance.GetResource("acct:bob@example.com")
		expected := driver.ResourceNotFound{ResourceName: "bob@example.com"}

		if err != expected {
			t.Errorf("should have failed to fetch resource, got: %v, err: %v", got, err)
		}
	})
}
