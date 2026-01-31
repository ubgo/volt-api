package aiembedding

import (
	"saas/pkg/ai/aiembedclients/aiembedclientollama"
	"saas/pkg/ai/aiembedclients/aiembedclientopenai"
	"saas/pkg/ai/aitypes"
	"saas/pkg/utils"

	"github.com/cohesivestack/valgo"
	"github.com/spf13/cast"
)

type EmbeddingProviderParams struct {
	Provider aitypes.AIModelProvider
	ModelID  string
	ApiKey   string
	BaseURL  string
	// Dims           int
	MaxDims        int
	MaxInputTokens int
}

type EmbeddingProvider struct {
	client aitypes.ItemEmbeddingClient
}

func NewEmbeddingProvider(params EmbeddingProviderParams) (*EmbeddingProvider, error) {
	val := valgo.Is(
		valgo.String(cast.ToString(params.Provider), "Provider").Not().Blank(),
		valgo.String(cast.ToString(params.ModelID), "ModelID").Not().Blank(),
		valgo.String(cast.ToString(params.ApiKey), "ApiKey").Not().Blank(),
		valgo.String(cast.ToString(params.BaseURL), "BaseURL").Not().Blank(),
		valgo.Int(params.MaxDims, "MaxDims").Not().Zero(),
		// valgo.Int(params.Dims, "Dims").Not().Zero().LessOrEqualTo(params.MaxDims),
		valgo.Int(params.MaxInputTokens, "MaxInputTokens").Not().Zero(),
	)

	if !val.Valid() {
		return nil, utils.ValgoFirstError(val)
	}

	var client aitypes.ItemEmbeddingClient
	var err error

	switch params.Provider {
	case aitypes.AIModelProviderOpenAI:
		client, err = aiembedclientopenai.New(aiembedclientopenai.ClientParams{
			BaseURL:        params.BaseURL,
			ModelID:        params.ModelID,
			ApiKey:         params.ApiKey,
			MaxDims:        params.MaxDims,
			MaxInputTokens: params.MaxInputTokens,
		})
		if err != nil {
			return nil, err
		}
	case aitypes.AIModelProviderOllama:
		client, err = aiembedclientollama.New(aiembedclientollama.OllamaClientParams{
			BaseURL:        params.BaseURL,
			ModelID:        params.ModelID,
			ApiKey:         params.ApiKey,
			MaxDims:        params.MaxDims,
			MaxInputTokens: params.MaxInputTokens,
		})
		if err != nil {
			return nil, err
		}
	}
	return &EmbeddingProvider{client: client}, nil
}

func (cls *EmbeddingProvider) GetEmbeddingClient() aitypes.ItemEmbeddingClient {
	return cls.client
}
