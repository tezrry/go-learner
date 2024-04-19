package sql

import (
	"fmt"
)

type DBCommand uint8

const (
	CmdNone DBCommand = iota
	CmdInsert
	CmdDeleteSingle
	CmdUpdateSingle
	CmdIncrBySingle
	CmdSelectSingle
	CmdMultiStart
	CmdSelectMulti
	CmdDeleteMulti
	CmdCountMulti
)

type mergeReqDataFunc func(prev *DBRequest, curr *DBRequest) bool

var mergeRedDataFuncList []mergeReqDataFunc

type Processor struct {
	driver       Driver
	nReq         int32
	nMerged      int32
	reqDummyHead *DBRequest
	reqTail      *DBRequest
}

type DBRequest struct {
	Command    DBCommand
	CanMerge   bool
	Processed  bool
	Sync       bool
	PreReq     bool
	Schema     *TableSchema
	ShardId    int32
	Keys       []string
	Data       interface{}
	RowContext *RowContext
	Custom     interface{}
	Reply      *DBReply

	next    *DBRequest
	merged  *dbMergedRequest
	brother *DBRequest
}

func (req *DBRequest) Reset() {
	req.Command = CmdNone
	req.CanMerge = false
	req.Processed = false
	req.Sync = false
	req.PreReq = false
	req.ShardId = 0
	req.Keys = nil
	req.Data = nil
	req.Schema = nil
	req.Reply = nil
	req.Custom = nil
	req.next = nil
	req.merged = nil
	req.brother = nil
	req.RowContext = nil
}

type DBReply struct {
	Data interface{}
	Err  error
	Msg  string
}

type dbMergedRequest struct {
	data interface{}
	//next *dbMergedRequest
}

type RowContext struct {
	lastReq *DBRequest
}

func init() {
	mergeRedDataFuncList = []mergeReqDataFunc{
		CmdNone:         mergeDefault,
		CmdInsert:       mergeDefault,
		CmdDeleteSingle: mergeDefault,
		CmdUpdateSingle: mergeUpdateSingle,
		CmdIncrBySingle: mergeIncrBySingle,
		CmdSelectSingle: mergeDefault,
		CmdMultiStart:   mergeDefault,
		CmdSelectMulti:  mergeDefault,
		CmdDeleteMulti:  mergeDefault,
		CmdCountMulti:   mergeDefault,
	}
}

func NewRowContext() *RowContext {
	return &RowContext{}
}

func (rc *RowContext) Reset() {
	rc.lastReq = nil
}

func (p *Processor) PendingReqNum() int32 {
	return p.nMerged
}

func (p *Processor) Empty() bool {
	return p.nReq == 0
}

func (p *Processor) Driver() Driver {
	return p.driver
}

func (p *Processor) AppendRequest(req *DBRequest) int32 {
	if req == nil {
		return p.nMerged
	}

	if p.reqTail == nil {
		p.reqDummyHead.next = req
		p.reqTail = req

	} else {
		p.reqTail.next = req
		p.reqTail = req
	}

	var prev *DBRequest
	reqCtx := req.RowContext
	if reqCtx == nil {
		//SHOULD NOT happen
		prev = nil

	} else {
		prev = reqCtx.lastReq
		reqCtx.lastReq = req
	}

	p.nReq++
	p.mergeRequest(prev, req)

	return p.nMerged
}

func (p *Processor) Execute() *DBRequest {
	head := p.reqDummyHead
	driver := p.driver

	curr := head.next
	if curr == nil {
		return nil
	}

	if curr.Reply == nil {
		data := curr.Data
		if curr.merged != nil {
			data = curr.merged.data
		}

		schema := curr.Schema
		shard := curr.Keys[curr.ShardId]
		var reply *DBReply
		switch curr.Command {
		case CmdSelectSingle:
			curr.Reply = driver.SelectSingle(schema, shard, curr.Keys)
			if curr.Reply.Err == nil {
				reply = curr.Reply
			}

		case CmdSelectMulti, CmdCountMulti:
			curr.Reply = driver.SelectMulti(schema, shard)
			if curr.Reply.Err == nil {
				reply = curr.Reply
			}

		case CmdIncrBySingle:
			curr.Reply = driver.IncrBySingle(schema, shard, curr.Keys, data.(*IncrByData))
			if curr.Reply.Err == nil {
				reply = curr.Reply
			}

		case CmdUpdateSingle:
			curr.Reply = driver.UpdateSingle(schema, shard, curr.Keys, data.([]string))
			if curr.Reply.Err == nil {
				reply = curr.Reply
			}

		case CmdInsert:
			curr.Reply = driver.Insert(schema, shard, data.([]string))
			if curr.Reply.Err == nil {
				reply = &DBReply{Data: int64(0), Msg: "duplicate key"}
			}

		case CmdDeleteSingle:
			curr.Reply = driver.DeleteSingle(schema, shard, curr.Keys)
			if curr.Reply.Err == nil {
				reply = &DBReply{Data: int64(0)}
			}

		case CmdDeleteMulti:
			curr.Reply = driver.DeleteMulti(schema, shard, data.(*MultiRequestData))
			if curr.Reply.Err == nil {
				reply = curr.Reply
			}

		case CmdNone:
			curr.Reply = &DBReply{}
			reply = curr.Reply

		default:
			curr.Reply = &DBReply{Err: fmt.Errorf("nonsupport cmd: %d", curr.Command)}
		}

		if reply == nil {
			//前置请求失败，后续请求同时失败
			if curr.PreReq && curr.next != nil {
				reply = curr.Reply
				curr.next.Reply = reply

				p.nMerged--
				req := curr.next.brother
				for req != nil {
					req.Reply = reply
					req.merged = nil

					tmp := req.brother
					req.brother = nil
					req = tmp
				}
			}

			if curr.Sync || curr.PreReq {
				reply = curr.Reply

			} else {
				//重试
				curr.Reply = nil
				return nil
			}
		}

		p.nMerged--
		req := curr.brother
		for req != nil {
			req.Reply = reply
			req.merged = nil

			tmp := req.brother
			req.brother = nil
			req = tmp
		}
	}

	head.next = curr.next
	if head.next == nil {
		p.reqTail = nil
	}

	curr.next = nil
	curr.merged = nil
	curr.brother = nil

	p.nReq--
	return curr
}

func (p *Processor) mergeRequest(prev *DBRequest, req *DBRequest) {
	if prev == nil || prev.Reply != nil || !req.CanMerge ||
		prev.Command != req.Command || prev.Sync != req.Sync {

		p.nMerged++
		return
	}

	if prev.merged == nil {
		prev.merged = &dbMergedRequest{}
	}

	f := mergeRedDataFuncList[req.Command]
	if !f(prev, req) {
		p.nMerged++
		return
	}

	req.merged = prev.merged
	prev.brother = req
}

func mergeDefault(prev *DBRequest, curr *DBRequest) bool {
	if prev.merged.data == nil {
		prev.merged.data = prev.Data
	}

	curr.Processed = true
	return true
}

func mergeUpdateSingle(prev *DBRequest, curr *DBRequest) bool {
	prevMerged := prev.merged
	if prevMerged.data == nil {
		prevData := prev.Data.([]string)
		mergedData := make([]string, len(prevData))
		for i, v := range prevData {
			mergedData[i] = v
		}
		prevMerged.data = mergedData
	}

	mergedData := prevMerged.data.([]string)
	nMerged := len(mergedData)
	currData := curr.Data.([]string)
	nCurr := len(currData)

	for i := 0; i < nCurr; i += 2 {
		exist := false
		name := currData[i]
		value := currData[i+1]

		for j := 0; j < nMerged; j += 2 {
			if name == mergedData[j] {
				mergedData[j+1] = value
				exist = true
				break
			}
		}

		if !exist {
			mergedData = append(mergedData, name)
			mergedData = append(mergedData, value)
		}
	}

	prevMerged.data = mergedData
	return true
}

func mergeIncrBySingle(prev *DBRequest, curr *DBRequest) bool {
	prevMerged := prev.merged
	if prevMerged.data == nil {
		prevData := prev.Data.(*IncrByData)
		prevMerged.data = &IncrByData{
			Column: prevData.Column,
			Delta:  prevData.Delta,
			Where:  prevData.Where,
		}
	}

	currData := curr.Data.(*IncrByData)
	prevData := prevMerged.data.(*IncrByData)
	if prevData.Column != currData.Column || currData.Where != "" || prevData.Where != "" {
		return false
	}

	prevData.Delta += currData.Delta
	return true
}
