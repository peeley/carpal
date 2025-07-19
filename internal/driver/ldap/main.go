package ldap

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"text/template"

	client "github.com/go-ldap/ldap/v3"
	"github.com/peeley/carpal/internal/config"
	"github.com/peeley/carpal/internal/driver"
	"github.com/peeley/carpal/internal/resource"
	"gopkg.in/yaml.v3"
)

type LdapClient interface {
	Bind(string, string) error
	Close() error
	Search(*client.SearchRequest) (*client.SearchResult, error)
}

type ldapDriver struct {
	Configuration config.Configuration
	Template      *template.Template
	ClientFunc    func() (LdapClient, error)
}

func NewLDAPDriver(conf config.Configuration) driver.Driver {
	d := ldapDriver{
		Configuration: conf,
	}
	d.Template = template.Must(template.ParseFiles(conf.LDAPConfiguration.Template))
	d.ClientFunc = func() (LdapClient, error) {
		return client.DialURL(conf.LDAPConfiguration.URL)
	}
	return d
}

func (d ldapDriver) GetResource(name string) (*resource.Resource, error) {
	var resource resource.Resource

	re, err := regexp.Compile("acct:([^@]+)")
	if err != nil {
		return nil, err
	}
	resourceName := re.FindStringSubmatch(name)
	if len(resourceName) == 0 {
		return nil, driver.ResourceNotFound{ResourceName: name}
	}

	if len(resourceName) < 2 {
		return nil, errors.New("Error breaking down resource")
	}
	username := resourceName[1]
	c, err := d.ClientFunc()
	if err != nil {
		return nil, err
	}
	defer c.Close()
	err = c.Bind(d.Configuration.LDAPConfiguration.BindUser, d.Configuration.LDAPConfiguration.BindPass)
	if err != nil {
		return nil, err
	}

	searchString := fmt.Sprintf("(%s=%s)", d.Configuration.LDAPConfiguration.UserAttr, username)
	if d.Configuration.LDAPConfiguration.Filter != "" {
		searchString = fmt.Sprintf("(&%v%v)", d.Configuration.LDAPConfiguration.Filter, searchString)
	}
	result, err := c.Search(client.NewSearchRequest(
		d.Configuration.LDAPConfiguration.BaseDN,
		client.ScopeWholeSubtree,
		client.NeverDerefAliases,
		0,
		0,
		false,
		searchString,
		d.Configuration.LDAPConfiguration.Attributes,
		nil,
	))
	if err != nil {
		return nil, err
	}

	if len(result.Entries) > 1 {
		return nil, fmt.Errorf("Error finding user: Wanted 1 result, got %v\n", len(result.Entries))
	} else if len(result.Entries) == 0 {
		return nil, driver.ResourceNotFound{ResourceName: username}
	}
	ldapUser := result.Entries[0]
	ldapAttrs := make(map[string]string)
	for _, v := range d.Configuration.LDAPConfiguration.Attributes {
		ldapAttrs[v] = ldapUser.GetAttributeValue(v)
	}
	var resourceFile bytes.Buffer
	err = d.Template.Execute(&resourceFile, ldapAttrs)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(resourceFile.Bytes(), &resource)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal file to JRD: %w", err)
	}

	resource.Subject = name
	return &resource, nil
}
