package sql

import (
	"strings"
)

type FakeColumn struct {
	Name string
	Type ColumnType
}

var mySqlCache map[string]*MySql

func GetMySql(url string) *MySql {
	if mySqlCache == nil {
		mySqlCache = make(map[string]*MySql)
	}

	inst, ok := mySqlCache[url]
	if !ok {
		inst = NewMySql(url)
		mySqlCache[url] = inst
	}

	return inst
}

func CreateFakeTableSchema(columns []FakeColumn, nPrimary int) *TableSchema {
	nField := len(columns)
	cs := make([]ColumnSchema, nField)
	m := make(map[string]int, nField)

	var idx int
	for idx = 0; idx < nField; idx++ {
		column := &cs[idx]
		column.Name = columns[idx].Name
		column.IsPrimaryKey = idx < nPrimary
		column.Index = idx
		column.Type = columns[idx].Type
		column.IsNumber = column.Type == ColumnTypeInt || column.Type == ColumnTypeFloat
		m[column.Name] = idx
	}

	return &TableSchema{
		Name:             "fake",
		ShardKey:         columns[0].Name,
		IsStringShardKey: columns[0].Type == ColumnTypeString,
		NumPrimaryKeys:   nPrimary,
		Columns:          cs,
		m:                m,
	}
}

func CreateFakeTableSchema2(columns []FakeColumn, pkIdxes []int, shardKey string) *TableSchema {
	nField := len(columns)
	cs := make([]ColumnSchema, nField)
	m := make(map[string]int, nField)

	nPrimary := len(pkIdxes)
	var idx int
	shardIndex := -1
	for idx = 0; idx < nField; idx++ {
		column := &cs[idx]
		column.Name = columns[idx].Name
		column.Index = idx
		column.Type = columns[idx].Type
		column.IsNumber = column.Type == ColumnTypeInt || column.Type == ColumnTypeFloat
		m[column.Name] = idx

		for _, i := range pkIdxes {
			if idx == i {
				column.IsPrimaryKey = true
			}
		}

		if column.Name == shardKey {
			if !column.IsPrimaryKey {
				panic("invalid shard key")
			}

			shardIndex = idx
		}
	}

	return &TableSchema{
		Name:              "fake",
		ShardKey:          shardKey,
		ShardIndex:        shardIndex,
		IsStringShardKey:  columns[shardIndex].Type == ColumnTypeString,
		NumPrimaryKeys:    nPrimary,
		PrimaryKeyIndexes: pkIdxes,
		Columns:           cs,
		m:                 m,
	}
}

func CreateMySqlTestTable(db *MySql, tableName string, columns []string, primaryKeys []string) {
	sql := "DROP TABLE IF EXISTS " + tableName
	_, err := db.Exec(sql)
	if err != nil {
		panic(err)
	}

	var builder strings.Builder
	builder.WriteString("CREATE TABLE IF NOT EXISTS ")
	builder.WriteString(tableName)
	builder.WriteString(" (")

	for _, c := range columns {
		builder.WriteString(c)
		builder.WriteByte(',')
	}

	builder.WriteString("PRIMARY KEY ")
	var sep byte = '('
	for _, c := range primaryKeys {
		builder.WriteByte(sep)
		builder.WriteString(c)
		sep = ','
	}

	builder.WriteString(")) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=Dynamic;")
	_, err = db.Exec(builder.String())
	if err != nil {
		panic(err)
	}
}

func CreateMySqlTestTable2(db *MySql, tableName string, createSql string) {
	sql := "DROP TABLE IF EXISTS " + tableName
	_, err := db.Exec(sql)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("CREATE TABLE " + tableName + createSql)
	if err != nil {
		panic(err)
	}
}
