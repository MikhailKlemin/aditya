package appliedcourses

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
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

//CourseExample is
type CourseExample struct {
	SubjectID   int      `json:"subjectId,omitempty"`
	SubjectName string   `json:"SubjectName,omitempty"`
	SubjectCode []string `json:"codes,omitempty"`

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
	//client := utils.GetClient()

	abbrs, err := GetAbbr()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", abbrs)
	//abbrs = []string{"CHEE"}
	cs, err := overAbbrs(abbrs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(cs))

	b, err := json.MarshalIndent(cs, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./assets/QU-applied-raw.json", b, 0600)
	if err != nil {
		log.Fatal(err)
	}

	err = Export(cs)
	if err != nil {
		log.Fatal(err)
	}

}

func overAbbrs(abbrs []string) (cs []CourseExample, err error) {
	client := utils.GetClient()

	ubase, err := url.ParseRequestURI("https://calendar.engineering.queensu.ca/")
	if err != nil {
		//log.Fatal(err)
		return
	}

	link := "https://calendar.engineering.queensu.ca/content.php?filter%%5B27%%5D=%s&filter%%5B29%%5D=&filter%%5Bcourse_type%%5D=-1&filter%%5Bkeyword%%5D=&filter%%5B32%%5D=1&filter%%5Bcpage%%5D=1&cur_cat_oid=9&expand=&navoid=233&search_database=Filter#acalog_template_course_filter"
	for i, abbr := range abbrs {
		/*resp, err := client.Get(fmt.Sprintf(link, abbr))
		if err != nil {
			log.Fatal(err)
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}*/
		var doc *goquery.Document
		counter := 0
		for {
			if counter > 10 {
				return cs, errors.New("Failed miserable")
			}

			resp, xerr := client.Get(fmt.Sprintf(link, abbr))

			if xerr != nil {
				//return c, err
				counter++
				continue
			}

			b, xerr := ioutil.ReadAll(resp.Body)
			if xerr != nil {
				counter++
				continue
			}

			//fmt.Println(string(b))
			doc, xerr = goquery.NewDocumentFromReader(bytes.NewReader(b))
			if xerr != nil {
				counter++
				continue
			}
			defer resp.Body.Close()
			break
		}

		subjT := doc.Find(`table.table_default td[colspan="2"] strong`).Last()
		subj := subjT.Text()
		//p := subjT.Parent().Parent().Parent().Parent()
		//fmt.Println(p.Html())
		doc.Find(`a[href^="preview_course_nopop.php?catoid="]`).Each(func(_ int, s *goquery.Selection) {
			fmt.Println(abbr, "\t", subj, "\t", s.Text())
			//var c CourseExample
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
			clink := fmt.Sprintf("https://calendar.engineering.queensu.ca/ajax/preview_course.php?catoid=%s&coid=%s&display_options=%s&show", catoID, coID, onclick)
			c, err := Parse(clink)
			if err != nil {
				log.Fatal(err)
			}
			c.SubjectID = i
			c.SubjectName = subj
			c.SubjectCode = []string{c.CourseCode}
			fmt.Printf("%#v\n", c)
			cs = append(cs, c)
		})

	}

	return
}

//GetAbbr reads abreviations
func GetAbbr() (abbrs []string, err error) {
	client := utils.GetClient()
	resp, err := client.Get("https://calendar.engineering.queensu.ca/content.php?filter%5B27%5D=-1&filter%5B29%5D=&filter%5Bcourse_type%5D=-1&filter%5Bkeyword%5D=&filter%5B32%5D=1&filter%5Bcpage%5D=1&cur_cat_oid=9&expand=&navoid=233&search_database=Filter#acalog_template_course_filter")
	if err != nil {
		return
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	doc.Find(`#courseprefix option`).Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}

		abbr, ok := s.Attr(`value`)
		if ok {
			abbrs = append(abbrs, abbr)
		}

	})

	return
}

//Parse a link
func Parse(link string) (c CourseExample, err error) {
	//https://calendar.engineering.queensu.ca/ajax/preview_course.php?catoid=9&coid=5315&display_options=a:2:{s:8:~location~;s:8:~template~;s:28:~course_program_display_field~;N;})&show
	//var c CourseExample
	client := utils.GetClient()
	var doc *goquery.Document
	counter := 0
	for {
		if counter > 10 {
			return c, errors.New("Failed miserable")
		}

		resp, xerr := client.Get(link)

		if xerr != nil {
			//return c, err
			counter++
			continue
		}

		b, xerr := ioutil.ReadAll(resp.Body)
		if xerr != nil {
			counter++
			continue
		}

		//fmt.Println(string(b))
		doc, xerr = goquery.NewDocumentFromReader(bytes.NewReader(b))
		if xerr != nil {
			counter++
			continue
		}
		defer resp.Body.Close()
		break
	}
	h, _ := doc.Html()

	h3 := strings.TrimSpace(doc.Find(`h3`).Text())

	m := regexp.MustCompile(`^([A-Z]+)\s*([A-Z]?\d+)\s*(.*?)\s*$`).FindStringSubmatch(h3)
	if len(m) > 0 {
		c.CourseCode = m[1]
		c.NumericCode = m[2]
		c.Name = m[3]
	}
	//Tutorial:[\s\d\.]+(.*)\s*Academic Units:

	/*m = regexp.MustCompile(`(?s)Tutorial:[\s\d\.Yes]+(.*)\s*Academic Units:`).FindStringSubmatch(doc.Text())
	if len(m) > 0 {
		c.Description = clean(m[1])
	}*/
	m = regexp.MustCompile(`(?s)Tutorial:.*?<br/>(.*?)(?:<br/>)+Academic`).FindStringSubmatch(h)
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

//Export -- exports into JSON
func Export(cs []CourseExample) (err error) {

	var subjs []Subject
	var se = make(map[string]bool)

	if len(cs) == 0 {
		b, xerr := ioutil.ReadFile("./assets/QU-applied-raw.json")
		if xerr != nil {
			//log.Fatal(err)
			return xerr
		}

		xerr = json.Unmarshal(b, &cs)
		if xerr != nil {
			//log.Fatal(err)
			return xerr
		}

	}
	var xcs []CourseExample
	xcounter := 0
	for _, c := range cs {
		if c.CourseCode == "" {
			continue
		}

		if _, ok := se[c.CourseCode]; !ok {
			//continue

			var subj Subject
			se[c.CourseCode] = true
			subj.SubjectID = c.SubjectID
			subj.SubjectName = c.SubjectName
			subj.SubjectCode = c.SubjectCode
			subjs = append(subjs, subj)

		}

		c.SubjectName = ""
		c.SubjectCode = []string{}
		c.Description = utils.Clean(c.Description)
		if c.CourseCode == "MNTC" {
			xcounter++
		}
		xcs = append(xcs, c)
	}
	fmt.Printf("Total %d MNTC\n", xcounter)
	b, err := json.MarshalIndent(subjs, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./assets/QU-applied-subjects.json", b, 0600)
	if err != nil {
		log.Fatal(err)
	}

	b, err = json.MarshalIndent(xcs, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./assets/QU-applied-courses.json", b, 0600)
	if err != nil {
		log.Fatal(err)
	}

	return
}
