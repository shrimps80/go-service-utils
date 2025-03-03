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

// Config 生成器配置
type Config struct {
	DB          *gorm.DB
	OutputPath  string
	PackageName string
	Template    string // 自定义模板，为空时使用默认模板
}

// Generator 模型生成器
type Generator struct {
	config *Config
}

// NewGenerator 创建生成器实例
func NewGenerator(config *Config) *Generator {
	if config.Template == "" {
		config.Template = defaultTemplate
	}
	return &Generator{config: config}
}

// GenerateModel 生成单个表的模型
func (g *Generator) GenerateModel(tableName string) error {
	// 获取表结构信息
	columns, err := g.getColumns(tableName)
	if err != nil {
		return err
	}

	// 生成结构体名称
	structName := g.toCamelCase(tableName)

	// 创建模板数据
	tmplData := struct {
		PackageName string
		StructName  string
		TableName   string
		Columns     []Column
	}{
		PackageName: g.config.PackageName,
		StructName:  structName,
		TableName:   tableName,
		Columns:     columns,
	}

	// 解析模板
	tmpl, err := template.New("model").Parse(g.config.Template)
	if err != nil {
		return err
	}

	// 创建输出目录
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

	// 执行模板
	return tmpl.Execute(file, tmplData)
}

// GenerateAllModels 生成所有表的模型
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

// Column 列信息结构
type Column struct {
	Field      string         // 字段名
	Type       string         // MySQL类型
	Default    *string        // 默认值
	Privileges sql.NullString // 权限
	Comment    *string        // 注释
	GoField    string         // Go字段名
	GoType     string         // Go类型
	Collation  *string        // 字符集
}

// 获取数据库中的所有表名
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

// 获取表的列信息
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
			null       string // YES/NO
			key        string // PRI/UNI/MUL
			defaultVal sql.NullString
			extra      string
			comment    sql.NullString
		)

		// 调整字段顺序与SHOW FULL COLUMNS结果一致
		err := rows.Scan(
			&col.Field,      // 1. Field
			&col.Type,       // 2. Type
			&collation,      // 3. Collation
			&null,           // 4. Null (YES/NO)
			&key,            // 5. Key (PRI/UNI/MUL)
			&defaultVal,     // 6. Default
			&extra,          // 7. Extra
			&col.Privileges, // 8. Privileges
			&comment,        // 9. Comment
		)
		if err != nil {
			return nil, err
		}

		// 处理默认值
		if defaultVal.Valid {
			col.Default = &defaultVal.String
		}

		// 处理注释
		if comment.Valid {
			col.Comment = &comment.String
		}

		col.GoField = g.toCamelCase(col.Field)
		col.GoType = g.mysqlTypeToGo(col.Type)
		columns = append(columns, col)
	}

	return columns, nil
}

// MySQL类型转Go类型
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

// 转换为驼峰命名
func (g *Generator) toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

// 默认模板
const defaultTemplate = `package {{.PackageName}}

import (
	"time"

	"gorm.io/gorm"
)

// {{.StructName}} {{.TableName}} 表的模型
type {{.StructName}} struct {
	{{range .Columns}}
	{{.GoField}} {{.GoType}} ` + "`" + `gorm:"column:{{.Field}}{{if .Comment}};comment:{{.Comment}}{{end}}"` + "`" + ` {{if .Comment}}// {{.Comment}}{{end}}{{end}}
}

// TableName 返回表名
func ({{.StructName}}) TableName() string {
	return "{{.TableName}}"
}
`