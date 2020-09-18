package ubc

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"sync"

	"github.com/MikhailKlemin/aditya/pkg/model"
	"github.com/MikhailKlemin/aditya/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

type aRow struct {
	code    string
	link    string
	title   string
	faculty string
	desc    string
	clinks  []cLink
}

type cLink struct {
	clink   string
	numeric string
	ccode   string
	cname   string
	//ctitl
}

//parsing h4 header like "BIOL 111 Introduction to Modern Biology"
var re1 = regexp.MustCompile(`^([A-Z]{4})\s*(\d+)\s*(.*)`)

//Split course code and numeric code in tabular view
var re2 = regexp.MustCompile(`^([A-Z]+)\s*(\d.*)`)

var proxies []string

//Start starts scrapping
func Start() {
	b, err := ioutil.ReadFile("/home/mike/Downloads/proxy_socks_ip.txt")

	if err != nil {
		log.Fatal(err)
	}
	proxies = strings.Split(string(b), "\n")
	//fmt.Printf("%#v\n", parseCourse(`https://courses.students.ubc.ca/cs/courseschedule?pname=subjarea&tname=subj-course&dept=BIOL&course=421`))
	ps := getRandomProxyFromFile()
	fmt.Println(ps)
	cs := collect()
	model.Export(cs, "ubc")
}

//uneffective and rude, bud sufficient at the moment
func getRandomProxyFromFile() string {

	return strings.TrimSpace(proxies[rand.Intn(len(proxies)-1)])

}

//collect collects courselinks
func collect() (cs []model.Course) {

	link := "https://courses.students.ubc.ca/cs/courseschedule?tname=subj-all-departments&campuscd=UBC&pname=subjarea"

	proxy := getRandomProxyFromFile()
	fmt.Println(proxy)
	var doc *goquery.Document
	var err error
	for {
		client := utils.CreateClientWithProxy(proxy)

		doc, err = client.Get(link)
		if err != nil {
			log.Println(err)
			continue
		}
		break
	}

	var rows []aRow
	doc.Find(`tr[class^="section"]`).Each(func(i int, s *goquery.Selection) {
		if i > 3 {
			//return
		}
		var r aRow
		s.Find(`td`).Each(func(i int, s2 *goquery.Selection) {
			switch i {
			case 0:
				r.code = strings.TrimSpace(s2.Text())
				r.link, _ = s2.Find(`a`).Attr(`href`)
				r.link = "https://courses.students.ubc.ca" + r.link
			case 1:
				r.title = strings.TrimSpace(s2.Text())
			case 2:
				r.faculty = strings.TrimSpace(s2.Text())
			}
		})
		if r.code != "" && r.link != "https://courses.students.ubc.ca" {
			//fmt.Println(r.link)
			rows = append(rows, r)
		}
	})

	sem := make(chan bool, 2)
	var mu sync.Mutex

	for i, row := range rows {
		if i != 48 {
			//continue
		}
		sem <- true
		go func(row aRow, i int) {
			defer func() { <-sem }()
			fmt.Printf("STARTING %d from %d \t link:%s\n", i, len(rows), row.link)
			xcs := parseLevel1(row)
			mu.Lock()
			cs = append(cs, xcs...)
			mu.Unlock()
		}(row, i)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	return
}

func parseLevel1(row aRow) (cs []model.Course) {
	client := utils.CreateClientWithProxy(getRandomProxyFromFile())

	doc, err := client.Get(row.link)
	if err != nil {
		log.Fatal(err)
	}

	row.desc = utils.Clean(doc.Find(`div.pagination`).First().Prev().Text())
	//var cl []string

	doc.Find(`tr[class^="section"]`).Each(func(_ int, s *goquery.Selection) {
		//cl = append(cl, "https://courses.students.ubc.ca"+href)
		var cl cLink
		s.Find(`td`).Each(func(i int, s2 *goquery.Selection) {
			switch i {
			case 0:
				href, _ := s2.Find(`a[href^="/cs/courseschedule?pname="]`).Attr(`href`)
				cl.clink = "https://courses.students.ubc.ca" + href
				txt := utils.Clean(s2.Text())
				if m := re2.FindStringSubmatch(txt); len(m) > 0 {
					cl.ccode = m[1]
					cl.numeric = m[2]
				}
			case 1:
				cl.cname = utils.Clean(s2.Text())
			}
		})
		row.clinks = append(row.clinks, cl)

	})

	//fmt.Println(row.desc, "\t", len(row.courseLinks))
	sem := make(chan bool, 3)
	var mu sync.Mutex

	for _, clink := range row.clinks {
		sem <- true
		go func(clink cLink) {
			defer func() { <-sem }()
			fmt.Println("\t", clink.clink)
			c := parseCourse(clink.clink)
			c.SubjectDescription = row.desc
			c.SubjectName = row.title
			c.CourseCode = clink.ccode
			c.NumericCode = clink.numeric
			c.Name = clink.cname
			mu.Lock()
			cs = append(cs, c)
			mu.Unlock()
		}(clink)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	return
}

func parseCourse(link string) (c model.Course) {
	//c.SubjectName = row.title
	//c.SubjectDescription = row.desc
	client := utils.CreateClientWithProxy(getRandomProxyFromFile())

	doc, err := client.Get(link)
	if err != nil {
		log.Fatal(err)
	}

	/*h4 := utils.Clean(doc.Find(`h4`).First().Text())
	if m := re1.FindStringSubmatch(h4); len(m) > 0 {
		c.CourseCode = m[1]
		c.Name = m[3]
		c.NumericCode = m[2]

	}
	*/
	c.Description = doc.Find(`h4+p`).Text()
	doc.Find(`div[role="main"] p`).EachWithBreak(func(_ int, s *goquery.Selection) bool {
		txt := utils.Clean(s.Text())
		if strings.HasPrefix(txt, "Pre-reqs:") {
			c.Prerequisite = strings.TrimSpace(strings.TrimPrefix(txt, "Pre-reqs:"))
			return false
		}
		return true
	})

	doc.Find(`div[role="main"] p`).EachWithBreak(func(_ int, s *goquery.Selection) bool {
		txt := utils.Clean(s.Text())
		if strings.HasPrefix(txt, "Equivalents:") {
			c.OneWayExclusion = strings.TrimSpace(strings.TrimPrefix(txt, "Equivalents:"))
			return false
		}
		return true
	})

	return
}
