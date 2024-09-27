package cmd

import (
	"fmt"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/gogo/protobuf/vanity/command"
	"golang.org/x/tools/imports"
)

type Plugin struct {
	*generator.Generator
	generator.PluginImports
	protoFiles   []*descriptor.FileDescriptorProto
	lockPriority map[string]map[string]int32
}

func NewPlugin(protoFiles []*descriptor.FileDescriptorProto) *Plugin {
	return &Plugin{
		protoFiles:   protoFiles,
		lockPriority: make(map[string]map[string]int32),
	}
}

func (p *Plugin) Name() string {
	return "custom"
}

func (p *Plugin) Init(g *generator.Generator) {
	p.Generator = g
}

func (p *Plugin) Generate(file *generator.FileDescriptor) {
	p.PluginImports = generator.NewPluginImports(p.Generator)

	for _, message := range file.Messages() {
		p.P("// message ", *message.Name)
	}
}

type grpcPlugin struct {
	*generator.Generator
	generator.PluginImports
}

func NewAgentPlugin() *grpcPlugin {
	return &grpcPlugin{}
}

func (p *grpcPlugin) Name() string {
	return "grpc_agent"
}

func (p *grpcPlugin) Init(g *generator.Generator) {
	p.Generator = g
}

func (p *grpcPlugin) Generate(file *generator.FileDescriptor) {
	p.PluginImports = generator.NewPluginImports(p.Generator)
	//contextPkg := p.NewImport("context")
	//grpcPkg := p.NewImport("google.golang.org/grpc")

	for _, service := range file.FileDescriptorProto.Service {
		methods := service.Method
		methodNames := make([]string, len(methods))

		for i, method := range methods {
			methodNames[i] = fmt.Sprintf("_Agent_%s_%s_Handler", *service.Name, *method.Name)

			inputType := *method.InputType
			p.Generator.RecordTypeUse(inputType)
		}

	}
}

func generate() {
	req := command.Read()
	p := NewPlugin(req.GetProtoFile())
	resp := command.GeneratePlugin(req, p, ".custom.go")
	for _, file := range resp.File {
		bs, err := imports.Process("", ([]byte)(*file.Content), nil)
		if err == nil {
			*file.Content = string(bs)
		}
	}
	command.Write(resp)
}
