package sql

import (
	"fmt"
	"strconv"
)

type DBStub struct {
	fakeReply bool
	schema    *TableSchema
	schemaErr error
	reply     *DBReply
	tables    map[string]map[string][]byte
}

func NewDBStub() *DBStub {
	return &DBStub{
		tables: make(map[string]map[string][]byte),
	}
}

func (db *DBStub) SetFakeSchemaReply(schema *TableSchema, err error) {
	db.schema = schema
	db.schemaErr = err
	db.fakeReply = true
}

func (db *DBStub) SetFakeReply(reply *DBReply) {
	db.reply = reply
	db.fakeReply = true
}

func (db *DBStub) LoadTableSchema(_ string) (*TableSchema, error) {
	if db.fakeReply {
		db.fakeReply = false
		return db.schema, db.schemaErr
	}

	return nil, fmt.Errorf("can not fake schema")
}

func (db *DBStub) Insert(schema *TableSchema, _ string, fields []string) *DBReply {
	if db.fakeReply {
		db.fakeReply = false
		return db.reply
	}

	row := NewRowDataFromSlice(schema, fields)
	if row == nil {
		return &DBReply{Data: int64(0), Msg: "invalid fields"}
	}

	key := GetRowKey(schema, row)
	table, ok := db.tables[schema.Name]
	if !ok {
		table = make(map[string][]byte)
		db.tables[schema.Name] = table
	}

	if _, ok = table[key]; ok {
		return &DBReply{Data: int64(0), Msg: "duplicate key"}
	}

	table[key] = row
	return &DBReply{Data: int64(1)}
}

func (db *DBStub) DeleteSingle(schema *TableSchema, _ string, keys []string) *DBReply {
	if db.fakeReply {
		db.fakeReply = false
		return db.reply
	}

	key := AssembleRowKey2(schema, keys)
	if key == "" {
		return &DBReply{Data: int64(0), Msg: "invalid primary keys"}
	}

	table, ok := db.tables[schema.Name]
	if !ok {
		return &DBReply{Data: int64(0)}
	}

	if _, ok = table[key]; ok {
		delete(table, key)
		return &DBReply{Data: int64(1)}
	}

	return &DBReply{Data: int64(0)}
}

func (db *DBStub) UpdateSingle(schema *TableSchema, _ string, keys []string, fields []string) *DBReply {
	if db.fakeReply {
		db.fakeReply = false
		return db.reply
	}

	key := AssembleRowKey2(schema, keys)
	if key == "" {
		return &DBReply{Data: int64(0), Msg: "invalid primary keys"}
	}

	nField := len(fields)
	if nField&1 != 0 {
		return &DBReply{Data: int64(0), Msg: "invalid fields"}
	}

	table, ok := db.tables[schema.Name]
	if !ok {
		return &DBReply{Data: int64(0)}
	}

	var row []byte
	if row, ok = table[key]; !ok {
		return &DBReply{Data: int64(0)}
	}

	mFields := RowData2Map(schema, row)
	for i := 0; i < nField; i += 2 {
		mFields[fields[i]] = fields[i+1]
	}

	row = NewRowDataFromMap(schema, mFields)
	if row == nil {
		return &DBReply{Data: int64(0), Msg: "invalid fields"}
	}

	table[key] = row
	return &DBReply{Data: int64(1)}

}

func (db *DBStub) IncrBySingle(schema *TableSchema, shard string, keys []string, data *IncrByData) *DBReply {
	if db.fakeReply {
		db.fakeReply = false
		return db.reply
	}

	key := AssembleRowKey2(schema, keys)
	if key == "" {
		return &DBReply{Data: int64(0), Msg: "invalid primary keys"}
	}

	table, ok := db.tables[schema.Name]
	if !ok {
		return &DBReply{Data: int64(0)}
	}

	var row []byte
	if row, ok = table[key]; !ok {
		return &DBReply{Data: int64(0)}
	}

	parser, err := CreateParser(schema, shard, data.Where, nil)
	if err != nil {
		return &DBReply{Data: int64(0), Msg: err.Error()}
	}
	if parser != nil && !parser.Check(row) {
		return &DBReply{Data: int64(0)}
	}

	mFields := RowData2Map(schema, row)
	v, ok := mFields[data.Column]
	if !ok {
		return &DBReply{Data: int64(0), Msg: "invalid field"}
	}

	iv, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return &DBReply{Data: int64(0), Msg: "invalid field"}
	}

	mFields[data.Column] = strconv.FormatInt(iv+data.Delta, 10)
	row = NewRowDataFromMap(schema, mFields)
	if row == nil {
		return &DBReply{Data: int64(0), Msg: "invalid fields"}
	}

	table[key] = row
	return &DBReply{Data: int64(1)}
}

func (db *DBStub) SelectSingle(schema *TableSchema, _ string, keys []string) *DBReply {
	if db.fakeReply {
		db.fakeReply = false
		return db.reply
	}

	key := AssembleRowKey2(schema, keys)
	if key == "" {
		return &DBReply{Data: int64(0), Msg: "invalid primary keys"}
	}

	var rowData []byte
	table, ok := db.tables[schema.Name]
	if !ok {
		return &DBReply{Data: rowData}
	}

	rowData, _ = table[key]
	return &DBReply{Data: rowData}
}

func (db *DBStub) SelectMulti(schema *TableSchema, shard string) *DBReply {
	if db.fakeReply {
		db.fakeReply = false
		return db.reply
	}

	var multiRowData [][]byte
	table, ok := db.tables[schema.Name]
	if !ok {
		return &DBReply{Data: multiRowData}
	}

	multiRowData = make([][]byte, 0)
	for _, rowData := range table {
		if shard == GetValueByIndex(schema, rowData, schema.ShardIndex) {
			multiRowData = append(multiRowData, rowData)
		}
	}

	return &DBReply{Data: multiRowData}
}

func (db *DBStub) DeleteMulti(schema *TableSchema, shard string, data *MultiRequestData) *DBReply {
	if db.fakeReply {
		db.fakeReply = false
		return db.reply
	}

	table, ok := db.tables[schema.Name]
	if !ok {
		return &DBReply{Data: int64(0)}
	}

	parser, err := CreateParser(schema, shard, data.Where, data.Params)
	if err != nil {
		return &DBReply{Data: int64(0), Msg: err.Error()}
	}

	keys := make([]string, 0)
	for k, rowData := range table {
		if shard != GetValueByIndex(schema, rowData, schema.ShardIndex) {
			continue
		}

		if parser == nil || parser.Check(rowData) {
			keys = append(keys, k)
		}
	}

	nRow := int64(len(keys))
	if nRow == 0 {
		return &DBReply{Data: int64(0)}
	}

	for _, k := range keys {
		delete(table, k)
	}

	return &DBReply{Data: nRow}
}
