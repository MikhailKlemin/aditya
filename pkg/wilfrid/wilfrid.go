package wilfrid

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type coursePair struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

//Start function begins the scraping
func Start() {
	ps, sessionID := getCourses()

	fmt.Println("JSESSIONID\t", sessionID)
	for i, p := range ps {
		fmt.Println(i, "\t", p.Code, "\t", p.Description)
	}

	resp := postSession(ps[1].Code, sessionID)
	fmt.Println(resp)
	terms := getTerm(sessionID)

	for i, p := range terms {
		fmt.Println(i, "\t", p.Code, "\t", p.Description)
	}

}

func getTerm(sessionID string) []coursePair {
	// https://loris.wlu.ca/register/ssb/classSearch/get_subject?searchTerm=&term=202101&offset=1&max=10&uniqueSessionId=j4h1m1598984856906&_=1598990860994
	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://loris.wlu.ca/register/ssb/classSearch/get_subject?searchTerm=&term=202101&offset=1&max=10&uniqueSessionId=%s&_=1598990860994", sessionID),
		nil)
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var p []coursePair

	if err := json.Unmarshal(body, &p); err != nil {
		log.Fatal(err)
	}

	return p

}

//postSession send post request to update
//state so the client will be able to browse
//classes
func postSession(term, sessionID string) string {
	client := &http.Client{}
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
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%s\n", bodyText)
	return string(bodyText)
}

//getCourses get list of courses and JSESSIONID for further querries.
//because everythign there depends on JSESSIONID (yeah stupid)
func getCourses() ([]coursePair, string) {
	var sessionID string
	client := &http.Client{}
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

/*
JSESSIONID       A4C8BDA5DAC5CECD93D1E033D521FC7B
0        202105          Spring 2021
1        202101          Winter 2021
2        202009          Fall 2020
3        202005          Spring 2020 (View Only)
4        202001          Winter 2020 (View Only)
5        201909          Fall 2019 (View Only)
6        201905          Spring 2019 (View Only)
7        201901          Winter 2019 (View Only)
8        201809          Fall 2018 (View Only)
9        201805          Spring 2018 (View Only)
10       201801          Winter 2018 (View Only)
11       201709          Fall 2017 (View Only)
12       201705          Spring 2017 (View Only)
13       201701          Winter 2017 (View Only)
14       201609          Fall 2016 (View Only)
15       201605          Spring 2016 (View Only)
16       201601          Winter 2016 (View Only)
17       201509          Fall 2015 (View Only)
18       201505          Spring 2015 (View Only)
19       201501          Winter 2015 (View Only)
20       201409          Fall 2014 (View Only)
21       201405          Spring 2014 (View Only)
22       201401          Winter 2014 (View Only)
23       201309          Fall 2013 (View Only)
24       201305          Spring 2013 (View Only)
25       201301          Winter 2013 (View Only)
26       201209          Fall 2012 (View Only)
27       201205          Spring 2012 (View Only)
28       201201          Winter 2012 (View Only)
29       201109          Fall 2011 (View Only)

*/

/*
0        AN      Anthropology
1        AB      Arabic
2        AR      Archaeology
3        AF      Arts Topic Seminar
4        AS      Astronomy
5        BH      Biological &amp; Chemical Sciences
6        BI      Biology
7        BF      Brantford Foundations
8        BU      Business
9        MB      Business Technology Management
*/
