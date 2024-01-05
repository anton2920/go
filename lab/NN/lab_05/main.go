package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"slices"
)

type Letter []int8

const (
	LettersDir = "./Letters"

	LetterWidth      = 9
	LetterHeight     = 9
	LetterResolution = LetterWidth * LetterHeight
)

func DecodeImage(filepath string) (Letter, error) {
	f, err := os.Open(fmt.Sprintf("%s/%s", LettersDir, filepath))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if (data[0] != 'B') && (data[1] != 'M') {
		return nil, fmt.Errorf("expected 'BM' header, found %c%c", data[0], data[1])
	}

	size := int(binary.LittleEndian.Uint32(data[2:]))
	if size != len(data) {
		return nil, fmt.Errorf("size mismatch %d != %d", size, len(data))
	}

	offset := int(binary.LittleEndian.Uint32(data[10:]))
	if offset > len(data) {
		return nil, fmt.Errorf("data offset %d > %d file size", offset, len(data))
	}

	width := int(binary.LittleEndian.Uint32(data[18:]))
	height := int(binary.LittleEndian.Uint32(data[22:]))
	if (width != LetterWidth) || (height != LetterHeight) {
		return nil, fmt.Errorf("resolution mismatch %dx%d != %dx%d", width, height, LetterWidth, LetterHeight)
	}

	data = data[offset:]
	img := make(Letter, width*height)

	j := 0
	rj := height - 1
	lineSize := ((width*1 + 31) / 32) * 4
	for j < height {
		for i := 0; i < width; i++ {
			pixel := int8(data[rj*lineSize+(i/8)] & (0x80 >> (i % 8)))
			if pixel == 0 {
				pixel = -1
			} else {
				pixel = 1
			}
			img[j*width+i] = pixel
		}

		j++
		rj--
	}

	return img, nil
}

func MustDecode(letter Letter, err error) Letter {
	if err != nil {
		Fatalf("Failed to decode image: %s\n", err.Error())
	}
	return letter
}

func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func PrintLetter(letter Letter, width, height int) {
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			var symbol byte

			pixel := letter[j*width+i]
			if pixel < 0 {
				symbol = '0'
			} else {
				symbol = '1'
			}
			fmt.Printf(" %c", symbol)
		}
		fmt.Printf("\n")
	}
}

func main() {
	letters := []Letter{
		MustDecode(DecodeImage("LetterA.bmp")),
		MustDecode(DecodeImage("LetterV.bmp")),
		MustDecode(DecodeImage("LetterP.bmp")),
	}

	weights := make([]int8, LetterResolution*LetterResolution)
	for i := 0; i < LetterResolution; i++ {
		for j := 0; j < LetterResolution; j++ {
			if i != j {
				for k := 0; k < len(letters); k++ {
					weights[i*LetterResolution+j] += letters[k][i] * letters[k][j]
				}
			}
		}
	}

	tests := []Letter{
		MustDecode(DecodeImage("LetterA_test.bmp")),
		MustDecode(DecodeImage("LetterV_test.bmp")),
		MustDecode(DecodeImage("LetterP_test.bmp")),
		MustDecode(DecodeImage("LetterP_test_inv.bmp")),
	}

	const maxSteps = 5000
	outputs := make([]int8, LetterResolution)
	for i, test := range tests {
		step := 0

		for {
			if step > maxSteps {
				fmt.Fprintf(os.Stderr, "Failed to restore image %d: exceeded %d steps\n", i, maxSteps)
				break
			}

			for i := 0; i < LetterResolution; i++ {
				var output int8
				for j := 0; j < LetterResolution; j++ {
					output += weights[i*LetterResolution+j] * test[j]
				}
				if output < 0 {
					outputs[i] = -1
				} else {
					outputs[i] = 1
				}
			}

			if slices.Equal(test, outputs) {
				break
			} else {
				copy(test, outputs)
			}
			step++
		}

		fmt.Printf("Test %d: decoded letter in %d steps:\n", i, step+1)
		PrintLetter(outputs, LetterWidth, LetterHeight)
		fmt.Println()
	}
}
