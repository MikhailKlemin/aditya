package concurent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"

	"github.com/MikhailKlemin/aditya/pkg/model"
	"github.com/MikhailKlemin/aditya/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

//Start starts scraping
func Start() {
	/*client := utils.GetClient()

	resp, err := client.Get("https://educ.queensu.ca/course-descriptions")
	if err != nil {
		log.Fatal(err)
	}

	outFile, err := os.Create("./assets/concurrent-source.html")
	// handle err
	defer outFile.Close()

	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	h, _ := doc.Html()

	_, err = io.Copy(outFile, strings.NewReader(h))
	if err != nil {
		log.Fatal(err)
	}
	// handle err
	*/
	cs, err := Parse()
	if err != nil {
		log.Fatal(err)
	}

	//b, _ := json.MarshalIndent(cs, "", "    ")
	//ioutil.WriteFile("./assets/QU-concurrent-courses.json", b, 0600)
	model.Export(cs, "QU-concurrent")
}

//Parse parse courses
func Parse() (cs []model.Course, err error) {
	client := utils.GetClient()

	var doc *goquery.Document
	counter := 0
	for {
		if counter > 10 {
			return cs, errors.New("Failed miserable")
		}

		resp, xerr := client.Get("https://educ.queensu.ca/course-descriptions")

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

	re := regexp.MustCompile(`^([A-Z]{4})\s*([\d/\.]+)\s*(.*)`)
	re2 := regexp.MustCompile(`</strong>.*\n(.*)`)

	re3 := regexp.MustCompile(`PREREQUISITE:\s*(.*)`)
	//var re = regexp.MustCompile(`^([A-Z]+)\s*([\d\./]+)\s*(.*)`)
	doc.Find(`p strong`).Each(func(i int, s *goquery.Selection) {
		var c model.Course
		h, _ := s.Html()

		h = utils.Clean(h)
		if m := re.FindStringSubmatch(h); len(m) > 0 {
			c.Name = m[3]
			c.NumericCode = m[2]
			c.CourseCode = m[1]
		} else {
			return
		}

		h, _ = s.Parent().Html()
		if m := re2.FindStringSubmatch(h); len(m) > 0 {
			//fmt.Println(m[1])
			c.Description = utils.Clean(m[1])
			//return
		}
		if c.Description == "" {
			h, _ = s.Parent().Next().Html()
			c.Description = utils.Clean(h)
		}
		fmt.Println("#####################")
		fmt.Println(i)
		//fmt.Println(h)

		if m := re3.FindStringSubmatch(h); len(m) > 0 {
			c.Prerequisite = utils.Clean(m[1])
			//c.Description = utils.Clean(re3.ReplaceAllString(c.Description, ""))
		}

		b, _ := json.MarshalIndent(c, "", "    ")
		fmt.Println(string(b))
		cs = append(cs, c)

	})

	return

}
