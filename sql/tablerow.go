package sql

import (
	"fmt"
	"go-learner/slice"

	"strconv"
	"strings"
	"time"
)

const (
	TableRowStateNone uint8 = iota
	TableRowStateInit
	TableRowStateNotExist
	TableRowStateValid
)

type Table struct {
	NumDBMultiReq int16
	NumDBReq      int16
	Idx           int32
	Schema        *TableSchema
	rows          []TableRow
	nRow          int32
	lastRowIdx    int32
	m             map[string]int32
	shardIdxes    []TableRowIndex
	nSI           int32
	lastSIIdx     int32
	mSI           map[string]int32
	lastHit       int64
	ttl           int64
	Outdated      bool
	initRowNum    int32
}

type TableRow struct {
	State           uint8
	HasShardIndex   bool
	HasShardPending bool
	NumDBReq        int16
	NumDBSyncReq    int16
	LastHitTime     int64
	Data            []byte
	DBContext       *RowContext
	prevIdx         int32
	nextIdx         int32
}

type TableRowIndex struct {
	State            uint8
	NumDBReq         int16
	NumDBSyncReq     int16
	NumPendingInsert int32
	DBContext        *RowContext
	LastHitTime      int64
	ShardKey         string
	Idxes            []int32
	prevIdx          int32
	nextIdx          int32
}

func (tb *Table) ReloadSchema(schema *TableSchema) (*Table, int32) {
	if n, err := tb.Reset(); err == nil {
		tb.Schema = schema
		return tb, n
	}

	inst := &Table{
		Idx:        tb.Idx,
		Schema:     schema,
		rows:       make([]TableRow, tb.initRowNum),
		m:          make(map[string]int32, tb.initRowNum),
		lastHit:    time.Now().UTC().Unix(),
		ttl:        tb.ttl,
		Outdated:   false,
		initRowNum: tb.initRowNum,
	}

	tb.Outdated = true
	return inst, tb.nRow
}

func (tb *Table) Reset() (int32, error) {
	if tb.NumDBMultiReq > 0 || tb.NumDBReq > 0 {
		return 0, fmt.Errorf("%s has DB request pending", tb.Schema.Name)
	}

	nRow := tb.nRow
	if nRow == 0 {
		return 0, nil
	}

	tb.NumDBReq = 0
	tb.NumDBMultiReq = 0
	tb.rows = make([]TableRow, tb.initRowNum)
	tb.nRow = 0
	tb.lastRowIdx = 0
	tb.m = make(map[string]int32, tb.initRowNum)
	tb.shardIdxes = nil
	tb.nSI = 0
	tb.lastSIIdx = 0
	tb.mSI = nil
	tb.lastHit = time.Now().UTC().Unix()

	return nRow, nil
}

func (tb *Table) HitMultiRow(shardKey string) (*TableRowIndex, int32, int32, int32) {
	currentTime := time.Now().UTC().Unix()
	tb.lastHit = currentTime

	if tb.mSI == nil {
		initNum := tb.initRowNum >> 2
		if initNum < 64 {
			initNum = 64
		}
		tb.mSI = make(map[string]int32, initNum)
		tb.shardIdxes = make([]TableRowIndex, initNum)
	}

	var nDBSyncReqRow, nDBReqRow int32
	shardIdxes := tb.shardIdxes
	if idx, ok := tb.mSI[shardKey]; ok {
		si := &shardIdxes[idx]
		if si.Idxes != nil {
			for _, i := range si.Idxes {
				row := tb.hitRow(i, currentTime)
				if row.NumDBSyncReq > 0 {
					nDBSyncReqRow++
				}
				if row.NumDBReq > 0 {
					nDBReqRow++
				}
			}
		}

		si.LastHitTime = currentTime
		if idx == tb.lastSIIdx {
			return si, idx, nDBSyncReqRow, nDBReqRow
		}

		shardIdxes[si.prevIdx].nextIdx = si.nextIdx
		shardIdxes[si.nextIdx].prevIdx = si.prevIdx

		last := &shardIdxes[tb.lastSIIdx]
		shardIdxes[last.nextIdx].prevIdx = idx
		si.nextIdx = last.nextIdx
		last.nextIdx = idx
		si.prevIdx = tb.lastSIIdx

		tb.lastSIIdx = idx
		return si, idx, nDBSyncReqRow, nDBReqRow
	}

	n := int32(len(shardIdxes))
	if tb.nSI >= n-1 {
		tb.Recycle(0, currentTime-tb.ttl)
		if tb.nSI >= n-1 {
			//n MUST > 1
			tb.shardIdxes = append(tb.shardIdxes, make([]TableRowIndex, n>>1)...)
			shardIdxes = tb.shardIdxes
		}
	}

	tb.nSI++

	last := &shardIdxes[tb.lastSIIdx]
	var si *TableRowIndex
	if last.nextIdx > 0 {
		si = &shardIdxes[last.nextIdx]
		tb.lastSIIdx = last.nextIdx

	} else {
		lastIdx := tb.nSI
		si = &shardIdxes[lastIdx]
		last.nextIdx = lastIdx
		si.prevIdx = tb.lastSIIdx
		tb.lastSIIdx = lastIdx
		shardIdxes[0].prevIdx = lastIdx
	}

	si.LastHitTime = currentTime
	si.ShardKey = shardKey
	//si.Idxes = make([]int, 0, sql.MaxMultiRowNum)

	tb.mSI[shardKey] = tb.lastSIIdx
	return si, tb.lastSIIdx, nDBSyncReqRow, nDBReqRow
}

func (tb *Table) InsertRowIndex(shardKey string, row *TableRow, rowIdx int32) bool {
	if tb.nSI < 1 {
		return false
	}

	idx, ok := tb.mSI[shardKey]
	if !ok {
		return false
	}

	si := &tb.shardIdxes[idx]
	return si.InsertRow(row, rowIdx)
}

func (tb *Table) GetShardDBReqNum(shardKey string) (int16, int16) {
	if tb.nSI < 1 {
		return 0, 0
	}

	idx, ok := tb.mSI[shardKey]
	if !ok {
		return 0, 0
	}

	si := &tb.shardIdxes[idx]
	return si.NumDBSyncReq, si.NumDBReq
}

func (tb *Table) GetShardIndex(shardKey string) *TableRowIndex {
	if tb.nSI < 1 {
		return nil
	}

	idx, ok := tb.mSI[shardKey]
	if !ok {
		return nil
	}

	return &tb.shardIdxes[idx]
}

func (tb *Table) GetShardIndexByIdx(idx int32) *TableRowIndex {
	return &tb.shardIdxes[idx]
}

func (tb *Table) HitRow(key string) (*TableRow, int32) {
	currentTime := time.Now().UTC().Unix()
	tb.lastHit = currentTime

	if idx, ok := tb.m[key]; ok {
		return tb.hitRow(idx, currentTime), idx
	}

	rows := tb.rows
	if tb.nRow >= int32(len(rows))-1 {
		if tb.Recycle(0, currentTime-tb.ttl) < 1 {
			tb.rows = append(tb.rows, make([]TableRow, (tb.nRow+1)>>1)...)
			rows = tb.rows
		}
	}

	tb.nRow++

	last := &rows[tb.lastRowIdx]
	var row *TableRow
	if last.nextIdx > 0 {
		row = &rows[last.nextIdx]
		tb.lastRowIdx = last.nextIdx

	} else {
		lastIdx := tb.nRow
		row = &rows[lastIdx]
		last.nextIdx = lastIdx
		row.prevIdx = tb.lastRowIdx
		tb.lastRowIdx = lastIdx
		rows[0].prevIdx = lastIdx
	}

	row.Data = slice.String2ByteSlice(key)
	row.LastHitTime = currentTime
	tb.m[key] = tb.lastRowIdx

	return row, tb.lastRowIdx
}

func (tb *Table) GetRowByIdx(idx int32) *TableRow {
	return &tb.rows[idx]
}

func (tb *Table) CheckAndClear(force bool) {
	if !tb.Outdated || tb.lastHit == 0 {
		return
	}

	if !force && (tb.NumDBMultiReq > 0 || tb.NumDBReq > 0) {
		return
	}

	tb.NumDBReq = 0
	tb.NumDBMultiReq = 0
	tb.rows = nil
	tb.nRow = 0
	tb.lastRowIdx = 0
	tb.m = nil
	tb.shardIdxes = nil
	tb.nSI = 0
	tb.lastSIIdx = 0
	tb.mSI = nil
	tb.lastHit = 0
	tb.Schema = nil
	return
}

func (tb *Table) Recycle(num int32, expireTime int64) int32 {
	if tb.Outdated {
		return 0
	}

	if tb.lastHit < expireTime {
		if n, err := tb.Reset(); err == nil {
			return n
		}
	}

	tb.recycleSI(num, expireTime)

	if num > tb.nRow || num < 1 {
		num = tb.nRow
	}

	rows := tb.rows
	head := &rows[0]

	currIdx := head.nextIdx
	curr := &rows[currIdx]

	lastIdx := tb.lastRowIdx
	last := &rows[lastIdx]
	end := &rows[last.nextIdx]

	var n int32
	for currIdx > 0 && n < num {
		if curr.LastHitTime > expireTime {
			break
		}

		if curr.NumDBReq > 0 || curr.HasShardIndex {
			// SHOULD NOT happen
			//break
		}

		key := GetRowKey(tb.Schema, curr.Data)
		delete(tb.m, key)
		curr.reset()
		n++

		if currIdx == lastIdx {
			tb.lastRowIdx = 0
			break
		}

		nextIdx := curr.nextIdx
		next := &rows[nextIdx]

		//remove from list
		head.nextIdx = nextIdx
		next.prevIdx = curr.prevIdx

		//add after lastIdx
		curr.prevIdx = lastIdx
		curr.nextIdx = last.nextIdx
		last.nextIdx = currIdx
		end.prevIdx = currIdx

		end = curr
		curr = next
		currIdx = nextIdx
	}

	tb.nRow -= n
	return n
}

func (tb *Table) RowDebugInfo(key string) (*TableRow, string) {
	idx, ok := tb.m[key]
	if !ok {
		return nil, ""
	}

	row := &tb.rows[idx]
	return row, row.DebugInfo(tb.Schema)
}

func (tb *Table) hitRow(idx int32, currentTime int64) *TableRow {
	rows := tb.rows
	row := &rows[idx]

	row.LastHitTime = currentTime
	if idx == tb.lastRowIdx {
		return row
	}

	rows[row.prevIdx].nextIdx = row.nextIdx
	rows[row.nextIdx].prevIdx = row.prevIdx

	last := &rows[tb.lastRowIdx]
	rows[last.nextIdx].prevIdx = idx
	row.nextIdx = last.nextIdx
	last.nextIdx = idx
	row.prevIdx = tb.lastRowIdx

	tb.lastRowIdx = idx
	return row
}

func (tb *Table) recycleSI(num int32, expireTime int64) int32 {
	if tb.nSI == 0 {
		return 0
	}

	if num > tb.nSI || num < 1 {
		num = tb.nSI
	}

	shardIdxes := tb.shardIdxes
	head := &shardIdxes[0]

	currIdx := head.nextIdx
	curr := &shardIdxes[currIdx]

	lastIdx := tb.lastSIIdx
	last := &shardIdxes[lastIdx]
	end := &shardIdxes[last.nextIdx]

	n := int32(0)
	for currIdx > 0 && n < num {
		if curr.LastHitTime > expireTime {
			break
		}

		if curr.NumDBReq > 0 {
			//break
		}

		delete(tb.mSI, curr.ShardKey)
		curr.reset(tb)
		n++

		if currIdx == lastIdx {
			tb.lastSIIdx = 0
			break
		}

		nextIdx := curr.nextIdx
		next := &shardIdxes[nextIdx]

		//remove from list
		head.nextIdx = nextIdx
		next.prevIdx = curr.prevIdx

		//add to lastRowIdx
		curr.prevIdx = lastIdx
		curr.nextIdx = last.nextIdx
		last.nextIdx = currIdx
		end.prevIdx = currIdx

		end = curr
		curr = next
		currIdx = nextIdx
	}

	tb.nSI -= n
	return n
}

func (tr *TableRow) Update(schema *TableSchema, fields map[string]string) (bool, error) {
	mapFields := RowData2Map(schema, tr.Data)
	if mapFields == nil {
		return false, fmt.Errorf("invalid row data")
	}

	//bChange := false
	for k, v := range fields {
		fieldSchema := schema.GetColumnSchema(k)
		if fieldSchema == nil {
			return false, fmt.Errorf("invalid field")
		}

		if fieldSchema.IsPrimaryKey {
			return false, fmt.Errorf("invalid field")
		}

		//check this maybe has performance issue for big value
		//if mapFields[k] != v {
		//	bChange = true
		//}
		mapFields[k] = v
	}

	//if !bChange {
	//	return false, nil
	//}

	for _, i := range schema.AutoMTimeFields {
		mapFields[schema.Columns[i].Name] = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	tr.Data = NewRowDataFromMap(schema, mapFields)
	return true, nil
}

func (tr *TableRow) Update2(schema *TableSchema, fields []string) (bool, string) {
	nUpdate := len(fields)
	if nUpdate&1 != 0 {
		return false, "invalid fields"
	}

	curr := RowData2Slice(schema, tr.Data)
	if curr == nil {
		return false, "invalid row data"
	}

	for i := 0; i < nUpdate; i += 2 {
		column := schema.GetColumnSchema(fields[i])
		if column == nil || column.IsPrimaryKey {
			return false, "invalid fields"
		}

		j := (column.Index << 1) + 1
		curr[j] = fields[i+1]
	}

	for _, i := range schema.AutoMTimeFields {
		curr[(i<<1)+1] = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	tr.Data = NewRowDataFromSlice(schema, curr)
	return true, ""
}

func (tr *TableRow) IncrBy(schema *TableSchema, idx int, delta int64) (bool, string) {
	value := GetValueByIndex(schema, tr.Data, idx)
	intV, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false, err.Error()
	}

	curr := RowData2Slice(schema, tr.Data)
	curr[(idx<<1)+1] = strconv.FormatInt(intV+delta, 10)

	for _, i := range schema.AutoMTimeFields {
		curr[(i<<1)+1] = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	tr.Data = NewRowDataFromSlice(schema, curr)
	return true, ""
}

func (tr *TableRow) DebugInfo(schema *TableSchema) string {
	states := [...]string{"None", "Init", "NotExist", "Valid"}
	var builder strings.Builder
	builder.WriteString("State:" + states[tr.State])
	builder.WriteString(fmt.Sprintf("; NumDBReq:%d", tr.NumDBReq))
	builder.WriteString(fmt.Sprintf("; NumDBSyncReq:%d", tr.NumDBSyncReq))
	builder.WriteString(fmt.Sprintf("; HasSI:%t", tr.HasShardIndex))
	if tr.Data == nil {
		builder.WriteString(", Data:nil")

	} else if tr.Data[0] == PrimaryKeySeparator {
		builder.WriteString(fmt.Sprintf(", Data:key"))

	} else {
		builder.WriteString(", Data:")
		prev := "{"
		fields := RowData2Map(schema, tr.Data)
		for k, v := range fields {
			column := schema.GetColumnSchema(k)
			if column.IsNumber {
				builder.WriteString(fmt.Sprintf("%s%s:%s", prev, k, v))

			} else {
				builder.WriteString(fmt.Sprintf("%s%s:\"%s\"", prev, k, v))
			}

			prev = ", "
		}
		builder.WriteByte('}')
	}

	sTime := fmt.Sprintf("; LastHitTime:\"%s\"",
		time.Unix(tr.LastHitTime, 0).UTC().Format("2006-01-02 15:04:05"))
	builder.WriteString(sTime)

	if tr.DBContext == nil {
		builder.WriteString("; DBContext:nil")

	} else {
		builder.WriteString(fmt.Sprintf("; DBContext:NotNil"))
	}

	return builder.String()
}

func (tr *TableRow) reset() {
	tr.LastHitTime = 0
	tr.Data = nil
	tr.State = TableRowStateNone
	tr.HasShardIndex = false
	tr.HasShardPending = false
	tr.NumDBReq = 0
	tr.NumDBSyncReq = 0
	tr.DBContext = nil
	//tr.nextIdx = 0
	//tr.prevIdx = 0
}

func (ti *TableRowIndex) reset(table *Table) {
	ti.Expire(table)
	ti.LastHitTime = 0
	ti.NumDBReq = 0
	ti.NumDBSyncReq = 0
	ti.NumPendingInsert = 0
	ti.DBContext = nil
	ti.ShardKey = ""
	//ti.nextIdx = 0
	//ti.prevIdx = 0
}

func (ti *TableRowIndex) Expire(table *Table) {
	if ti.Idxes != nil {
		for _, i := range ti.Idxes {
			row := table.GetRowByIdx(i)
			if row == nil {
				//SHOULD NOT happen
				continue
			}

			row.HasShardIndex = false
		}
	}

	ti.State = TableRowStateNone
	ti.Idxes = nil
}

func (ti *TableRowIndex) InsertRow(row *TableRow, rowIdx int32) bool {
	if ti.State == TableRowStateNotExist {
		ti.State = TableRowStateValid
	}

	//状态未定，没有必要插入
	if row.HasShardIndex || ti.State != TableRowStateValid {
		return false
	}

	if ti.Idxes == nil {
		ti.Idxes = make([]int32, 0, 256)
	}

	ti.Idxes = append(ti.Idxes, rowIdx)
	row.HasShardIndex = true

	return true
}
