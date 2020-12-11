build:
	go build -o main main.go agent.go
	mpirun -np 2 --hostfile hostfile main 4 100 setup/14.tsp 0 

clean:
	rm main
