package ldap

import (
	"errors"
	"fmt"
	"testing"
	"text/template"

	"github.com/go-ldap/ldap/v3"
	client "github.com/go-ldap/ldap/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/peeley/carpal/internal/config"
	"github.com/peeley/carpal/internal/driver"
	"github.com/peeley/carpal/internal/resource"
)

var testLdapTempl = `aliases:
  - "mailto:{{ index . "mail" }}"
  - "https://mastodon/{{ index . "uid" }}"
properties:
  'http://webfinger.example/ns/name': '{{ index . "cn" }}'
links:
  - rel: "http://webfinger.example/rel/profile-page"
    href: "https://www.example.com/~{{ index . "uid" }}/"
    `

type testLdapConn struct {
	driver ldapDriver
	user   string
	cn     string
	domain string
}

func (testLdapConn) Bind(_ string, _ string) (_ error) {
	return nil
}

func (testLdapConn) Close() (_ error) {
	return nil
}

func (t testLdapConn) Search(req *client.SearchRequest) (_ *client.SearchResult, _ error) {
	var res client.SearchResult
	if req.Filter != fmt.Sprintf("(%v=%v)", t.driver.Configuration.LDAPConfiguration.UserAttr, t.user) {
		return nil, &ldap.Error{
			ResultCode: ldap.LDAPResultNoSuchObject,
		}
	}
	entry := &ldap.Entry{
		DN: fmt.Sprintf("%v=%v,%v", t.driver.Configuration.LDAPConfiguration.UserAttr, t.user, t.driver.Configuration.LDAPConfiguration.BaseDN),
		Attributes: []*client.EntryAttribute{
			{
				Name:   "uid",
				Values: []string{t.user},
			},
			{
				Name:   "mail",
				Values: []string{fmt.Sprintf("%v@%v", t.user, t.domain)},
			},
			{
				Name:   "cn",
				Values: []string{t.cn},
			},
		},
	}
	res.Entries = append(res.Entries, entry)

	return &res, nil
}

func TestLdapDriverGetResource(t *testing.T) {
	conf := config.Configuration{
		Driver: "ldap",
		LDAPConfiguration: &config.LDAPConfiguration{
			URL:        "ldaps://ldap.example.com",
			BindUser:   "cn=root,dc=example,dc=com",
			BindPass:   "password",
			BaseDN:     "ou=Users,dc=example,dc=com",
			UserAttr:   "uid",
			Attributes: []string{"uid", "mail", "cn"},
		},
	}
	d := ldapDriver{
		Configuration: conf,
	}
	tmpl := template.New("test")
	d.Template = template.Must(tmpl.Parse(testLdapTempl))
	d.ClientFunc = func() (LdapClient, error) {
		return testLdapConn{d, "bob", "Bob", "foobar.com"}, nil
	}

	t.Run("can get resource from ldap", func(t *testing.T) {
		got, err := d.GetResource("acct:bob@foobar.com")
		if err != nil {
			t.Fatal(err)
		}
		profilePage := "https://www.example.com/~bob/"
		want := &resource.Resource{
			Subject:    "acct:bob@foobar.com",
			Aliases:    []string{"mailto:bob@foobar.com", "https://mastodon/bob"},
			Properties: map[string]any{"http://webfinger.example/ns/name": "Bob"},
			Links: []resource.Link{
				{
					Rel:  "http://webfinger.example/rel/profile-page",
					Href: &profilePage,
				},
			},
		}

		if !cmp.Equal(got, want) {
			t.Errorf("\n got: %+v \n want: %+v", got, want)
		}
	})

	t.Run("missing resource files should throw error", func(t *testing.T) {
		resource, err := d.GetResource("missingno")

		if err == nil {
			t.Errorf("should have gotten error, instead got resource: %+v", resource)
		}

		if !errors.As(err, &driver.ResourceNotFound{}) {
			t.Errorf("error should be ResourceNotFound: %+v", err)
		}
	})
}
