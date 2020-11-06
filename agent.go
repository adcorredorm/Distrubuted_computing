package main

import (
	"fmt"
	"math/rand"
)

//Agent type
type Agent struct {
	size    int
	genome  []int
	fitness float32
}

//RandomAgent returns a new random agent
func RandomAgent(size int) Agent {
	g := make([]int, size)
	for i := 0; i < size; i++ {
		g[i] = i
	}

	rand.Shuffle(len(g), func(i, j int) { g[i], g[j] = g[j], g[i] })

	return Agent{
		size:    size,
		genome:  g,
		fitness: 0.0,
	}
}

//NewAgent returns an agent from a given genome
func NewAgent(genome []int) Agent {
	return Agent{
		size:    len(genome),
		genome:  genome,
		fitness: 0.0,
	}
}

//Evaluate an Agent
func (agent *Agent) Evaluate(function func([]int) float32) {
	agent.fitness = function(agent.genome)
}

//Mutate alter a single agent's genome
func Mutate(agent *Agent) {

	p1 := rand.Intn(agent.size)
	for {
		p2 := rand.Intn(agent.size)
		if p1 != p2 {
			agent.genome[p1], agent.genome[p2] = agent.genome[p2], agent.genome[p1]
			break
		}
	}
}

func find(collection []int, element int) int {
	for i, e := range collection {
		if e == element {
			return i
		}
	}
	return -1
}

//crossPermitation mixes 2 given permutations into a new valid permutation
func crossPermutation(p1 []int, p2 []int, point int) []int {
	result := make([]int, point, len(p1))
	copy(result, p1)

	for len(result) < len(p1) {
		if find(result, p2[point]) < 0 {
			result = append(result, p2[point])
		}
		point = (point + 1) % len(p1)
	}
	return result
}

//Crossover takes 2 agents to produce an offspring
func Crossover(p1 *Agent, p2 *Agent) (Agent, Agent) {
	crossPoint := rand.Intn(p1.size-2) + 1

	a1 := NewAgent(crossPermutation(p1.genome, p2.genome, crossPoint))
	a2 := NewAgent(crossPermutation(p2.genome, p1.genome, crossPoint))

	return a1, a2
}

//PrintAgent to console for quick check
func (agent Agent) PrintAgent() {
	fmt.Printf("%v %0.2f\n", agent.genome, agent.fitness)
}
