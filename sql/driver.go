package sql

import (
	"encoding/binary"
	"fmt"
	"go-learner/slice"
	"math"
	"strings"
	"sync"
	"time"
)

var ColumnNumberTypes map[string]ColumnType
var ColumnTimeTypes map[string]ColumnType

// PrimaryKeySeparator ASCII HT = 9
const PrimaryKeySeparator byte = 9

type ColumnType uint8

const (
	ColumnTypeString ColumnType = iota
	ColumnTypeInt
	ColumnTypeFloat
	ColumnTypeTime
)

type Driver interface {
	LoadTableSchema(tableName string) (*TableSchema, error)
	Insert(schema *TableSchema, shard string, fields []string) *DBReply
	DeleteSingle(schema *TableSchema, shard string, keys []string) *DBReply
	UpdateSingle(schema *TableSchema, shard string, keys []string, fields []string) *DBReply
	IncrBySingle(schema *TableSchema, shard string, keys []string, data *IncrByData) *DBReply
	SelectSingle(schema *TableSchema, shard string, keys []string) *DBReply
	SelectMulti(schema *TableSchema, shard string) *DBReply
	DeleteMulti(schema *TableSchema, shard string, data *MultiRequestData) *DBReply
}

type TableSchemaManager struct {
	sync.RWMutex

	schemas map[string]*TableSchema
	driver  Driver
}

type MultiRequestData struct {
	Where  string
	Params []string
	Parser *Parser
}

type IncrByData struct {
	Column string
	Delta  int64
	Where  string
}

func (tsm *TableSchemaManager) LoadSchema(name string) (*TableSchema, error) {
	tsm.RLock()
	schema := tsm.schemas[name]
	tsm.RUnlock()

	if schema != nil {
		return schema, nil
	}

	tsm.Lock()
	defer tsm.Unlock()

	schema = tsm.schemas[name]
	if schema != nil {
		return schema, nil
	}

	var err error
	schema, err = tsm.driver.LoadTableSchema(name)
	if err != nil {
		return nil, err
	}

	tsm.schemas[name] = schema
	return schema, nil
}

func (tsm *TableSchemaManager) ReloadSchema(name string, force bool) (*TableSchema, error) {
	tsm.Lock()
	defer tsm.Unlock()

	schema1, ok := tsm.schemas[name]
	if !ok {
		return nil, fmt.Errorf("not found table %s in memory, no need reload for a new table", name)
	}

	var err error
	schema2, err := tsm.driver.LoadTableSchema(name)
	if err != nil {
		return nil, err
	}

	if schema1.NumPrimaryKeys != schema2.NumPrimaryKeys {
		// 这里没有作严格的检查，因为通常不允许有删除列的操作
		return nil, fmt.Errorf("can not alter key")
	}

	if !force {
		nL1 := len(schema1.Columns)
		nL2 := len(schema2.Columns)
		if nL1 == nL2 {
			eq := true
			for i := 0; i < nL1; i++ {
				if !schema1.Columns[i].Equal(&schema2.Columns[i]) {
					eq = false
					break
				}
			}
			if eq {
				return nil, nil
			}
		}
	}

	tsm.schemas[name] = schema2
	return schema2, nil
}

func (tsm *TableSchemaManager) Driver() Driver {
	return tsm.driver
}

type TableSchema struct {
	Name              string
	ShardKey          string
	ShardIndex        int
	IsStringShardKey  bool
	Columns           []ColumnSchema
	NumPrimaryKeys    int
	PrimaryKeyIndexes []int
	AutoMTimeFields   []int

	m                 map[string]int
	whereSingleClause string
	selectSingle      string
	selectMulti       string
	updatePrefix      string
	insertClause1     string
	insertClause2     string
	deleteSingle      string
	deleteMultiPrefix string
}

type ColumnSchema struct {
	IsNumber     bool
	IsPrimaryKey bool
	Index        int
	Name         string
	Type         ColumnType
	DefaultValue string
}

func (inst *ColumnSchema) Equal(other *ColumnSchema) bool {
	if inst.Type != other.Type {
		return false
	}
	if inst.IsPrimaryKey != other.IsPrimaryKey {
		return false
	}
	if inst.Name != other.Name {
		return false
	}
	if inst.DefaultValue != inst.DefaultValue {
		return false
	}

	return true
}

func NewRowData(schema *TableSchema, fieldData [][]byte) []byte {
	nColumn := len(fieldData)
	if nColumn == 0 {
		return nil
	}

	columns := schema.Columns
	if len(columns) < nColumn {
		return nil
	}

	var nData int
	for i := 0; i < nColumn; i++ {
		nData += len(fieldData[i])
	}

	if nData == 0 {
		return nil
	}

	nHeader := nColumn*4 + 1
	nPrimaryKey := schema.NumPrimaryKeys
	nData += nHeader + nPrimaryKey
	if nData > math.MaxUint32 {
		return nil
	}

	rowData := make([]byte, nData)
	rowData[0] = 0

	headerStart := uint32(1)
	dataStart := uint32(nHeader)

	for i := 0; i < nColumn; i++ {
		data := fieldData[i]
		dataEnd := dataStart + uint32(len(data))
		//primary keys will start with PrimaryKeySeparator
		if columns[i].IsPrimaryKey {
			dataEnd++
			rowData[dataStart] = PrimaryKeySeparator
			dataStart++
		}

		headerEnd := headerStart + 4
		binary.LittleEndian.PutUint32(rowData[headerStart:headerEnd], dataEnd)
		headerStart = headerEnd

		copy(rowData[dataStart:dataEnd], data)
		dataStart = dataEnd
	}

	return rowData
}

func NewRowDataFromMap(schema *TableSchema, fields map[string]string) []byte {
	columns := schema.Columns
	nColumn := len(columns)

	nField := 0
	b := make([][]byte, nColumn)
	for name, value := range fields {
		column := schema.GetColumnSchema(name)
		if column == nil {
			return nil
		}

		nField++
		b[column.Index] = slice.String2ByteSlice(value)
	}
	for _, idx := range schema.AutoMTimeFields {
		if b[idx] == nil {
			b[idx] = slice.String2ByteSlice(time.Now().UTC().Format("2006-01-02 15:04:05"))
		}
	}

	if nField > nColumn {
		return nil
	}

	return NewRowData(schema, b)
}

func NewRowDataFromSlice(schema *TableSchema, fields []string) []byte {
	columns := schema.Columns
	nColumn := len(columns)

	nData := len(fields)
	if nData&1 == 1 || nData>>1 > nColumn {
		return nil
	}

	b := make([][]byte, nColumn)
	for i := 0; i < nData; i += 2 {
		column := schema.GetColumnSchema(fields[i])
		if column == nil {
			return nil
		}

		b[column.Index] = slice.String2ByteSlice(fields[i+1])
	}
	for _, idx := range schema.AutoMTimeFields {
		if b[idx] == nil {
			b[idx] = slice.String2ByteSlice(time.Now().UTC().Format("2006-01-02 15:04:05"))
		}
	}

	return NewRowData(schema, b)
}

func RowData2Map(schema *TableSchema, rowData []byte) map[string]string {
	if len(rowData) < 2 || rowData[0] != 0 {
		return nil
	}

	nColumn := len(schema.Columns)
	columns := schema.Columns

	ret := make(map[string]string, nColumn)
	headerStart := uint32(1)
	dataStart := uint32(4*nColumn + 1)

	for i := 0; i < nColumn; i++ {
		headerEnd := headerStart + 4
		dataEnd := binary.LittleEndian.Uint32(rowData[headerStart:headerEnd])
		headerStart = headerEnd

		cs := &columns[i]
		//attention: no copy
		var v string
		if cs.IsPrimaryKey {
			v = slice.ByteSlice2String(rowData[dataStart+1 : dataEnd])

		} else {
			v = slice.ByteSlice2String(rowData[dataStart:dataEnd])
		}
		if v == "" {
			v = cs.DefaultValue
		}

		dataStart = dataEnd
		//if v == "" && columns[i].Type != ColumnTypeString {
		//	continue
		//}

		ret[columns[i].Name] = v
	}

	return ret
}

func RowData2Slice(schema *TableSchema, rowData []byte) []string {
	if len(rowData) < 2 || rowData[0] != 0 {
		return nil
	}

	nColumn := len(schema.Columns)
	columns := schema.Columns

	ret := make([]string, nColumn<<1)
	headerStart := uint32(1)
	dataStart := uint32(4*nColumn + 1)

	j := 0
	for i := 0; i < nColumn; i++ {
		headerEnd := headerStart + 4
		dataEnd := binary.LittleEndian.Uint32(rowData[headerStart:headerEnd])
		headerStart = headerEnd

		//attention: no copy
		cs := &columns[i]
		ret[j] = cs.Name

		var v string
		if cs.IsPrimaryKey {
			v = slice.ByteSlice2String(rowData[dataStart+1 : dataEnd])

		} else {
			v = slice.ByteSlice2String(rowData[dataStart:dataEnd])
		}
		if v == "" {
			v = cs.DefaultValue
		}
		ret[j+1] = v

		dataStart = dataEnd
		j += 2
	}

	return ret
}

func GetRowKey(schema *TableSchema, rowData []byte) string {
	if len(rowData) < 2 {
		return ""
	}

	//row在状态为None或NotExist时，data只存储key的信息
	if rowData[0] == PrimaryKeySeparator {
		return slice.ByteSlice2String(rowData)
	}

	//key is sorted
	if schema.PrimaryKeyIndexes == nil {
		headerStart := (schema.NumPrimaryKeys-1)*4 + 1
		dataStart := uint32(4*len(schema.Columns) + 1)
		dataEnd := binary.LittleEndian.Uint32(rowData[headerStart : headerStart+4])

		return slice.ByteSlice2String(rowData[dataStart:dataEnd])
	}

	var builder strings.Builder
	for _, i := range schema.PrimaryKeyIndexes {
		builder.WriteByte(PrimaryKeySeparator)
		builder.WriteString(GetValueByIndex(schema, rowData, i))
	}

	return builder.String()
}

func AssembleRowKey(schema *TableSchema, keys map[string]string) string {
	nPrimaryKey := schema.NumPrimaryKeys
	if len(keys) != nPrimaryKey {
		return ""
	}

	if schema.PrimaryKeyIndexes == nil {
		pks := make([][]byte, nPrimaryKey)
		n := 0
		for name, value := range keys {
			column := schema.GetColumnSchema(name)
			if column == nil {
				return ""
			}

			if !column.IsPrimaryKey {
				return ""
			}

			n += len(value)
			pks[column.Index] = slice.String2ByteSlice(value)
		}

		b := make([]byte, 0, n+nPrimaryKey)
		for i := 0; i < nPrimaryKey; i++ {
			b = append(b, PrimaryKeySeparator)
			b = append(b, pks[i]...)
		}

		return slice.ByteSlice2String(b)
	}

	var builder strings.Builder
	columns := schema.Columns
	for _, i := range schema.PrimaryKeyIndexes {
		v, ok := keys[columns[i].Name]
		if !ok {
			return ""
		}

		builder.WriteByte(PrimaryKeySeparator)
		builder.WriteString(v)
	}

	return builder.String()
}

func AssembleRowKey2(schema *TableSchema, keys []string) string {
	nPrimaryKey := schema.NumPrimaryKeys
	if len(keys) != nPrimaryKey {
		return ""
	}

	nLen := nPrimaryKey
	for i := 0; i < nPrimaryKey; i++ {
		nLen += len(keys[i])
	}

	bValue := make([]byte, nLen)
	th, i := 0, 0
	for th < nPrimaryKey {
		bValue[i] = PrimaryKeySeparator
		i++

		i += copy(bValue[i:], slice.String2ByteSlice(keys[th]))
		th++
	}

	return slice.ByteSlice2String(bValue)
}

func GetValueByIndex(schema *TableSchema, rowData []byte, idx int) string {
	if len(rowData) < 2 || rowData[0] == PrimaryKeySeparator {
		return ""
	}

	//invoker need guarantee idx is valid
	headerStart := idx*4 + 1
	dataEnd := binary.LittleEndian.Uint32(rowData[headerStart : headerStart+4])

	var dataStart uint32
	if idx > 0 {
		dataStart = binary.LittleEndian.Uint32(rowData[headerStart-4 : headerStart])

	} else {
		dataStart = uint32(len(schema.Columns)*4 + 1)
	}

	cs := &schema.Columns[idx]
	if cs.IsPrimaryKey {
		dataStart += 1
	}

	v := slice.ByteSlice2String(rowData[dataStart:dataEnd])
	if v == "" {
		v = cs.DefaultValue
	}

	return v
}

func CheckFields(schema *TableSchema, fields []string) bool {
	columns := schema.Columns
	nColumn := len(columns)

	nData := len(fields)
	if nData&1 == 1 || nData>>1 > nColumn {
		return false
	}

	for i := 0; i < nData; i += 2 {
		column := schema.GetColumnSchema(fields[i])
		if column == nil {
			return false
		}
	}

	return true
}

func (ts *TableSchema) GetColumnSchema(name string) *ColumnSchema {
	if idx, ok := ts.m[name]; ok {
		return &ts.Columns[idx]
	}

	return nil
}
