package main

import (
	"fmt"
	"time"

	"github.com/MikhailKlemin/aditya/pkg/wilfrid"
)

func main() {
	t := time.Now()
	cs := wilfrid.Start()
	wilfrid.Export(cs)
	fmt.Printf("Took %s to finish\n", time.Since(t))
}
