package main

import (
	"fmt"
	"time"
)

func main() {
	t := time.Now()
	//cs := wilfrid.Start()
	//wilfrid.Export(cs)
	//maincourses.Start()
	//appliedcourses.Start()
	//appliedcourses.Export([]appliedcourses.CourseExample{})
	//appliedcourses.Parse("https://calendar.engineering.queensu.ca/ajax/preview_course.php?catoid=9&coid=5315&display_options=a:2:{s:8:~location~;s:8:~template~;s:28:~course_program_display_field~;N;})&show")
	fmt.Printf("Took %s to finish\n", time.Since(t))
}
