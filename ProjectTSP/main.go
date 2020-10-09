package main

import (
	"fmt"
	"math/rand"
	"time"
	"bufio"
	"os"
	"log"
	"strconv"
	"strings"
)

var indSize int
const popSize = 20
const generations = 50
var matrix [][]float64

func testfunc(genome []int) float32 {
	var rta float32 = 0.0
	for i :=0 ; i < len(genome) ; i++ {
		if i < len(genome)-1{
			rta += float32(matrix[genome[i]][genome[i+1]])
		}else{
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

func chargeTest(){
	file,err := os.Open("./setup/100.tsp")
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
		for k := 0 ; k < len(arr) ; k++ {
			temp,err := strconv.ParseFloat(arr[k],64)
			if err != nil {
				log.Fatal(err)
			}	
			matrix[j][k] = temp
		}
		j++
	}
}

func main() {

	chargeTest()
	rand.Seed(time.Now().UnixNano())
	population := initPopulation(testfunc)

	i := 0
	var offspring [popSize]Agent
	for i < generations {
		fmt.Printf("Best %d: ", i)
		getBest(population[:]...).PrintAgent()
		for j := range population {
			if rand.Float32() < 1 {
				pair := rand.Intn(popSize)
				n1, n2 := Crossover(&population[j], &population[pair])
				n1.Evaluate(testfunc)
				n2.Evaluate(testfunc)
				best := getBest(n1, n2)
				Mutate(&best)
				best.Evaluate(testfunc)
				offspring[j] = getBest(population[j], best)
			} else {
				offspring[j] = population[j]
			}
		}
		population = offspring
		i++
	}
	fmt.Printf("Best %d: ", i)
	getBest(population[:]...).PrintAgent()
}

