package main

import (
	"fmt"
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
	testNN             NN
	testTrainingData11 [][]float32
	testTrainingData01 [][]float32
)

func benchmarkNeuronTrain(b *testing.B, nlayers, nneurons int, trainingRate float32, functionID int, trainingData [][]float32) {
	var count int
	var err error
	var nn NN

	b.Helper()

	if nlayers < 2 {
		b.Fatalf("Number of layers must be at least 2 (provided %d)", nlayers)
	}

	nn.Layers = append(nn.Layers, Layer{Neurons: make([]Neuron, nneurons), FunctionID: functionID})
	for i := 0; i < nlayers-2; i++ {
		nn.Layers = append(nn.Layers, Layer{Neurons: make([]Neuron, nneurons), FunctionID: functionID})
	}
	nn.Layers = append(nn.Layers, Layer{Neurons: make([]Neuron, 5), FunctionID: functionID})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count, err = nn.Train(trainingData, Ninputs, trainingRate, 50000)
		if err != nil {
			b.Fatalf("Failed to train NN: %s", err.Error())
		}
	}
	b.Logf("%d epochs", count)
}

func BenchmarkNeuronTrain(b *testing.B) {
	fts := [...]struct {
		FuncName string
		Func     int
		Data     [][]float32
	}{
		{"tanh", FunctionTh, testTrainingData11},
		{"sigmoid", FunctionSigmoid, testTrainingData01},
	}

	for _, ft := range fts {
		f := ft.Func
		t := ft.Data
		name := ft.FuncName

		for l := 2; l <= 5; l++ {
			for n := 5; n <= 20; n *= 2 {
				for r := 0.05; r <= 0.4; r *= 2 {
					b.Run(fmt.Sprintf("%dlayers,%dneurons,%.2frate,%s", l, n, r, name), func(b *testing.B) {
						b.Helper()
						benchmarkNeuronTrain(b, l, n, float32(r), f, t)
					})
				}
			}
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
			if math.Abs(float64(output-city[2+i])) > 1e-1 {
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

	testTrainingData11, err = ReadTrainingData(TrainingFile)
	if err != nil {
		Fatalf("Failed to read training data: %s\n", err.Error())
	}
	testNN.MinVector, testNN.MaxVector = NormalizeTrainingData11(testTrainingData11, Ninputs)

	testTrainingData01, err = ReadTrainingData(TrainingFile)
	if err != nil {
		Fatalf("Failed to read training data: %s\n", err.Error())
	}
	NormalizeTrainingData01(testTrainingData01, Ninputs)
	for i := 0; i < len(testTrainingData01); i++ {
		for j := Ninputs; j < len(testTrainingData01[i]); j++ {
			if testTrainingData01[i][j] < 0 {
				testTrainingData01[i][j] = 0
			}
		}
	}

	if _, err := testNN.Train(testTrainingData11, Ninputs, 0.1, 100000); err != nil {
		Fatalf("Failed to train NN: %s\n", err.Error())
	}

	os.Exit(m.Run())
}
