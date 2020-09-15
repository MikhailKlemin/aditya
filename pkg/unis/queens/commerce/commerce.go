package commerce

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/MikhailKlemin/aditya/pkg/model"
	"github.com/MikhailKlemin/aditya/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

//pair to hold informaiton fromm main course listing
type pair struct {
	num  string
	desc string
	link string
}

//Start starting
func Start() {
	//Parse(pair{"162", "some crap", "https://smith.queensu.ca/bcom/academic_calendar/courses/2018_19/COMM317.php"})
	//return
	client := utils.GetClient()
	re := regexp.MustCompile(`COMM\s*(.*):\s*(.*)`)

	/*
		resp, err := client.Get(`https://smith.queensu.ca/bcom/the_program/curriculum/all_courses.php`)
		if err != nil {
			log.Println(err)
			return
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}

		defer resp.Body.Close()
	*/

	var doc *goquery.Document
	counter := 0
	for {
		if counter > 10 {
			log.Fatal("Failed miserable")
		}

		resp, xerr := client.Get("https://smith.queensu.ca/bcom/the_program/curriculum/all_courses.php")

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
	var ps []pair
	doc.Find(`a.modal_frame`).Each(func(i int, s *goquery.Selection) {
		//fmt.Printf("%d\t%s\n", i, s.Text())
		var p pair
		p.link, _ = s.Attr(`href`)
		p.link = "https://smith.queensu.ca" + p.link
		data, _ := s.Attr(`data-title`)

		if m := re.FindStringSubmatch(data); len(m) > 0 {
			p.num = m[1]
			p.desc = utils.Clean(m[2])
		}
		ps = append(ps, p)
	})

	var cs []model.Course
	var mu sync.Mutex
	sem := make(chan bool, 3)

	for _, p := range ps {
		sem <- true
		go func(p pair) {
			defer func() { <-sem }()
			c, err := Parse(p)
			if err != nil {
				log.Println(err)
				return
			}
			mu.Lock()
			if strings.Contains(c.NumericCode, "/") {
				m := regexp.MustCompile(`(\d+)`).FindAllStringSubmatch(c.NumericCode, -1)
				for i := 0; i < len(m); i++ {
					c.NumericCode = m[i][0]
					cs = append(cs, c)
				}

			} else {
				cs = append(cs, c)
			}
			fmt.Printf("%#v\n", c)
			mu.Unlock()
		}(p)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	//b, _ := json.MarshalIndent(cs, "", "   ")
	//ioutil.WriteFile("./assets/QU-commerce-courses.json", b, 0600)
	model.Export(cs, "QU-commerce")

}

//Parse parses link
func Parse(p pair) (c model.Course, err error) {
	client := utils.GetClient()
	count := 0
	var doc *goquery.Document

	for {
		if count > 10 {
			return c, errors.New("failed\t" + p.link)
		}
		resp, err := client.Get(p.link)
		if err != nil {
			log.Println(err)
			count++
			continue
		}

		doc, err = goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println(err)
			count++
			continue
		}
		defer resp.Body.Close()
		break
	}

	c.Description = doc.Find(`description p`).First().Text()
	h, _ := doc.Find(`description`).Html()
	//fmt.Println(h)
	if m := regexp.MustCompile(`Prerequisite\s*-\s*(.*?)\s*(?:<br/>|</p>)`).FindStringSubmatch(h); len(m) > 0 {
		c.Prerequisite = utils.Clean(m[1])
	}

	if m := regexp.MustCompile(`Corequisite\s*-\s*(.*)\s*`).FindStringSubmatch(h); len(m) > 0 {
		c.Corequisite = utils.Clean(m[1])
	}

	if m := regexp.MustCompile(`Exclusions\s*-\s*(.*)\s*`).FindStringSubmatch(h); len(m) > 0 {
		c.Antirequisite = utils.Clean(m[1])
	}

	c.Name = p.desc
	c.NumericCode = p.num
	c.CourseCode = "COMM"

	return

}
