package main

import (
	"fmt"
	"math"
)

func main() {
	var cylinderRadius, cylinderHeight float64
	for {
		fmt.Print("Type in radius and height of cylinder (space separated): ")
		_, err := fmt.Scanf("%g %g", &cylinderRadius, &cylinderHeight)
		if err != nil {
			fmt.Println("Please type in both cylinder parameters!")
		} else {
			break
		}
	}
	prismBaseSideLength := math.Sqrt(3) * cylinderRadius
	prismBaseArea := math.Sqrt(3) * prismBaseSideLength * prismBaseSideLength / 4
	prismSideArea := 3 * prismBaseSideLength * cylinderHeight
	prismFullArea := 2*prismBaseArea + prismSideArea
	prismVolume := prismBaseArea * cylinderHeight
	fmt.Printf("Prism volume = %.6g, prism full area = %.6g, prism side area = %.6g\n", prismVolume, prismFullArea, prismSideArea)
}
