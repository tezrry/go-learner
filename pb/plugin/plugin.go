package plugin

import (
	"fmt"
	"io"
	"os"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	plugin_go "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/gogo/protobuf/vanity/command"
)

type Generator interface {
	Generate(p *Plugin, file *generator.FileDescriptor)
}

type Plugin struct {
	*generator.Generator
	generator.PluginImports
	name string
	g    Generator
	req  plugin_go.CodeGeneratorRequest
}

func NewPlugin(name string, g Generator) *Plugin {
	return &Plugin{
		name: name,
		g:    g,
	}
}

func (inst *Plugin) Name() string {
	return inst.name
}

func (inst *Plugin) Init(g *generator.Generator) {
	inst.Generator = g
	inst.PluginImports = generator.NewPluginImports(g)
}

func (inst *Plugin) Generate(file *generator.FileDescriptor) {
	inst.g.Generate(inst, file)
}

func (inst *Plugin) Read(r io.Reader) error {

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	var req generator.FileDescriptor
	req.FileDescriptorProto = new(descriptor.FileDescriptorProto)
	err = proto.Unmarshal(b, &req)
	if err != nil {
		return err
	}

	if len(inst.req.FileToGenerate) == 0 {
		return fmt.Errorf("no files to generate")
	}

	return nil
}

func (inst *Plugin) DryRun(fp string) error {
	file, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	if err = inst.Read(file); err != nil {
		return err
	}

	resp := command.GeneratePlugin(&inst.req, inst, fmt.Sprintf(".%s.go", inst.name))
	for _, f := range resp.File {
		os.Stdout.Write(([]byte)(*f.Content))
		//bs, err := imports.Process("", ([]byte)(*f.Content), nil)
		//if err != nil {
		//	return err
		//}
		//*f.Content = string(bs)
	}

	return nil
}
