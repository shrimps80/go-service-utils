package genmodel

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gorm.io/gorm"
)

type Config struct {
	DB          *gorm.DB
	OutputPath  string
	PackageName string
	Template    string
}

type Generator struct {
	config *Config
}

func NewGenerator(config *Config) *Generator {
	if config.Template == "" {
		config.Template = defaultTemplate
	}
	return &Generator{config: config}
}

func (g *Generator) GenerateModel(tableName string) error {
	columns, err := g.getColumns(tableName)
	if err != nil {
		return err
	}

	structName := g.toCamelCase(tableName)

	tmplData := struct {
		PackageName string
		StructName  string
		TableName   string
		Columns     []Column
		Fields      []string
	}{
		PackageName: g.config.PackageName,
		StructName:  structName,
		TableName:   tableName,
		Columns:     columns,
		Fields:      g.getFieldList(columns),
	}

	tmpl, err := template.New("model").Parse(g.config.Template)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(g.config.OutputPath, 0755); err != nil {
		return err
	}

	// 创建输出文件
	filePath := filepath.Join(g.config.OutputPath, strings.ToLower(tableName)+".go")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, tmplData)
}

func (g *Generator) getFieldList(columns []Column) []string {
	fields := make([]string, 0, len(columns))
	for _, col := range columns {
		fields = append(fields, col.Field)
	}
	return fields
}

func (g *Generator) GenerateAllModels() error {
	tables, err := g.getTables()
	if err != nil {
		return err
	}

	for _, table := range tables {
		if err := g.GenerateModel(table); err != nil {
			return fmt.Errorf("生成表 %s 的模型失败: %v", table, err)
		}
	}

	return nil
}

type Column struct {
	Field      string
	Type       string
	Default    *string
	Privileges sql.NullString
	Comment    *string
	GoField    string
	GoType     string
	Collation  *string
	IsPrimary  bool
}

func (g *Generator) getTables() ([]string, error) {
	var tables []string
	rows, err := g.config.DB.Raw("SHOW TABLES").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

func (g *Generator) getColumns(tableName string) ([]Column, error) {
	rows, err := g.config.DB.Raw("SHOW FULL COLUMNS FROM " + tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		var (
			collation  sql.NullString
			null       string
			key        string
			defaultVal sql.NullString
			extra      string
			comment    sql.NullString
		)

		err := rows.Scan(
			&col.Field,
			&col.Type,
			&collation,
			&null,
			&key,
			&defaultVal,
			&extra,
			&col.Privileges,
			&comment,
		)
		if err != nil {
			return nil, err
		}

		if defaultVal.Valid {
			col.Default = &defaultVal.String
		}

		if comment.Valid {
			col.Comment = &comment.String
		}

		col.IsPrimary = key == "PRI"
		col.GoField = g.toCamelCase(col.Field)
		col.GoType = g.mysqlTypeToGo(col.Type)
		columns = append(columns, col)
	}

	return columns, nil
}

func (g *Generator) mysqlTypeToGo(mysqlType string) string {
	mysqlType = strings.ToLower(mysqlType)
	switch {
	case strings.Contains(mysqlType, "bigint"):
		return "uint64"
	case strings.Contains(mysqlType, "int"):
		return "int"
	case strings.Contains(mysqlType, "decimal"), strings.Contains(mysqlType, "numeric"), strings.Contains(mysqlType, "float"), strings.Contains(mysqlType, "double"):
		return "float64"
	case strings.Contains(mysqlType, "datetime"), strings.Contains(mysqlType, "timestamp"):
		return "time.Time"
	case strings.Contains(mysqlType, "bool"):
		return "bool"
	default:
		return "string"
	}
}

func (g *Generator) toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

const defaultTemplate = `package {{.PackageName}}

import (
	"time"
	"gorm.io/gorm"
)

type {{.StructName}} struct {
	{{range .Columns -}}
	{{.GoField}} {{.GoType}} ` + "`gorm:\"column:{{.Field}}{{if .IsPrimary}};primaryKey{{end}}{{if .Comment}};comment:{{.Comment}}{{end}}\"`" + `
	{{end}}
}

func ({{.StructName}}) TableName() string {
	return "{{.TableName}}"
}

func ({{.StructName}}) GetFields() []string {
	return []string{
		{{range $field := .Fields -}}
		"{{$field}}",
		{{end}}
	}
}
`
