package runbook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

type Service struct {
	logger *zerolog.Logger
}

func NewService(logger *zerolog.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) NewRegistryFromDir(dirName string) (*Registry, error) {
	entries, err := os.ReadDir(dirName)

	if err != nil {
		return nil, fmt.Errorf("Unable to read content from: %s", dirName)
	}

	var registry Registry
	registry.runbooks = make(map[string]Runbook)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		serviceName := entry.Name()
		data, err := os.ReadFile(filepath.Join(dirName, serviceName))
		if err != nil {
			s.logger.Error().Msg("Unable to read data the file")
			continue
		}

		runbooks, err := newRunbook(data)
		if err != nil {
			s.logger.Error().Msg("Unable to unmarshal runbook")
			continue
		}

		for service, runbook := range *runbooks {
			_, ok := registry.runbooks[service]
			if ok {
				continue
			}

			registry.runbooks[service] = runbook
		}
	}

	return &registry, nil
}

func newRunbook(data []byte) (*map[string]Runbook, error) {

	var runbooks []Runbook
	if err := json.Unmarshal(data, &runbooks); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal runbook data. Error: %w", err)
	}

	runbooksMap := make(map[string]Runbook, len(runbooks))
	for _, runbook := range runbooks {
		service := strings.ToLower(runbook.Service)
		runbooksMap[service] = runbook
	}

	return &runbooksMap, nil
}
