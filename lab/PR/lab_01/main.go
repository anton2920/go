package main

import (
	"encoding/csv"
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
}

const (
	FunctionSigmoid = iota
	FunctionTh
	FunctionReLU
)

const (
	Ninputs          = 2
	TrainingFile     = "training.csv"
	EPS              = 0.3
	Rate             = 0.15
	MaxTrainingCount = 1000000
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

func (nn *NN) Query(inputs []float32) []float32 {
	outputs := inputs
	for l := 0; l < len(nn.Layers); l++ {
		layer := &nn.Layers[l]
		outputs = layer.Query(outputs)
	}
	return outputs
}

func (nn *NN) Train(inputs [][]float32, outputs [][]float32, trainingRate float32, maxTrainingCount int) (int, error) {
	var done, needsTraining bool
	var count int

	rng := rand.New(rand.NewSource(6585))
	for l := 0; l < len(nn.Layers); l++ {
		layer := &nn.Layers[l]
		layer.Outputs = nil

		for n := 0; n < len(layer.Neurons); n++ {
			neuron := &layer.Neurons[n]

			var nweights int
			if l == 0 {
				nweights = len(inputs[0])
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
		for i := 0; i < len(inputs); i++ {
			results := nn.Query(inputs[i])

			// fmt.Println(inputs, outputs, results)

			needsTraining = false
			for j := 0; j < len(outputs[i]); j++ {
				if math.Abs(float64(outputs[i][j]-results[j])) > EPS {
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
							coef[n] = Derivatives[layer.FunctionID](layer.Outputs[n]) * (outputs[i][n] - results[n])
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
								*currWeight += trainingRate*coef[n]*inputs[i][w] + 0.5*(*currWeight-*prevWeight)
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

func NormalizeTrainingData01(trainingData [][]float32) ([]float32, []float32) {
	ninputs := len(trainingData[0])
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

func NormalizeTrainingData11(trainingData [][]float32) ([]float32, []float32) {
	ninputs := len(trainingData[0])
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

func EuclidianDistance(p1, p2 []float32) float32 {
	var distance float32

	for i := 0; i < len(p1); i++ {
		distance += (p1[i] - p2[i]) * (p1[i] - p2[i])
	}

	return distance
}

func FindMostDistantPointIndicies(points [][]float32) (int, int) {
	var maxDistance float32
	var pindex1, pindex2 int

	for i := 0; i < len(points); i++ {
		for j := 0; j < len(points); j++ {
			if i == j {
				continue
			}
			distance := EuclidianDistance(points[i], points[j])
			if distance > maxDistance {
				pindex1 = i
				pindex2 = j
				maxDistance = distance
			}
		}
	}

	return pindex1, pindex2
}

func RemoveAtIndex[T any](vs []T, i int) []T {
	if (vs == nil) || (len(vs) == 0) {
		return vs
	}
	if i < len(vs)-1 {
		copy(vs[i:], vs[i+1:])
	}
	return vs[:len(vs)-1]
}

func RemoveByValue(vs [][]float32, v []float32) [][]float32 {
	if (vs == nil) || (len(vs) == 0) {
		return vs
	}
	var found bool
	var i int
	for ; i < len(vs); i++ {
		found = true
		for j := 0; j < len(vs[i]); j++ {
			if vs[i][j] != v[j] {
				found = false
				break
			}
		}
		if found {
			break
		}
	}
	if i < len(vs)-1 {
		copy(vs[i:], vs[i+1:])
	}
	return vs[:len(vs)-1]
}

func main() {
	generationFlag := flag.Bool("g", false, "generate training data for NN")
	flag.Parse()

	if *generationFlag {
		if err := GenerateTrainingData(TrainingFile, [][]float32{
			{53.2521, 34.3717}, /* Bryansk. */
			{52.9651, 36.0785}, /* Orel. */
			{54.7818, 32.0401}, /* Smolensk. */
			{54.5293, 36.2754}, /* Kaluga. */
			{54.1961, 37.6182}, /* Tula. */
		}, 0.1, Ninputs, 100); err != nil {
			Fatalf("Failed to generate training data: %s\n", err.Error())
		}
	}

	nn := NN{
		Layers: []Layer{
			{Neurons: make([]Neuron, 10), FunctionID: FunctionSigmoid},
			{Neurons: make([]Neuron, 5), FunctionID: FunctionSigmoid},
			{Neurons: make([]Neuron, 2), FunctionID: FunctionSigmoid},
		},
	}

	trainingData, err := ReadTrainingData(TrainingFile)
	if err != nil {
		Fatalf("Failed to read training data: %s\n", err.Error())
	}

	nn.MinVector, nn.MaxVector = NormalizeTrainingData01(trainingData)

	inputs := trainingData
	testInputs := [][]float32{
		{53.2521, 34.3717}, /* Bryansk. */
		{52.9651, 36.0785}, /* Orel. */
		{54.7818, 32.0401}, /* Smolensk. */
		{54.5293, 36.2754}, /* Kaluga. */
		{54.1961, 37.6182}, /* Tula. */
	}
	for i := 0; i < len(testInputs); i++ {
		for j := 0; j < len(testInputs[0]); j++ {
			testInputs[i][j] = (testInputs[i][j] - nn.MinVector[j]) / (nn.MaxVector[j] - nn.MinVector[j])
			// testInputs[i][j] = (testInputs[i][j] - 0.5*(nn.MaxVector[j]+nn.MinVector[j])) / (0.5 * (nn.MaxVector[j] - nn.MinVector[j]))
		}
	}

	var clusters [][]float32
	var points [][]float32

	fmt.Println("Initial clusterization...")
	pindex1, pindex2 := FindMostDistantPointIndicies(inputs)
	points = append(points, inputs[pindex1], inputs[pindex2])
	clusters = [][]float32{{1, 0}, {0, 1}}

	count, err := nn.Train(points, clusters, Rate, MaxTrainingCount)
	if err != nil {
		Fatalf("Failed to train NN: %s\n", err.Error())
	}
	fmt.Printf("Trained after %d epochs\n", count)

	inputs = RemoveByValue(inputs, points[0])
	inputs = RemoveByValue(inputs, points[1])

	for i := 0; i < len(inputs); i++ {
		result := nn.Query(inputs[i])
		fmt.Println(result)

		var clusterNumber int
		newCluster := true
		for j := 0; j < len(result); j++ {
			if result[j] > 0.7 {
				newCluster = false
				clusterNumber = j
				break
			}
		}

		if newCluster {
			fmt.Println("New cluster")

			for j := 0; j < len(clusters); j++ {
				clusters[j] = append(clusters[j], 0)
			}

			cluster := make([]float32, len(clusters[0]))
			cluster[len(cluster)-1] = 1
			clusters = append(clusters, cluster)
			points = append(points, inputs[i])

			nn.Layers[len(nn.Layers)-1].Neurons = make([]Neuron, len(clusters[0]))
			count, err := nn.Train(points, clusters, Rate, MaxTrainingCount)
			if err != nil {
				Fatalf("Failed to train NN: %s\n", err.Error())
			}
			fmt.Printf("Trained after %d epochs\n", count)
		} else {
			fmt.Println("Using existing cluster")

			cluster := make([]float32, len(clusters[0]))
			cluster[clusterNumber] = 1

			clusters = append(clusters, cluster)
			points = append(points, inputs[i])
		}
	}

	fmt.Println("Number of initial clusters: ", len(clusters[0]))
	for i := 0; i < len(points); i++ {
		for j := 0; j < len(points[i]); j++ {
			point := points[i][j]*(nn.MaxVector[j]-nn.MinVector[j]) + nn.MinVector[j]
			// point := points[i][j]*0.5*(nn.MaxVector[j]-nn.MinVector[j]) + 0.5*(nn.MaxVector[j]+nn.MinVector[j])
			fmt.Printf("%f,", point)
		}

		for j := 0; j < len(clusters[i]); j++ {
			if clusters[i][j] == 1 {
				fmt.Println(j)
			}
		}
	}

	/* NOTE(anton2920): merging neighbouring clusters. */
	done := false
	for !done {
		done = true

		var averages [][]float32
		var counts map[int]int
		var sums []float32

		counts = make(map[int]int)

		for c := 0; c < len(clusters[0]); c++ {
			var sum float32
			var count int

			for j := 0; j < len(points[0]); j++ {
				for i := 0; i < len(points); i++ {
					if clusters[i][c] == 1 {
						sum += points[i][j]
						count++
					}
				}
				sums = append(sums, sum/float32(count))
			}
			averages = append(averages, sums)
			counts[c] = count
			sums = nil
		}

		// fmt.Println(averages)

		var indicies [2]int
		for i := 0; i < len(averages)-1; i++ {
			for j := i + 1; j < len(averages); j++ {
				distance := EuclidianDistance(averages[i], averages[j])
				// fmt.Println(i, j, distance)

				if distance < 0.005 {
					indicies[0] = i
					indicies[1] = j
					done = false
					break
				}
			}
			if !done {
				break
			}
		}

		if !done {
			var keepIndex, removeIndex int
			if counts[indicies[0]] > counts[indicies[1]] {
				keepIndex = indicies[0]
				removeIndex = indicies[1]
			} else {
				keepIndex = indicies[1]
				removeIndex = indicies[0]
			}

			for i := 0; i < len(clusters); i++ {
				if clusters[i][removeIndex] == 1 {
					clusters[i][keepIndex] = 1
				}
				clusters[i] = RemoveAtIndex(clusters[i], removeIndex)
			}

			fmt.Printf("Merged cluster %d with %d\n", removeIndex, keepIndex)
			// fmt.Println(clusters)
		}
	}

	fmt.Println("Number of clusters after merging: ", len(clusters[0]))
	for i := 0; i < len(points); i++ {
		for j := 0; j < len(points[i]); j++ {
			point := points[i][j]*(nn.MaxVector[j]-nn.MinVector[j]) + nn.MinVector[j]
			// point := points[i][j]*0.5*(nn.MaxVector[j]-nn.MinVector[j]) + 0.5*(nn.MaxVector[j]+nn.MinVector[j])
			fmt.Printf("%f,", point)
		}

		for j := 0; j < len(clusters[i]); j++ {
			if clusters[i][j] == 1 {
				fmt.Println(j)
			}
		}
	}

	fmt.Println("Final training...")
	nn.Layers[len(nn.Layers)-1].Neurons = make([]Neuron, len(clusters[0]))
	count, err = nn.Train(points, clusters, Rate, MaxTrainingCount)
	if err != nil {
		Fatalf("Failed to train NN: %s\n", err.Error())
	}
	fmt.Printf("Trained after %d epochs\n", count)

	fmt.Println("Testing...")
	for i := 0; i < len(testInputs); i++ {
		result := nn.Query(testInputs[i])

		for j := 0; j < len(testInputs[i]); j++ {
			point := testInputs[i][j]*(nn.MaxVector[j]-nn.MinVector[j]) + nn.MinVector[j]
			fmt.Printf("%f,", point)
		}

		var clusterNumber int
		newCluster := true
		for j := 0; j < len(result); j++ {
			if result[j] > 0.7 {
				newCluster = false
				clusterNumber = j
				break
			}
		}
		if newCluster {
			fmt.Println("FAILED TO PREDICT!!!")
		} else {
			fmt.Printf("%d\n", clusterNumber)
		}
	}
}
