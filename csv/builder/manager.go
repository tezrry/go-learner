package builder

import (
	"sync/atomic"
)

var currentIdx atomic.Int64
var versions [2]*Manager

type IRow interface {
	GlobalId() ID
	TableId() ID
	RowId() ID
}

type ITable interface {
	TableId() ID
	MD5() string
	GetByRowId(rid ID) IRow
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

func GlobalId(tableId, rowId ID) ID {
	return (tableId << rowIdOffset) | rowId
}

func TableId(gid ID) ID {
	return (gid & tableMask) >> rowIdOffset
}

func RowId(gid ID) ID {
	return gid & MaxRowId
}

func Current() *Manager {
	return versions[currentIdx.Load()]
}

func (inst *Manager) reload(ver int64, tbs []ITable) {
}

func (inst *Manager) GetByGlobalId(gid ID) IRow {
	tb := inst.tbs[TableId(gid)]
	return tb.GetByRowId(RowId(gid))
}

func (inst *Manager) GetRow(tableId ID, rowId ID) IRow {
	return inst.GetTable(tableId).GetByRowId(rowId)
}

func (inst *Manager) GetTable(tableId ID) ITable {
	return inst.tbs[tableId]
}

func (inst *Manager) ForeachRow(tableId ID, f func(row IRow)) {
	inst.GetTable(tableId).Foreach(f)
}
