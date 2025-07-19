package sql

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/peeley/carpal/internal/config"
	"github.com/peeley/carpal/internal/driver"
	"github.com/peeley/carpal/internal/resource"
	"gopkg.in/yaml.v3"
)

type SQLClient interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Close() error
}

type sqlDriver struct {
	Configuration config.Configuration
	Template      *template.Template
	DB            *sql.DB
}

func NewSQLDriver(conf config.Configuration) (driver.Driver, error) {
	tmpl := template.Must(template.ParseFiles(conf.DatabaseConfiguration.Template))

	db, err := sql.Open(conf.DatabaseConfiguration.Driver, conf.DatabaseConfiguration.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &sqlDriver{
		Configuration: conf,
		Template:      tmpl,
		DB:            db,
	}, nil
}

func (d *sqlDriver) GetResource(name string) (*resource.Resource, error) {
	re, err := regexp.Compile("acct:(.+@.+)")
	if err != nil {
		return nil, err
	}
	resourceName := re.FindStringSubmatch(name)
	if len(resourceName) == 0 {
		return nil, driver.ResourceNotFound{ResourceName: name}
	}
	email := resourceName[1]

	column_names := strings.Join(d.Configuration.DatabaseConfiguration.ColumnNames, ",")
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1",
		column_names,
		d.Configuration.DatabaseConfiguration.Table,
		d.Configuration.DatabaseConfiguration.KeyColumn,
	)

	row := d.DB.QueryRow(query, email)
	if row == nil {
		return nil, driver.ResourceNotFound{ResourceName: email}
	}

	columns := make([]sql.NullString, len(d.Configuration.DatabaseConfiguration.ColumnNames))
	columnPointers := make([]interface{}, len(columns))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	if err := row.Scan(columnPointers...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, driver.ResourceNotFound{ResourceName: email}
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	data := make(map[string]string)
	for i, colName := range d.Configuration.DatabaseConfiguration.ColumnNames {
		if columns[i].Valid {
			data[colName] = columns[i].String
		} else {
			data[colName] = ""
		}
	}

	var resourceFile bytes.Buffer
	if err := d.Template.Execute(&resourceFile, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	var res resource.Resource
	if err := yaml.Unmarshal(resourceFile.Bytes(), &res); err != nil {
		return nil, fmt.Errorf("could not unmarshal YAML to resource: %w", err)
	}

	res.Subject = name
	return &res, nil
}
