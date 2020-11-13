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

var popSize int
var generations int

var matrix [][]float64

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

	pS, err := strconv.Atoi(os.Args[1])
	if err != nil || pS < 1 {
		panic("First argument must be an integer > 0")
	}
	popSize = pS

	g, err := strconv.Atoi(os.Args[2])
	if err != nil || g < 1 {
		panic("Second argument must be an integer > 0")
	}
	generations = g

	chargeTest(os.Args[3])

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

	fo, err := os.Create(os.Args[1] + " " + os.Args[2] + " " + strconv.Itoa(indSize) + " " + os.Args[4] + ".txt")
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	fo.Write([]byte("\tbest\tworst\tmean\tmedian\tstDeviation"))
	for i := 0; i < generations*5; i++ {
		if i%5 == 0 {
			fo.Write([]byte("\n"))
		}
		fo.Write([]byte("\t"))
		fo.Write([]byte(fmt.Sprintf("%f", results[i])))
	}
	fmt.Printf("%f\t", elapsed.Seconds())
}
