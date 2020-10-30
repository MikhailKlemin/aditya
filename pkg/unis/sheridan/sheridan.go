package sheridan

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"log"

	"github.com/MikhailKlemin/aditya/pkg/model"
	"github.com/MikhailKlemin/aditya/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

/*
TRUE: https://ulysses.sheridanc.on.ca/coutline/coutlineview_sal.jsp?appver=sal&subjectCode=WKSP&courseCode=78142&version=2019090300&sec=0&reload=true
	  https://ulysses.sheridanc.on.ca/coutline/coutlineview_sal.jsp?appver=ba&subjectCode=VISM&courseCode=4013&version=3.0&sec=0&reload=true
	  https://ulysses.sheridanc.on.ca/coutline/coutlineview.jsp?appver=ba&subjectCode=VISM&courseCode=4013&version=3.0&sec=0&reload=true
*/

type theLink struct {
	Link        string
	SubjectName string
	Code        string
}

//Start starts scrapping
func Start() {
	if false {
		ps := getProgramms()
		var links []theLink

		for i, p := range ps {
			fmt.Println("Getting ", p)
			ls, err := get(p, "")
			if err != nil {
				log.Println(err)
				continue
			}
			links = append(links, ls...)
			fmt.Printf("Done %d from %d\n", i, len(ps))
		}

		b, _ := json.MarshalIndent(links, "", "    ")
		ioutil.WriteFile("./assets/sheridan-links.json", b, 0600)
	}

	//os.Exit(1)
	links := loadLinks()
	//links = links[:10]
	var cs []model.Course
	var mu sync.Mutex
	sem := make(chan bool, 5)
	for i, l := range links {
		/*
			if !strings.Contains(l.Link, "subjectCode=ACCG&courseCode=16971") {
				continue
			}
		*/
		sem <- true
		go func(i int, l theLink) {
			defer func() { <-sem }()
			fmt.Println("Done:\t", i)
			c, err := Parse(l.Link)
			if err != nil {
				log.Println(err)
				//continue
				return
			}
			c.SubjectName = l.SubjectName
			c.SubjectCode = l.Code
			mu.Lock()
			cs = append(cs, c)
			mu.Unlock()

		}(i, l)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	model.Export(cs, "sheridan")

}

//Parse parsing
func Parse(link string) (c model.Course, err error) {
	//coutlineview.jsp?appver=ba&subjectCode=VISM&courseCode=4013&version=3.0&sec=0&reload=true
	client := utils.CreateClient()

	prefix := "https://ulysses.sheridanc.on.ca/coutline/"
	link = prefix + link
	u, err := url.Parse(link)
	q := u.Query()
	c.NumericCode = q.Get("courseCode")
	c.CourseCode = q.Get("subjectCode")
	c.Link = link

	doc, err := client.Get(link)
	if err != nil {
		return
	}

	c.Name = utils.Clean(doc.Find(`.CourseTitleHeading`).Text())
	if c.Name == "" {
		c.Name = utils.Clean(doc.Find(`font[size="5"]`).Text())
	}

	txt := doc.Text()
	//fmt.Println(doc.Html())
	if m := regexp.MustCompile(`(?s)Detailed Description(.*?)Program Context`).FindStringSubmatch(txt); len(m) > 0 {
		c.Description = utils.Clean(m[1])
	}

	if m := regexp.MustCompile(`(?s) Program\(s\):\s*(.*?)\s*Program Coordinator\(s\)`).FindStringSubmatch(txt); len(m) > 0 {
		c.SubjectName = utils.Clean(m[1])
	}

	fmt.Printf("%#v\n", c)
	return
}

func loadLinks() (links []theLink) {
	b, _ := ioutil.ReadFile("./assets/sheridan-links.json")

	json.Unmarshal(b, &links)

	fmt.Println(len(links))
	return links
}

func get(p string, season string) (links []theLink, err error) {
	/*
		seasons:
		09 -- Fall
		01 -- Winter
		05  -- Summer
	*/
	if season == "" {
		season = "09"
	}
	url := "https://ulysses.sheridanc.on.ca/coutline/results.jsp"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf("appver=ps&subjectCode=&courseNumber=&title=&program_ps=%s&program_old=&program=%s&programName=&season=%s&year=2020&Search=Search", p, p, season))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:81.0) Gecko/20100101 Firefox/81.0")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Origin", "https://ulysses.sheridanc.on.ca")
	req.Header.Add("DNT", "1")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://ulysses.sheridanc.on.ca/coutline/results.jsp")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Cookie", "JSESSIONID=166B847771276FBF00A217A0DB7BD7BB")

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()
	/*body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}*/

	//fmt.Println(string(body))
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return
	}
	doc.Find(`#detail_table a`).Each(func(i int, s *goquery.Selection) {

		href, _ := s.Attr(`href`)
		links = append(links, theLink{href, strings.TrimSpace(s.Parent().Next().Next().Text()), p})

	})

	return
}

func getProgramms() (ps []string) {
	client := utils.CreateClient()
	doc, err := client.Get(`https://ulysses.sheridanc.on.ca/coutline/results.jsp`)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(`select[name="program_ps"] option`).Each(func(i int, s *goquery.Selection) {
		val, _ := s.Attr(`value`)
		if val != "" {
			fmt.Println(val)
			ps = append(ps, val)
		}
	})

	return
}
