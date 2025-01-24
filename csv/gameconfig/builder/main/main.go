package main

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"

	"csv/gameconfig/builder/lexer"
	"csv/gameconfig/builder/reader"
	"csv/gameconfig/infra/metafile"
)

type Source struct {
	tbs map[string]*lexer.Table
}

func main() {
	var cfgName string
	flag.StringVar(&cfgName, "c", "builder/test/config/config.json", "")
	f, err := os.Open(cfgName)
	if err != nil {
		panic(err)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		panic(err)
	}
	_ = f.Close()

	var outputDir string
	flag.StringVar(&outputDir, "o", "builder/generated", "output directory")
	f, err = os.Open(outputDir)
	if err != nil {
		panic(err)
	}

	tbg, err := metafile.LoadTableGroup(outputDir)
	if err != nil {
		panic(err)
	}

	var srcDir string
	flag.StringVar(&srcDir, "s", "builder/test", "source directory")
	f, err = os.Open(srcDir)
	if err != nil {
		panic(err)
	}

	srcFiles, err := f.Readdir(-1)
	if err != nil {
		panic(err)
	}
	_ = f.Close()

	var src Source
	src.tbs = make(map[string]*lexer.Table, len(srcFiles))

	csv := new(reader.CSV)
	for _, fi := range srcFiles {
		if fi.IsDir() || filepath.Ext(fi.Name()) != csv.Suffix() {
			continue
		}

		name := strings.TrimSuffix(fi.Name(), csv.Suffix())
		ver := csv.Version(name)
		meta := tbg.Table(name)
		if meta == nil {
			meta, err = tbg.CreateTable(name, ver)
			if err != nil {
				panic(err)
			}
		} else {
			if ver == meta.Version() {
				meta.Close()
				continue
			}

			err = meta.LoadData()
			if err != nil {
				panic(err)
			}
		}

		rows, err := csv.Read(filepath.Join(srcDir, fi.Name()))
		if err != nil {
			panic(err)
		}

		tb, err := lexer.InitTable(name, rows)
		if err != nil {
			panic(err)
		}

		gen := lexer.NewGenerator(tb)
		gen.Save()
	}

}
