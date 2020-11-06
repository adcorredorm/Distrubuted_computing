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

var indSize int

const popSize = 128
const generations = 100

var matrix [][]float64
var threads int

func fitnessFunction(genome []int) float32 {
	var rta float32 = 0.0

	for i := 0; i < len(genome); i++ {
		if i < len(genome)-1 {
			rta += float32(matrix[genome[i]][genome[i+1]])
		} else {
			rta += float32(matrix[genome[i]][genome[0]])
		}
	}
	return rta
}

func initPopulation(function func([]int) float32) [popSize]Agent {
	var population [popSize]Agent
	for i := range population {
		population[i] = RandomAgent(indSize)
		population[i].Evaluate(function)
	}
	return population
}

func getBest(agents ...Agent) Agent {
	best := agents[0]
	for _, agent := range agents {
		if agent.fitness < best.fitness {
			best = agent
		}
	}
	return best
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

func evaluateGen(population *[popSize]Agent, offspring *[popSize]Agent,
	id int, rate float32, ch chan int) {

	if rate < 0 || rate > 1 {
		panic("Crossover rata must be in [0, 1]")
	}

	bash := int(popSize / threads)
	init := id * bash
	end := init + bash
	if id == threads-1 {
		end = popSize
	}

	for j := init; j < end; j++ {
		if rand.Float32() < rate {
			pair := rand.Intn(popSize)
			n1, n2 := Crossover(&population[j], &population[pair])
			Mutate(&n1)
			Mutate(&n2)
			n1.Evaluate(fitnessFunction)
			n2.Evaluate(fitnessFunction)
			offspring[j] = getBest(population[j], n1, n2)
		} else {
			offspring[j] = population[j]
		}
	}

	ch <- 0
}

func main() {
	start := time.Now()

	th, err := strconv.Atoi(os.Args[1])
	if err != nil || th < 1 {
		panic("First argument must be an integer > 0")
	}
	threads = th
	chargeTest(os.Args[2])
	rand.Seed(time.Now().UnixNano())

	population := initPopulation(fitnessFunction)

	i := 0
	var offspring [popSize]Agent
	for i < generations {
		//fmt.Printf("Best %d: ", i)
		//getBest(population[:]...).PrintAgent()

		ch := make(chan int, threads)

		for j := 0; j < threads; j++ {
			go evaluateGen(&population, &offspring, j, 0.7, ch)
		}

		for j := 0; j < threads; j++ {
			<-ch
		}

		population = offspring
		i++
	}
	//fmt.Printf("Best %d: ", i)
	//getBest(population[:]...).PrintAgent()
	elapsed := time.Since(start)
	fmt.Printf("Time: %f\n", elapsed.Seconds())
}
