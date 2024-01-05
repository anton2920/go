package main

import (
	"math"
	"math/rand"
	"os"
	"testing"
)

const (
	MaxChecks = 100
	MaxOffset = 0.1
)

var (
	testNN           NN
	testTrainingData [][]float32
)

func BenchmarkNeuronTrain(b *testing.B) {
	var nn NN

	for i := 0; i < b.N; i++ {
		for j := 0; j < len(nn.Neurons); j++ {
			if err := nn.Neurons[j].Train(testTrainingData, Ninputs, j, 0.05, 10000, StepFunction); err != nil {
				b.Fatalf("Failed to train neuron: %s\n", err.Error())
			}
		}
	}
}

func testRandomFrom(x float32, maxOffset float32) float32 {
	return x + maxOffset*rand.Float32()
}

func testCity(t *testing.T, city [6]float32) {
	var inputs [][]float32

	for i := 0; i < MaxChecks; i++ {
		inputs = append(inputs, []float32{testRandomFrom(city[0], MaxOffset), testRandomFrom(city[1], MaxOffset)})
	}

	for i := 0; i < len(inputs); i++ {
		for j := 0; j < len(inputs[i]); j++ {
			inputs[i][j] = (inputs[i][j] - 0.5*(testNN.MaxVector[j]+testNN.MinVector[j])) / (0.5 * (testNN.MaxVector[j] - testNN.MinVector[j]))
		}

		for n := 0; n < len(testNN.Neurons); n++ {
			output := StepFunction(&testNN.Neurons[n], inputs[i])
			if math.Abs(float64(output-city[2+n])) > EPS {
				t.Errorf("Neuron failed to predict city at [%f; %f]: expected %.2f, got %.2f", inputs[i][0], inputs[i][1], city[2], output)
			}
		}
	}
}

func TestBryansk(t *testing.T) {
	testCity(t, [...]float32{53.2521, 34.3717, 1, -1, -1, -1})
}

func TestOrel(t *testing.T) {
	testCity(t, [...]float32{52.9651, 36.0785, -1, 1, -1, -1})
}

func TestSmolensk(t *testing.T) {
	testCity(t, [...]float32{54.7818, 32.0401, -1, -1, 1, -1})
}

func TestKaluga(t *testing.T) {
	testCity(t, [...]float32{54.5293, 36.2754, -1, -1, -1, 1})
}

func TestMain(m *testing.M) {
	var err error

	testTrainingData, err = ReadTrainingData(TrainingFile)
	if err != nil {
		Fatalf("Failed to read training data: %s\n", err.Error())
	}

	testNN.MinVector, testNN.MaxVector = NormalizeTrainingData(testTrainingData, Ninputs)

	for i := 0; i < len(testNN.Neurons); i++ {
		if err := testNN.Neurons[i].Train(testTrainingData, Ninputs, i, 0.05, 10000, StepFunction); err != nil {
			Fatalf("Failed to train neuron: %s\n", err.Error())
		}
	}

	os.Exit(m.Run())
}
