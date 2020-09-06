package queens

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

//CourseExample as per https://www.notion.so/Main-Course-Listing-eb9adf609af84e51af4727d6ae63aff0
type CourseExample struct {
	SubjectID     string `json:"subjectId"`
	Name          string `json:"name"`
	NumericCode   string `json:"numericCode"`
	CourseCode    string `json:"courseCode"`
	Description   string `json:"description"`
	Prerequisite  string `json:"prerequisite"`
	Antirequisite string `json:"antirequisite"`
}

var rgx = struct {
	reCourseCode    *regexp.Regexp //has NumericCode and Name
	reDescription   *regexp.Regexp
	rePrerequisite  *regexp.Regexp
	reAntirequisite *regexp.Regexp
}{
	regexp.MustCompile(`^[^A-Z]?([A-Z]+)\s+(\d+/\d+)\.\d+\s*([A-Z][^\n]*)`),
	regexp.MustCompile(`(?s)[^A-Z]?[A-Z]+\s+\d+/\d+\.\d+\s*[A-Z][^\n]*\s*(.*?)LEARNING`),
	regexp.MustCompile(`(?s)PREREQUISITE\s*(.*?)(?:EXCLUSION)?$`),
	regexp.MustCompile(`(?s)EXCLUSION\s*(?:\([Ss]\))?(.*)\s*`),
}

//Start starts scrapping
func Start() {

	cs, err := parsePDF()
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.MarshalIndent(cs, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./assets/QU-main-courses.json", b, 0600)
	if err != nil {
		log.Fatal(err)
	}
}

func parsePDF() (cs []CourseExample, err error) {

	/* 	Create Temp by downloading PDF from amazon file  */
	fmt.Println("[INFO] ", "Downloading...")
	tempfile, err := ioutil.TempFile("/tmp", "queens-*.pdf")
	if err != nil {
		//log.Fatal(err)
		return
	}

	defer os.Remove(tempfile.Name())

	resp, err := http.Get("https://s3.us-west-2.amazonaws.com/secure.notion-static.com/f739e36d-43d9-4f48-9259-179b115e524a/Main_Course_Listing.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAT73L2G45O3KS52Y5%2F20200905%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20200905T192634Z&X-Amz-Expires=86400&X-Amz-Signature=3e462e276791100e4ec9293d8a1ee53c5a67eba02b2255f99b9b65f8c08aca72&X-Amz-SignedHeaders=host&response-content-disposition=filename%20%3D%22Main%2520Course%2520Listing.pdf%22")
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
		//log.Fatal(err)
		return
	}

	fmt.Println(file.Name())

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

func parseBlock(block string) (c CourseExample, err error) {

	if m := rgx.reCourseCode.FindStringSubmatch(block); len(m) != 0 {
		c.CourseCode = m[1]
		c.NumericCode = m[2]
		c.Name = m[3]
	}

	if m := rgx.reDescription.FindStringSubmatch(block); len(m) != 0 {
		c.Description = m[1]
	}

	if m := rgx.rePrerequisite.FindStringSubmatch(block); len(m) != 0 {
		c.Prerequisite = m[1]
	}

	if m := rgx.reAntirequisite.FindStringSubmatch(block); len(m) != 0 {
		c.Antirequisite = m[1]
	}

	return

}
