package main

import (
	"fmt"
	"time"

	"github.com/MikhailKlemin/aditya/pkg/queens/maincourses"
)

func main() {
	t := time.Now()
	//cs := wilfrid.Start()
	//wilfrid.Export(cs)
	maincourses.Start()
	//appliedcourses.Start()
	fmt.Printf("Took %s to finish\n", time.Since(t))
}
