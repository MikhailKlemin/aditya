package appliedcourses

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var coursemap = map[string]string{
	"APSC":  "Applied Science",
	"BCHM":  "Biochemistry",
	"BIOL":  "Biology",
	"CHEE":  "Chemical Engineering",
	"CIVIL": "Civil Engineering",
	"CMPE":  "Computer Engineering",
	"ELEC":  "Electrical Engineering",
	"ENCH":  "Engineering Chemistry",
	"ENPH":  "Engineering Physics",
	"GEOE":  "Geological Engineering",
	"GEOL":  "Geology",
	"GISC":  "Geographic Information Science",
	"GPHY":  "Geography",
	"MBIO":  "Microbiology",
	"MDEP":  "Multi-department Courses",
	"MECH":  "Mechanical Engineering",
	"MINE":  "Mining Engineering",
	"MNTC":  "Mining Technology",
	"MTHE":  "Mathematics and Engineering",
	"SURP":  "School of Urban and Regional Planning",
}

//preview_course_nopop.php?catoid=9&coid=5765
//showCourse('9', '5765',this, 'a:2:{s:8:~location~;s:8:~template~;s:28:~course_program_display_field~;N;}'); return false;
//https://calendar.engineering.queensu.ca/ajax/preview_course.php?catoid=9&coid=5765&display_options=a:2:{s:8:~location~;s:8:~template~;s:28:~course_program_display_field~;N;}&show

//Start starts appliedcourses @queens
func Start() {
	client := getClient()

	link := "https://calendar.engineering.queensu.ca/content.php?filter%5B27%5D=-1&filter%5B29%5D=&filter%5Bcourse_type%5D=-1&filter%5Bkeyword%5D=&filter%5B32%5D=1&filter%5Bcpage%5D=1&cur_cat_oid=9&expand=&navoid=233&search_database=Filter#acalog_template_course_filter"
	//aink := "https://calendar.engineering.queensu.ca/content.php?catoid=9&navoid=233&filter%5B27%5D=-1&filter%5B29%5D=&filter%5Bcourse_type%5D=-1&filter%5Bkeyword%5D=&filter%5B32%5D=1&filter%5Bcpage%5D=2&filter%5Bitem_type%5D=3&filter%5Bonly_active%5D=1&filter%5B3%5D=1#acalog_template_course_filter"

	resp, err := client.Get(link)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	var pages int
	doc.Find(`td`).EachWithBreak(func(_ int, s *goquery.Selection) bool {
		txt := strings.TrimSpace(s.Text())
		if strings.HasPrefix(txt, "Page") {
			//fmt.Println(txt)
			pages, err = strconv.Atoi(s.Find(`a`).Last().Text())
			if err != nil {
				log.Fatal(err)
			}
			return false
		}
		return true
	})

	fmt.Println("Total Pages\t", pages)
	//var courses [][]string
	doc.Find(`a[href="preview_course_nopop.php?catoid="]`).Each(func(_ int, s *goquery.Selection) {

	})

}

func getClient() *http.Client {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 15 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 15 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 30,
		Transport: netTransport,
	}
	return netClient
}
