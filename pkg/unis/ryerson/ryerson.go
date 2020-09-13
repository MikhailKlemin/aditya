package ryerson

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/MikhailKlemin/aditya/pkg/model"
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

	//Shuffling for fun sake, but actually it was blocling me on science link, so tried to see if that connected anyhow to since, apperared that it is not, but I kept the shuffle just in case.
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(sbjs), func(i, j int) { sbjs[i], sbjs[j] = sbjs[j], sbjs[i] })

	fmt.Println(len(sbjs), "\t", sbjs[0])
	//return
	//sbjs = sbjs[:2]
	var cs []model.Course
	for _, sbj := range sbjs {
		cs = append(cs, Parse(sbj)...)
	}

	model.Export(cs, "ryerson-undergraduate")
}

func getLinks() (sjs []subjs) {
	client := utils.GetClient()
	var doc *goquery.Document
	for {
		resp, err := client.Get("https://www.ryerson.ca/calendar/2020-2021/courses/")
		if err != nil {
			log.Println(err)
			continue
		}

		/*resp, err := os.Open("./assets/concurrent-source.html")
		if err != nil {
			return
		}
		defer resp.Close()
		*/
		doc, err = goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println(err)
			continue
		}
		resp.Body.Close()
		break
	}
	var re = regexp.MustCompile(`\s*\(.*$`)

	doc.Find(`a[href^="/content/ryerson/calendar/2020-2021/courses/"]`).Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr(`href`)
		txt := utils.Clean(s.Text())
		txt = re.ReplaceAllString(txt, "")
		sjs = append(sjs, subjs{txt, "http://www.ryerson.ca" + href})
	})
	return
}

//Parse is
func Parse(sj subjs) (cs []model.Course) {
	time.Sleep(5 * time.Second)
	client := utils.GetClient()
	var doc *goquery.Document
	for {
		resp, err := client.Get(sj.link)
		if err != nil {
			log.Println(err)
			time.Sleep(600 * time.Second)
			continue
		}

		/*resp, err := os.Open("./assets/concurrent-source.html")
		if err != nil {
			return
		}
		defer resp.Close()
		*/
		doc, err = goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println(err)
			time.Sleep(600 * time.Second)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()
		break
	}
	var re = regexp.MustCompile(`^([A-Z]+)\s*(\d+)`)
	doc.Find(`li.courseListItem`).Each(func(_ int, s *goquery.Selection) {
		var c model.Course
		var pre []string
		s.Find(`.courseListPrerequisites a`).Each(func(_ int, s2 *goquery.Selection) {
			pre = append(pre, strings.TrimSpace(s2.Text()))
		})
		c.Prerequisite = strings.Join(pre, ";")

		pre = []string{}
		s.Find(`.courseListAntirequisites a`).Each(func(_ int, s2 *goquery.Selection) {
			pre = append(pre, strings.TrimSpace(s2.Text()))
		})

		c.Antirequisite = strings.Join(pre, ";")
		c.Description = utils.Clean(s.Find(`.courseListDescription`).Text())
		code := strings.TrimSpace(s.Find(`.courseListCourseCode`).Text())
		if m := re.FindStringSubmatch(code); len(m) > 0 {
			c.CourseCode = m[1]
			c.NumericCode = m[2]
		}
		c.SubjectName = sj.name
		c.Name = utils.Clean(s.Find(`.courseListCourseName`).Text())
		cs = append(cs, c)
		b, _ := json.MarshalIndent(c, "", "")
		fmt.Println(string(b))

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

	err = ioutil.WriteFile("./assets/ryerson-undergraduate-subjects.json", b, 0600)
	if err != nil {
		log.Fatal(err)
	}

	b, err = json.MarshalIndent(xcs, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./assets/ryerson-undergraduate-courses.json", b, 0600)
	if err != nil {
		log.Fatal(err)
	}

	return
}
*/
