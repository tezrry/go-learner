package cmd

import "go-learner/orm"

type TestKey struct {
	I64 int64
}

type testMetadata struct {
}

type TestRecord struct {
	orm.Header
	body TestPB
}

type TestPB struct {
	I64 int64
	Str string
}

func (inst *TestRecord) LoadByIndex(idx int, v []byte) {
	switch idx {
	case 0:
	}
}

func (inst *TestRecord) Set_i64(v int64) {
	inst.body.I64 = v
	inst.Header.SetDirty(0)
}

func (inst *TestRecord) Set_str(v string) {
	inst.body.Str = v
	inst.Header.SetDirty(1)
}

func (inst *TestRecord) Get_i64() int64 {
	return inst.body.I64
}

func (inst *TestRecord) Get_str() string {
	return inst.body.Str
}
