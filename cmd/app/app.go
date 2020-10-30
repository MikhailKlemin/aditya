package main

import (
	"fmt"
	"log"
	"time"

	"github.com/MikhailKlemin/aditya/pkg/unis/seneca"
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
	seneca.Start()
	//seneca.ParseCourseLink([]string{"Crap", "https://apps.senecacollege.ca/ssos/findOutline.do?subjectCode=YKC100"})
	//seneca.Start()
	//sheridan.Start()
	//sheridan.Parse("coutlineview.jsp?appver=ba&subjectCode=VISM&courseCode=4013&version=3.0&sec=0&reload=true")
	//sheridan.Start()
	fmt.Printf("Took %s to finish\n", time.Since(t))

}
