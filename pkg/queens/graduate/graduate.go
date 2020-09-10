package graduate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/MikhailKlemin/aditya/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

type subjs struct {
	name string
	link string
}

//Start is
func Start() {
	//Parse(subjs{"NameIS", "https://www.queensu.ca/sgs/graduate-calendar/courses-instruction/biology-courses"})
	//return
	sbjs := getLinks()
	fmt.Println(len(sbjs), "\t", sbjs[0])
	//sbjs = sbjs[:2]
	var cs []utils.CourseExample
	for _, sbj := range sbjs {
		cs = append(cs, Parse(sbj)...)
	}

	Export(cs)
}

func getLinks() (sjs []subjs) {
	client := utils.GetClient()
	resp, err := client.Get("https://www.queensu.ca/sgs/graduate-calendar/courses-instruction")
	if err != nil {
		log.Fatal(err)
	}

	/*resp, err := os.Open("./assets/concurrent-source.html")
	if err != nil {
		return
	}
	defer resp.Close()
	*/
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	doc.Find(`a[href^="/sgs/graduate-calendar/courses-instruction/"]`).Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr(`href`)
		sjs = append(sjs, subjs{s.Text(), "https://www.queensu.ca" + href})
	})
	return
}

//Parse is
func Parse(sj subjs) (cs []utils.CourseExample) {
	client := utils.GetClient()
	resp, err := client.Get(sj.link)
	if err != nil {
		log.Fatal(err)
	}

	/*resp, err := os.Open("./assets/concurrent-source.html")
	if err != nil {
		return
	}
	defer resp.Close()
	*/
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	re := regexp.MustCompile(`([A-Z]{4})-(\d+)\*\s*(.*)`)
	re2 := regexp.MustCompile(`PREREQUISITE:\s*(?:OR\s*COREQUISITES:)?\s*(.*)`)
	re3 := regexp.MustCompile(`EXCLUSION:\s*(.*)`)

	doc.Find(`p strong`).Each(func(_ int, s *goquery.Selection) {
		var c utils.CourseExample
		txt := strings.TrimSpace(s.Text())
		m := re.FindStringSubmatch(txt)
		if len(m) == 0 {
			return
		}
		c.CourseCode = m[1]
		c.NumericCode = m[2]
		c.Name = utils.Clean(m[3])
		c.SubjectName = sj.name
		descs := strings.Split(strings.TrimSpace(s.Parent().Clone().Children().Remove().End().Text()), "\n")
		if c.NumericCode == "926" {
			fmt.Println(descs)
		}
		c.Description = descs[0]

		//fmt.Printf("%#v\n", c)
		if len(descs) > 0 {
			for _, d := range descs[1:] {
				//fmt.Println(d)
				if m1 := re2.FindStringSubmatch(d); len(m1) > 0 {
					c.Prerequisite = strings.TrimSuffix(strings.TrimSpace(m1[1]), ".")
				} else {
					//c.Description = c.Description + " " + d
				}

				if m1 := re3.FindStringSubmatch(d); len(m1) > 0 {
					c.Antirequisite = strings.TrimSuffix(strings.TrimSpace(m1[1]), ".")

				} else {
					//c.Description = c.Description + " " + d
				}

			}
		}
		if c.Description == "" {
			c.Description = utils.Clean(s.Parent().Next().Text())
		}

		//b, _ := json.MarshalIndent(c, "", "    ")
		//fmt.Println(string(b))
		cs = append(cs, c)
		//fmt.Println(s.Parent().Clone().Children().Remove().End().Html())

	})
	return
}

func dedup(in []string) (out []string) {
	if len(in) <= 1 {
		return in
	}

	keys := make(map[string]bool)

	for _, val := range in {
		if _, ok := keys[val]; !ok {
			out = append(out, val)
			keys[val] = true
		}
	}
	return
}

//Export Exports
func Export(cs []utils.CourseExample) (err error) {
	var subjs []utils.Subject
	var se = make(map[string]utils.Subject)
	var xcs []utils.CourseExample
	scount := 0
	for _, c := range cs {
		if val, ok := se[c.SubjectName]; !ok {
			var subj utils.Subject
			c.SubjectID = scount
			subj.SubjectID = scount
			subj.SubjectName = c.SubjectName
			subj.SubjectCode = []string{c.CourseCode}
			se[c.SubjectName] = subj //adding to "exists" map
			//fmt.Println(c.SubjectName, "\t", scount, "\t", val.SubjectName)
			scount++
		} else {
			val.SubjectCode = append(val.SubjectCode, c.CourseCode)
			se[c.SubjectName] = val
			//fmt.Println(c.SubjectName, "\t", se[c.SubjectName].SubjectName, "\t", se[c.SubjectName].SubjectID)
		}
		//c.SubjectName = ""
		//fmt.Println(c.SubjectName, "\t", se[c.SubjectName].SubjectName, "\t", se[c.SubjectName].SubjectID)

		//c.CourseCode =
		//fmt.Println(se[c.SubjectName].SubjectID)
		c.SubjectID = se[c.SubjectName].SubjectID
		xcs = append(xcs, c)
	}
	for _, val := range se {
		val.SubjectCode = dedup(val.SubjectCode)
		subjs = append(subjs, val)
	}
	b, err := json.MarshalIndent(subjs, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./assets/QU-graduate-subjects.json", b, 0600)
	if err != nil {
		log.Fatal(err)
	}

	b, err = json.MarshalIndent(xcs, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./assets/QU-graduate-courses.json", b, 0600)
	if err != nil {
		log.Fatal(err)
	}

	return
}
