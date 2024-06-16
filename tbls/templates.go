package tbls

import (
	"fmt"
	"strings"

	"github.com/k1LoW/tbls/schema"
)

const (
	DefaultPromtTmpl = `Answer the questions in the Question assuming the following DDL.
{{ .DatabaseVersion }}

## DDL ( Data Definition Language )

{{ .QuoteStart }}
{{ .DDL }}
{{ .QuoteEnd }}
`
	DefaultQueryPromptTmpl = `Answer the SQL query in the "Explanation of the query to be created" section, assuming the database was created with the following DDL.
{{ .DatabaseVersion }}

## DDL ( Data Definition Language )

{{ .QuoteStart }}
{{ .DDL }}
{{ .QuoteEnd }}
`
)

func GenerateDDLRoughly(s *schema.Schema) string {
	var ddl string
	for _, t := range s.Tables {
		if t.Type == "VIEW" {
			continue
		}
		ddl += fmt.Sprintf("CREATE TABLE %s (", t.Name)
		td := []string{}
		for _, c := range t.Columns {
			d := fmt.Sprintf("%s %s", c.Name, c.Type)
			if c.Default.String != "" {
				d += fmt.Sprintf(" DEFAULT %s", c.Default.String)
			}
			if c.Comment != "" {
				d += fmt.Sprintf(" COMMENT %q", c.Comment)
			}
			td = append(td, d)
		}
		for _, c := range t.Constraints {
			switch c.Type {
			case "PRIMARY KEY", "UNIQUE KEY":
				continue
			default:
				if strings.Contains(c.Def, "FOREIGN KEY") {
					d := " CONSTRAINT " + c.Def
					td = append(td, d)
				}
			}
		}
		ddl += strings.Join(td, ",")
		if t.Comment != "" {
			ddl += fmt.Sprintf(") COMMENT = %q;\n", t.Comment)
		} else {
			ddl += ");\n"
		}
	}
	return ddl
}

func DatabaseVersion(s *schema.Schema) string {
	var n string
	switch s.Driver.Name {
	case "mysql":
		n = "MySQL"
	case "sqlite":
		n = "SQLite"
	case "postgres":
		n = "PostgreSQL"
	default:
		n = s.Driver.Name
	}
	if s.Driver.DatabaseVersion != "" {
		n += " " + s.Driver.DatabaseVersion
	}
	if n == "" {
		n = "unknown"
	}
	return fmt.Sprintf("Database is %s.", n)
}
