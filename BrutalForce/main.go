package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var indSize int
var matrix [][]float64
var min float32

func testfunc(genome []int) {
	var rta float32 = 0.0

	for i := 0; i < len(genome); i++ {
		if i < len(genome)-1 {
			rta += float32(matrix[genome[i]][genome[i+1]])
		} else {

			rta += float32(matrix[genome[i]][genome[0]])
		}
	}
	if rta < min {
		min = rta
	}
}

func chargeTest() {
	file, err := os.Open("../setup/14.tsp")

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		indSize, err = strconv.Atoi(scanner.Text())
		if err != nil {
			log.Fatal(err)
		}
		break
	}
	matrix = make([][]float64, indSize)
	for i := 0; i < indSize; i++ {
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

func permute(s []int, emit func([]int)) {
	if len(s) == 0 {
		emit(s)
		return
	}
	var rc func(int)
	rc = func(np int) {
		if np == 1 {
			emit(s)
			return
		}
		np1 := np - 1
		pp := len(s) - np1
		rc(np1)
		for i := pp; i > 0; i-- {
			s[i], s[i-1] = s[i-1], s[i]
			rc(np1)
		}
		w := s[0]
		copy(s, s[1:pp+1])
		s[pp] = w
	}
	rc(len(s))
}

func main() {

	chargeTest()
	arr := make([]int, indSize)
	for i := 0; i < len(arr); i++ {
		arr[i] = i
	}
	min = 100000

	permute(arr, func(p []int) { go testfunc(p) })

	fmt.Println("done")
	fmt.Println(min)
}
