package concordia

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/MikhailKlemin/aditya/pkg/model"
	"github.com/MikhailKlemin/aditya/pkg/utils"
	"github.com/PuerkitoBio/goquery"
)

//TextBlockIndex holds text and corresponding index
type TextBlockIndex struct {
	Pos   int
	Value string
}

//re1 split course title
var re1 = regexp.MustCompile(`^(?:<b>)?([A-Z]{4})\s*(\d+)<i>\s*(.*?)</i>`)

//re2 deal with Prerequest
var re2 = regexp.MustCompile(`^(?:Prerequisite\s*:\s*)(.*?)\.(.*)`)

//Start Starts Scrapping
func Start() {
	Parse("https://www.concordia.ca/academics/undergraduate/calendar/current/sec31/31-060.html")
	os.Exit(1)
	//model.Export(cs, "concrodia")

	links := collectLinks()
	var cs []model.Course
	for i, link := range links {
		if i > 3 {
			//break
		}
		cs = append(cs, Parse(link)...)
	}

	model.Export(cs, "concordia")

}

func collectLinks() (links []string) {
	//https://www.concordia.ca/academics/undergraduate/calendar/current/sec26.html#unss-courses
	re := regexp.MustCompile(`#.*$`)
	client := utils.CreateClient()

	doc, err := client.Get("https://www.concordia.ca/academics/undergraduate/calendar/current/courses-quick-links.html")
	if err != nil {
		log.Fatal(err)
	}
	umap := make(map[string]bool)
	doc.Find(`a[href^="/academics/undergraduate/calendar/current/sec"]`).Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr(`href`)
		href = re.ReplaceAllString(href, "")
		if _, ok := umap[href]; !ok {
			umap[href] = true
			fmt.Println(href)
			links = append(links, "https://www.concordia.ca"+href)
		}
	})
	return
}

// GetSubjIndexes takes subj indexes like "Art Therapy"
func GetSubjIndexes(src string) (idxs []TextBlockIndex) {
	subRe := regexp.MustCompile(`<a\s*(?:name=".*?")\s*id=".*?"\s*>.*?</a>.*<br>`)

	indexes := subRe.FindAllStringIndex(src, -1)

	for _, index := range indexes {
		//fmt.Println(src[index[0]:index[1]])
		//subj[utils.Clean(src[index[0]:index[1]])] = index
		var idx TextBlockIndex
		idx.Pos = index[0]
		idx.Value = utils.Clean(src[index[0]:index[1]])
		idx.Value = strings.TrimSuffix(idx.Value, ":")
		idxs = append(idxs, idx)
	}

	return idxs
}

// Parse parses course page like
// https://www.concordia.ca/academics/undergraduate/calendar/current/sec31/31-010.html
func Parse(link string) []model.Course {

	client := utils.CreateClient()

	doc, b, err := client.GetWithByte(link)
	if err != nil {
		log.Fatal(err)
	}

	/*b, err := ioutil.ReadFile("/tmp/concordia.html")
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))

	if err != nil {
		log.Fatal(err)
	}
	*/
	var cs []model.Course

	html := string(b)

	subjectName := utils.Clean(doc.Find(`h1`).Text())
	/*

	   <a\s*(?:name=".*?")?\s*id=".*?">.*?</a><b>.*?</b><br>\n<br>\n<b>[A-Z]

	*/
	/*
		if m := regexp.MustCompile(`(?s)<a\s*(?:name=".*?")?\s*id=".*?"\s*>(.*)<hr>`).FindStringSubmatch(html); len(m) > 0 {
			html = m[1]
		} else {
			log.Println("no match!\t", link)
			return cs
		}
	*/

	// <a\s*(?:name=".*?")?\s*id=".*?">.*?</a><b>.*?</b><br>\n<br>\n<b>[A-Z]
	startIndexBlockRe := regexp.MustCompile(`<a\s*(?:name="courses")?\s*id=".*?courses">`)
	endIndexBlockRe := regexp.MustCompile(`(?s).*<hr>`)
	startIndexBlock := startIndexBlockRe.FindStringIndex(html)
	if len(startIndexBlock) == 0 {
		startIndexBlock = regexp.MustCompile(`<b>Courses</b>`).FindStringIndex(html)

	}
	endIndexBlock := endIndexBlockRe.FindStringIndex(html)

	/*
		if len(startIndexBlock) != 0 && len(endIndexBlock) != 0 {
		if startIndexBlock[0] > 500 {
			startIndexBlock[0] = startIndexBlock[0] - 1500
		}
	*/
	if len(startIndexBlock) != 0 && len(endIndexBlock) != 0 {

		html = html[startIndexBlock[0]:endIndexBlock[1]]
	} else {
		log.Println("no index\t", link)
		return []model.Course{}
	}
	//fmt.Println(html)
	//os.Exit(1)
	indexes := regexp.MustCompile(`(?m)^(?:<b>)?([A-Z]{4})\s*(\d+)<i>\s*(.*?)</i>`).FindAllStringIndex(html, -1)
	//var blocks []string
	var blocks []TextBlockIndex

	for i := 0; i < len(indexes); i++ {
		if i != len(indexes)-1 {
			blocks = append(blocks, TextBlockIndex{indexes[i][0], html[indexes[i][0]:indexes[i+1][0]]})
		} else {
			blocks = append(blocks, TextBlockIndex{indexes[i][0], html[indexes[i][0]:]})
		}
	}

	subIdxs := GetSubjIndexes(html)
	//fmt.Println(subIdxs)
	for _, ss := range subIdxs {
		fmt.Printf("%#v\n", ss)
	}
	for i, block := range blocks {
		if i > 2 {
			//break
		}
		//fmt.Println(block)
		//os.Exit(1)
		lines := strings.Split(block.Value, "<br>")
		var c model.Course

		for _, line := range lines {
			clean := strings.TrimSpace(line)
			//			fmt.Printf("Line:%d, Text:\"%s\"\n", i, line)
			if clean != "" {
				//fmt.Println(clean)
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
		c.SubjectDescription = subjectName
		c.Link = link
		//fmt.Printf("%#v\n", c)
		//c.Pos = block.Pos

		for j := 0; j < len(subIdxs); j++ {
			if j < len(subIdxs)-1 {
				if block.Pos > subIdxs[j].Pos && block.Pos < subIdxs[j+1].Pos {
					c.SubjectName = subIdxs[j].Value
					break
				}
			} else {
				c.SubjectName = subIdxs[j].Value
			}
		}
		if c.SubjectName == "" || c.SubjectName == "Courses" {
			c.SubjectName = subjectName
		}
		xb, _ := json.MarshalIndent(c, "", "    ")
		fmt.Printf("%s\n", xb)

		cs = append(cs, c)

	}

	return cs
}
