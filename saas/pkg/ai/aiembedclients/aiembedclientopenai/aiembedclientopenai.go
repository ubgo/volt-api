package aiembedclientopenai

import (
	"context"
	"errors"
	"saas/pkg/ai/aitypes"
	"saas/pkg/utils"

	"github.com/cohesivestack/valgo"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/pandodao/tokenizer-go"
	"github.com/spf13/cast"
)

type OpenAIClient struct {
	params ClientParams
	client *openai.Client
}

type ClientParams struct {
	BaseURL        string
	ModelID        string
	ApiKey         string
	MaxDims        int
	MaxInputTokens int
}

func New(params ClientParams) (*OpenAIClient, error) {
	val := valgo.Is(
		valgo.String(cast.ToString(params.BaseURL), "BaseURL").Not().Blank(),
		valgo.String(cast.ToString(params.ModelID), "ModelID").Not().Blank(),
		valgo.String(cast.ToString(params.ApiKey), "ApiKey").Not().Blank(),
		valgo.Int(params.MaxDims, "MaxDims").Not().Zero(),
		valgo.Int(params.MaxInputTokens, "MaxInputTokens").Not().Zero(),
	)

	if !val.Valid() {
		return nil, utils.ValgoFirstError(val)
	}

	client := openai.NewClient(
		option.WithBaseURL(params.BaseURL),
		option.WithAPIKey(params.ApiKey), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)

	return &OpenAIClient{params: params, client: &client}, nil
}

func (c *OpenAIClient) GetEmbeddings(values []string, dims int) ([]aitypes.ItemEmbeddings, error) {
	// Validate each value separately for token count
	for i, value := range values {
		t, _ := tokenizer.CalToken(value)
		if t > c.params.MaxInputTokens-100 {
			return nil, errors.New("max tokens exceeded for value at index " + cast.ToString(i))
		}
	}

	if dims > c.params.MaxDims {
		return nil, errors.New("dims exceeded than max dims")
	}

	clientai := c.client
	result, err := clientai.Embeddings.New(context.Background(), openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: values,
		},
		Model:          c.params.ModelID,
		Dimensions:     openai.Opt(int64(dims)),
		EncodingFormat: openai.EmbeddingNewParamsEncodingFormatFloat,
	})

	// fmt.Println(err)

	if err != nil {
		return nil, err
	}

	var itemEmbeddings []aitypes.ItemEmbeddings

	for i, d := range result.Data {
		var embedding []float32
		for _, e := range d.Embedding {
			embedding = append(embedding, float32(e))
		}

		// if len(embedding) > dims {
		// 	embedding = embedding[:dims]
		// }

		itemEmbeddings = append(itemEmbeddings, aitypes.ItemEmbeddings{Content: values[i], Embedding: embedding})
	}

	return itemEmbeddings, nil
}
