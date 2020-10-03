package main

import (
	"fmt"
	"log"
	"time"

	"github.com/MikhailKlemin/aditya/pkg/unis/concordia"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	t := time.Now()
	//cs := wilfrid.Start()
	//wilfrid.Export(cs)
	//maincourses.Start()
	//appliedcourses.Start()
	//appliedcourses.Export([]appliedcourses.CourseExample{})
	//appliedcourses.Parse("https://calendar.engineering.queensu.ca/ajax/preview_course.php?catoid=9&coid=5315&display_options=a:2:{s:8:~location~;s:8:~template~;s:28:~course_program_display_field~;N;})&show")
	//commerce.Start()
	//commerce.Parse(nil)

	//concurent.Start()
	//ubc.Start()
	//commerce.Start()
	//mcgill.Start()
	//mcgill.ToCSV()
	concordia.Start()
	fmt.Printf("Took %s to finish\n", time.Since(t))

}
