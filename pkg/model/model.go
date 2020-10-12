package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

//Course holds all fields for all Unies.
type Course struct {
	SubjectID          int    `json:"subjectId"`
	SubjectName        string `json:"SubjectName,omitempty"`
	SubjectDescription string `json:"-"`
	Link               string `json:"Link,omitempty"`
	//Pos                int    `json:"Pos,omitempty"`

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

	Enrollment *struct {
		Enrolled struct {
			Actual, Max, Available int
		}
		Waitlist struct {
			Actual, Max, Available int
		}
	} `json:"enrollment,omitempty"`
}

//Subject holds terms
type Subject struct {
	SubjectID   int      `json:"subjectId"`
	TermID      int      `json:"termID,omitempty"`
	SubjectName string   `json:"name"`
	SubjectCode []string `json:"codes"`
	Description string   `json:"description,omitempty"`
}

// Export exporting the Course into ./assets dir,
// files are prefixed with prefix
func Export(cs []Course, prefix string) (err error) {
	//writing raw formatt
	rb, _ := json.MarshalIndent(cs, "", "    ")
	ioutil.WriteFile(fmt.Sprintf("./assets/%s-raw.json", prefix), rb, 0600)

	var subjs []Subject
	var se = make(map[string]Subject)
	var xcs []Course
	scount := 0
	for _, c := range cs {
		if val, ok := se[c.SubjectName]; !ok {
			var subj Subject
			c.SubjectID = scount
			subj.SubjectID = scount
			subj.SubjectName = c.SubjectName
			subj.Description = c.SubjectDescription
			subj.SubjectCode = []string{c.CourseCode}
			se[c.SubjectName] = subj //adding to "exists" map
			scount++
		} else {
			val.SubjectCode = append(val.SubjectCode, c.CourseCode)
			se[c.SubjectName] = val
		}
		c.SubjectID = se[c.SubjectName].SubjectID
		xcs = append(xcs, c)
	}
	for _, val := range se {
		val.SubjectCode = dedup(val.SubjectCode)
		subjs = append(subjs, val)
	}
	b, err := json.MarshalIndent(subjs, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	//prefix = ryerson-undergraduate for example
	err = ioutil.WriteFile(fmt.Sprintf("./assets/%s-subjects.json", prefix), b, 0600)
	if err != nil {
		log.Fatal(err)
	}

	b, err = json.MarshalIndent(xcs, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(fmt.Sprintf("./assets/%s-courses.json", prefix), b, 0600)
	if err != nil {
		log.Fatal(err)
	}

	return
}

//dedup deleting duplicates in slice,
//for making
func dedup(in []string) (out []string) {
	if len(in) <= 1 {
		return in
	}

	keys := make(map[string]bool)

	for _, val := range in {
		if _, ok := keys[val]; !ok {
			out = append(out, val)
			keys[val] = true
		}
	}
	return
}
