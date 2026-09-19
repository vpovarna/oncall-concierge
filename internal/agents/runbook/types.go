package runbook

type Registry struct {
	runbooks map[string]Runbook
}

type Runbook struct {
	Service string   `json:"service"`
	Title   string   `json:"title"`
	Steps   []string `json:"steps"`
}

type GetRunbookInput struct {
	Service string `json:"service"`
}

type ListRunbooksOutput struct {
	Services []string `json:"services"`
}
