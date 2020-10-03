package concordia

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/MikhailKlemin/aditya/pkg/model"
	"github.com/MikhailKlemin/aditya/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

//re1 split course title
var re1 = regexp.MustCompile(`^<b>([A-Z]{4})\s*(\d+)<i>\s*(.*?)</i>`)

//re2 deal with Prerequest
var re2 = regexp.MustCompile(`^(?:Prerequisite\s*:\s*)(.*?)\.(.*)`)

//Start Starts Scrapping
func Start() {
	cs := Parse("https://www.concordia.ca/academics/undergraduate/calendar/current/sec31/31-060.html")
	model.Export(cs, "concrodia")
}

// Parse parses course page like
// https://www.concordia.ca/academics/undergraduate/calendar/current/sec31/31-010.html
func Parse(link string) []model.Course {

	client := utils.CreateClient()
	/*
		b, err := ioutil.ReadFile("/tmp/concordia.html")
		if err != nil {
			log.Fatal(err)
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))

		if err != nil {
			log.Fatal(err)
		}
	*/

	doc, err := client.Get(link)
	if err != nil {
		log.Fatal(err)
	}
	subjectName := utils.Clean(doc.Find(`h1`).Text())

	var re = regexp.MustCompile(`[A-Z]{4}\s*\d+`)
	//coursesDoc := doc.Find(`#courses`).Parent().Parent()
	var coursesDoc *goquery.Selection
	doc.Find(`p`).EachWithBreak(func(_ int, s *goquery.Selection) bool {
		txt, _ := s.Html()
		if re.MatchString(strings.TrimSpace(txt)) {
			coursesDoc = s.Parent()
			return false
		}
		return true
	})

	html, _ := coursesDoc.Html()

	//splitted := regexp.MustCompile(`(?m)^<b>[A-Z]{4}\s*(\d+)<i>\s*(.*?)</i>`).Split(html, -1)
	indexes := regexp.MustCompile(`(?m)^<b>([A-Z]{4})\s*(\d+)<i>\s*(.*?)</i>`).FindAllStringIndex(html, -1)
	var blocks []string

	for i := 1; i < len(indexes); i++ {
		if i != len(indexes)-1 {
			blocks = append(blocks, html[indexes[i][0]:indexes[i+1][0]])
		} else {
			blocks = append(blocks, html[indexes[i][0]:])
		}
	}

	var cs []model.Course
	for _, block := range blocks {
		lines := strings.Split(block, "<br/>")
		var c model.Course

		for _, line := range lines {
			clean := strings.TrimSpace(line)
			if clean != "" {
				//fmt.Println(j, "\t", clean)
				if m := re1.FindStringSubmatch(clean); len(m) > 0 {
					c.Name = utils.Clean(m[3])
					c.NumericCode = utils.Clean(m[2])
					c.CourseCode = utils.Clean(m[1])
					continue
				}

				if m := re2.FindStringSubmatch(clean); len(m) > 0 {
					c.Prerequisite = m[1]
					c.Description = m[2]
					continue
				}

				if strings.HasPrefix(clean, "<i>NOTE:") {
					if strings.Contains(clean, "may not take this course") {
						c.Antirequisite = utils.Clean(strings.TrimPrefix(clean, "<i>NOTE:"))
					}
					continue
				}

				c.Description = c.Description + line

			}

		}
		c.Description = utils.Clean(c.Description)
		c.SubjectName = subjectName
		fmt.Printf("%#v\n", c)
		cs = append(cs, c)

	}

	return cs
}
