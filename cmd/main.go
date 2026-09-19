package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/vpovarna/oncall-concierge/internal/agents/runbook"
	"github.com/vpovarna/oncall-concierge/internal/logger"
)

const defaultModel = "gpt-4o-mini"

func main() {
	_ = context.Background()

	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file loaded, using ambient environment: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")

	if apiKey == "" && baseURL == "" {
		log.Fatal("set OPENAI_API_KEY (for OpenAI) or OPENAI_BASE_URL (for an OpenAI-compatible endpoint)")
	}

	modelName := os.Getenv("OPENAI_MODEL")
	if modelName == "" {
		modelName = defaultModel
	}

	// Load runbooks
	logger := logger.New("info")
	runbookService := runbook.NewService(&logger)
	registry, err := runbookService.NewRegistryFromDir("runbooks")

	if err != nil {
		logger.Error().Err(err).Msg("Unable to load registry jos files")
	}

	fmt.Printf("Registry: %s", registry)

	

	// model, err := openaimodel.NewModel(ctx, modelName, &openaimodel.ClientConfig{
	// 	APIKey:  apiKey,
	// 	BaseURL: baseURL,
	// })

	// if err != nil {
	// 	log.Fatalf("Failed to create model: %v", err)
	// }

	// a, err := llmagent.New(llmagent.Config{
	// 	Name:        "oncall_concierge",
	// 	Model:       model,
	// 	Description: "Helps an on-call engineer reason about incidents and runbooks.",
	// 	Instruction: "You are an on-call assistant. Answer the engineer's questions clearly and concisely.",
	// })
	// if err != nil {
	// 	log.Fatalf("Failed to create agent: %v", err)
	// }

	// config := &launcher.Config{
	// 	AgentLoader: agent.NewSingleLoader(a),
	// }

	// l := full.NewLauncher()
	// if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
	// 	log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	// }
}
