package runbook

import (
	"fmt"
	"strings"

	"google.golang.org/adk/v2/agent"
)

func (r *Registry) GetRunbookTool() func(agent.Context, GetRunbookInput) (Runbook, error) {
	return func(_ agent.Context, in GetRunbookInput) (Runbook, error) {
		service := strings.ToLower(in.Service)
		rb, ok := r.runbooks[service]
		if !ok {
			return Runbook{}, fmt.Errorf("no runbook found for service: %s", service)
		}

		return rb, nil

	}
}

func (r *Registry) ListRunbooksTool() func(agent.Context, struct{}) (ListRunbooksOutput, error) {
	return func(ctx agent.Context, s struct{}) (ListRunbooksOutput, error) {
		services := make([]string, 0, len(r.runbooks))
		for s := range r.runbooks {
			services = append(services, s)
		}
		return ListRunbooksOutput{Services: services}, nil
	}
}
