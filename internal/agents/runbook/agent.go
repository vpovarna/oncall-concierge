package runbook

import (
	"fmt"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

func NewAgent(m model.LLM, registry *Registry) (agent.Agent, error) {

	getRunbook, err := functiontool.New(functiontool.Config{
		Name:        "get_runbook",
		Description: "Get runbook for a specific service",
	}, registry.GetRunbookTool())
	if err != nil {
		return nil, fmt.Errorf("failed to create get_runbook tool: %w", err)
	}

	listRunbooks, err := functiontool.New(functiontool.Config{
		Name:        "list_runbooks",
		Description: "List Runbooks",
	}, registry.ListRunbooksTool())
	if err != nil {
		return nil, fmt.Errorf("failed to create list_runbooks tool: %w", err)
	}

	return llmagent.New(llmagent.Config{
		Name:        "oncall_concierge",
		Model:       m,
		Description: "Helps an on-call engineer reason about incidents and runbooks.",
		Instruction: "You are an on-call assistant. Use list_runbooks to discover available services, and get_runbook to retrieve steps for a specific service.",
		Tools:       []tool.Tool{getRunbook, listRunbooks},
	})
}
