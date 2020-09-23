package mcgill

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/MikhailKlemin/aditya/pkg/model"
	"github.com/MikhailKlemin/aditya/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

/*
All the subjects are listed on the right hand side sidebar (image) attached.
You can use this to gather all the subject information for mcgill-subjects.json.
The courses are paginated and will require navigating all pages of the site.
Each course page has all the similar information as before.
(name, courseCode, numericCode, description, prerequisite, corequisite, antirequisite (called 'restriction' on this site))
*/

//re1 split title on Code/NumericCode/Name
var re1 = regexp.MustCompile(`([A-Z]{4})\s*([\dA-Za-z]+)\s*(.*?)\s*\d*\s*\(`)

//re2 split desctiption on subjectName and rest
var re2 = regexp.MustCompile(`^(.*?)\s*:\s*(.*)`)

var re3 = regexp.MustCompile(`^Prerequisite\s*:\s*(.*)`)

var re4 = regexp.MustCompile(`^Restriction\s*:\s*(.*)`)

//Start is starting
func Start() error {
	links, err := collect()
	if err != nil {
		return err
	}
	var cs []model.Course
	for i, link := range links {
		fmt.Printf("Processing %d link\n", i)
		c, err := Parse(link)
		if err != nil {
			log.Println(err)
			continue
		}
		if c.Name != "" {
			cs = append(cs, c)
		}
	}

	model.Export(cs, "mcgill")
	return nil
}

func collect() (links []string, err error) {

	client := utils.CreateClientWithTOR()

	doc, err := client.Get("https://www.mcgill.ca/study/2020-2021/courses/search")
	if err != nil {
		return
	}

	doc.Find(`.view-search-courses .views-row`).Each(func(_ int, s *goquery.Selection) {
		//fmt.Println(s.Text())
		class, _ := s.Attr(`class`)
		if !strings.Contains(class, "not-offered") {
			href, _ := s.Find(`a`).Attr(`href`)
			//txt := s.Find(`a`).Text()
			//fmt.Println(txt)
			links = append(links, "https://www.mcgill.ca"+href)

		}
		//fmt.Println(class)

	})

	maxattr, _ := doc.Find(`.pager-last.last a`).Attr(`href`)
	fmt.Println("MAX_ATTR:\t", maxattr)
	var max int
	if m := regexp.MustCompile(`.*page=(\d+)`).FindStringSubmatch(maxattr); len(m) > 0 {
		max, err = strconv.Atoi(m[1])
		if err != nil {
			return
		}
	}

	for i := 1; i <= max; i++ {
		if i > 5 {
			break
		}
		fmt.Printf("Collecting links page %d collected %d links\n", i, len(links))
		doc, err = client.Get(fmt.Sprintf("https://www.mcgill.ca/study/2020-2021/courses/search?page=%d", i))
		if err != nil {
			return
		}
		doc.Find(`.view-search-courses .views-row`).Each(func(_ int, s *goquery.Selection) {
			class, _ := s.Attr(`class`)
			if !strings.Contains(class, "not-offered") {
				href, _ := s.Find(`a`).Attr(`href`)
				links = append(links, "https://www.mcgill.ca"+href)

			}
		})

	}
	return
}

//Parse parses entity
func Parse(link string) (c model.Course, err error) {
	client := utils.CreateClient()

	doc, err := client.Get(link)
	if err != nil {
		return
	}

	title := utils.Clean(doc.Find(`#page-title`).Text())

	if m := re1.FindStringSubmatch(title); len(m) > 0 {
		c.CourseCode = m[1]
		c.NumericCode = m[2]
		c.Name = m[3]
	}

	var desc string
	doc.Find(`.content`).Each(func(_ int, s *goquery.Selection) {
		if strings.Contains(s.Find(`h3`).Text(), "Overview") {
			desc = utils.Clean(s.Find(`h3`).Next().Text())
		}
	})

	if m := re2.FindStringSubmatch(desc); len(m) > 0 {
		c.SubjectName = utils.Clean(m[1])
		c.Description = utils.Clean(m[2])
	}

	doc.Find(`.catalog-notes p`).Each(func(_ int, s *goquery.Selection) {
		//if strings.Contains(s.Find(`h3`).Text(), "Overview") {
		desc = utils.Clean(s.Find(`h3`).Next().Text())
		//}
		txt := utils.Clean(s.Text())
		if m := re3.FindStringSubmatch(txt); len(m) > 0 {
			c.Prerequisite = utils.Clean(m[1])
		}
		if m := re4.FindStringSubmatch(txt); len(m) > 0 {
			c.Antirequisite = utils.Clean(m[1])
		}

	})

	//fmt.Printf("%#v\n", c)
	return c, nil
}
