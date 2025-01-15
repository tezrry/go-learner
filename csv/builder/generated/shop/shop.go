package shop

import (
	"go-learner/csv/builder/infra"
)

const TableId infra.ID = 1

type ItemType infra.ENUM

const (
	ItemType_Chest ItemType = iota
	ItemType_Equip
	ItemType_End_
)

var ItemTypeName = [ItemType_End_]string{
	"chest", "equip",
}

const (
	RowId_None infra.ID = iota
	RowId_infantry_health
	RowId_End_
)

var ptr = &Table{
	data: []Row{{},
		// infantry_health:1
		{_gid: 262145},
	},
}

type Table struct {
	data []Row
}

type Row struct {
	_gid  infra.ID
	_type ItemType
}

func init() {
	//generated.Register(TableId, ptr)
}

func RowByGlobalId(gid infra.ID) *Row {
	return &ptr.data[gid&infra.MaxRowId]
}

func RowById(rid infra.ID) *Row {
	return &ptr.data[rid]
}

func GlobalId(rid infra.ID) infra.ID {
	return infra.GlobalId(TableId, rid)
}

func (inst *Row) GlobalId() infra.ID {
	return inst._gid
}

func (inst *Row) TableId() infra.ID {
	return TableId
}

func (inst *Row) RowId() infra.ID {
	return inst._gid & infra.MaxRowId
}

func (inst *Row) Type() ItemType {
	return inst._type
}

//func GetTable() *Table {
//	//return (*Table)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&ptr))))
//	return ptr
//}
//
//func Reload(filePath string) error {
//	//atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&ptr)), unsafe.Pointer(config))
//	return nil
//}

func (inst *Table) TableId() infra.ID {
	return TableId
}

func (inst *Table) MD5() string {
	return "12345"
}

func (it ItemType) ToString() string {
	return ItemTypeName[it]
}
