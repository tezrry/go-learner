package test

import (
	"go-learner/sql"
	"sync"
)

type ReqTracer struct {
	Req     interface{}
	Invalid bool

	expectedFields map[string]string
	expectedNum    int64
	expectedRows   []map[string]string
	next           *ReqTracer
}

func NewReqTracer(req interface{}, invalid bool) *ReqTracer {
	return &ReqTracer{Req: req, Invalid: invalid}
}

type TableRowTracer struct {
	nTotalReq int
	key       string
	shard     string
	//lastExpectedFields map[string]string
	currFields map[string]string
	reqHead    *ReqTracer
	reqTail    *ReqTracer
}

func (tr *TableRowTracer) onRequest(reqTracer *ReqTracer) {
	if tr.reqTail == nil {
		tr.reqHead.next = reqTracer
		tr.reqTail = reqTracer

	} else {
		tr.reqTail.next = reqTracer
		tr.reqTail = reqTracer
	}
}

func (tr *TableRowTracer) onReply(req interface{}, skipOrder bool) *ReqTracer {
	prev := tr.reqHead
	curr := prev.next
	if curr == nil {
		return nil
	}

	if curr.Req == req {
		prev.next = curr.next
		if prev.next == nil {
			prev.expectedFields = curr.expectedFields
			tr.reqTail = nil
		}

		curr.next = nil
		return curr
	}

	if !skipOrder {
		return nil
	}

	for {
		prev = curr
		curr = curr.next
		if curr == nil {
			return nil
		}

		if curr.Req == req {
			prev.next = curr.next
			if prev.next == nil {
				tr.reqTail = prev
			}

			curr.next = nil
			return curr
		}
	}
}

func (tr *TableRowTracer) getPrevExpectedFields() map[string]string {
	if tr.reqTail == nil {
		return tr.reqHead.expectedFields
	}

	return tr.reqTail.expectedFields
}

type TableShardIndexTracer struct {
	nTotalReq int
	shard     string
	rows      map[string]*TableRowTracer
	reqHead   *ReqTracer
	reqTail   *ReqTracer
}

func (si *TableShardIndexTracer) onRequest(reqTracer *ReqTracer) {
	if si.reqTail == nil {
		si.reqHead.next = reqTracer
		si.reqTail = reqTracer

	} else {
		si.reqTail.next = reqTracer
		si.reqTail = reqTracer
	}
}

func (si *TableShardIndexTracer) onReply(req interface{}, invalid bool) *ReqTracer {
	prev := si.reqHead
	curr := prev.next
	if curr == nil {
		return nil
	}

	if curr.Req == req {
		prev.next = curr.next
		if prev.next == nil {
			si.reqTail = nil
		}

		curr.next = nil
		return curr
	}

	if !invalid {
		return nil
	}

	for {
		prev = curr
		curr = curr.next
		if curr == nil {
			return nil
		}

		if curr.Req == req {
			prev.next = curr.next
			if prev.next == nil {
				si.reqTail = prev
			}

			curr.next = nil
			return curr
		}
	}
}

type TableTracer struct {
	sync.Mutex
	nTotalReq int
	rows      map[string]*TableRowTracer
	sis       map[string]*TableShardIndexTracer
}

func NewTableTracer() *TableTracer {
	return &TableTracer{
		nTotalReq: 0,
		rows:      make(map[string]*TableRowTracer),
		sis:       make(map[string]*TableShardIndexTracer),
	}
}

func (table *TableTracer) HasRow(schema *sql.TableSchema, primaryKeys []string) bool {
	key := sql.AssembleRowKey2(schema, primaryKeys)
	_, ok := table.rows[key]
	return ok
}

func (table *TableTracer) checkCondition(schema *sql.TableSchema, fields map[string]string, shard string, where string, params []string) bool {
	if fields == nil {
		return false
	}

	parser, err := sql.CreateParser(schema, shard, where, params)
	if err != nil {
		panic(err)
	}

	if parser == nil {
		return true
	}

	return parser.Check(sql.NewRowDataFromMap(schema, fields))
}

func (table *TableTracer) getRowByFields(schema *sql.TableSchema, fields map[string]string) *TableRowTracer {
	data := sql.NewRowDataFromMap(schema, fields)
	key := sql.GetRowKey(schema, data)
	if row, ok := table.rows[key]; ok {
		return row
	}

	panic(fields)
}

func (table *TableTracer) getRow(schema *sql.TableSchema, shardId int32, primaryKeys []string) *TableRowTracer {
	key := sql.AssembleRowKey2(schema, primaryKeys)
	if row, ok := table.rows[key]; ok {
		return row
	}

	row := &TableRowTracer{
		key:     key,
		shard:   primaryKeys[shardId],
		reqHead: &ReqTracer{},
	}

	table.rows[key] = row
	return row
}

func (table *TableTracer) getShardIndex(shard string, createIfNone bool) *TableShardIndexTracer {
	if si, ok := table.sis[shard]; ok {
		return si
	}

	if !createIfNone {
		return nil
	}

	si := &TableShardIndexTracer{
		shard:   shard,
		rows:    make(map[string]*TableRowTracer),
		reqHead: &ReqTracer{},
	}

	for k, row := range table.rows {
		if row.shard == si.shard {
			si.rows[k] = row
		}
	}

	table.sis[shard] = si
	return si
}

func (table *TableTracer) checkFields(actual []string, expected map[string]string) bool {
	nExpect := len(expected)
	if nExpect == 0 {
		return len(actual) == 0
	}

	mActual := slice2Map(actual)
	nActual := len(mActual)
	if nActual < nExpect {
		return false
	}

	for k, v1 := range expected {
		if k == "mtime" {
			continue
		}

		v2, ok := mActual[k]
		if !ok {
			return false
		}
		if v1 != v2 {
			return false
		}
	}

	return true
}

func slice2Map(data []string) map[string]string {
	n := len(data)
	ret := make(map[string]string, n>>1)

	for i := 0; i < n; i += 2 {
		ret[data[i]] = data[i+1]
	}

	return ret
}

func fillDefaultFields(schema *sql.TableSchema, prevFields map[string]string) map[string]string {
	nColumn := len(schema.Columns)
	if nColumn <= len(prevFields) {
		return prevFields
	}

	expectedFields := make(map[string]string, nColumn)
	for k, v := range prevFields {
		expectedFields[k] = v
	}
	for i := range schema.Columns {
		fs := &schema.Columns[i]
		if _, ok := prevFields[fs.Name]; !ok {
			expectedFields[fs.Name] = fs.DefaultValue
		}
	}

	return expectedFields
}
