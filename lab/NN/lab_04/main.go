package main

import (
	"encoding/csv"
	"encoding/gob"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

type ActivationFunction func(float32) float32

type Neuron struct {
	Weights         []float32
	PreviousWeights []float32
	Bias            float32
}

type Layer struct {
	Neurons    []Neuron
	Outputs    []float32
	FunctionID int
}

type NN struct {
	Layers    []Layer
	MinVector []float32
	MaxVector []float32
	Trained   bool
}

const (
	FunctionSigmoid = iota
	FunctionTh
	FunctionReLU
)

const (
	Ninputs      = 2
	TrainingFile = "training.csv"
	NetworkFile  = "nn.bin"
	EPS          = 1e-1
)

var (
	Functions = []ActivationFunction{
		Sigmoid,
		Th,
		ReLU,
	}

	Derivatives = []ActivationFunction{
		SigmoidPrime,
		ThPrime,
		ReLUPrime,
	}
)

func Sigmoid(x float32) float32 {
	return 1 / (1 + float32(math.Exp(float64(-x))))
}

func SigmoidPrime(x float32) float32 {
	return x * (1 - x)
}

func Th(x float32) float32 {
	return float32(math.Tanh(float64(x)))
}

func ThPrime(x float32) float32 {
	return 1 - x*x
}

func ReLU(x float32) float32 {
	if x < 0 {
		x *= 0.01
	}
	return x
}

func ReLUPrime(x float32) float32 {
	if x < 0 {
		return 0.01
	} else {
		return 1
	}
}

func (n *Neuron) Query(inputs []float32) float32 {
	var output float32

	for i := 0; i < len(inputs); i++ {
		output += inputs[i] * n.Weights[i]
	}
	output += n.Bias

	return output
}

func (l *Layer) Query(inputs []float32) []float32 {
	if l.Outputs == nil {
		l.Outputs = make([]float32, len(l.Neurons))
	}

	for n := 0; n < len(l.Neurons); n++ {
		neuron := &l.Neurons[n]
		l.Outputs[n] = Functions[l.FunctionID](neuron.Query(inputs))
	}
	return l.Outputs
}

func (nn *NN) Load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	gobDecoder := gob.NewDecoder(f)
	if err := gobDecoder.Decode(&nn); err != nil {
		return err
	}

	return nil
}

func (nn *NN) Store(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	gobEncoder := gob.NewEncoder(f)
	if err := gobEncoder.Encode(&nn); err != nil {
		return err
	}

	return nil
}

func (nn *NN) Query(inputs []float32) []float32 {
	outputs := inputs
	for l := 0; l < len(nn.Layers); l++ {
		layer := &nn.Layers[l]
		outputs = layer.Query(outputs)
	}
	return outputs
}

func (nn *NN) Train(trainingData [][]float32, ninputs int, trainingRate float32, maxTrainingCount int) (int, error) {
	var done, needsTraining bool
	var count int

	rng := rand.New(rand.NewSource(6585))
	for l := 0; l < len(nn.Layers); l++ {
		layer := &nn.Layers[l]

		for n := 0; n < len(layer.Neurons); n++ {
			neuron := &layer.Neurons[n]

			var nweights int
			if l == 0 {
				nweights = Ninputs
			} else {
				nweights = len(nn.Layers[l-1].Neurons)
			}
			neuron.Weights = make([]float32, nweights)

			for w := 0; w < len(neuron.Weights); w++ {
				neuron.Weights[w] = (rng.Float32() - 0.5) / 10
			}
			neuron.Bias = (rng.Float32() - 0.5) / 10

			neuron.PreviousWeights = make([]float32, len(neuron.Weights))
			copy(neuron.PreviousWeights, neuron.Weights)
		}
	}

	for !done {
		if count > maxTrainingCount {
			return 0, fmt.Errorf("count exceeded %d", maxTrainingCount)
		}

		done = true
		for _, row := range trainingData {
			inputs := row[:ninputs]
			correctOutputs := row[ninputs:]

			outputs := nn.Query(inputs)

			// fmt.Println(inputs, correctOutputs, outputs)

			needsTraining = false
			for j := 0; j < len(correctOutputs); j++ {
				if math.Abs(float64(correctOutputs[j]-outputs[j])) > EPS {
					done = false
					needsTraining = true
					break
				}
			}

			if needsTraining {
				var coef, prevCoef []float32
				for l := len(nn.Layers) - 1; l >= 0; l-- {
					layer := &nn.Layers[l]

					coef = make([]float32, len(layer.Neurons))
					for n := 0; n < len(layer.Neurons); n++ {
						if l == len(nn.Layers)-1 {
							coef[n] = Derivatives[layer.FunctionID](layer.Outputs[n]) * (correctOutputs[n] - outputs[n])
						} else {
							var temp float32
							nextLayer := &nn.Layers[l+1]
							for i := 0; i < len(nextLayer.Neurons); i++ {
								temp += prevCoef[i] * nextLayer.Neurons[i].Weights[n]
							}
							coef[n] = Derivatives[layer.FunctionID](layer.Outputs[n]) * temp
						}

						neuron := &layer.Neurons[n]
						for w := 0; w < len(neuron.Weights); w++ {
							currWeight := &neuron.Weights[w]
							prevWeight := &neuron.PreviousWeights[w]

							if l == 0 {
								*currWeight += trainingRate*coef[n]*inputs[w] + 0.5*(*currWeight-*prevWeight)
							} else {
								prevLayer := &nn.Layers[l-1]
								*currWeight += trainingRate*coef[n]*prevLayer.Outputs[w] + 0.5*(*currWeight-*prevWeight)
							}

							*prevWeight = *currWeight
						}
						neuron.Bias += trainingRate * coef[n]
					}

					prevCoef = coef
				}
			}
		}

		count++
	}

	return count, nil
}

func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func GenerateTrainingDataRow(csvWriter *csv.Writer, row []string, basis [][]float32, ninputs int, i int, maxOffset float32) error {
	var j int
	for ; j < ninputs; j++ {
		row[j] = strconv.FormatFloat(float64(basis[i][j]+maxOffset*rand.Float32()), 'f', 4, 32)
	}

	for ; j < len(basis[i]); j++ {
		row[j] = strconv.Itoa(int(basis[i][j]))
	}

	if err := csvWriter.Write(row); err != nil {
		return err
	}

	return nil
}

func GenerateTrainingData(trainingFilename string, basis [][]float32, maxOffset float32, ninputs, count int) error {
	f, err := os.Create(trainingFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	csvWriter := csv.NewWriter(f)
	defer csvWriter.Flush()

	row := make([]string, len(basis[0]))
	for i := 0; i < len(basis); i++ {
		if err := GenerateTrainingDataRow(csvWriter, row, basis, ninputs, i, maxOffset); err != nil {
			return err
		}
	}

	for k := 0; k < count-len(basis); k++ {
		i := rand.Int() % len(basis)
		if err := GenerateTrainingDataRow(csvWriter, row, basis, ninputs, i, maxOffset); err != nil {
			return err
		}
	}

	return nil
}

func ReadTrainingData(trainingFile string) ([][]float32, error) {
	f, err := os.Open(trainingFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	trainingStrings, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	trainingData := make([][]float32, len(trainingStrings))
	for i := 0; i < len(trainingData); i++ {
		trainingData[i] = make([]float32, len(trainingStrings[i]))
		for j := 0; j < len(trainingStrings[i]); j++ {
			value, err := strconv.ParseFloat(strings.TrimSpace(trainingStrings[i][j]), 32)
			if err != nil {
				return nil, err
			}
			trainingData[i][j] = float32(value)
		}
	}

	return trainingData, nil
}

func NormalizeTrainingData01(trainingData [][]float32, ninputs int) ([]float32, []float32) {
	minVector := make([]float32, ninputs)
	maxVector := make([]float32, ninputs)

	for j := 0; j < ninputs; j++ {
		minVector[j] = trainingData[0][j]
		maxVector[j] = trainingData[0][j]
	}

	for i := 0; i < len(trainingData); i++ {
		for j := 0; j < ninputs; j++ {
			minVector[j] = min(minVector[j], float32(math.Abs(float64(trainingData[i][j]))))
			maxVector[j] = max(maxVector[j], float32(math.Abs(float64(trainingData[i][j]))))
		}
	}

	for i := 0; i < len(trainingData); i++ {
		for j := 0; j < ninputs; j++ {
			trainingData[i][j] = (trainingData[i][j] - minVector[j]) / (maxVector[j] - minVector[j])
			// trainingData[i][j] = (trainingData[i][j] - 0.5*(maxVector[j]+minVector[j])) / (0.5 * (maxVector[j] - minVector[j]))
		}
	}

	return minVector, maxVector
}

func NormalizeTrainingData11(trainingData [][]float32, ninputs int) ([]float32, []float32) {
	minVector := make([]float32, ninputs)
	maxVector := make([]float32, ninputs)

	for j := 0; j < ninputs; j++ {
		minVector[j] = trainingData[0][j]
		maxVector[j] = trainingData[0][j]
	}

	for i := 0; i < len(trainingData); i++ {
		for j := 0; j < ninputs; j++ {
			minVector[j] = min(minVector[j], float32(math.Abs(float64(trainingData[i][j]))))
			maxVector[j] = max(maxVector[j], float32(math.Abs(float64(trainingData[i][j]))))
		}
	}

	for i := 0; i < len(trainingData); i++ {
		for j := 0; j < ninputs; j++ {
			// trainingData[i][j] = (trainingData[i][j] - minVector[j]) / (maxVector[j] - minVector[j])
			trainingData[i][j] = (trainingData[i][j] - 0.5*(maxVector[j]+minVector[j])) / (0.5 * (maxVector[j] - minVector[j]))
		}
	}

	return minVector, maxVector
}

func main() {
	var nn NN

	generationFlag := flag.Bool("g", false, "generate training data for NN")
	trainingFlag := flag.Bool("t", false, fmt.Sprintf("train NN with data from '%s' file", TrainingFile))
	flag.Parse()

	nn.Load(NetworkFile)

	if *generationFlag {
		if err := GenerateTrainingData(TrainingFile, [][]float32{
			{53.2521, 34.3717, 1, -1, -1, -1, -1}, /* Bryansk. */
			{52.9651, 36.0785, -1, 1, -1, -1, -1}, /* Orel. */
			{54.7818, 32.0401, -1, -1, 1, -1, -1}, /* Smolensk. */
			{54.5293, 36.2754, -1, -1, -1, 1, -1}, /* Kaluga. */
			{54.1961, 37.6182, -1, -1, -1, -1, 1}, /* Tula. */
		}, 0.15, Ninputs, 20); err != nil {
			Fatalf("Failed to generate training data: %s\n", err.Error())
		}
	}

	if (!nn.Trained) || (*trainingFlag) {
		nn = NN{
			Layers: []Layer{
				{Neurons: make([]Neuron, 5), FunctionID: FunctionTh},
				{Neurons: make([]Neuron, 5), FunctionID: FunctionTh},
				{Neurons: make([]Neuron, 5), FunctionID: FunctionTh},
				{Neurons: make([]Neuron, 5), FunctionID: FunctionTh},
			},
		}

		trainingData, err := ReadTrainingData(TrainingFile)
		if err != nil {
			Fatalf("Failed to read training data: %s\n", err.Error())
		}

		nn.MinVector, nn.MaxVector = NormalizeTrainingData11(trainingData, Ninputs)

		count, err := nn.Train(trainingData, Ninputs, 0.1, 100000)
		if err != nil {
			Fatalf("Failed to train NN: %s\n", err.Error())
		}
		fmt.Printf("Trained after %d epochs\n", count)

		nn.Trained = true

		if err := nn.Store(NetworkFile); err != nil {
			Fatalf("Failed to store NN: %s\n", err.Error())
		}
	}

	if !nn.Trained {
		Fatalf("NN must be trained before it can process data\n")
	}

	inputs := make([]float32, 2)
	for i := 0; i < len(inputs); i++ {
		fmt.Printf("Type value %d: ", i+1)
		_, _ = fmt.Scanf("%f", &inputs[i])
		inputs[i] = (inputs[i] - 0.5*(nn.MaxVector[i]+nn.MinVector[i])) / (0.5 * (nn.MaxVector[i] - nn.MinVector[i]))
	}

	for i, output := range nn.Query(inputs) {
		fmt.Printf("Answer from neuron #%d: %f\n", i, output)
	}
}
