package assets

import _ "embed"

//go:embed rag/knowledge.json
var KnowledgeJSON []byte

//go:embed prompts/base-coach-system-prompt.txt
var BaseCoachSystemPrompt string

//go:embed prompts/bd-queries-generator.txt
var CreateSemanticQueriesPrompt string

//go:embed prompts/final-prompt.txt
var FinalUserQuestionPrompt string

//go:embed prompts/heroes-processing-prompt.txt
var HeroesProcessingPrompt string

//go:embed prompts/aganim-processing-prompt
var AganimProcessingPrompt string

//go:embed prompts/items-processing-prompt
var ItemsProcessingPrompt string

//go:embed prompts/auto-user-question
var AutoUserQuestion string
