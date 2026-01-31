// Code generated from Pkl module `Nats`. DO NOT EDIT.
package natsconfig

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Nats struct {
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Nats
func LoadFromPath(ctx context.Context, path string) (ret Nats, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Nats
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Nats, error) {
	var ret Nats
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
