package test

import (
	"fmt"
	"go-learner/slice"
	"go-learner/sql"
	"math/rand"
	"strconv"
	"sync"
)

type FieldGenerator struct {
	MinValue int64
	MaxValue int64
	MinLen   int64
	MaxLen   int64
}

func (fc *FieldGenerator) RandomString() string {
	rLen := rand.Int63n(fc.MaxLen-fc.MinLen+1) + fc.MinLen
	if rLen < 1 {
		return ""
	}

	b := make([]byte, rLen)
	return slice.ByteSlice2String(b)
}

func (fc *FieldGenerator) RandomInt() string {
	rV := rand.Int63n(fc.MaxValue-fc.MinValue+1) + fc.MinValue
	return strconv.FormatInt(rV, 10)
}

type TableReqGenerator struct {
	schema       *sql.TableSchema
	fieldGens    []FieldGenerator
	updateFields []string
	incrByFields []string
	selectWheres []string
	deleteWheres []string
}

func NewTableReqGenerator(schema *sql.TableSchema, fieldGenerators []FieldGenerator) *TableReqGenerator {
	if fieldGenerators != nil && len(schema.Columns) != len(fieldGenerators) {
		panic("invalid field generators: " + schema.Name)
	}

	ret := &TableReqGenerator{
		schema:       schema,
		fieldGens:    fieldGenerators,
		updateFields: make([]string, 0, 1),
		incrByFields: make([]string, 0, 1),
		selectWheres: make([]string, 0, 1),
		deleteWheres: make([]string, 0, 1),
	}

	nColumn := len(schema.Columns)
	for i := 0; i < nColumn; i++ {
		column := &schema.Columns[i]
		if column.IsPrimaryKey {
			continue
		}

		switch column.Type {
		case sql.ColumnTypeInt:
			ret.AddUpdateColumns(column.Name)
			ret.AddIncrByColumn(column.Name)

			ret.AddSelectWhere(column.Name)
			ret.AddDeleteWhere(column.Name)

		case sql.ColumnTypeString:
			ret.AddUpdateColumns(column.Name)

		default:
			// do nothing
		}
	}

	return ret
}

func (tg *TableReqGenerator) AddUpdateColumns(column string) {
	tg.updateFields = append(tg.updateFields, column)
}

func (tg *TableReqGenerator) AddIncrByColumn(column string) {
	tg.incrByFields = append(tg.incrByFields, column)
}

func (tg *TableReqGenerator) AddSelectWhere(column string) {
	tg.selectWheres = append(tg.selectWheres, column)
}

func (tg *TableReqGenerator) AddDeleteWhere(column string) {
	tg.deleteWheres = append(tg.deleteWheres, column)
}

func (tg *TableReqGenerator) RandomKey() (int32, []string) {
	n := tg.schema.NumPrimaryKeys
	primaryKeys := make([]string, n)
	var hintId int32
	if tg.schema.PrimaryKeyIndexes == nil {
		for i := 0; i < n; i++ {
			primaryKeys[i] = tg.randomFieldValue(i)
		}

		hintId = int32(tg.schema.ShardIndex)

	} else {
		for i, idx := range tg.schema.PrimaryKeyIndexes {
			primaryKeys[i] = tg.randomFieldValue(idx)
			if idx == tg.schema.ShardIndex {
				hintId = int32(i)
			}
		}
	}

	return hintId, primaryKeys
}

func (tg *TableReqGenerator) randomFieldValue(idx int) string {
	if tg.fieldGens == nil {
		panic("no field generator")
	}

	if tg.schema.Columns[idx].IsNumber {
		return tg.fieldGens[idx].RandomInt()
	}

	return tg.fieldGens[idx].RandomString()
}

func (tg *TableReqGenerator) GetKeysRange() [][2]int64 {
	if tg.fieldGens == nil {
		panic("no field generator")
	}

	ret := make([][2]int64, tg.schema.NumPrimaryKeys)
	if tg.schema.PrimaryKeyIndexes == nil {
		for i := 0; i < tg.schema.NumPrimaryKeys; i++ {
			ret[i] = [2]int64{tg.fieldGens[i].MinValue, tg.fieldGens[i].MaxValue}
		}

	} else {
		for i, idx := range tg.schema.PrimaryKeyIndexes {
			ret[i] = [2]int64{tg.fieldGens[idx].MinValue, tg.fieldGens[idx].MaxValue}
		}
	}

	return ret
}

func (tg *TableReqGenerator) GetSchema() *sql.TableSchema {
	return tg.schema
}

type ReqGenerator struct {
	sync.RWMutex
	tableReqGens   []*TableReqGenerator
	tableTracers   []*TableTracer
	multiKeyIdxes  []int
	singleKeyIdxes []int
	m              map[string]int
}

func NewReqGenerator() *ReqGenerator {
	pre := 128
	ret := &ReqGenerator{
		tableReqGens:   make([]*TableReqGenerator, 0, pre),
		tableTracers:   make([]*TableTracer, 0, pre),
		multiKeyIdxes:  make([]int, 0, pre),
		singleKeyIdxes: make([]int, 0, pre),
		m:              make(map[string]int, pre),
	}

	return ret
}

func (rg *ReqGenerator) AddTableReqGenerators(schemaMgr *sql.TableSchemaManager, tableInfo ITableInfo, trace bool) error {
	rg.Lock()
	defer rg.Unlock()

	count := tableInfo.GetCount()
	fields := tableInfo.GetFields()
	for i := 0; i < count; i++ {
		tableName := tableInfo.GetName(i)
		if _, ok := rg.m[tableName]; ok {
			return fmt.Errorf("duplicate table")
		}

		schema, err := schemaMgr.LoadSchema(tableName)
		if err != nil {
			return err
		}

		idx := len(rg.tableReqGens)
		rg.tableReqGens = append(rg.tableReqGens, NewTableReqGenerator(schema, fields))
		if trace {
			rg.tableTracers = append(rg.tableTracers, NewTableTracer())

		} else {
			rg.tableTracers = append(rg.tableTracers, nil)
		}

		if schema.NumPrimaryKeys == 1 {
			rg.singleKeyIdxes = append(rg.singleKeyIdxes, idx)

		} else {
			rg.multiKeyIdxes = append(rg.multiKeyIdxes, idx)
		}

		rg.m[tableName] = idx
	}

	return nil
}

func (rg *ReqGenerator) GetTableReqGenerator(tableIdx int) *TableReqGenerator {
	rg.RLock()
	defer rg.RUnlock()

	if tableIdx < 0 || tableIdx >= len(rg.tableReqGens) {
		panic("invalid table idx")
	}

	return rg.tableReqGens[tableIdx]
}

func (rg *ReqGenerator) GetTableTracer(tableIdx int) *TableTracer {
	rg.RLock()
	defer rg.RUnlock()

	if tableIdx < 0 || tableIdx >= len(rg.tableTracers) {
		panic("invalid table idx")
	}

	return rg.tableTracers[tableIdx]
}

func (rg *ReqGenerator) ResetTableReqGenerator(tg *TableReqGenerator, idx int) {
	rg.Lock()
	defer rg.Unlock()

	if idx >= len(rg.tableReqGens) {
		panic("exceed table num")
	}
	rg.tableReqGens[idx] = tg
}

func (rg *ReqGenerator) RandomSingleKeyTableIdxes(num int) []int {
	rg.RLock()
	defer rg.RUnlock()

	l := len(rg.singleKeyIdxes)
	idxes := make([]int, num)
	for i := 0; i < num; i++ {
		idxes[i] = rg.singleKeyIdxes[rand.Intn(l)]
	}

	return idxes
}

func (rg *ReqGenerator) RandomMultiKeyTableIdxes(num int) []int {
	rg.RLock()
	defer rg.RUnlock()

	l := len(rg.multiKeyIdxes)
	idxes := make([]int, num)
	for i := 0; i < num; i++ {
		idxes[i] = rg.multiKeyIdxes[rand.Intn(l)]
	}

	return idxes
}

func (rg *ReqGenerator) RandomTableIdxes(num int) []int {
	rg.RLock()
	defer rg.RUnlock()

	l := len(rg.tableReqGens)
	idxes := make([]int, num)
	for i := 0; i < num; i++ {
		idxes[i] = rand.Intn(l)
	}

	return idxes
}
