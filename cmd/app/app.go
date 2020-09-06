package main

import (
	"fmt"
	"time"

	"github.com/MikhailKlemin/aditya/pkg/queens"
)

func main() {
	t := time.Now()
	//cs := wilfrid.Start()
	//wilfrid.Export(cs)
	queens.Start()
	fmt.Printf("Took %s to finish\n", time.Since(t))
}
