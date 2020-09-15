package maincourses

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/MikhailKlemin/aditya/pkg/model"
)

//CourseExample as per https://www.notion.so/Main-Course-Listing-eb9adf609af84e51af4727d6ae63aff0

var rgx = struct {
	reCourseCode    *regexp.Regexp //has NumericCode and Name
	reDescription   *regexp.Regexp
	reDescription2  *regexp.Regexp
	rePrerequisite  *regexp.Regexp
	reAntirequisite *regexp.Regexp
	reOneWay        *regexp.Regexp
}{
	regexp.MustCompile(`^[^A-Z]?([A-Z]+)\s+(\d+/\d+)\.\d+\s*([A-Z][^\n]*)`),
	regexp.MustCompile(`(?s)^[^A-Z]?[A-Z]+\s+\d+/\d+\.\d+\s*[A-Z][^\n]*(.*?)(?:PREREQUISITE|EXCLUSION|ONE-WAY)`),
	regexp.MustCompile(`(?s)^[^A-Z]?[A-Z]+\s+\d+/\d+\.\d+\s*[A-Z][^\n]*(.*)`), //If Desription comes empty try this one
	regexp.MustCompile(`PREREQUISITE(.*)`),
	regexp.MustCompile(`EXCLUSION\s*\(S\)\s*(.*)`),
	regexp.MustCompile(`ONE-WAY\s*EXCLUSION\s*(.*)`),
}

//Start starts scrapping
func Start() {

	cs, err := parsePDF()
	if err != nil {
		log.Fatal(err)
	}

	/*
		b, err := json.MarshalIndent(cs, "", "    ")
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile("./assets/QU-main-courses.json", b, 0600)
		if err != nil {
			log.Fatal(err)
		}
	*/
	model.Export(cs, "QU-main")
}

func parsePDF() (cs []model.Course, err error) {

	/* 	Create Temp by downloading PDF from amazon file  */
	fmt.Println("[INFO] ", "Downloading..")
	tempfile, err := ioutil.TempFile("/tmp", "queens-*.pdf")
	if err != nil {
		//log.Fatal(err)
		return
	}

	defer os.Remove(tempfile.Name())

	resp, err := http.Get("https://www.queensu.ca/artsci/sites/default/files/a_courses_of_instruction_2.pdf")
	if err != nil {
		return
	}

	_, err = io.Copy(tempfile, resp.Body)
	if err != nil {
		return
	}
	fmt.Println("[INFO] ", "Done downloading, parsing...")

	/* Done with temp file downloading */

	file, err := ioutil.TempFile("/tmp", "queens-*.txt")
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println(file.Name(), "\t", tempfile.Name())

	defer os.Remove(file.Name())
	_, err = exec.Command("pdftotext", "-layout", tempfile.Name(), file.Name()).Output()
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(output)

	b, err := ioutil.ReadFile(file.Name())
	text := string(b)
	lines := strings.Split(text, "\n")

	var re = regexp.MustCompile(`^[^A-Z]?([A-Z]+)\s+(\d+/\d+)\.\d+\s*([A-Z][^\n]*)`)
	//newrecord = true
	var block []string
	var blocks [][]string
	for _, line := range lines {
		if re.MatchString(line) {
			if len(block) != 0 {
				blocks = append(blocks, block)
			}
			block = []string{}
		}
		block = append(block, line)
	}
	blocks = append(blocks, block)
	//fmt.Println(strings.Join(blocks[len(blocks)-1], "\n"))

	for _, block := range blocks {
		sb := strings.Join(block, "\n")
		c, _ := parseBlock(sb)
		cs = append(cs, c)
	}
	return
}

func parseBlock(block string) (c model.Course, err error) {

	t := func(in string) string {
		return strings.TrimSpace(in)
	}
	if m := rgx.reCourseCode.FindStringSubmatch(block); len(m) != 0 {
		c.CourseCode = t(m[1])
		c.NumericCode = t(m[2])
		c.Name = t(m[3])
	}

	if m := rgx.reDescription.FindStringSubmatch(block); len(m) != 0 {
		c.Description = t(m[1])
	}

	if c.Description == "" {
		if m := rgx.reDescription2.FindStringSubmatch(block); len(m) != 0 {
			c.Description = t(m[1])
		}

	}

	if m := rgx.rePrerequisite.FindStringSubmatch(block); len(m) != 0 {
		c.Prerequisite = t(m[1])
	}

	if m := rgx.reAntirequisite.FindStringSubmatch(block); len(m) != 0 {
		c.Antirequisite = t(m[1])
	}

	if m := rgx.reOneWay.FindStringSubmatch(block); len(m) != 0 {
		c.OneWayExclusion = t(m[1])
	}

	return

}
