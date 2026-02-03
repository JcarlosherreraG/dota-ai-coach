package retriever

import (
	"context"
	"fmt"
	"github.com/nlpodyssey/cybertron/pkg/tasks"
	"github.com/nlpodyssey/cybertron/pkg/tasks/textencoding"
	"github.com/philippgille/chromem-go"
)

const modelsDir = "./storage"

func CreateEmbedFunction() (chromem.EmbeddingFunc, error) {
	conf := &tasks.Config{
		ModelName: textencoding.DefaultModel,
		ModelsDir: modelsDir,
	}

	model, err := tasks.Load[textencoding.Interface](conf)
	if err != nil {
		return nil, fmt.Errorf("cybertron load error: %w", err)
	}

	return func(ctx context.Context, text string) ([]float32, error) {
		result, err := model.Encode(ctx, text, 0)
		if err != nil {
			return nil, err
		}

		vec64 := result.Vector.Data().F64()

		vec32 := make([]float32, len(vec64))
		for i, v := range vec64 {
			vec32[i] = float32(v)
		}

		return vec32, nil
	}, nil
}
