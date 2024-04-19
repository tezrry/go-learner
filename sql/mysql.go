package sql

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

func init() {
	ColumnNumberTypes = map[string]ColumnType{
		"tinyint":   ColumnTypeInt,
		"boolean":   ColumnTypeInt,
		"smallint":  ColumnTypeInt,
		"mediumint": ColumnTypeInt,
		"int":       ColumnTypeInt,
		"bigint":    ColumnTypeInt,
		"float":     ColumnTypeFloat,
		"double":    ColumnTypeFloat,
		"decimal":   ColumnTypeString,
		"bit":       ColumnTypeString,
	}

	ColumnTimeTypes = map[string]ColumnType{
		"timestamp": ColumnTypeTime,
		"datetime":  ColumnTypeTime,
		"date":      ColumnTypeString,
		"time":      ColumnTypeString,
	}
}

type MySql struct {
	db *sql.DB
}

func NewMySql(url string) *MySql {
	db, err := sql.Open("mysql", url)
	if err != nil {
		return nil
	}

	return &MySql{
		db: db,
	}
}

func (db *MySql) LoadTableSchema(tableName string) (*TableSchema, error) {
	rows, err := db.db.Query("DESC " + tableName)
	if err != nil {
		return nil, err
	}

	rowData := make([][]string, 0, 32)
	for rows.Next() {
		var fieldNameDesc, fieldTypeDesc, nullDesc, keyDesc, defaultDesc, extraDesc sql.NullString
		err := rows.Scan(&fieldNameDesc, &fieldTypeDesc, &nullDesc, &keyDesc, &defaultDesc, &extraDesc)
		if err != nil {
			return nil, errors.Wrap(err, "scan")
		}

		rowData = append(rowData, []string{
			fieldNameDesc.String, fieldTypeDesc.String, nullDesc.String,
			keyDesc.String, defaultDesc.String, extraDesc.String,
		})
	}

	nFields := len(rowData)
	schema := &TableSchema{
		Name:            tableName,
		Columns:         make([]ColumnSchema, nFields),
		m:               make(map[string]int, nFields),
		AutoMTimeFields: make([]int, 0, 1),
	}

	var selectBuilder, whereBuilder strings.Builder
	selectBuilder.WriteString("SELECT")
	whereBuilder.WriteString(" WHERE")

	selectLastWord := " `"
	whereLastWord := " `"
	fieldSchemas := schema.Columns

	nPrimary := 0
	pkIndexes := make([]int, 0, 2)
	bNoSortedKey := false
	for i := 0; i < nFields; i++ {
		fieldData := rowData[i]
		fieldName := fieldData[0]

		fieldType := parseFieldType(fieldData[1])
		category, isNumberType := ColumnNumberTypes[fieldType]
		if !isNumberType {
			var isTimeField bool
			if category, isTimeField = ColumnTimeTypes[fieldType]; !isTimeField {
				category = ColumnTypeString
			}
		}

		field := &fieldSchemas[i]
		field.Index = i
		field.Name = fieldName
		field.Type = category
		field.IsNumber = isNumberType

		if strings.Contains(fieldData[5], "on update CURRENT_TIMESTAMP") {
			schema.AutoMTimeFields = append(schema.AutoMTimeFields, i)

		} else {
			field.DefaultValue = fieldData[4]
		}

		selectBuilder.WriteString(selectLastWord)
		selectBuilder.WriteString(fieldName)
		selectBuilder.WriteByte('`')
		selectLastWord = ",`"

		if fieldData[3] == "PRI" {
			pkIndexes = append(pkIndexes, i)
			if nPrimary != i {
				bNoSortedKey = true
				//return nil, fmt.Errorf("table: %s primary key MUST preposition", tableName)
			}

			nPrimary++
			field.IsPrimaryKey = true

			whereBuilder.WriteString(whereLastWord)
			whereBuilder.WriteString(fieldName)
			//mysql string value need not wrap with ''
			whereBuilder.WriteString("`=?")
			whereLastWord = " AND `"
		}

		schema.m[fieldName] = i
	}

	if nPrimary == 0 {
		return nil, fmt.Errorf("table: %s not found primary key", tableName)
	}

	schema.NumPrimaryKeys = nPrimary
	//mysql没有分表信息，默认使用第一个主键
	shardIndex := 0
	if bNoSortedKey {
		schema.PrimaryKeyIndexes = pkIndexes
		shardIndex = pkIndexes[0]
	}

	schema.ShardKey = schema.Columns[shardIndex].Name
	schema.ShardIndex = shardIndex

	var updateBuilder, insertBuilder, deleteBuilder strings.Builder
	whereClause := whereBuilder.String()
	schema.whereSingleClause = whereClause

	insertBuilder.WriteString("INSERT INTO ")
	insertBuilder.WriteString(tableName)
	schema.insertClause1 = insertBuilder.String()
	schema.insertClause2 = " VALUES"

	deleteBuilder.WriteString("DELETE FROM ")
	deleteBuilder.WriteString(tableName)
	deleteBuilder.WriteString(whereClause)
	schema.deleteSingle = deleteBuilder.String()

	updateBuilder.WriteString("UPDATE ")
	updateBuilder.WriteString(tableName)
	updateBuilder.WriteString(" SET")
	schema.updatePrefix = updateBuilder.String()

	selectBuilder.WriteString("SELECT * FROM ")
	selectBuilder.WriteString(tableName)
	if nPrimary == 1 {
		selectBuilder.WriteString(whereClause)
		schema.selectSingle = selectBuilder.String()

	} else {
		var selectMultiBuilder strings.Builder
		selectMultiBuilder.WriteString(selectBuilder.String())
		selectMultiBuilder.WriteString(" WHERE `")
		selectMultiBuilder.WriteString(schema.ShardKey)
		selectMultiBuilder.WriteString("`=?")
		//selectMultiBuilder.WriteString("=? LIMIT ?")
		//selectMultiBuilder.WriteString(strconv.Itoa(MaxMultiRowNum + 1))
		schema.selectMulti = selectMultiBuilder.String()

		selectBuilder.WriteString(whereClause)
		schema.selectSingle = selectBuilder.String()

		var deleteMultiBuilder strings.Builder
		deleteMultiBuilder.WriteString("DELETE FROM ")
		deleteMultiBuilder.WriteString(tableName)
		deleteMultiBuilder.WriteString(" WHERE ")
		schema.deleteMultiPrefix = deleteMultiBuilder.String()
	}

	return schema, nil
}

func (db *MySql) Insert(schema *TableSchema, _ string, fields []string) *DBReply {
	var sql1, sql2 strings.Builder
	sql1.WriteString(schema.insertClause1)
	sql2.WriteString(schema.insertClause2)

	nField := len(fields)
	if nField&1 != 0 {
		return &DBReply{
			Data: int64(0),
			Msg:  "invalid fields for INSERT parameters",
		}
	}

	var lastSign byte = '('
	params := make([]interface{}, 0, nField>>1)
	for i := 0; i < nField; i += 2 {
		name := fields[i]
		value := fields[i+1]

		cs := schema.GetColumnSchema(name)
		if cs == nil {
			return &DBReply{
				Data: int64(0),
				Msg:  "invalid field of INSERT parameter",
			}
		}

		if cs.IsNumber && value == "" {
			continue
		}

		sql1.WriteByte(lastSign)
		sql1.WriteByte('`')
		sql1.WriteString(name)
		sql1.WriteByte('`')

		sql2.WriteByte(lastSign)
		sql2.WriteByte('?')

		params = append(params, value)
		lastSign = ','
	}
	sql1.WriteByte(')')
	sql2.WriteByte(')')
	sql1.WriteString(sql2.String())

	_, err := db.db.Exec(sql1.String(), params...)
	if err != nil {
		sqlErr, ok := err.(*mysql.MySQLError)
		if ok {
			return &DBReply{Data: int64(0), Msg: sqlErr.Message}
		}

		return &DBReply{Err: sqlErr}
	}

	return &DBReply{Data: int64(1)}
}

func (db *MySql) DeleteSingle(schema *TableSchema, _ string, keys []string) *DBReply {
	nKey := len(keys)
	if nKey != schema.NumPrimaryKeys {
		return &DBReply{Data: int64(0), Msg: "invalid primary keys"}
	}

	params := make([]interface{}, nKey)
	for i := 0; i < nKey; i++ {
		params[i] = keys[i]
	}

	ret, err := db.db.Exec(schema.deleteSingle, params...)
	if err != nil {
		sqlErr, ok := err.(*mysql.MySQLError)
		if ok {
			return &DBReply{Data: int64(0), Msg: sqlErr.Message}
		}

		return &DBReply{Err: err}
	}

	num, _ := ret.RowsAffected()
	return &DBReply{Data: num}
}

func (db *MySql) UpdateSingle(schema *TableSchema, _ string, keys []string, fields []string) *DBReply {
	n := len(fields)
	if n&1 != 0 {
		return &DBReply{Data: int64(0), Msg: "invalid fields for Update"}
	}
	nKey := len(keys)
	if nKey != schema.NumPrimaryKeys {
		return &DBReply{Data: int64(0), Msg: "invalid primary keys"}
	}

	nField := n >> 1
	params := make([]interface{}, nField+schema.NumPrimaryKeys)
	var builder strings.Builder
	builder.WriteString(schema.updatePrefix)
	var prevSign byte = ' '
	idx := 0
	for i := 0; i < n; i += 2 {
		name := fields[i]
		value := fields[i+1]

		builder.WriteByte(prevSign)
		builder.WriteByte('`')
		builder.WriteString(name)
		builder.WriteByte('`')
		builder.WriteString("=?")

		params[idx] = value
		idx++

		prevSign = ','
	}
	builder.WriteString(schema.whereSingleClause)

	for i := 0; i < nKey; i++ {
		params[i+nField] = keys[i]
	}

	ret, err := db.db.Exec(builder.String(), params...)
	if err != nil {
		sqlErr, ok := err.(*mysql.MySQLError)
		if ok {
			return &DBReply{Data: int64(0), Msg: sqlErr.Message}
		}

		return &DBReply{Err: err}
	}

	num, _ := ret.RowsAffected()
	return &DBReply{Data: num}
}

func (db *MySql) IncrBySingle(schema *TableSchema, _ string, keys []string, data *IncrByData) *DBReply {
	nKey := len(keys)
	if nKey != schema.NumPrimaryKeys {
		return &DBReply{Data: int64(0), Msg: "invalid primary keys"}
	}

	params := make([]interface{}, nKey)
	for i := 0; i < nKey; i++ {
		params[i] = keys[i]
	}

	var builder strings.Builder
	builder.WriteString(schema.updatePrefix)
	//builder.WriteByte(' ')
	builder.WriteString(" `")
	builder.WriteString(data.Column)
	//builder.WriteByte('=')
	builder.WriteString("`=`")
	builder.WriteString(data.Column)
	//builder.WriteByte('+')
	builder.WriteString("`+")
	builder.WriteString(strconv.FormatInt(data.Delta, 10))
	builder.WriteString(schema.whereSingleClause)

	if data.Where != "" {
		builder.WriteString(" AND ")
		builder.WriteString(data.Where)
	}

	ret, err := db.db.Exec(builder.String(), params...)
	if err != nil {
		sqlErr, ok := err.(*mysql.MySQLError)
		if ok {
			return &DBReply{Data: int64(0), Msg: sqlErr.Message}
		}

		return &DBReply{Err: err}
	}

	num, _ := ret.RowsAffected()
	return &DBReply{Data: num}
}

func (db *MySql) SelectSingle(schema *TableSchema, _ string, keys []string) *DBReply {
	var rowData []byte
	nKey := len(keys)
	if nKey != schema.NumPrimaryKeys {
		return &DBReply{Data: rowData, Msg: "invalid primary keys"}
	}

	params := make([]interface{}, nKey)
	for i := 0; i < nKey; i++ {
		params[i] = keys[i]
	}

	rows, err := db.db.Query(schema.selectSingle, params...)
	if err != nil {
		sqlErr, ok := err.(*mysql.MySQLError)
		if ok {
			return &DBReply{Data: rowData, Msg: sqlErr.Message}
		}

		return &DBReply{Data: rowData, Err: err}
	}

	nField := len(schema.Columns)
	wrapper := make([]interface{}, nField)
	rawBytes := make([]sql.RawBytes, nField)

	rowNum := 0
	for rows.Next() {
		for i := range wrapper {
			wrapper[i] = &rawBytes[i]
		}

		err := rows.Scan(wrapper...)
		if err != nil {
			_ = rows.Close()
			return &DBReply{Data: rowData, Err: err}
		}

		rowData = NewRowData(schema, rawBytes2Bytes(rawBytes))
		if rowData == nil {
			_ = rows.Close()
			return &DBReply{Data: rowData, Msg: "inconsistent columns returned, check table schema"}
		}
		rowNum++
	}

	if rowNum > 1 {
		return &DBReply{Data: rowData, Msg: "multiple row returned"}
	}

	return &DBReply{Data: rowData}
}

func (db *MySql) SelectMulti(schema *TableSchema, shard string) *DBReply {
	params := []interface{}{shard}
	var multiRowData [][]byte

	rows, err := db.db.Query(schema.selectMulti, params...)
	if err != nil {
		sqlErr, ok := err.(*mysql.MySQLError)
		if ok {
			return &DBReply{Data: multiRowData, Msg: sqlErr.Message}
		}

		return &DBReply{Data: multiRowData, Err: err}
	}

	nField := len(schema.Columns)
	wrapper := make([]interface{}, nField)
	rawBytes := make([]sql.RawBytes, nField)

	multiRowData = make([][]byte, 0, 512)
	nRow := 0
	for rows.Next() {
		for i := range wrapper {
			wrapper[i] = &rawBytes[i]
		}

		err := rows.Scan(wrapper...)
		if err != nil {
			_ = rows.Close()
			return &DBReply{Data: multiRowData, Err: err}
		}

		rowData := NewRowData(schema, rawBytes2Bytes(rawBytes))
		if rowData == nil {
			_ = rows.Close()
			return &DBReply{Data: rowData, Msg: "inconsistent columns returned, check table schema"}
		}

		multiRowData = append(multiRowData, rowData)
		nRow++
	}

	return &DBReply{Data: multiRowData}
}

func (db *MySql) DeleteMulti(schema *TableSchema, _ string, data *MultiRequestData) *DBReply {
	n := len(data.Params)
	params := make([]interface{}, n)
	for i := 0; i < n; i++ {
		params[i] = data.Params[i]
	}

	var builder strings.Builder
	builder.WriteString(schema.deleteMultiPrefix)
	builder.WriteString(data.Where)

	ret, err := db.db.Exec(builder.String(), params...)
	if err != nil {
		sqlErr, ok := err.(*mysql.MySQLError)
		if ok {
			return &DBReply{Data: int64(0), Msg: sqlErr.Message}
		}

		return &DBReply{Err: err}
	}

	num, _ := ret.RowsAffected()
	return &DBReply{Data: num}
}

func (db *MySql) Query(query string, params ...interface{}) ([][][]byte, error) {
	rows, err := db.db.Query(query, params...)
	if err != nil {
		return nil, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	nField := len(columns)
	wrapper := make([]interface{}, nField)
	multiRowData := make([][][]byte, 0, 512)
	for rows.Next() {
		rawBytes := make([][]byte, nField)
		for i := range wrapper {
			wrapper[i] = &rawBytes[i]
		}

		err := rows.Scan(wrapper...)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}

		multiRowData = append(multiRowData, rawBytes)
	}

	return multiRowData, nil
}

func (db *MySql) Exec(query string, params ...interface{}) (int64, error) {
	ret, err := db.db.Exec(query, params...)
	if err != nil {
		return 0, err
	}

	return ret.RowsAffected()
}

func parseFieldType(typeDesc string) string {
	e := strings.IndexByte(typeDesc, ' ')
	if e < 0 {
		e = len(typeDesc)
	}

	ret := typeDesc[:e]
	e = strings.IndexByte(ret, '(')
	if e > 0 {
		return ret[:e]
	}

	return ret
}

func rawBytes2Bytes(b []sql.RawBytes) [][]byte {
	return *(*[][]byte)(unsafe.Pointer(&b))
}
