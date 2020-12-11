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

type agentGroup struct {
	gen [][]int
	fit []float64
}

var indSize int

const popSize = 8

var matrix [][]float64
var threads int
var generations int
var bestInd [popSize]float64
var stats [][]float64

var population agentGroup

func fitnessFunction(genome []int) float64 {
	var rta float64 = 0.0

	for i := 0; i < len(genome); i++ {
		if i < len(genome)-1 {
			rta += float64(matrix[genome[i]][genome[i+1]])
		} else {
			rta += float64(matrix[genome[i]][genome[0]])
		}
	}
	return rta
}

func createRandomAgents(n int) agentGroup {
	var agents Agent
	inds := make([][]int, 0)
	fit := make([]float64, 0)
	for i := 0; i < n; i++ {
		agents = RandomAgent(indSize)
		agents.Evaluate(fitnessFunction)
		inds = append(inds, agents.genome)
		fit = append(fit, agents.fitness)
	}
	return agentGroup{gen: inds, fit: fit}
}

func recvAgents(src int, bash int, ch chan agentGroup) {
	comm := mpi.NewCommunicator(nil)

	genomes := make([][]int, bash)
	fitness := make([]float64, bash)
	for i := 0; i < bash; i++ {
		genomes[i] = make([]int, indSize)
		comm.RecvI(genomes[i], src)
		// fmt.Printf("Recieved: %v from %d\n", genomes[i], src)
	}
	comm.Recv(fitness, src)
	// fmt.Printf("Recieved: %v from %d\n", fitness, src)

	ch <- agentGroup{gen: genomes, fit: fitness}
}

func recvPopulation(firstAgents agentGroup, nodes int, bash int) {
	if mpi.WorldRank() == 0 {
		population = firstAgents
		ch := make(chan agentGroup, nodes)

		for i := 1; i < nodes; i++ {
			go recvAgents(i, bash, ch)
		}
		for i := 1; i < nodes; i++ {
			newAgents := <-ch
			population.gen = append(population.gen, newAgents.gen...)
			population.fit = append(population.fit, newAgents.fit...)
		}
	} else {
		comm := mpi.NewCommunicator(nil)
		population.gen = make([][]int, popSize)
		population.fit = make([]float64, popSize)

		for j := 0; j < popSize; j++ {
			population.gen[j] = make([]int, indSize)
			comm.RecvI(population.gen[j], 0)
			// fmt.Printf("Recieved %v in %d\n", population.gen[j], mpi.WorldRank())
		}
		comm.Recv(population.fit, 0)
		// fmt.Printf("Recieved %v in %d\n", population.fit, mpi.WorldRank())
	}
}

func sendPopulation() {
	comm := mpi.NewCommunicator(nil)

	for i := 1; i < mpi.WorldSize(); i++ {
		for j := 0; j < popSize; j++ {
			// fmt.Printf("Sending %v to %d\n", population.gen[j], i)
			comm.SendI(population.gen[j], i)
		}
		// fmt.Printf("Sending %v to %d\n", population.fit, i)
		comm.Send(population.fit, i)
	}
}

func sendAgents(agents agentGroup) {
	comm := mpi.NewCommunicator(nil)
	for i := 0; i < len(agents.gen); i++ {
		// fmt.Printf("Sending %v from %d\n", agents.gen[i], mpi.WorldRank())
		comm.SendI(agents.gen[i], 0)
	}
	// fmt.Printf("Sending %v from %d\n", agents.fit, mpi.WorldRank())
	comm.Send(agents.fit, 0)
}

func initPopulation(function func([]int) float64) {
	nodes := mpi.WorldSize()
	bash := popSize / nodes
	agents := createRandomAgents(bash)
	comm := mpi.NewCommunicator(nil)

	if mpi.WorldRank() == 0 {
		recvPopulation(agents, nodes, bash)
	} else {
		sendAgents(agents)
	}
	comm.Barrier()
	// fmt.Printf("After Barrier %d\n", mpi.WorldRank())
	if mpi.WorldRank() == 0 {
		sendPopulation()
	} else {
		recvPopulation(agentGroup{}, -1, -1)
	}

	fmt.Printf("Node %d agent 0", mpi.WorldRank())
	Agent{size: 14, genome: population.gen[0], fitness: population.fit[0]}.PrintAgent()
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

func statitstics(agents ...Agent) []float64 {

	fitnesses := make([]float64, 0, popSize)

	var best, worst, median, mean, stdDv float64

	for _, agent := range agents {
		fitnesses = append(fitnesses, agent.fitness)
	}
	sort.Slice(fitnesses, func(i, j int) bool { return fitnesses[i] < fitnesses[j] })
	best = fitnesses[0]
	worst = fitnesses[popSize-1]
	median = fitnesses[int(popSize/2)]

	var total float64 = 0.0
	for _, fitness := range fitnesses {
		total += fitness
	}
	mean = total / popSize

	total2 := 0.0
	for _, fitness := range fitnesses {
		total2 += math.Pow(float64(fitness-mean), 2)
	}

	total2 /= popSize

	stdDv = float64(math.Sqrt(total2))

	return []float64{best, worst, mean, median, stdDv}
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

func evaluateGen(population *[popSize]Agent, offspring *[popSize]Agent, rate float64) {

	if rate < 0 || rate > 1 {
		panic("Crossover rate must be in [0, 1]")
	}

	id := mpi.WorldRank()
	ws := mpi.WorldSize()

	bash := int(popSize / ws)
	init := id * bash
	end := init + bash
	if id == ws-1 {
		end = popSize
	}

	fmt.Printf("Hello from %d: init %d, end: %d\n", id, init, end)

	for j := init; j < end; j++ {
		if rand.Float64() < rate {
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
	mpi.Start()
	defer mpi.Stop()

	// start := time.Now()

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

	initPopulation(fitnessFunction)

	// i := 0
	// var offspring [popSize]Agent

	// for i < generations {
	// 	population[0].PrintAgent()
	// fmt.Printf("Best %d: ", i)
	// getBest(population[:]...).PrintAgent()
	// stats = append(stats, statitstics(population[:]...))

	// evaluateGen(&population, &offspring, 0.7)

	// population = offspring
	// 	i++
	// 	comm.Barrier()
	// }
	// fmt.Printf("Best %d: ", i)
	// getBest(population[:]...).PrintAgent()
	// stats = append(stats, statitstics(population[:]...))

	// elapsed := time.Since(start)

	// fo, err := os.Create("resultsEvolution/" + strconv.Itoa(popSize) + " " +
	// 	strconv.Itoa(generations) + " " +
	// 	strconv.Itoa(indSize) + " " +
	// 	strconv.Itoa(threads) + " " +
	// 	os.Args[4] + ".txt")
	// if err != nil {
	// 	panic(err)
	// }
	// defer fo.Close()
	// fo.Write([]byte("\tbest\tworst\tmean\tmedian\tstDeviation\n"))
	// for _, gen := range stats {
	// 	fo.Write([]byte(fmt.Sprintf("\t%f\t%f\t%f\t%f\t%f\n", gen[0], gen[1], gen[2], gen[3], gen[4])))
	// }

	// fmt.Printf("%f\t", elapsed.Seconds())
}
