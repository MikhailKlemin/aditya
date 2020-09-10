package utils

import (
	"html"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

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

//CourseExample holds courses information
type CourseExample struct {
	SubjectID   int      `json:"subjectId"`
	SubjectName string   `json:"SubjectName,omitempty"`
	SubjectCode []string `json:"codes,omitempty"`

	Name          string `json:"name"`
	NumericCode   string `json:"numericCode"`
	CourseCode    string `json:"courseCode"`
	Description   string `json:"description"`
	Prerequisite  string `json:"prerequisite"`
	Antirequisite string `json:"antirequisite"`
	Corequisite   string `json:"corequisite,omitempty"` //commerce queen

	OneWayExclusion string `json:"oneWayExclusion,omitempty"`
}

//Terms holds terms
type Terms struct {
	TermID   int
	TermName string
	TermCode string
}

//Subject holds terms
type Subject struct {
	SubjectID   int      `json:"subjectId"`
	TermID      int      `json:"termID,omitempty"`
	SubjectName string   `json:"name"`
	SubjectCode []string `json:"codes"`
}

//GetClient creates simple http client
func GetClient() *http.Client {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 15 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 15 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 30,
		Transport: netTransport,
	}
	return netClient
}

//Clean removes HTML, double spaces and other
func Clean(in string) (out string) {
	re := regexp.MustCompile(`<[^>]*>`)
	re2 := regexp.MustCompile(`\s+`)
	out = re.ReplaceAllString(in, " ")
	out = html.UnescapeString(out)
	out = strings.ReplaceAll(out, "\u00a0", " ")
	out = re2.ReplaceAllString(out, " ")
	return strings.TrimSpace(out)

}
