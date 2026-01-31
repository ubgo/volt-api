package aitypes

type AIModelProvider string

const (
	AIModelProviderOpenAI AIModelProvider = "openai"
	AIModelProviderOllama AIModelProvider = "ollama"
)

type ItemEmbeddings struct {
	Content   string
	Embedding []float32
}

type ItemEmbeddingClient interface {
	GetEmbeddings(values []string, dims int) ([]ItemEmbeddings, error)
}
