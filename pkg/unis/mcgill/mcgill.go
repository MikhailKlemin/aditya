package mcgill

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MikhailKlemin/aditya/pkg/model"
	"github.com/MikhailKlemin/aditya/pkg/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocarina/gocsv"
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

//re1b split title on Code/NumericCode/Name but when no credit info is published
var re1b = regexp.MustCompile(`([A-Z]{4})\s*([\dA-Za-z]+)\s*(.*)\s*`)

//re2 split desctiption on subjectName and rest
var re2 = regexp.MustCompile(`^(.*?)\s*:\s*(.*)`)

var re3 = regexp.MustCompile(`^Prerequisite\s*:\s*(.*)`)

var re4 = regexp.MustCompile(`^Restriction.*?:\s*(.*)`)

var mp *utils.MyProxy

//Start is starting
func Start() error {
	mp = utils.NewProxySet("/home/mike/Downloads/proxy_socks_ip.txt")
	if false {
		c, _ := Parse("https://www.mcgill.ca/study/2020-2021/courses/urbp-708d1")
		b, _ := json.MarshalIndent(c, "", "   ")
		fmt.Printf("%#v\n", c)

		fmt.Printf("%s\n", b)
		return nil
	}

	links, err := collect()
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(links, "", "    ")

	ioutil.WriteFile("./assets/mccgill-links.json", b, 0600)

	var cs []model.Course

	sem := make(chan bool, 10)
	var mu sync.Mutex

	for i, link := range links {
		sem <- true
		go func(i int, link string) {
			defer func() { <-sem }()
			fmt.Printf("Processing %d link\n", i)
			c, err := Parse(link)
			if err != nil {
				log.Println(err)
				//continue
				return
			}
			if c.Name != "" {
				mu.Lock()
				cs = append(cs, c)
				mu.Unlock()
			}
		}(i, link)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	model.Export(cs, "mcgill")
	return nil
}

func collect() (links []string, err error) {

	umap := make(map[string]bool)
	var doc *goquery.Document
	count := 0
	for {
		client := utils.CreateClientWithProxy(mp.GetRandom())

		if count > 10 {
			return links, errors.New("too many errors")
		}
		doc, err = client.Get("https://www.mcgill.ca/study/2020-2021/courses/search")
		if err != nil {
			//return
			count++
			continue
		}
		break
	}

	doc.Find(`.view-search-courses .views-row`).Each(func(_ int, s *goquery.Selection) {
		//fmt.Println(s.Text())
		//class, _ := s.Attr(`class`)
		//if !strings.Contains(class, "not-offered") {
		href, _ := s.Find(`a`).Attr(`href`)
		//txt := s.Find(`a`).Text()
		//fmt.Println(txt)
		links = append(links, "https://www.mcgill.ca"+href)

		//}
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

	sem := make(chan bool, 10)
	var mu sync.Mutex

	for i := 1; i <= max; i++ {
		sem <- true
		go func(i int) {
			defer func() { <-sem }()
			fmt.Printf("Collecting links page %d collected %d links\n", i, len(links))
			count = 0
			for {
				client := utils.CreateClientWithProxy(mp.GetRandom())
				if count > 0 {
					fmt.Printf("Retrying %d %d times\n", i, count+1)
				}
				if count > 20 {
					//return links, errors.New("too many errors")
					log.Printf("Error count 20 on link %d\n", i)
					//continue
					break
				}
				doc, err = client.Get(fmt.Sprintf("https://www.mcgill.ca/study/2020-2021/courses/search?page=%d", i))
				if err != nil {
					log.Println(err)
					time.Sleep(time.Second)
					count++
					continue
				}
				break
			}

			doc.Find(`.view-search-courses .views-row`).Each(func(_ int, s *goquery.Selection) {
				/*class, _ := s.Attr(`class`)
				if !strings.Contains(class, "not-offered") {
					href, _ := s.Find(`a`).Attr(`href`)
					links = append(links, "https://www.mcgill.ca"+href)

				}*/

				href, _ := s.Find(`a`).Attr(`href`)
				mu.Lock()
				if _, ok := umap[href]; !ok {
					links = append(links, "https://www.mcgill.ca"+href)
					umap[href] = true
				}
				mu.Unlock()

			})
		}(i)

	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return
}

//Parse parses entity
func Parse(link string) (c model.Course, err error) {

	var doc *goquery.Document
	count := 0
	for {
		client := utils.CreateClientWithProxy(mp.GetRandom())

		if count > 10 {
			return c, errors.New("too many errors")
		}
		doc, err = client.Get(link)
		if err != nil {
			//return
			log.Println(err)
			count++
			continue
		}
		break
	}

	title := utils.Clean(doc.Find(`#page-title`).Text())

	if m := re1.FindStringSubmatch(title); len(m) > 0 {
		c.CourseCode = m[1]
		c.NumericCode = m[2]
		c.Name = m[3]
	} else if m := re1b.FindStringSubmatch(title); len(m) > 0 {
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

	fmt.Println("Description:\t", desc)

	if m := re2.FindStringSubmatch(desc); len(m) > 0 {
		c.SubjectName = utils.Clean(m[1])
		c.Description = utils.Clean(m[2])
	} else {
		c.SubjectName = utils.Clean(desc)
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

//ToCSV just for debug
func ToCSV() {
	type Course struct {
		SubjectID          int    `json:"subjectId"`
		SubjectName        string `json:"SubjectName,omitempty"`
		SubjectDescription string `json:"-"`
		//TermID      int    `json:"TermID,omitempty"`
		//TermName    string `json:"TermName,omitempty"`

		//SubjectCode []string `json:"codes,omitempty"`

		Name          string `json:"name"`
		NumericCode   string `json:"numericCode"`
		CourseCode    string `json:"courseCode"`
		Description   string `json:"description"`
		Prerequisite  string `json:"prerequisite"`
		Antirequisite string `json:"antirequisite"`
		Corequisite   string `json:"corequisite,omitempty"` //commerce queen

		OneWayExclusion string `json:"oneWayExclusion,omitempty"`

		CrossListed string `json:"crosslisted,omitempty"`
	}
	var cs []Course
	b, _ := ioutil.ReadFile("./assets/mcgill-courses.json")

	json.Unmarshal(b, &cs)

	b, _ = gocsv.MarshalBytes(cs)
	ioutil.WriteFile("./assets/mcgill-courses.csv", b, 0600)
}
