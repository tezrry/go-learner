package sql

type DBAccessor struct {
	processor *Processor
	req       chan *DBRequest
	rtn       chan *DBRequest
}

func (dba *DBAccessor) Run() {
	go func() {
		for {
			dba.processRequest(dba.Empty())
		}
	}()
}

func (dba *DBAccessor) Driver() Driver {
	return dba.processor.Driver()
}

func (dba *DBAccessor) PendingReqNum() int32 {
	return dba.processor.PendingReqNum()
}

func (dba *DBAccessor) Empty() bool {
	return dba.processor.Empty()
}

func (dba *DBAccessor) getReq(wait bool) *DBRequest {
	if wait {
		return <-dba.req
	}

	select {
	case req := <-dba.req:
		return req

	default:
		return nil
	}
}

func (dba *DBAccessor) appendRequest(wait bool) (int32, bool) {
	processor := dba.processor
	req := dba.getReq(wait)
	if req == nil {
		return processor.PendingReqNum(), false
	}

	nAppending := processor.AppendRequest(req)
	if req.PreReq {
		req := dba.getReq(true)
		nAppending = processor.AppendRequest(req)
	}

	return nAppending, true
}

func (dba *DBAccessor) execute() {
	processor := dba.processor
	req := processor.Execute()
	for req != nil {
		dba.rtn <- req
		req = processor.Execute()
	}
}

func (dba *DBAccessor) processRequest(wait bool) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()

	nAppending, _ := dba.appendRequest(wait)
	maxN := int32(256)
	for nAppending < maxN {
		var ok bool
		if nAppending, ok = dba.appendRequest(false); !ok {
			break
		}
	}

	dba.execute()
}
