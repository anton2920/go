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
	nn := NN{
		Layers: []Layer{
			{Neurons: make([]Neuron, 5), FunctionID: FunctionTh},
			{Neurons: make([]Neuron, 5), FunctionID: FunctionTh},
			{Neurons: make([]Neuron, 5), FunctionID: FunctionTh},
		},
	}

	for i := 0; i < b.N; i++ {
		if err := nn.Train(testTrainingData, Ninputs, 0.05, 50000); err != nil {
			b.Fatalf("Failed to train NN: %s", err.Error())
		}
	}
}

func testRandomFrom(x float32, maxOffset float32) float32 {
	return x + maxOffset*rand.Float32()
}

func testCity(t *testing.T, city [7]float32) {
	var inputs [][]float32

	for i := 0; i < MaxChecks; i++ {
		inputs = append(inputs, []float32{testRandomFrom(city[0], MaxOffset), testRandomFrom(city[1], MaxOffset)})
	}

	for i := 0; i < len(inputs); i++ {
		for j := 0; j < len(inputs[i]); j++ {
			inputs[i][j] = (inputs[i][j] - 0.5*(testNN.MaxVector[j]+testNN.MinVector[j])) / (0.5 * (testNN.MaxVector[j] - testNN.MinVector[j]))
		}

		outputs := testNN.Query(inputs[i])
		for i, output := range outputs {
			if math.Abs(float64(output-city[2+i])) > EPS {
				t.Errorf("Neuron failed to predict city at [%f; %f]: expected %.2f, got %.2f", inputs[i][0], inputs[i][1], city[2], output)
			}
		}
	}
}

func TestBryansk(t *testing.T) {
	testCity(t, [...]float32{53.2521, 34.3717, 1, -1, -1, -1, -1})
}

func TestOrel(t *testing.T) {
	testCity(t, [...]float32{52.9651, 36.0785, -1, 1, -1, -1, -1})
}

func TestSmolensk(t *testing.T) {
	testCity(t, [...]float32{54.7818, 32.0401, -1, -1, 1, -1, -1})
}

func TestKaluga(t *testing.T) {
	testCity(t, [...]float32{54.5293, 36.2754, -1, -1, -1, 1, -1})
}

func TestTula(t *testing.T) {
	testCity(t, [...]float32{54.1961, 37.6182, -1, -1, -1, -1, 1})
}

func TestMain(m *testing.M) {
	var err error

	testNN = NN{
		Layers: []Layer{
			{Neurons: make([]Neuron, 5), FunctionID: FunctionTh},
			{Neurons: make([]Neuron, 5), FunctionID: FunctionTh},
			{Neurons: make([]Neuron, 5), FunctionID: FunctionTh},
		},
	}

	testTrainingData, err = ReadTrainingData(TrainingFile)
	if err != nil {
		Fatalf("Failed to read training data: %s\n", err.Error())
	}

	testNN.MinVector, testNN.MaxVector = NormalizeTrainingData(testTrainingData, Ninputs)

	if err := testNN.Train(testTrainingData, Ninputs, 0.05, 50000); err != nil {
		Fatalf("Failed to train NN: %s\n", err.Error())
	}

	os.Exit(m.Run())
}
