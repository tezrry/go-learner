package infra

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
)

const FileVersionLen = 20
const Delim = '\n'

type Metadata struct {
	Version string
	IdName  []string
	TableId ID
	m       map[string]ID
	f       *os.File
	dirty   bool
}

func initMetadata(filePath string) (*Metadata, error) {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			f, err2 := os.Create(filePath)
			if err2 != nil {
				return nil, fmt.Errorf("failed to create file: %w", err2)
			}

			return &Metadata{f: f, dirty: true}, nil
		}

		return nil, fmt.Errorf("checking file: %w", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(f)
	version, err := reader.ReadString(Delim)
	if err != nil {
		return nil, err
	}

	if len(version) != FileVersionLen {
		return nil, fmt.Errorf("invalid file version length: %d", len(version))
	}

	s, err := reader.ReadString(Delim)
	if err != nil {
		return nil, err
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, err
	}

	if i < 0 || i > MaxTableId {
		return nil, fmt.Errorf("invalid table id: %d", i)
	}

	return &Metadata{Version: version, TableId: ID(i), f: f}, nil
}

func (inst *Metadata) LoadData() error {
	initSize := 2048
	inst.IdName = make([]string, 0, initSize)
	inst.m = make(map[string]ID, initSize)

	reader := bufio.NewReader(inst.f)
	rid := ID(0)
	for {
		line, err := reader.ReadString(Delim)
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		if line == "" {
			return fmt.Errorf("empty id in line %d", rid+2)
		}

		rid++
		if rid > MaxRowId {
			return fmt.Errorf("beyond max row id: %d", MaxRowId)
		}

		inst.IdName = append(inst.IdName, line)
		inst.m[line] = rid
	}
}

func (inst *Metadata) SaveIdName(name string) error {
	if inst.m[name] != InvalidID {
		return nil
	}

	_, err := inst.f.WriteString(name)
	if err != nil {
		return err
	}

	inst.IdName = append(inst.IdName, name)
	inst.m[name] = ID(len(inst.IdName))
	return nil
}

func (inst *Metadata) NameToGlobalID(name string) ID {
	rid := inst.m[name]
	if rid != InvalidID {
		return GlobalId(inst.TableId, rid)

	}

	return InvalidID
}

func (inst *Metadata) Close() {
	_ = inst.f.Close()
}
