// Code generated from Pkl module `RootConfig`. DO NOT EDIT.
package rootconfig

import (
	"context"

	"config/gen/appconfig"
	"config/gen/coreconfig"
	"config/gen/databaseconfig"
	"config/gen/graphqlconfig"
	"config/gen/loggerconfig"
	"config/gen/natsconfig"
	"config/gen/redisconfig"
	"config/gen/serverconfig"
	"github.com/apple/pkl-go/pkl"
)

type RootConfig struct {
	Server serverconfig.Setting `pkl:"server"`

	Core coreconfig.Setting `pkl:"core"`

	Database databaseconfig.Setting `pkl:"database"`

	Nats natsconfig.Setting `pkl:"nats"`

	Graphql graphqlconfig.Setting `pkl:"graphql"`

	Otel loggerconfig.Otel `pkl:"otel"`

	Logger loggerconfig.Setting `pkl:"logger"`

	Redis redisconfig.Setting `pkl:"redis"`

	App appconfig.Setting `pkl:"app"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a RootConfig
func LoadFromPath(ctx context.Context, path string) (ret RootConfig, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return ret, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a RootConfig
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (RootConfig, error) {
	var ret RootConfig
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
