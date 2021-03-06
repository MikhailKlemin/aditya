package graduate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/MikhailKlemin/aditya/pkg/model"
	"github.com/MikhailKlemin/aditya/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

/*NOTES
https://www.queensu.ca/sgs/graduate-calendar/courses-instruction/urban-and-regional-planning-courses
https://www.notion.so/Graduate-Courses-3d49a70d024b4e3f94ef7c6587674f4f
*/
// Problems:
// Some courses have double course name like
// SURP-891*/892*     Directed Study in Advanced Aspects of Urban and Regional Planning
// so have to catch the whole  891*/892* and split during export

type subjs struct {
	name string
	link string
}

//Start is
func Start() {
	//Parse(subjs{"NameIS", "https://www.queensu.ca/sgs/graduate-calendar/courses-instruction/urban-and-regional-planning-courses"})
	//return
	sbjs := getLinks()
	fmt.Println(len(sbjs), "\t", sbjs[0])
	//sbjs = sbjs[:2]
	var cs []model.Course
	for _, sbj := range sbjs {
		cs = append(cs, Parse(sbj)...)
	}

	model.Export(cs, "QU-graduate")
}

func getLinks() (sjs []subjs) {
	client := utils.GetClient()

	/*
	   	resp, err := client.Get("https://www.queensu.ca/sgs/graduate-calendar/courses-instruction")
	   	if err != nil {
	   		log.Fatal(err)
	   	}

	   	doc, err := goquery.NewDocumentFromReader(resp.Body)
	   	if err != nil {
	   		return
	   	}
	   	defer resp.Body.Close()

	   }
	*/
	var doc *goquery.Document
	counter := 0
	for {
		if counter > 10 {
			log.Fatal("Failed miserable")
		}

		resp, xerr := client.Get("https://www.queensu.ca/sgs/graduate-calendar/courses-instruction")

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

	doc.Find(`a[href^="/sgs/graduate-calendar/courses-instruction/"]`).Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr(`href`)
		sjs = append(sjs, subjs{s.Text(), "https://www.queensu.ca" + href})
	})
	return
}

//Parse is
func Parse(sj subjs) (cs []model.Course) {
	client := utils.GetClient()
	var doc *goquery.Document
	counter := 0
	for {
		if counter > 10 {
			log.Fatal("Failed miserable")
		}

		resp, xerr := client.Get(sj.link)

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

	re := regexp.MustCompile(`.?([A-Z]{4})-([\d*/]+)\s*(.*)`) //Course Code
	re2 := regexp.MustCompile(`PREREQUISITE:\s*(?:OR\s*COREQUISITES:)?\s*(.*)`)
	re3 := regexp.MustCompile(`EXCLUSION:\s*(.*)`)
	re4 := regexp.MustCompile(`(\d+)`)
	re5 := regexp.MustCompile(`COREQUISITE:\s*(.*)`)

	doc.Find(`p strong`).Each(func(_ int, s *goquery.Selection) {
		var c model.Course
		txt := strings.TrimSpace(s.Text())
		m := re.FindStringSubmatch(txt)
		if len(m) == 0 {
			return
		}
		c.CourseCode = m[1]
		c.NumericCode = m[2]
		c.Name = utils.Clean(m[3])
		c.SubjectName = sj.name
		desc, _ := s.Parent().Html()
		ds := strings.Split(desc, "<br/>")
		if len(ds) < 1 {
			return
		}
		for n, val := range ds {
			if n == 0 {
				continue
			}
			//if strings.HasPrefix("")
			val = strings.TrimSpace(val)

			if m := re2.FindStringSubmatch(val); len(m) > 0 {
				c.Prerequisite = utils.Clean(m[1])
				continue
			}
			if m := re3.FindStringSubmatch(val); len(m) > 0 {
				c.Antirequisite = utils.Clean(m[1])
				continue
			}
			if m := re5.FindStringSubmatch(val); len(m) > 0 {
				c.Prerequisite = utils.Clean(c.Prerequisite + " " + m[1])
				continue
			}

			c.Description = utils.Clean(c.Description + " " + val)

		}

		// Splitting courses if they have double name
		ms := re4.FindAllStringSubmatch(c.NumericCode, -1)
		for _, m := range ms {
			c.NumericCode = m[1]
			cs = append(cs, c)
			xb, _ := json.MarshalIndent(c, "", "    ")
			fmt.Printf("%s\n", xb)
		}

	})
	return
}

/*
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
*/
