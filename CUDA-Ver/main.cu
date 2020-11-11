#include <iostream>
#include <stdio.h>
#include <cuda.h>
#include <cstdlib>
#include <curand_kernel.h>
#include <bits/stdc++.h>

using namespace std;

const int numberNodes = 14;

struct Agent{
    int size;
    float fitness;
    int genome[numberNodes];
};

/* Arrange the N elements of ARRAY in random order.
   Only effective if N is much smaller than RAND_MAX;
   if this may not be the case, use a better random
   number generator. */
void shuffle(int *array, size_t n)
{
    if (n > 1)
    {
        size_t i;
        for (i = 0; i < n - 1; i++)
        {
            size_t j = i + rand() / (RAND_MAX / (n - i) + 1);
            int t = array[j];
            array[j] = array[i];
            array[i] = t;
        }
    }
}

Agent RandomAgent(int size){
    int g[size];
    for (int i = 0; i < size ; i++){
        g[i] = i;
    }
    shuffle(g,size);
    struct Agent RAgent{};

    RAgent.size = size;
    RAgent.fitness = 0.0;
    for (int i = 0; i < size ; i++){
        RAgent.genome[i] = g[i];
    }
    return RAgent;
}

__device__ Agent NewAgent (const int genome[]){
    struct Agent nw{};
    nw.size = numberNodes;
    nw.fitness = 0.0;
    for (int i = 0; i < numberNodes ; i++){
        nw.genome[i] = genome[i];
    }
    return nw;
}

__device__ void Mutate(struct Agent *agent, curandState state) {
    unsigned int p1 =  curand(&state) % numberNodes;
    while (true) {
        unsigned int p2 = curand(&state) % numberNodes;
        if (p1 != p2) {
            int temp = agent->genome[p1];
            agent->genome[p1] = agent->genome[p2];
            agent->genome[p2] = temp;
            break;
        }
    }
}

__device__ int Find(const int arr[], int e, unsigned int size){
    for (int i = 0 ; i < size ;i++){
        if (arr[i]== e){
            return e;
        }
    }
    return -1;
}

__device__ void CrossPermutation(int a[], int b[], curandState state){
    unsigned int crossPoint = curand(&state)%(numberNodes-2)+1;
    unsigned int tempSize = numberNodes-crossPoint-1;
    int tempA[numberNodes];
    int tempB[numberNodes];
    int k = 0;
    for (unsigned int i=crossPoint+1;i<numberNodes;i++){
        tempA[k] = a[i];
        tempB[k] = b[i];
        k++;
    }
    unsigned int idenA = crossPoint + 1;
    unsigned int idenB = crossPoint + 1;
    for (int i = 0;i < numberNodes ; i++) {
        if (Find(tempA, b[i], tempSize) != -1) {
            a[idenA] = b[i];
            idenA++;
        }
    }
    for (int i = 0;i < numberNodes ; i++) {
        if (Find(tempB, a[i], tempSize) != -1) {
            b[idenB] = a[i];
            idenB++;
        }
    }

};

__device__ void FitnessFunction(struct Agent *agent, const float *distance){
    float fitness = 0;
    for (int i = 0; i < numberNodes ; i++){
        if (i < numberNodes - 1 ){
            fitness += distance[agent->genome[i]+agent->genome[i+1]*numberNodes];
        }else{
            fitness += distance[agent->genome[i]+agent->genome[0]*numberNodes];
        }
    }
    agent->fitness = fitness;
}

__device__ Agent GetBest(struct Agent a1,struct Agent a2,struct Agent a3){
    if(a1.fitness < a2.fitness && a1.fitness < a3.fitness){
        return a1;
    }else if (a2.fitness < a3.fitness){
        return a2;
    }else{
        return a3;
    }
}

void PrintAgent(struct Agent agent){
    cout<<agent.fitness<< " ";
    for (int i : agent.genome){
        cout<< i << " ";
    }
    cout << endl;
}

__global__ void EvaluateGen(float *DDistance, struct Agent *DIPopulation, struct Agent *DFPopulation, int popSize, float rate) {
    unsigned int tId = threadIdx.x + (blockIdx.x * blockDim.x);
    curandState state;
    curand_init((unsigned long long)clock() , tId, 0, &state);
    if(tId < popSize){
        FitnessFunction(&DIPopulation[tId], DDistance);
        if(curand_uniform(&state) < rate){
            unsigned int pair =  curand(&state)%popSize;
            int n1[numberNodes];
            int n2[numberNodes];
            for (int i=0;i<numberNodes;i++){
                n1[i] = DIPopulation[pair].genome[i];
                n2[i] = DIPopulation[tId].genome[i];
            }
            CrossPermutation(n1,n2,state);
            struct Agent a1 = NewAgent(n1);
            struct Agent a2 = NewAgent(n2);
            Mutate(&a1, state);
            Mutate(&a2, state);
            FitnessFunction(&a1,DDistance);
            FitnessFunction(&a2,DDistance);
            DFPopulation[tId] = GetBest(a1,a2,DIPopulation[tId]);
        }else{
            DFPopulation[tId] = DIPopulation[tId];
        }
    }

}

void CondensedResult(float current[], float *results, float mean, int popSize, int generation){
    float median = 0, best = 0, worst = 0, stDeviation = 0;
    sort(current,current +popSize);

    best = current[0];
    worst = current[popSize-1];
    if (popSize % 2 == 0){
        median = (current[int(popSize/2)]+current[int((popSize/2)+1)])/2;
    } else{
        median = current[int((popSize/2)+1)];
    }
    for(int i = 0;i<popSize;i++){
        stDeviation += pow(current[i]-mean,2);
    }
    stDeviation /= popSize;
    stDeviation = sqrt(stDeviation);
    results[generation*5] = best;
    results[generation*5+1] = worst;
    results[generation*5+2] = mean;
    results[generation*5+3] = median;
    results[generation*5+4] = stDeviation;
}

extern "C" {
    void evaluateGen(float *distance, float *results, int popSize, int generations, float rate) {
        srand(time(nullptr));
        struct Agent population[popSize];
        for (int i = 0; i< popSize;i++){
            population[i] = RandomAgent(numberNodes);

        }
        unsigned long DDistanceSize = (numberNodes*numberNodes)*sizeof(float);
        float* DDistance;
        cudaMalloc((void**)&DDistance,DDistanceSize);
        cudaMemcpy(DDistance,distance,DDistanceSize,cudaMemcpyHostToDevice);

        unsigned long DPopulationSize = (sizeof(int)+sizeof(float)+(sizeof(int)*numberNodes))*popSize;
        struct Agent* DIPopulation;
        cudaMalloc((void**)&DIPopulation,DPopulationSize);

        struct Agent* DFPopulation;
        cudaMalloc((void**)&DFPopulation,DPopulationSize);
        for (int i = 0; i < generations;i++){
            float popResults[popSize];
            float mean = 0.0;

            cudaMemcpy(DIPopulation,population,DPopulationSize,cudaMemcpyHostToDevice);

            EvaluateGen<<<256,int(popSize/256)+1>>>(DDistance, DIPopulation, DFPopulation, popSize, rate);

            cudaMemcpy(population,DFPopulation,DPopulationSize,cudaMemcpyDeviceToHost);

            for (int j = 0; j< popSize;j++){
                mean += population[j].fitness;
                popResults[j] = population[j].fitness;
            }
            mean /= popSize;
            CondensedResult(popResults,results,mean, popSize,i);
        }
        cudaFree(DDistance);
        cudaFree(DIPopulation);
        cudaFree(DFPopulation);

    }
}


