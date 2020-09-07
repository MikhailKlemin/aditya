package appliedcourses

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/MikhailKlemin/aditya/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

var coursemap = map[string]string{
	"ANAT":  "Anatomy of the Human Body",
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
//https://calendar.engineering.queensu.ca/ajax/preview_course.php?catoid=9&coid=5316&display_options=a:2:{s:8:~location~;s:8:~template~;s:28:~course_program_display_field~;N;}')&show

//CourseExample is
type CourseExample struct {
	SubjectID       string `json:"subjectId,omitempty"`
	Name            string `json:"name"`
	NumericCode     string `json:"numericCode"`
	CourseCode      string `json:"courseCode"`
	Description     string `json:"description"`
	Prerequisite    string `json:"prerequisite"`
	Antirequisite   string `json:"antirequisite"`
	OneWayExclusion string `json:"oneWayExclusion,omitempty"`
}

//Subject holds terms
type Subject struct {
	SubjectID   int      `json:"subjectId"`
	TermID      int      `json:"termID,omitempty"`
	SubjectName string   `json:"name"`
	SubjectCode []string `json:"codes"`
}

//Start starts appliedcourses @queens
func Start() {
	client := utils.GetClient()

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
	ubase, err := url.ParseRequestURI("https://calendar.engineering.queensu.ca/")
	if err != nil {
		log.Fatal(err)
	}
	links := []string{}
	//var courses [][]string
	doc.Find(`a[href^="preview_course_nopop.php?catoid="]`).Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr(`href`)
		onclick := `a:2:{s:8:~location~;s:8:~template~;s:28:~course_program_display_field~;N;})`
		u, err := url.Parse(href)
		if err != nil {
			log.Fatal(err)
		}
		u = ubase.ResolveReference(u)
		q := u.Query()
		coID := q.Get("coid")
		catoID := q.Get("catoid")
		link := fmt.Sprintf("https://calendar.engineering.queensu.ca/ajax/preview_course.php?catoid=%s&coid=%s&display_options=%s&show", catoID, coID, onclick)
		//fmt.Println(link)
		links = append(links, link)
		//fmt.Println(onclick)

	})

	for i := 2; i <= pages; i++ {
		link = "https://calendar.engineering.queensu.ca/content.php?filter%5B27%5D=-1&filter%5B29%5D=&filter%5Bcourse_type%5D=-1&filter%5Bkeyword%5D=&filter%5B32%5D=1&filter%5Bcpage%5D=" + strconv.Itoa(i) + "&cur_cat_oid=9&expand=&navoid=233&search_database=Filter#acalog_template_course_filter"
		ls, err := readit(link)
		if err != nil {
			log.Println(err)
			continue
		}
		links = append(links, ls...)
	}

	fmt.Printf("Total %d links\n", len(link))
	var cs []CourseExample
	for _, link := range links {
		c, err := Parse(link)
		if err != nil {
			log.Println(err)
			continue
		}
		cs = append(cs, c)
	}

	if err := export(cs); err != nil {
		log.Fatal(err)
	}

}

func readit(link string) (links []string, err error) {
	client := utils.GetClient()

	ubase, err := url.ParseRequestURI("https://calendar.engineering.queensu.ca/")
	if err != nil {
		//log.Fatal(err)
		return
	}

	resp, err := client.Get(link)
	if err != nil {
		//log.Fatal(err)
		return
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	doc.Find(`a[href^="preview_course_nopop.php?catoid="]`).Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr(`href`)
		onclick := `a:2:{s:8:~location~;s:8:~template~;s:28:~course_program_display_field~;N;})`
		u, err := url.Parse(href)
		if err != nil {
			log.Fatal(err)
		}
		u = ubase.ResolveReference(u)
		q := u.Query()
		coID := q.Get("coid")
		catoID := q.Get("catoid")
		link := fmt.Sprintf("https://calendar.engineering.queensu.ca/ajax/preview_course.php?catoid=%s&coid=%s&display_options=%s&show", catoID, coID, onclick)
		//fmt.Println(link)
		links = append(links, link)

		//fmt.Println(onclick)

	})

	return

}

//Parse a link
func Parse(link string) (c CourseExample, err error) {
	//https://calendar.engineering.queensu.ca/ajax/preview_course.php?catoid=9&coid=5315&display_options=a:2:{s:8:~location~;s:8:~template~;s:28:~course_program_display_field~;N;})&show
	//var c CourseExample
	client := utils.GetClient()
	resp, err := client.Get(link)

	if err != nil {
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	//fmt.Println(string(b))
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if err != nil {
		return
	}

	h3 := strings.TrimSpace(doc.Find(`h3`).Text())

	m := regexp.MustCompile(`^([A-Z]+)\s*(\d+)\s*(.*)\s+(?:F|W)`).FindStringSubmatch(h3)
	if len(m) > 0 {
		c.CourseCode = m[1]
		c.NumericCode = m[2]
		c.Name = m[3]
	}
	//Tutorial:[\s\d\.]+(.*)\s*Academic Units:

	m = regexp.MustCompile(`Tutorial:[\s\d\.]+(.*)\s*Academic Units:`).FindStringSubmatch(doc.Text())
	if len(m) > 0 {
		c.Description = clean(m[1])
	}

	//PREREQUISITE\(S\):\s*(.*?)EXCLUSION\(S\)
	m = regexp.MustCompile(`PREREQUISITE\(S\):\s*(.*?)EXCLUSION\(S\)`).FindStringSubmatch(doc.Text())
	if len(m) > 0 {
		c.Prerequisite = clean(m[1])
	}

	if c.Prerequisite == "" {
		m = regexp.MustCompile(`(?m)PREREQUISITE\(S\):\s*(.*?)$`).FindStringSubmatch(doc.Text())
		if len(m) > 0 {
			c.Prerequisite = clean(m[1])
		}
	}

	//EXCLUSION\(S\):\s*(.*?)$

	m = regexp.MustCompile(`(?m)EXCLUSION\(S\):\s*(.*?)$`).FindStringSubmatch(doc.Text())
	if len(m) > 0 {
		c.Antirequisite = clean(m[1])
	}

	fmt.Printf("%#v\n", c)
	//fmt.Println(doc.Text())
	return
}

func clean(in string) string {
	out := regexp.MustCompile("\u00a0").ReplaceAllString(in, " ")
	out = regexp.MustCompile(`\s+`).ReplaceAllString(out, " ")
	return strings.TrimSpace(out)
}

func export(cs []CourseExample) (err error) {
	var countermap = make(map[string]int)
	var sbjs []Subject
	counter := 0
	for key, val := range coursemap {
		var se Subject
		se.SubjectID = counter
		se.SubjectName = val
		se.SubjectCode = []string{key}
		sbjs = append(sbjs, se)
		countermap[key] = counter
		counter++
	}
	b, err := json.MarshalIndent(sbjs, "", "    ")
	if err != nil {
		return
	}
	err = ioutil.WriteFile("./assets/QU-applied-subjects.json", b, 0600)
	if err != nil {
		return
	}

	for i := 0; i < len(cs); i++ {
		if cs[i].CourseCode == "" {
			continue
		}
		xc, ok := countermap[cs[i].CourseCode]
		if ok {
			cs[i].SubjectID = strconv.Itoa(xc)
		}
	}

	b, err = json.MarshalIndent(cs, "", "    ")
	if err != nil {
		return
	}
	err = ioutil.WriteFile("./assets/QU-applied-courses.json", b, 0600)
	if err != nil {
		return
	}

	return
}
