package test

import (
	"fmt"
	"go-learner/slice"
	"go-learner/sql"
	"sync/atomic"

	"strconv"
)

type ServerInfo struct {
	ServerId int
	Host     string
	Port     int
	User     string
	Password string
	Timeout  int
}

type ITableInfo interface {
	GetName(idx int) string
	GetCount() int
	GetColumns() []string
	GetPrimaryKeys() []string
	GetFields() []FieldGenerator
}

type TableInfo struct {
	NamePrefix  string
	Count       int
	Columns     []string
	Fields      []FieldGenerator
	PrimaryKeys []string
}

func (inst *TableInfo) GetName(idx int) string {
	return fmt.Sprintf("%s%d", inst.NamePrefix, idx)
}

func (inst *TableInfo) GetCount() int {
	return inst.Count
}

func (inst *TableInfo) GetColumns() []string {
	return inst.Columns
}

func (inst *TableInfo) GetFields() []FieldGenerator {
	return inst.Fields
}

func (inst *TableInfo) GetPrimaryKeys() []string {
	return inst.PrimaryKeys
}

type SpecificTableInfo struct {
	TableInfo
	ColumnPrefix string
	NumColumn    int
	Field        FieldGenerator
}

func (si *SpecificTableInfo) GetColumns() []string {
	n := len(si.TableInfo.Columns)
	num := n + si.NumColumn
	columns := make([]string, num)
	for i := 0; i < n; i++ {
		columns[i] = si.TableInfo.Columns[i]
	}

	for i := n; i < num; i++ {
		columns[i] = fmt.Sprintf("%s_%d BLOB", si.ColumnPrefix, i-n)
	}

	return columns
}

func (si *SpecificTableInfo) GetFields() []FieldGenerator {
	n := len(si.TableInfo.Fields)
	num := n + si.NumColumn
	columns := make([]FieldGenerator, num)
	for i := 0; i < n; i++ {
		columns[i] = si.TableInfo.Fields[i]
	}

	for i := n; i < num; i++ {
		columns[i] = si.Field
	}

	return columns
}

type AlterSchemaDef struct {
	Num          atomic.Int64
	ColumnPrefix string
	MaxCount     int64
	Columns      []string
	Fields       []FieldGenerator
}

func (inst *AlterSchemaDef) NextColumn() (string, *FieldGenerator) {
	n := inst.Num.Add(1)
	if n > inst.MaxCount {
		return "", nil
	}

	idx := (n - 1) % int64(len(inst.Columns))
	return fmt.Sprintf("`%s%d` %s", inst.ColumnPrefix, n, inst.Columns[idx]), &inst.Fields[idx]
}

func GetDBProxyServerInfo(db *sql.MySql) []ServerInfo {
	rows, err := db.Query("SELECT server_id, host, port, user, passwd, timeout FROM server_info")
	if err != nil {
		panic(err)
	}

	ret := make([]ServerInfo, len(rows))
	for i := range ret {
		row := rows[i]
		ret[i].ServerId, err = strconv.Atoi(slice.ByteSlice2String(row[0]))
		if err != nil {
			panic(err)
		}
		ret[i].Host = slice.ByteSlice2String(row[1])
		ret[i].Port, err = strconv.Atoi(slice.ByteSlice2String(row[2]))
		if err != nil {
			panic(err)
		}
		ret[i].User = slice.ByteSlice2String(row[3])
		ret[i].Password = slice.ByteSlice2String(row[4])
		ret[i].Timeout, err = strconv.Atoi(slice.ByteSlice2String(row[5]))
		if err != nil {
			panic(err)
		}
	}

	return ret
}

func ResetDBProxy(db *sql.MySql, servers []ServerInfo) []ServerInfo {
	sqlStr := "TRUNCATE TABLE server_info"
	_, err := db.Exec(sqlStr)
	if err != nil {
		panic(err)
	}

	sqlStr = "INSERT INTO server_info (host, port, user, passwd, timeout) VALUES(?,?,?,?,?)"
	for _, srv := range servers {
		_, err := db.Exec(sqlStr, srv.Host, srv.Port, srv.User, srv.Password, srv.Timeout)
		if err != nil {
			panic(err)
		}
	}

	sqlStr = "TRUNCATE TABLE table_info"
	_, err = db.Exec(sqlStr)
	if err != nil {
		panic(err)
	}

	sqlStr = "TRUNCATE TABLE split_table_info"
	_, err = db.Exec(sqlStr)
	if err != nil {
		panic(err)
	}

	return GetDBProxyServerInfo(db)
}

func CreateDBProxyTableInfo(db *sql.MySql, servers []ServerInfo, dbPrefix string,
	dbShard int, tableShard int, tableInfo ITableInfo, realTableDBInfo map[string]map[string]*sql.MySql) {
	count := tableInfo.GetCount()
	if count == 0 {
		return
	}

	sqlStr := "INSERT INTO table_info(table_name,split_type,table_count,hint_field) VALUES(?,?,?,?)"
	sqlStr2 := "INSERT INTO split_table_info(table_name,table_number,server_id,database_name) VALUES(?,?,?,?)"

	nServer := len(servers)
	nShard := nServer * dbShard * tableShard
	tmp := make(map[string]struct{})

	for i := 0; i < count; i++ {
		tbName := tableInfo.GetName(i)
		_, err := db.Exec(sqlStr, tbName, 0, nShard, tableInfo.GetPrimaryKeys()[0])
		if err != nil {
			panic(err)
		}

		var dbList map[string]*sql.MySql
		if realTableDBInfo != nil {
			dbList = make(map[string]*sql.MySql, nShard)
			realTableDBInfo[tbName] = dbList
		}

		var srvIdx, dbIdx int
		for i := 0; i < nShard; i++ {
			srv := &servers[srvIdx%nServer]
			dbName := fmt.Sprintf("%s_%d", dbPrefix, dbIdx)

			_, err = db.Exec(sqlStr2, tbName, i, srv.ServerId, dbName)
			if err != nil {
				panic(err)
			}

			srvURL := fmt.Sprintf("%s:%s@tcp(%s:%d)/", srv.User, srv.Password, srv.Host, srv.Port)
			dbURL := srvURL + dbName
			if _, ok := tmp[dbURL]; !ok {
				dataDB := sql.GetMySql(srvURL)
				_, err = dataDB.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
				if err != nil {
					panic(err)
				}

				tmp[dbURL] = struct{}{}
			}

			tbRealName := fmt.Sprintf("%s_%d", tbName, i)
			dataDB := sql.GetMySql(dbURL)
			sql.CreateMySqlTestTable(dataDB, tbRealName, tableInfo.GetColumns(), tableInfo.GetPrimaryKeys())

			if dbList != nil {
				dbList[tbRealName] = dataDB
			}

			dbIdx++
			if dbIdx == dbShard {
				srvIdx++
				dbIdx = 0
			}
		}
	}
}

func AlterTableSchema(realTableDBInfo map[string]map[string]*sql.MySql,
	def *AlterSchemaDef, tg *TableReqGenerator) ([]FieldGenerator, error) {
	col, fg := def.NextColumn()
	if col == "" {
		return nil, nil
	}

	schema := tg.schema
	var afterIdx int
	if schema.PrimaryKeyIndexes != nil {
		afterIdx = schema.PrimaryKeyIndexes[schema.NumPrimaryKeys-1]

	} else {
		afterIdx = schema.NumPrimaryKeys - 1
	}
	afterName := schema.Columns[afterIdx].Name
	dbMap, ok := realTableDBInfo[schema.Name]
	if !ok {
		return nil, fmt.Errorf("not found table")
	}

	oldFields := tg.fieldGens
	fields := make([]FieldGenerator, 0, len(oldFields)+1)
	fields = append(fields, oldFields[:afterIdx+1]...)
	fields = append(fields, *fg)
	fields = append(fields, oldFields[afterIdx+1:]...)

	for realName, db := range dbMap {
		_, err := db.Exec(
			fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN %s AFTER `%s`", realName, col, afterName))
		if err != nil {
			return nil, err
		}
	}

	return fields, nil
}

func ClearDataDB(db *sql.MySql) {
	serverInfo := GetDBProxyServerInfo(db)
	cache := make(map[string]map[string]struct{})

	lastId := 0
	limit := 512
	sqlStr := "SELECT id, server_id, database_name FROM split_table_info where id > ? order by id asc limit ?"
	for {
		rows, err := db.Query(sqlStr, lastId, limit)
		if err != nil {
			panic(err)
		}

		for i := range rows {
			row := rows[i]
			lastId, err = strconv.Atoi(slice.ByteSlice2String(row[0]))
			if err != nil {
				panic(err)
			}

			serverId, err := strconv.Atoi(slice.ByteSlice2String(row[1]))
			if err != nil {
				panic(err)
			}

			dbName := slice.ByteSlice2String(row[2])
			for i := range serverInfo {
				if serverInfo[i].ServerId == serverId {
					srvURL := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
						serverInfo[i].User, serverInfo[i].Password,
						serverInfo[i].Host, serverInfo[i].Port)

					tmp, ok := cache[srvURL]
					if !ok {
						tmp = make(map[string]struct{})
						cache[srvURL] = tmp
					}
					tmp[dbName] = struct{}{}
				}
			}
		}

		println(lastId)
		if len(rows) < limit {
			break
		}
	}

	for url, m := range cache {
		db := sql.GetMySql(url)
		for dbName := range m {
			_, err := db.Exec("DROP DATABASE IF EXISTS " + dbName)
			if err != nil {
				panic(err)
			}
		}
	}
}
