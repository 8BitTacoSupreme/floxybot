package claude

import (
	_ "embed"
	"strings"

	"github.com/8BitTacoSupreme/floxybot/internal/floxctx"
	"github.com/8BitTacoSupreme/floxybot/internal/rag"
)

//go:embed system_prompt.txt
var basePrompt string

// BuildSystemPrompt constructs the full system prompt with RAG context and Flox env context.
func BuildSystemPrompt(ragChunks []rag.RetrievedChunk, floxCtx *floxctx.Context) string {
	prompt := basePrompt

	contextBlock := ""
	if floxCtx != nil {
		contextBlock = floxCtx.ForPrompt()
	}
	prompt = strings.ReplaceAll(prompt, "{{CONTEXT}}", contextBlock)

	ragBlock := rag.FormatForPrompt(ragChunks)
	prompt = strings.ReplaceAll(prompt, "{{RAG_CONTEXT}}", ragBlock)

	return prompt
}
