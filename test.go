package main

import (
	"fmt"

	"github.com/cpmech/gosl/mpi"
)

func mpiF() {
	mpi.Start()
	defer mpi.Stop()

	fmt.Println("Inicie")
}

func maintest() {

	for i := 0; i < 5; i++ {
		mpiF()
	}

}
