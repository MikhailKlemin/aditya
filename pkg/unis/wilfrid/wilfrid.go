package wilfrid

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/MikhailKlemin/aditya/pkg/model"
	"github.com/PuerkitoBio/goquery"
	strip "github.com/grokify/html-strip-tags-go"
)

type theClass struct {
	Success              bool      `json:"success"`
	TotalCount           int       `json:"totalCount"`
	Data                 []theData `json:"data"`
	PageOffset           int       `json:"pageOffset"`
	PageMaxSize          int       `json:"pageMaxSize"`
	SectionsFetchedCount int       `json:"sectionsFetchedCount"`
	PathMode             string    `json:"pathMode"`
	SearchResultsConfigs []struct {
		Config  string `json:"config"`
		Display string `json:"display"`
		Title   string `json:"title"`
		Width   string `json:"width"`
	} `json:"searchResultsConfigs"`
	ZtcEncodedImage string `json:"ztcEncodedImage"`
}

type theData struct {
	ID                      int     `json:"id"`
	Term                    string  `json:"term"`
	TermDesc                string  `json:"termDesc"`
	CourseReferenceNumber   string  `json:"courseReferenceNumber"`
	PartOfTerm              string  `json:"partOfTerm"`
	CourseNumber            string  `json:"courseNumber"`
	Subject                 string  `json:"subject"`
	SubjectDescription      string  `json:"subjectDescription"`
	SequenceNumber          string  `json:"sequenceNumber"`
	CampusDescription       string  `json:"campusDescription"`
	ScheduleTypeDescription string  `json:"scheduleTypeDescription"`
	CourseTitle             string  `json:"courseTitle"`
	CreditHours             float64 `json:"creditHours"`
	MaximumEnrollment       int     `json:"maximumEnrollment"`

	Enrollment     int `json:"enrollment"`
	SeatsAvailable int `json:"seatsAvailable"`

	WaitCapacity  int `json:"waitCapacity"`
	WaitCount     int `json:"waitCount"`
	WaitAvailable int `json:"waitAvailable"`

	CrossList           interface{}   `json:"crossList"`
	CrossListCapacity   interface{}   `json:"crossListCapacity"`
	CrossListCount      interface{}   `json:"crossListCount"`
	CrossListAvailable  interface{}   `json:"crossListAvailable"`
	CreditHourHigh      float64       `json:"creditHourHigh"`
	CreditHourLow       int           `json:"creditHourLow"`
	CreditHourIndicator string        `json:"creditHourIndicator"`
	OpenSection         bool          `json:"openSection"`
	LinkIdentifier      interface{}   `json:"linkIdentifier"`
	IsSectionLinked     bool          `json:"isSectionLinked"`
	SubjectCourse       string        `json:"subjectCourse"`
	Faculty             []interface{} `json:"faculty"`
	MeetingsFaculty     []struct {
		Category              string        `json:"category"`
		Class                 string        `json:"class"`
		CourseReferenceNumber string        `json:"courseReferenceNumber"`
		Faculty               []interface{} `json:"faculty"`
		MeetingTime           struct {
			BeginTime              interface{} `json:"beginTime"`
			Building               interface{} `json:"building"`
			BuildingDescription    interface{} `json:"buildingDescription"`
			Campus                 interface{} `json:"campus"`
			CampusDescription      interface{} `json:"campusDescription"`
			Category               string      `json:"category"`
			Class                  string      `json:"class"`
			CourseReferenceNumber  string      `json:"courseReferenceNumber"`
			CreditHourSession      float64     `json:"creditHourSession"`
			EndDate                string      `json:"endDate"`
			EndTime                interface{} `json:"endTime"`
			Friday                 bool        `json:"friday"`
			HoursWeek              float64     `json:"hoursWeek"`
			MeetingScheduleType    string      `json:"meetingScheduleType"`
			MeetingType            string      `json:"meetingType"`
			MeetingTypeDescription string      `json:"meetingTypeDescription"`
			Monday                 bool        `json:"monday"`
			Room                   interface{} `json:"room"`
			Saturday               bool        `json:"saturday"`
			StartDate              string      `json:"startDate"`
			Sunday                 bool        `json:"sunday"`
			Term                   string      `json:"term"`
			Thursday               bool        `json:"thursday"`
			Tuesday                bool        `json:"tuesday"`
			Wednesday              bool        `json:"wednesday"`
		} `json:"meetingTime"`
		Term string `json:"term"`
	} `json:"meetingsFaculty"`
	ReservedSeatSummary interface{} `json:"reservedSeatSummary"`
	SectionAttributes   []struct {
		Class                 string `json:"class"`
		Code                  string `json:"code"`
		CourseReferenceNumber string `json:"courseReferenceNumber"`
		Description           string `json:"description"`
		IsZTCAttribute        bool   `json:"isZTCAttribute"`
		TermCode              string `json:"termCode"`
	} `json:"sectionAttributes"`
}

type coursePair struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

//Course represents the Course infrmation this is Exlusively for Wilfrid
type Course struct {
	TermCode string `json:",omitempty"`
	TermName string `json:",omitempty"`
	TermID   int

	SubjectCode string `json:",omitempty"`
	SubjectName string `json:",omitempty"`
	SubjectID   int    `json:"subjectId"`

	CourseID              int
	CourseCode            string // subject
	NumericCode           string // courseNumber
	Campus                string // campusDescription
	CourseTitle           string //courseTitle
	Description           string //separate request
	CourseReferenceNumber string // CourseReferenceNumber
	Section               string // sequenceNumber
	Prerequisite          string
	CrossListed           string

	Enrollment struct {
		Enrolled struct {
			Actual, Max, Available int
		}
		Waitlist struct {
			Actual, Max, Available int
		}
	}
}

//Terms holds terms
type Terms struct {
	TermID   int
	TermName string
	TermCode string
}

var client *http.Client

//Start function begins the scraping
func Start() {

	//creating global  http client
	cookieJar, _ := cookiejar.New(nil)
	client = &http.Client{
		Jar: cookieJar,
	}

	// Getting  list of terms and JSESSIONID

	fmt.Println("Getting Terms and  SessionID")
	ps, sessionID := getCourses()

	// Printing for test purposes
	fmt.Printf("SeessionID is %s and there are %d terms the 1st one is %s with code %s\n", sessionID, len(ps), ps[0].Description, ps[0].Code)
	var cs []Course

	for i, p := range ps {
		if i > 2 {
			break
		}

		fmt.Printf("Processing %d term with code %s and description %s\n", i, p.Code, p.Description)
		fmt.Println("Sending post request to switch term for current Session ID")
		resp := postSession(p.Code, sessionID)
		fmt.Printf("Post response is %s no errors\n", strings.TrimSpace(string(resp)))
		fmt.Printf("Getting classes for the %s term\n", p.Code)
		classes := getClasses(p.Code, sessionID)
		if len(classes) > 0 {
			fmt.Printf("For %s term we got %d classes from %s to %s \n", p.Code, len(classes), classes[0].Description, classes[len(classes)-1].Description)
		} else {
			fmt.Println("No classes for ", p.Code)
			continue
		}

		fmt.Printf("Walking classes for %s term\n", p.Code)

		for j, class := range classes {
			if j > 5 {
				//break
			}
			tds := browseClasses(class.Code, p.Code, sessionID)
			fmt.Printf("Got %d results for class %s\n", len(tds), class.Code)
			for _, td := range tds {
				var c Course
				fmt.Println(td.Term, "\t", td.CourseReferenceNumber)
				getCourseDesc(&c, td.Term, td.CourseReferenceNumber)
				setCommonFields(&c, td)
				getPreprequests(&c, td.Term, td.CourseReferenceNumber)
				c.TermName = p.Description
				c.TermCode = p.Code

				c.SubjectCode = class.Code
				c.SubjectName = class.Description

				fmt.Printf("%s\t%s\n", p.Description, c.CourseTitle)
				cs = append(cs, c)
			}
			time.Sleep(time.Second)
			postSession(p.Code, sessionID)

		}

	}

	b, _ := json.MarshalIndent(cs, "", "    ")
	ioutil.WriteFile("./assets/wlu-raw-data.json", b, 0600)
	//return cs
	Export(cs)

}

func setCommonFields(c *Course, td theData) {
	if td.CrossList != nil {
		fmt.Printf("%#v\n", td.CrossList)
	}

	c.CourseID = td.ID
	c.CourseCode = td.Subject
	c.NumericCode = td.CourseNumber
	c.Campus = td.CampusDescription
	c.CourseTitle = html.UnescapeString(td.CourseTitle)
	//c.Description
	c.CourseReferenceNumber = td.CourseReferenceNumber
	c.Section = td.SequenceNumber
	//c.Prerequisite
	//c.Crosslisted

	c.Enrollment.Enrolled.Actual = td.Enrollment
	c.Enrollment.Enrolled.Max = td.MaximumEnrollment
	c.Enrollment.Enrolled.Available = td.MaximumEnrollment - td.Enrollment

	c.Enrollment.Waitlist.Actual = td.WaitCount
	c.Enrollment.Waitlist.Max = td.WaitCapacity
	c.Enrollment.Waitlist.Available = td.WaitCapacity - td.WaitCount

}

func getPreprequests(c *Course, term, courseReferenceNumber string) {
	fCount := 0
	for {
		if fCount > 10 {
			log.Print(fmt.Errorf("Persistent error on getting Course prerequest for %s reference and %s term", courseReferenceNumber, term))
		}

		if fCount > 0 {
			time.Sleep(2 * time.Second)
		}

		var data = strings.NewReader(fmt.Sprintf(`term=%s&courseReferenceNumber=%s&first=first`, term, courseReferenceNumber))
		req, err := http.NewRequest("POST", "https://loris.wlu.ca/register/ssb/searchResults/getSectionPrerequisites", data)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0")
		req.Header.Set("Accept", "text/html; q=0.01")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("X-Synchronizer-Token", "2239a4b4-d4dc-42da-9ca3-7065bf3f158f")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Origin", "https://loris.wlu.ca")
		req.Header.Set("DNT", "1")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Referer", "https://loris.wlu.ca/register/ssb/classSearch/classSearch")
		//req.Header.Set("Cookie", "f5avrbbbbbbbbbbbbbbbb=CHCGGIOPBBFJCDBDDFELOANJPKPLNJEMIFMBGALIFKAKJHNLHAODKGEEKCMFJPCJGICDOLAPHHHAGBNOOEPAOLDNFPNHHCLGPOCFOOODIONBGKGANNGIDKMONIDHIGKN; f5_cspm=1234; JSESSIONID=A4B0E71002DB9CE6DA8C812C5C1796BC; _ga=GA1.2.803802487.1598984858; _gid=GA1.2.991047408.1598984858; BIGipServerpool_prodlorisregister=1096157706.24353.0000; BIGipServerpool_prodlorisbanextension=1029048842.18213.0000")
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			fCount++
			continue

		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println(err)
			fCount++
			continue
		}
		//fmt.Println(doc.Html())
		var prereq []string
		doc.Find("tr").Each(func(i int, s *goquery.Selection) {
			cells := []string{}
			s.Find("td").Each(func(_ int, s2 *goquery.Selection) {
				cell := html.UnescapeString(strings.TrimSpace(s2.Text()))
				if cell != "" {
					cells = append(cells, cell)
				}
			})

			prereq = append(prereq, strings.TrimSpace(strings.Join(cells, " ")))
		})
		c.Prerequisite = strings.TrimPrefix(strings.Join(prereq, ";"), ";")
		break
	}

}

// getCourseDesc get descritption of the course
func getCourseDesc(c *Course, term, courseReferenceNumber string) {
	fCount := 0
	for {

		if fCount > 10 {
			log.Fatal(fmt.Errorf("Persistent error on getting Course description for %s reference and %s term", courseReferenceNumber, term))
		}

		if fCount > 0 {
			time.Sleep(2 * time.Second)
		}

		var data = strings.NewReader(fmt.Sprintf(`term=%s&courseReferenceNumber=%s`, term, courseReferenceNumber))
		req, err := http.NewRequest("POST", "https://loris.wlu.ca/register/ssb/searchResults/getCourseDescription", data)
		if err != nil {
			log.Println(err)
			fCount++
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0")
		req.Header.Set("Accept", "text/html; q=0.01")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("X-Synchronizer-Token", "2239a4b4-d4dc-42da-9ca3-7065bf3f158f")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Origin", "https://loris.wlu.ca")
		req.Header.Set("DNT", "1")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Referer", "https://loris.wlu.ca/register/ssb/classSearch/classSearch")
		//req.Header.Set("Cookie", "f5avrbbbbbbbbbbbbbbbb=CHCGGIOPBBFJCDBDDFELOANJPKPLNJEMIFMBGALIFKAKJHNLHAODKGEEKCMFJPCJGICDOLAPHHHAGBNOOEPAOLDNFPNHHCLGPOCFOOODIONBGKGANNGIDKMONIDHIGKN; f5_cspm=1234; JSESSIONID=A4B0E71002DB9CE6DA8C812C5C1796BC; _ga=GA1.2.803802487.1598984858; _gid=GA1.2.991047408.1598984858; BIGipServerpool_prodlorisregister=1096157706.24353.0000; BIGipServerpool_prodlorisbanextension=1029048842.18213.0000")
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			fCount++
			continue
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println(err)
			fCount++
			continue
		}
		c.Description = html.UnescapeString(strings.TrimSpace(strip.StripTags(doc.Text())))
		break
	}

}

func browseClasses(subject, term, sessionID string) []theData {
	fCount := 0
	var body []byte
	for {

		if fCount > 10 {
			log.Fatal(fmt.Errorf("Persistent error on getting list of Classes for %s term", term))
		}

		if fCount > 0 {
			time.Sleep(2 * time.Second)
		}

		link := fmt.Sprintf("https://loris.wlu.ca/register/ssb/searchResults/searchResults?txt_subject=%s&txt_term=%s&pageOffset=0&pageMaxSize=100&sortColumn=subjectDescription&sortDirection=asc",
			subject, term)
		//		fmt.Println(link)

		//client := &http.Client{}
		req, err := http.NewRequest("GET", link, nil)
		if err != nil {
			log.Println(err)
			fCount++
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("DNT", "1")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Upgrade-Insecure-Requests", "1")
		req.Header.Set("Cache-Control", "max-age=0")
		//req.Header.Set("Cookie", "f5avrbbbbbbbbbbbbbbbb=FHGBEJFABJPIMMAHKOBAGGEANDKJHMGACDPAKLHDECJOMBLIPLEKJFMNNCGBDCGOEGADNNHOODJMHDEKHNJAJPHGENEODLPLKKFFFMIKOCILDDMLBAILGEIDGPFKBMLD; f5_cspm=1234; JSESSIONID="+sessionID+"; _ga=GA1.2.803802487.1598984858; _gid=GA1.2.991047408.1598984858; BIGipServerpool_prodlorisregister=1096157706.24353.0000; BIGipServerpool_prodlorisbanextension=1029048842.18213.0000")
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			fCount++
			continue
		}
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			fCount++
			continue
		}
		break

	}

	//fmt.Println(string(body))
	var g theClass
	if err := json.Unmarshal(body, &g); err != nil {
		log.Fatal(err)
	}

	return g.Data

}

func getClasses(term, sessionID string) []coursePair {

	fCount := 0
	var p []coursePair

	for {

		if fCount > 10 {
			log.Fatal(fmt.Errorf("Persistent error on getting classes for %s term", term))
		}

		if fCount > 0 {
			time.Sleep(2 * time.Second)
		}

		req, err := http.NewRequest("GET",
			fmt.Sprintf("https://loris.wlu.ca/register/ssb/classSearch/get_subject?searchTerm=&term=%s&offset=1&max=150&uniqueSessionId=%s&=1598990860994", term, sessionID),
			nil)
		if err != nil {
			log.Println(err)
			fCount++
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0")
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("X-Synchronizer-Token", "ff256186-80f4-4565-ba44-99003635ffe8")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("DNT", "1")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Referer", "https://loris.wlu.ca/register/ssb/term/termSelection?mode=search")
		req.Header.Set("Cookie", "f5avrbbbbbbbbbbbbbbbb=IEDKECFBPIFCBOCPOBLHMONHNIPLANEMPBKCMDNHMDOFIALDHKIEGLOKLHNAFGPIGEIDAOGPOCPHMAJGENDACCGPEJPOPGMOEBLBEJIAMGFEGBEPJBPAMOKIFIILOIOF; JSESSIONID="+sessionID+"; BIGipServerpool_prodlorisregister=1045826058.24353.0000; BIGipServerpool_prodlorisbanextension=1029048842.18213.0000; _ga=GA1.2.803802487.1598984858; _gid=GA1.2.991047408.1598984858; _gat=1")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			fCount++
			continue
		}

		if err := json.Unmarshal(body, &p); err != nil {
			log.Println(err)
			fCount++
			continue
		}

		break

	}

	return p

}

//postSession send post request to update
//state so the client will be able to browse
//classes
func postSession(term, sessionID string) string {
	//client := &http.Client{}
	var bodyText []byte
	fCount := 0

	for {
		if fCount > 10 {
			break
		}
		var data = strings.NewReader(`term=` + term + `&studyPath=&studyPathText=&startDatepicker=&endDatepicker=&uniqueSessionId=` + sessionID)
		req, err := http.NewRequest("POST", "https://loris.wlu.ca/register/ssb/term/search?mode=search", data)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("X-Synchronizer-Token", "ff256186-80f4-4565-ba44-99003635ffe8")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Origin", "https://loris.wlu.ca")
		req.Header.Set("DNT", "1")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Referer", "https://loris.wlu.ca/register/ssb/term/termSelection?mode=search")
		//req.Header.Set("Cookie", "f5avrbbbbbbbbbbbbbbbb=PEKHPCHAFNHONHICPEIIOJGFPJBCPMHJKLPNBCEEBNLNJMLMLNBKKEELMNEKNKFDHDIDANOFPJNILAHIINLABDKAMJHIMDCBGIEHKJLMGGJPCEIBIIPCMDNIBCDONJFK; f5_cspm=1234; JSESSIONID=E786E31604C131FFFB3216F2FFA7D9FC; BIGipServerpool_prodlorisregister=1045826058.24353.0000; BIGipServerpool_prodlorisbanextension=1029048842.18213.0000; _ga=GA1.2.803802487.1598984858; _gid=GA1.2.991047408.1598984858; _gat=1")
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			fCount++
			continue
		}
		bodyText, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			fCount++
			continue
		}
		break
	}
	//fmt.Printf("%s\n", bodyText)
	return string(bodyText)
}

//getCourses get list of courses and JSESSIONID for further querries.
//because everythign there depends on JSESSIONID (yeah stupid)
func getCourses() ([]coursePair, string) {
	var sessionID string
	//client := &http.Client{}
	req, err := http.NewRequest("GET", "https://loris.wlu.ca/register/ssb/classSearch/getTerms?searchTerm=&offset=1&max=30&_=1598984877083", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:80.0) Gecko/20100101 Firefox/80.0")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("X-Synchronizer-Token", "ff256186-80f4-4565-ba44-99003635ffe8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://loris.wlu.ca/register/ssb/term/termSelection?mode=search")
	//req.Header.Set("Cookie", "f5avrbbbbbbbbbbbbbbbb=IEDKECFBPIFCBOCPOBLHMONHNIPLANEMPBKCMDNHMDOFIALDHKIEGLOKLHNAFGPIGEIDAOGPOCPHMAJGENDACCGPEJPOPGMOEBLBEJIAMGFEGBEPJBPAMOKIFIILOIOF; JSESSIONID=E786E31604C131FFFB3216F2FFA7D9FC; BIGipServerpool_prodlorisregister=1045826058.24353.0000; BIGipServerpool_prodlorisbanextension=1029048842.18213.0000; _ga=GA1.2.803802487.1598984858; _gid=GA1.2.991047408.1598984858; _gat=1")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	cookies := resp.Cookies()
	for _, c := range cookies {
		//fmt.Printf("%#v\n", c)
		if c.Name == "JSESSIONID" {
			sessionID = c.Value
			break
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var p []coursePair

	if err := json.Unmarshal(body, &p); err != nil {
		log.Fatal(err)
	}

	return p, sessionID

}

//Export exports into appropirate format
func Export(cs []Course) {
	if len(cs) == 0 {
		b, err := ioutil.ReadFile("./assets/wlu-raw-data.json")
		if err != nil {
			log.Fatal(err)
		}

		//var cs []utils.Course

		if err = json.Unmarshal(b, &cs); err != nil {
			log.Fatal(err)
		}
	}

	type iTerm struct {
		ID   int
		Name string
		Code string
	}

	type iSubj struct {
		ID     int
		TermID int
		Name   string
		Code   string
	}
	var termMap = make(map[string]iTerm)
	termCount := 1

	var subjectMap = make(map[string]iSubj)
	subjCount := 1
	var ecs []Course

	for _, c := range cs {
		if val, ok := termMap[c.TermCode]; ok {
			c.TermID = val.ID
		} else {
			termMap[c.TermCode] = iTerm{termCount, c.TermName, c.TermCode}
			c.TermID = termCount
			termCount++
		}

		if val, ok := subjectMap[c.TermCode+c.SubjectCode]; ok {
			c.SubjectID = val.ID
		} else {
			subjectMap[c.TermCode+c.SubjectCode] = iSubj{subjCount, c.TermID, c.SubjectName, c.SubjectCode}
			c.SubjectID = subjCount
			subjCount++
		}

		c.TermCode = ""
		c.TermName = ""

		c.SubjectCode = ""
		c.SubjectName = ""

		ecs = append(ecs, c)
	}

	b, _ := json.MarshalIndent(ecs, "", "    ")
	ioutil.WriteFile("./assets/wlu-courses.json", b, 0600)

	var terms []Terms
	for _, val := range termMap {
		var t Terms
		t.TermID = val.ID
		t.TermName = val.Name
		t.TermCode = val.Code
		terms = append(terms, t)
	}
	b, _ = json.MarshalIndent(terms, "", "    ")
	ioutil.WriteFile("./assets/wlu-terms.json", b, 0600)

	var subjs []model.Subject
	for _, val := range subjectMap {
		var t model.Subject
		t.SubjectID = val.ID
		t.SubjectName = html.UnescapeString(val.Name)
		t.SubjectCode = []string{val.Code}
		t.TermID = val.TermID
		subjs = append(subjs, t)
	}
	b, _ = json.MarshalIndent(subjs, "", "    ")
	ioutil.WriteFile("./assets/wlu-subjects.json", b, 0600)

}
