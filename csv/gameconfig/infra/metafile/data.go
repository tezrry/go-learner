package metafile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"csv/gameconfig/infra/ctype"
	"go-learner/slice"
)

const VersionLen = 20
const Delim = '\n'
const FileName = "metadata.txt"

type TableData struct {
	ver   string
	ids   []string
	tid   ctype.TID
	m     map[string]ctype.RID
	f     *os.File
	idPos int
}

type TableGroupData struct {
	dir   string
	m     map[string]*TableData
	maxId ctype.TID
}

func LoadTableGroup(rootDir string) (*TableGroupData, error) {
	f, err := os.Open(rootDir)
	if err != nil {
		return nil, err
	}
	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	_ = f.Close()

	m := make(map[string]*TableData, len(fis))
	var mtd *TableData
	maxId := ctype.TID(0)
	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}

		mtd, err = LoadTable(rootDir, fi.Name())
		if err != nil {
			return nil, err
		}

		m[fi.Name()] = mtd
		if mtd.TableId() > maxId {
			maxId = mtd.TableId()
		}
	}

	return &TableGroupData{dir: rootDir, m: m, maxId: maxId}, nil
}

func createFile(rootDir, version, name string, id ctype.TID) (*TableData, error) {
	if len(version) != VersionLen {
		return nil, fmt.Errorf("invalid version length %d", len(version))
	}

	dir := filepath.Join(rootDir, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
			return nil, fmt.Errorf("failed to create directory: %w", mkErr)
		}
	}

	fn := filepath.Join(dir, FileName)
	if _, err := os.Stat(fn); !os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s already exists", fn)
	}

	f, err := os.Create(fn)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	var n int
	n, err = f.WriteString(fmt.Sprintf("%s%c%d%c", version, Delim, id, Delim))
	if err != nil {
		return nil, err
	}

	rtn := &TableData{ver: version, tid: id, f: f, idPos: n}
	if err = rtn.LoadData(); err != nil {
		return nil, err
	}

	return rtn, nil
}

func LoadTable(rootDir, name string) (*TableData, error) {
	fn := filepath.Join(rootDir, name, FileName)
	f, err := os.OpenFile(fn, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	bVer := make([]byte, VersionLen+1)
	_, err = f.Read(bVer)
	if err != nil {
		return nil, err
	}

	if bVer[VersionLen] != Delim {
		return nil, fmt.Errorf("invalid file version: %s", string(bVer))
	}

	bId := make([]byte, 0, 64)
	buf := []byte{0}
	for {
		_, err = f.Read(buf)
		if err != nil {
			return nil, err
		}

		if buf[0] == Delim {
			break
		}

		bId = append(bId, buf[0])
	}

	i, err := strconv.ParseInt(slice.ByteSlice2String(bId), 10, 64)
	if err != nil {
		return nil, err
	}

	if i < 1 || i > ctype.MaxTableId {
		return nil, fmt.Errorf("invalid table id: %d", i)
	}

	n := VersionLen + 1 + len(bId) + 1
	return &TableData{
		ver: slice.ByteSlice2String(bVer[:VersionLen]),
		tid: ctype.TID(i), f: f, idPos: n,
	}, nil
}

func (inst *TableData) LoadData() error {
	initSize := 2048
	inst.ids = make([]string, 0, initSize)
	inst.m = make(map[string]ctype.RID, initSize)

	_, err := inst.f.Seek(int64(inst.idPos), io.SeekStart)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(inst.f)
	rid := ctype.RID(0)
	for {
		line, err := reader.ReadString(Delim)
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		if len(line) < 2 {
			return fmt.Errorf("empty id in line %d", rid+2)
		}

		line = line[:len(line)-1]
		rid++
		if rid > ctype.MaxRowId {
			return fmt.Errorf("beyond max row id: %d", ctype.MaxRowId)
		}

		inst.ids = append(inst.ids, line)
		inst.m[line] = rid
	}
}

func (inst *TableData) SaveAndClose(version string) error {
	_, err := inst.f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = inst.f.WriteString(version[0:VersionLen])
	if err != nil {
		return err
	}

	inst.ver = version
	inst.Close()
	return nil
}

func (inst *TableData) AddId(idName string) error {
	if inst.m[idName] != ctype.InvalidID {
		return nil
	}

	_, err := inst.f.WriteString(idName + string(Delim))
	if err != nil {
		return err
	}

	inst.ids = append(inst.ids, idName)
	inst.m[idName] = ctype.RID(len(inst.ids))
	return nil
}

func (inst *TableData) GlobalID(idName string) ctype.ID {
	rid := inst.m[idName]
	if rid != ctype.InvalidID {
		return ctype.GlobalId(inst.tid, rid)
	}

	return ctype.InvalidID
}

func (inst *TableData) Version() string {
	return inst.ver
}

func (inst *TableData) TableId() ctype.TID {
	return inst.tid
}

func (inst *TableData) Close() {
	_ = inst.f.Close()
}

func (inst *TableGroupData) Table(name string) *TableData {
	return inst.m[name]
}

func (inst *TableGroupData) CreateTable(name, version string) (*TableData, error) {
	td := inst.m[name]
	if td != nil {
		return nil, fmt.Errorf("table %s already exists", name)
	}

	maxId := inst.maxId + 1
	td, err := createFile(inst.dir, version, name, maxId)
	if err != nil {
		return nil, err
	}

	inst.maxId = maxId
	inst.m[name] = td
	return td, nil
}
