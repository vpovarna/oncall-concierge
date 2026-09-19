package runbook

import (
	"fmt"
	"strings"

	"google.golang.org/adk/v2/agent"
)

func GetRunbook(_ agent.Context, in GetRunbookInput) (Runbook, error) {
	return Runbook{}, fmt.Errorf("not implemented: GetRunbook(%s)", strings.ToLower(in.Service))
}

func ListRunbooks(_ agent.Context, in struct{}) (ListRunbooksOutput, error) {
	return ListRunbooksOutput{}, fmt.Errorf("not implemented: ListRunbooks")
}
