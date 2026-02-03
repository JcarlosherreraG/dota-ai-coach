package prompt

import (
	"github.com/BrightGir/game-ai-helper/assets"
	"text/template"
)

type Templates struct {
	QueryGenerator *template.Template
	FinalPrompt    *template.Template
}

func LoadTemplates() *Templates {
	return &Templates{
		QueryGenerator: template.Must(template.New("queryGen").Parse(assets.CreateSemanticQueriesPrompt)),
		FinalPrompt:    template.Must(template.New("finalPrompt").Parse(assets.FinalUserQuestionPrompt)),
	}
}
