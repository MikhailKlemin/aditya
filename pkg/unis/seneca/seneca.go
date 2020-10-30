package seneca

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/MikhailKlemin/aditya/pkg/model"
	"github.com/MikhailKlemin/aditya/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

var cleanRe = regexp.MustCompile("[[:^ascii:]]")

//Start starts
func Start() {
	//xx := parseMainLink("https://www.senecacollege.ca/programs/fulltime/DAN/courses.html#pre-content_menu")
	//fmt.Println(xx[len(xx)-1], "\t", len(xx))
	//os.Exit(1)
	links := getMainLinks()
	//fmt.Println(links[0], "\t", links[len(links)-1])
	var cs []model.Course
	var mu sync.Mutex
	sem := make(chan bool, 3)
	for i, link := range links {

		pairs := parseMainLink(link)
		for _, pair := range pairs {
			sem <- true
			go func(pair []string) {
				defer func() { <-sem }()
				c := ParseCourseLink(pair)
				if c.Name != "" {
					mu.Lock()
					cs = append(cs, c)
					mu.Unlock()
				}
			}(pair)
		}
		fmt.Printf("Done %d from %d\n", i, len(links))
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	//dedupe
	var deCS []model.Course
	exists := func(c model.Course, deCS []model.Course) bool {
		for _, x := range deCS {
			if c.SubjectName == x.SubjectName &&
				c.Name == x.Name &&
				c.NumericCode == x.NumericCode &&
				c.CourseCode == x.CourseCode &&
				c.Description == x.Description {
				return true
			}
		}
		return false
	}
	for _, c := range cs {
		if !exists(c, deCS) {
			deCS = append(deCS, c)
		}
	}

	model.Export(deCS, "seneca")
	//fmt.Println(pairs)
	//parseCourseLink("https://apps.senecacollege.ca/ssos/findOutline.do?subjectCode=ACT351")
}

func getMainLinks() (links []string) {

	//https://www.senecacollege.ca/programs/fulltime/AUC/courses.html#pre-content_menu
	client := utils.CreateClient()
	//um := make(map[string]bool)

	doc, err := client.Get("https://www.senecacollege.ca/programs/alphabetical.html")
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(`a[href^="/programs/fulltime/"]`).Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr(`href`)
		//	if _, ok := um[href]; !ok {
		//		um[href] = true
		href = strings.Replace(href, ".html", "/courses.html", 1)
		links = append(links, "https://www.senecacollege.ca"+href)
		//	}
	})

	return
}

func parseMainLink(link string) (pairs [][]string) {
	client := utils.CreateClient()

	doc, err := client.Get(link)
	if err != nil {
		log.Fatal(err)
	}

	subj := utils.Clean(doc.Find(`h1`).Clone().Children().Remove().End().Text())
	//fmt.Println(subj)
	um := make(map[string]bool)

	//var links []string
	doc.Find(`a[href^="https://apps.senecacollege.ca/ssos/findOutline.do?subjectCode="]`).Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr(`href`)
		if _, ok := um[href]; !ok {
			um[href] = true
			pairs = append(pairs, []string{subj, href})

		}
		//href = strings.Replace(href, ".html", "/courses.html", 1)
		//links = append(links, "https://www.senecacollege.ca"+href)
		//links = append(links, href)
		//fmt.Println(href)
	})

	//href = strings.Replace(href, ".html", "/courses.html", 1)
	//links = append(links, "https://www.senecacollege.ca"+href)
	//links = append(links, href)
	//fmt.Println(href)

	return

}

//ParseCourseLink parses a link
func ParseCourseLink(pair []string) (c model.Course) {
	client := utils.CreateClient()

	doc, err := client.Get(pair[1])
	if err != nil {
		log.Fatal(err)
	}

	na := doc.Find(`h1`).Text()
	if strings.Contains(na, "No Subject Outline could be found.") {
		fmt.Printf("%s has no outline\n", pair[1])
		return
	}
	doc.Find(`strong`).EachWithBreak(func(_ int, s *goquery.Selection) bool {
		txt := utils.Clean(s.Text())
		if txt == `Subject Description` {
			//p := s.Parent().Clone().Children().Remove().End().Text()
			p := strings.Replace(s.Parent().Text(), "Subject Description", "", 1)
			p = cleanRe.ReplaceAllString(p, " ")
			desc := utils.Clean(p)
			if desc == "" {
				//p = cleanRe.ReplaceAllString(p, " ")
				desc = utils.Clean(s.Parent().Next().Text())
			}
			if desc == "" {
				p = s.Parent().Parent().Find(`span`).Text()
				p = cleanRe.ReplaceAllString(p, "")
				desc = utils.Clean(p)
			}
			//fmt.Println(desc)
			c.Description = desc
			return false
		}
		return true
	})

	name := utils.Clean(doc.Find(`h1 p`).Text())

	if m := regexp.MustCompile(`^([A-Z]{3})(\d{3})\s*-\s*(.*)$`).FindStringSubmatch(name); len(m) > 0 {
		c.CourseCode = m[1]
		c.NumericCode = m[2]
		c.Name = m[3]
	}

	c.SubjectName = pair[0]
	c.Link = pair[1]

	b, _ := json.MarshalIndent(c, "", "    ")
	fmt.Printf("%s\n", string(b))

	return

}
