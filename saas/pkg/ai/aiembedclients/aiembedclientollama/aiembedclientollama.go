package aiembedclientollama

import (
	"errors"
	"lace/httpreq"
	"saas/pkg/ai/aitypes"
	"saas/pkg/utils"

	"github.com/cohesivestack/valgo"
	"github.com/pandodao/tokenizer-go"
	"github.com/spf13/cast"
)

type OllamaEmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

type OllamaEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type OllamaClient struct {
	params OllamaClientParams
}

type OllamaClientParams struct {
	BaseURL string
	ModelID string
	ApiKey  string
	// Dims           int // truncate embedding to this dims
	MaxDims        int
	MaxInputTokens int
}

func New(params OllamaClientParams) (*OllamaClient, error) {
	val := valgo.Is(
		valgo.String(cast.ToString(params.BaseURL), "BaseURL").Not().Blank(),
		valgo.String(cast.ToString(params.ModelID), "ModelID").Not().Blank(),
		valgo.String(cast.ToString(params.ApiKey), "ApiKey").Not().Blank(),
		// valgo.Int(params.Dims, "Dims").Not().Zero(),
		valgo.Int(params.MaxDims, "MaxDims").Not().Zero(),
		valgo.Int(params.MaxInputTokens, "MaxInputTokens").Not().Zero(),
	)

	if !val.Valid() {
		return nil, utils.ValgoFirstError(val)
	}

	return &OllamaClient{params: params}, nil
}

func (c *OllamaClient) GetEmbeddings(values []string, dims int) ([]aitypes.ItemEmbeddings, error) {
	// Validate each value separately for token count
	for i, value := range values {
		t, _ := tokenizer.CalToken(value)
		if t > c.params.MaxInputTokens-100 {
			return nil, errors.New("max tokens exceeded for value at index " + cast.ToString(i))
		}
	}

	// dims := c.params.Dims
	if dims > c.params.MaxDims {
		return nil, errors.New("dims exceeded than max dims")
	}

	var apiResp OllamaEmbeddingResponse
	_, err := httpreq.Do(httpreq.RequestParam{
		EndPoint: c.params.BaseURL + "/api/embed",
		Method:   "POST",
		Args:     OllamaEmbeddingRequest{Model: c.params.ModelID, Input: values},
	}, &apiResp)

	if err != nil {
		return nil, err
	}

	embeddings := make([]aitypes.ItemEmbeddings, len(apiResp.Embeddings))
	for i, embedding := range apiResp.Embeddings {

		// truncate embedding to max dims
		if len(embedding) > dims {
			embedding = embedding[:dims]
		}

		embeddings[i] = aitypes.ItemEmbeddings{
			Content:   values[i],
			Embedding: embedding,
		}
	}
	return embeddings, nil
}
