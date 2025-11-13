package adk

import "context"

type Agent interface {
	Name() string
	Execute(ctx context.Context, input string) (string, error)
	RegisterTools(registry ToolRegistry)
}

type ToolRegistry interface {
	RegisterTool(name string, tool Tool)
	GetTool(name string) (Tool, bool)
}

type Tool interface {
	Name() string
	Execute(ctx context.Context, input map[string]interface{}) (interface{}, error)
}

type BaseAgent struct {
	name  string
	tools map[string]Tool
}

func NewBaseAgent(name string) *BaseAgent {
	return &BaseAgent{
		name:  name,
		tools: make(map[string]Tool),
	}
}

func (a *BaseAgent) Name() string {
	return a.name
}

func (a *BaseAgent) RegisterTools(registry ToolRegistry) {
	for name, tool := range a.tools {
		registry.RegisterTool(name, tool)
	}
}

func (a *BaseAgent) AddTool(tool Tool) {
	a.tools[tool.Name()] = tool
}

type SimpleToolRegistry struct {
	tools map[string]Tool
}

func NewToolRegistry() *SimpleToolRegistry {
	return &SimpleToolRegistry{
		tools: make(map[string]Tool),
	}
}

func (r *SimpleToolRegistry) RegisterTool(name string, tool Tool) {
	r.tools[name] = tool
}

func (r *SimpleToolRegistry) GetTool(name string) (Tool, bool) {
	tool, exists := r.tools[name]
	return tool, exists
}