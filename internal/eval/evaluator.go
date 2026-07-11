package eval

import "agent-demo/internal/model"

type Evaluator interface {
	Evaluate(answer string, sources []model.Source) model.Quality
}
