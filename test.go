package main

import (
	"fmt"

	"github.com/cpmech/gosl/mpi"
)

func maint() {
	mpi.Start()
	defer mpi.Stop()
	fmt.Printf("Hello World! from %d of %d\n", mpi.WorldRank(), mpi.WorldSize())
}
