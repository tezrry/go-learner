package user

import "go-learner/csv/gameconfig"

const TableId gameconfig.ID = 1

const (
	RowId_None gameconfig.ID = iota
	RowId_infantry_health
	RowId_infantry_attack
	RowId_infantry_defend
	RowId_cavalry_health
	RowId_cavalry_attack
	RowId_cavalry_defend
	RowId_ranged_health
	RowId_ranged_attack
	RowId_ranged_defend
	RowId_siege_health
	RowId_siege_attack
	RowId_siege_defend
	RowId_army_health
	RowId_army_attack
	RowId_army_defend
	RowId_infantry_incr_dmg
	RowId_cavalry_incr_dmg
	RowId_ranged_incr_dmg
	RowId_siege_incr_dmg
	RowId_army_incr_dmg
	RowId_infantry_decr_dmg
	RowId_cavalry_decr_dmg
	RowId_ranged_decr_dmg
	RowId_siege_decr_dmg
	RowId_army_decr_dmg
	RowId_attacker_incr_dmg
	RowId_attacker_decr_dmg
	RowId_defender_incr_dmg
	RowId_defender_decr_dmg
	RowId_hero_skill_incr_dmg
	RowId_hero_skill_decr_dmg
	RowId_End_
)

var ptr = &Table{
	data: []Row{{},
		// infantry_health:1
		{gid: 262145},
		// infantry_attack:2
		{gid: 262146},
		// infantry_defend:3
		{gid: 262147},
		// cavalry_health:4
		{gid: 262148},
		// cavalry_attack:5
		{gid: 262149},
		// cavalry_defend:6
		{gid: 262150},
		// ranged_health:7
		{gid: 262151},
		// ranged_attack:8
		{gid: 262152},
		// ranged_defend:9
		{gid: 262153},
		// siege_health:10
		{gid: 262154},
		// siege_attack:11
		{gid: 262155},
		// siege_defend:12
		{gid: 262156},
		// army_health:13
		{gid: 262157},
		// army_attack:14
		{gid: 262158},
		// army_defend:15
		{gid: 262159},
		// infantry_incr_dmg:16
		{gid: 262160},
		// cavalry_incr_dmg:17
		{gid: 262161},
		// ranged_incr_dmg:18
		{gid: 262162},
		// siege_incr_dmg:19
		{gid: 262163},
		// army_incr_dmg:20
		{gid: 262164},
		// infantry_decr_dmg:21
		{gid: 262165},
		// cavalry_decr_dmg:22
		{gid: 262166},
		// ranged_decr_dmg:23
		{gid: 262167},
		// siege_decr_dmg:24
		{gid: 262168},
		// army_decr_dmg:25
		{gid: 262169},
		// attacker_incr_dmg:26
		{gid: 262170},
		// attacker_decr_dmg:27
		{gid: 262171},
		// defender_incr_dmg:28
		{gid: 262172},
		// defender_decr_dmg:29
		{gid: 262173},
		// hero_skill_incr_dmg:30
		{gid: 262174},
		// hero_skill_decr_dmg:32
		{gid: 262175},
	},
}

type Table struct {
	data []Row
}

type Row struct {
	gid gameconfig.ID
}

func init() {
	generated.Register(TableId, ptr)
}

func GlobalId(rid gameconfig.ID) gameconfig.ID {
	return gameconfig.GlobalId(TableId, rid)
}

func (inst *Row) GlobalId() gameconfig.ID {
	return inst.gid
}

func (inst *Row) TableId() gameconfig.ID {
	return TableId
}

func (inst *Row) RowId() gameconfig.ID {
	return inst.gid & gameconfig.MaxRowId
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

func (inst *Table) TableId() gameconfig.ID {
	return TableId
}

func (inst *Table) GetByRowId(rid gameconfig.ID) gameconfig.IRow {
	return &inst.data[rid]
}

func (inst *Table) MD5() string {
	return "12345"
}

func (inst *Table) Foreach(f func(row gameconfig.IRow)) {
	for i := range inst.data {
		f(&inst.data[i])
	}
}
