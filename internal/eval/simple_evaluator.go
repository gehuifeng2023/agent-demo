package eval

import (
	"agent-demo/internal/model"
	"strings"
)

type SimpleEvaluator struct{}

func (e SimpleEvaluator) Evaluate(answer string, sources []model.Source) model.Quality {
	q := model.Quality{HasSources: len(sources) > 0, Score: 0.5}
	if q.HasSources {
		q.Score += 0.3
	} else {
		q.Warnings = append(q.Warnings, "没有引用来源，回答可能缺少知识库依据")
	}
	if strings.Contains(answer, "无法确定") {
		q.Score += 0.1
		q.Warnings = append(q.Warnings, "回答包含不确定性说明")
	}
	if len([]rune(answer)) < 20 {
		q.Score -= 0.2
		q.Warnings = append(q.Warnings, "回答过短")
	}
	if q.Score < 0 {
		q.Score = 0
	}
	if q.Score > 1 {
		q.Score = 1
	}
	return q
}
