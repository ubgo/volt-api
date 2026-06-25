package pubsublisten

import (
	"context"
	"dbent/gen/ent"
	"encoding/json"
	"saas/pkg/pubsub"

	"github.com/ThreeDotsLabs/watermill/message"
)

// Message represents a generic pub/sub message
type Message struct {
	ActionType string          `json:"action_type"`
	Payload    json.RawMessage `json:"payload"`
}

// MessageHandler is a function that handles a message
type MessageHandler func(client *ent.Client) func(payload []byte) error

// messageHandlers maps action types to their handlers
var messageHandlers = make(map[string]MessageHandler)

// RegisterHandler registers a new message handler for a specific action type
func RegisterHandler(actionType string, handler MessageHandler) {
	messageHandlers[actionType] = handler
}

// SetupSubscribers sets up all the subscribers for the application
func SetupSubscribers(client *ent.Client) error {
	// Subscribe to all topics
	messages, err := pubsub.GetPublisher().Subscribe(context.Background(), "all")
	if err != nil {
		return err
	}

	// Start processing messages in a goroutine
	go func(messages <-chan *message.Message) {
		for msg := range messages {
			var message Message
			if err := json.Unmarshal(msg.Payload, &message); err != nil {
				msg.Nack()
				continue
			}

			handler, exists := messageHandlers[message.ActionType]
			if !exists {
				msg.Nack()
				continue
			}

			// Pass only the inner payload to the handler
			if err := handler(client)(message.Payload); err != nil {
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}(messages)

	return nil
}

// init registers all message handlers
func init() {
	// Register handlers here
	// Example: RegisterHandler("user.created", HandleUserCreated)
}
