package main

import (
	"encoding/csv"
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

type Neuron struct {
	Weights []float32
	Bias    float32
}

type NN struct {
	Neurons             [1]Neuron
	NormalizationVector []float32
	Trained             bool
}

const (
	Ninputs      = 2
	TrainingFile = "training.csv"
	NetworkFile  = "nn.bin"
	EPS          = 1e-9
)

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

func (n *Neuron) Train(trainingData [][]float32, ninputs, relOutputPos int, trainingRate float32, maxTrainingCount int, activationFunction func(*Neuron, []float32) float32) error {
	var done = false
	var count int

	if len(trainingData) == 0 {
		return errors.New("no training data provided")
	}

	n.Weights = make([]float32, ninputs)
	for i := 0; i < len(n.Weights); i++ {
		n.Weights[i] = (rand.Float32() - 0.5) / 10
	}
	n.Bias = rand.Float32() / 10

	for !done {
		if count > maxTrainingCount {
			return fmt.Errorf("count exceeded %d", maxTrainingCount)
		}

		done = true
		for _, row := range trainingData {
			inputs := row[:ninputs]
			correctOutput := row[ninputs+relOutputPos]
			output := activationFunction(n, inputs)

			if math.Abs(float64(output-correctOutput)) > EPS {
				done = false
				for j := 0; j < len(n.Weights); j++ {
					n.Weights[j] += trainingRate * (correctOutput - output) * inputs[j]
				}
			}
		}
		count++
	}

	return nil
}

func StepFunction(n *Neuron, inputs []float32) float32 {
	var output float32

	for i := range inputs {
		output += inputs[i] * n.Weights[i]
	}
	output += n.Bias

	if output < 0 {
		return -1
	} else {
		return 1
	}
}

func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
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
	for k := 0; k < count; k++ {
		i := rand.Int() % len(basis)

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
	}

	return nil
}

func ReadTrainingData(trainingFilename string) ([][]float32, error) {
	f, err := os.Open(trainingFilename)
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

func NormalizeTrainingData(trainingData [][]float32, ninputs int) []float32 {
	normalizationVector := make([]float32, ninputs)

	for i := 0; i < len(trainingData); i++ {
		for j := 0; j < ninputs; j++ {
			normalizationVector[j] = max(normalizationVector[j], float32(math.Abs(float64(trainingData[i][j]))))
		}
	}

	for i := 0; i < len(trainingData); i++ {
		for j := 0; j < ninputs; j++ {
			trainingData[i][j] /= normalizationVector[j]
		}
	}

	return normalizationVector
}

func main() {
	var nn NN

	generationFlag := flag.Bool("g", false, "generate training data for NN")
	trainingFlag := flag.Bool("t", false, "train NN with data from file")
	flag.Parse()

	nn.Load(NetworkFile)

	if *generationFlag {
		if err := GenerateTrainingData(TrainingFile, [][]float32{
			{53.2521, 34.3717, 1},  /* Bryansk. */
			{52.9651, 36.0785, -1}, /* Orel. */
		}, 0.15, Ninputs, 100); err != nil {
			Fatalf("Failed to generate training data: %s\n", err.Error())
		}
	}

	if (!nn.Trained) || (*trainingFlag) {
		trainingData, err := ReadTrainingData(TrainingFile)
		if err != nil {
			Fatalf("Failed to read training data: %s\n", err.Error())
		}

		nn.NormalizationVector = NormalizeTrainingData(trainingData, Ninputs)

		for i := 0; i < len(nn.Neurons); i++ {
			if err := nn.Neurons[i].Train(trainingData, Ninputs, i, 0.05, 5000, StepFunction); err != nil {
				Fatalf("Failed to train neuron #%d: %s\n", i, err.Error())
			}
		}
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
		inputs[i] /= nn.NormalizationVector[i]
	}

	for i := 0; i < len(nn.Neurons); i++ {
		fmt.Printf("Answer from neuron #%d: %f\n", i, StepFunction(&nn.Neurons[i], inputs))
	}
}
