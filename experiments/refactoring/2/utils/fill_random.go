package utils

import (
	"math/rand"

	"../json"
	"../types"
)

func FillRandomPerformances(perfs []types.Performance, seed int64) {
	keys := make([]string, 0, len(json.Plays))
	for k := range json.Plays {
		keys = append(keys, k)
	}

	rng := rand.New(rand.NewSource(seed))
	for i := 0; i < len(perfs); i++ {
		perfs[i] = types.Performance{PlayID: keys[rng.Intn(len(json.Plays))], Audience: rng.Intn(500)}
	}
}
