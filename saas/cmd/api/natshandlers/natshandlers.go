package natshandlers

import (
	"lace/gozap"
	"lace/natso"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/nats-io/nats.go/micro"
)

type Config struct {
	Natso natso.Natso
}

func SchemaFor(t any) string {
	schema := jsonschema.Reflect(t)
	data, _ := schema.MarshalJSON()
	return string(data)
}

func New(config Config) {
	nc := config.Natso.GetConn()

	if nc == nil {
		gozap.GetLogger().Sugar().Errorf("FATAL NATS CONNECTION - %s", "nats connection is nil")
		return
	}

	srv, err := micro.AddService(nc, micro.Config{
		Name:    "boilerplate",
		Version: "0.0.1",
	})

	if err != nil {
		gozap.GetLogger().Sugar().Fatalf("FATAL ADD NATS SERVICE - %s", err.Error())
	}

	m := srv.AddGroup("boilerplate")

	// CLI: nats req boilerplate.ping '{}'
	m.AddEndpoint(
		"ping",
		micro.HandlerFunc(func(req micro.Request) {
			time.Sleep(time.Second * 1)
			req.Respond([]byte("pong"))
		}),
		micro.WithEndpointMetadata(map[string]string{
			"format": "application/json",
		}),
	)

	// Add your NATS handlers here
}
