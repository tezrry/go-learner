package driver

import "go-learner/orm"

type DataServer struct {
}

func (inst *DataServer) Get(rt *orm.RecordType, keys ...any) [][]byte {
	rtn := make([][]byte, len(keys))
	return rtn
}
