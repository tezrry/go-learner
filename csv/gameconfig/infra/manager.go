package infra

import (
	"sync/atomic"

	"csv/gameconfig/infra/ctype"
)

var currentIdx atomic.Int64
var versions [2]*Manager

type IRow interface {
	GlobalId() ctype.ID
	TableId() ctype.ID
	RowId() ctype.ID
}

type ITable interface {
	TableId() ctype.ID
	MD5() string
	GetByRowId(rid ctype.ID) IRow
	Foreach(f func(row IRow))
}

type Manager struct {
	ver int64
	tbs []ITable
}

func Init(ver int64, tbs []ITable) {
	versions[0] = &Manager{ver, tbs}
}

func Reload(ver int64, tbs []ITable) {
	nextIdx := (currentIdx.Load() + 1) & 1
	versions[nextIdx].reload(ver, tbs)
	currentIdx.Store(nextIdx)
}

func Current() *Manager {
	return versions[currentIdx.Load()]
}

func (inst *Manager) reload(ver int64, tbs []ITable) {
}

func (inst *Manager) GetByGlobalId(gid ctype.ID) IRow {
	tb := inst.tbs[ctype.TableId(gid)]
	return tb.GetByRowId(ctype.RowId(gid))
}

func (inst *Manager) GetRow(tableId ctype.ID, rowId ctype.ID) IRow {
	return inst.GetTable(tableId).GetByRowId(rowId)
}

func (inst *Manager) GetTable(tableId ctype.ID) ITable {
	return inst.tbs[tableId]
}

func (inst *Manager) ForeachRow(tableId ctype.ID, f func(row IRow)) {
	inst.GetTable(tableId).Foreach(f)
}
