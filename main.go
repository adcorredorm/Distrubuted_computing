package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cpmech/gosl/mpi"
)

var indSize int

const popSize = 128

var matrix [][]float64
var threads int
var generations int
var bestInd [popSize]float32
var stats [][]float32

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

func statitstics(agents ...Agent) []float32 {

	fitnesses := make([]float32, 0, popSize)

	var best, worst, median, mean, stdDv float32

	for _, agent := range agents {
		fitnesses = append(fitnesses, agent.fitness)
	}
	sort.Slice(fitnesses, func(i, j int) bool { return fitnesses[i] < fitnesses[j] })
	best = fitnesses[0]
	worst = fitnesses[popSize-1]
	median = fitnesses[int(popSize/2)]

	var total float32 = 0.0
	for _, fitness := range fitnesses {
		total += fitness
	}
	mean = total / popSize

	total2 := 0.0
	for _, fitness := range fitnesses {
		total2 += math.Pow(float64(fitness-mean), 2)
	}

	total2 /= popSize

	stdDv = float32(math.Sqrt(total2))

	return []float32{best, worst, mean, median, stdDv}
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

func evaluateGen(population *[popSize]Agent, offspring *[popSize]Agent, rate float32) {

	if rate < 0 || rate > 1 {
		panic("Crossover rate must be in [0, 1]")
	}

	mpi.Start()
	defer mpi.Stop()

	id := mpi.WorldRank()

	bash := int(popSize / mpi.WorldSize())
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
}

func main() {
	start := time.Now()

	th, err := strconv.Atoi(os.Args[1])
	if err != nil || th < 1 {
		panic("First argument must be an integer > 0")
	}
	threads = th

	gens, err := strconv.Atoi(os.Args[2])
	if err != nil || th < 1 {
		panic("First argument must be an integer > 0")
	}
	generations = gens

	chargeTest(os.Args[3])
	rand.Seed(time.Now().UnixNano())

	population := initPopulation(fitnessFunction)

	i := 0
	var offspring [popSize]Agent
	for i < generations {
		//fmt.Printf("Best %d: ", i)
		//getBest(population[:]...).PrintAgent()
		stats = append(stats, statitstics(population[:]...))

		for j := 0; j < threads; j++ {
			go evaluateGen(&population, &offspring, 0.7)
		}

		population = offspring
		i++
	}
	//fmt.Printf("Best %d: ", i)
	//getBest(population[:]...).PrintAgent()
	stats = append(stats, statitstics(population[:]...))

	elapsed := time.Since(start)

	fo, err := os.Create("resultsEvolution/" + strconv.Itoa(popSize) + " " +
		strconv.Itoa(generations) + " " +
		strconv.Itoa(indSize) + " " +
		strconv.Itoa(threads) + " " +
		os.Args[4] + ".txt")
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	fo.Write([]byte("\tbest\tworst\tmean\tmedian\tstDeviation\n"))
	for _, gen := range stats {
		fo.Write([]byte(fmt.Sprintf("\t%f\t%f\t%f\t%f\t%f\n", gen[0], gen[1], gen[2], gen[3], gen[4])))
	}

	fmt.Printf("%f\t", elapsed.Seconds())
}
