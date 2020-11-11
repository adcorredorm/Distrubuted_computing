package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
void evaluateGen(float *distance, float *results, int popSize, int generations, float rate);
#cgo LDFLAGS: -L. -L./ -lmain
*/
import "C"

var indSize int

const popSize = 1280
const generations = 100

var matrix [][]float64
var threads int

//EvaluateGen takes parameters of go a put it on c function
func EvaluateGen(distance []C.float, results []C.float, popSize int, generations int, rate float32) {
	C.evaluateGen(&distance[0], &results[0], C.int(popSize), C.int(generations), C.float(rate))
}

func chargeTest(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	indSize, err = strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatal(err)
	}

	matrix = make([][]float64, indSize)
	for i := range matrix {
		matrix[i] = make([]float64, indSize)
	}

	j := 0
	for scanner.Scan() {
		arr := strings.Fields(scanner.Text())

		for k := 0; k < len(arr); k++ {
			temp, err := strconv.ParseFloat(arr[k], 64)
			if err != nil {
				log.Fatal(err)
			}
			matrix[j][k] = temp
		}
		j++
	}
}

func main() {
	start := time.Now()

	th, err := strconv.Atoi(os.Args[1])
	if err != nil || th < 1 {
		panic("First argument must be an integer > 0")
	}
	threads = th
	chargeTest(os.Args[2])

	var position []C.float = make([]C.float, indSize*indSize)
	for i := 0; i < indSize; i++ {
		for j := 0; j < indSize; j++ {
			position[i+j*indSize] = C.float(matrix[i][j])
		}
	}
	var results []C.float = make([]C.float, generations*5)

	EvaluateGen(position, results, popSize, generations, 0.8)

	rand.Seed(time.Now().UnixNano())

	elapsed := time.Since(start)

	for i := 0; i < generations*5; i++ {
		if i%5 == 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("\t")
		fmt.Print(results[i])
	}
	fmt.Printf("%f\t", elapsed.Seconds())
}
