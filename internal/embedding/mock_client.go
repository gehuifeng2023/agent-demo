package embedding

import (
	"context"
	"math"
)

type MockClient struct {
	Dim int
}

func (c MockClient) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	dim := c.Dim
	if dim <= 0 {
		dim = 8
	}
	out := make([][]float64, len(texts))
	for i, text := range texts {
		v := make([]float64, dim)
		for _, r := range text {
			v[int(r)%dim] += 1
		}
		out[i] = normalize(v)
	}
	return out, nil
}

func normalize(v []float64) []float64 {
	var sum float64
	for _, x := range v {
		sum += x * x
	}
	if sum == 0 {
		return v
	}
	n := math.Sqrt(sum)
	for i := range v {
		v[i] /= n
	}
	return v
}
