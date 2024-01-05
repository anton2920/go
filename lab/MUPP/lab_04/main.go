package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Neuron struct {
	Weights       []float32
	X, Y          float32
	Width, Height float32
}

type SOM struct {
	Neurons   []Neuron
	MinVector []float32
	MaxVector []float32
	Trained   bool
}

const (
	Ninputs           = 2
	TrainingFile      = "training.csv"
	TrainingImageFile = "training.png"

	NetworkFile      = "nn.bin"
	NetworkImageFile = "nn.png"

	ImageWidth  = 400
	ImageHeight = 400

	NRows = 20
	NCols = 20
)

var Tmpls *template.Template

var (
	ReloadPageError    = errors.New("whoops... Something went wrong. Please reload this page or try again")
	TryAgainLaterError = errors.New("whoops... Something went wrong. Please try again later")
)

func (n *Neuron) Render(img *image.RGBA) {
	var color color.RGBA
	color.R = uint8(n.Weights[0] * 255)
	color.G = uint8(n.Weights[1] * 255)
	color.A = 255

	for y := int(n.Y - n.Height*0.5); y < int(n.Y+n.Height*0.5); y += 1 {
		for x := int(n.X - n.Width*0.5); x < int(n.X+n.Width*0.5); x += 1 {
			img.Set(x, y, color)
		}
	}
}

func (n *Neuron) DistanceTo(inputs []float32) float32 {
	var distance float32

	for i := 0; i < len(n.Weights); i++ {
		distance += (inputs[i] - n.Weights[i]) * (inputs[i] - n.Weights[i])
	}

	return distance
}

func (n *Neuron) AdjustWeights(inputs []float32, rate, influence float32) {
	for i := 0; i < len(n.Weights); i++ {
		n.Weights[i] += rate * influence * (inputs[i] - n.Weights[i])
	}
}

/* FindBMU returns Best Matching Node: node which is the closest to input. */
func (s *SOM) FindBMU(inputs []float32) int {
	bmuIndex := 0
	minDist := s.Neurons[0].DistanceTo(inputs)

	for i := 1; i < len(s.Neurons); i++ {
		neuron := &s.Neurons[i]
		dist := neuron.DistanceTo(inputs)

		if dist < minDist {
			minDist = dist
			bmuIndex = i
		}
	}

	return bmuIndex
}

func (s *SOM) Train(trainingData [][]float32, ninputs int, width, height int, nrows, ncols int, maxCount int, startingRate float32) {
	var count int

	rate := startingRate
	neuronWidth := float32(width) / float32(ncols)
	neuronHeight := float32(height) / float32(nrows)
	mapRadius := float32(max(width, height)) / 2
	timeConstant := float32(float64(maxCount) / math.Log(float64(mapRadius)))

	rng := rand.New(rand.NewSource(6585))
	for row := 0; row < nrows; row++ {
		for col := 0; col < ncols; col++ {
			var neuron Neuron
			neuron.Weights = make([]float32, ninputs)
			for i := 0; i < ninputs; i++ {
				neuron.Weights[i] = (rng.Float32() - 0.5) / 10
			}
			neuron.X = neuronWidth * (float32(col) + 0.5)
			neuron.Y = neuronHeight * (float32(row) + 0.5)
			neuron.Width = neuronWidth
			neuron.Height = neuronHeight
			s.Neurons = append(s.Neurons, neuron)
		}
	}

	for count < maxCount {
		inputs := trainingData[rand.Int()%len(trainingData)]
		bmuIndex := s.FindBMU(inputs)
		bmu := &s.Neurons[bmuIndex]

		neighbourhoodRadius := mapRadius * float32(math.Exp(float64(-count)/float64(timeConstant)))

		for i := 0; i < len(s.Neurons); i++ {
			neuron := &s.Neurons[i]

			distanceSquared := (bmu.X-neuron.X)*(bmu.X-neuron.X) + (bmu.Y-neuron.Y)*(bmu.Y-neuron.Y)
			radiusSquared := neighbourhoodRadius * neighbourhoodRadius

			if distanceSquared < radiusSquared {
				influence := float32(math.Exp(float64(-distanceSquared / (2 * radiusSquared))))
				neuron.AdjustWeights(inputs, rate, influence)
			}
		}

		rate = startingRate * float32(math.Exp(float64(-count)/float64(maxCount)))
		count++
	}
}

func (s *SOM) Render(img *image.RGBA) {
	for i := 0; i < len(s.Neurons); i++ {
		s.Neurons[i].Render(img)
	}
}

func GenerateTrainingData(trainingFilename string, basis [][]float32, maxOffset float32, count int) error {
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

		for j := 0; j < len(row); j++ {
			row[j] = strconv.FormatFloat(float64(basis[i][j]+maxOffset*rand.Float32()), 'f', 4, 32)
		}

		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func ReadTrainingData(trainingFile io.Reader) ([][]float32, error) {
	csvReader := csv.NewReader(trainingFile)
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

func NormalizeTrainingData(trainingData [][]float32) ([]float32, []float32) {
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

func WriteTemplate(w http.ResponseWriter, tmpl string, respCode int, payload url.Values, err error) {
	response := new(bytes.Buffer)
	if e := Tmpls.ExecuteTemplate(response, tmpl, struct {
		Error   error
		Payload url.Values
	}{err, payload}); e != nil {
		log.Printf("ERROR: failed to execute '%s' template: %s", tmpl, e.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if err := Tmpls.ExecuteTemplate(w, "error.tmpl", TryAgainLaterError); err != nil {
			log.Printf("ERROR: failed to execute 'error.tmpl' after failing to execute another template")
		}
		return
	}

	w.WriteHeader(respCode)
	if _, err := w.Write(response.Bytes()); err != nil {
		log.Print("ERROR: failed to write template: ", err)
	}
}

func WriteTemplateAny[T any](w http.ResponseWriter, tmpl string, respCode int, s T) {
	response := new(bytes.Buffer)
	if e := Tmpls.ExecuteTemplate(response, tmpl, s); e != nil {
		log.Printf("ERROR: failed to execute '%s' template: %s", tmpl, e.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if err := Tmpls.ExecuteTemplate(w, "error.tmpl", TryAgainLaterError); err != nil {
			log.Printf("ERROR: failed to execute 'error.tmpl' after failing to execute another template")
		}
		return
	}
	w.WriteHeader(respCode)
	if _, err := w.Write(response.Bytes()); err != nil {
		log.Print("ERROR: failed to write template: ", err)
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		WriteTemplate(w, "index.tmpl", http.StatusOK, nil, nil)
	case http.MethodPost:
		var som SOM

		if err := r.ParseForm(); err != nil {
			WriteTemplate(w, "index.tmpl", http.StatusBadRequest, r.Form, ReloadPageError)
			return
		}

		trainingData, err := ReadTrainingData(strings.NewReader(r.Form.Get("Data")))
		if err != nil {
			WriteTemplate(w, "index.tmpl", http.StatusBadRequest, r.Form, fmt.Errorf("invalid input data: %s", err.Error()))
			return
		}
		som.MinVector, som.MaxVector = NormalizeTrainingData(trainingData)
		som.Train(trainingData, Ninputs, ImageWidth, ImageHeight, NRows, NCols, 5000, 0.1)
		som.Trained = true

		img := image.NewRGBA(image.Rect(0, 0, ImageWidth, ImageHeight))
		som.Render(img)

		imgBuffer := new(bytes.Buffer)
		if err := png.Encode(imgBuffer, img); err != nil {
			WriteTemplate(w, "index.tmpl", http.StatusInternalServerError, r.Form, TryAgainLaterError)
			return
		}

		r.Form.Set("Image", base64.StdEncoding.EncodeToString(imgBuffer.Bytes()))

		WriteTemplate(w, "index.tmpl", http.StatusOK, r.Form, nil)
	}
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
			// {54.1961, 37.6182}, /* Tula. */
		}, 0.15, 50); err != nil {
			log.Fatalf("Failed to generate training data: %s\n", err.Error())
		}
	}

	Tmpls = template.Must(template.ParseGlob("*.tmpl"))

	log.Print("Listening on 0.0.0.0:8080...")
	http.HandleFunc("/", IndexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
