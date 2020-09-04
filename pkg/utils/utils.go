package utils

//Course represents the Course infrmation
type Course struct {
	TermCode string `json:",omitempty"`
	TermName string `json:",omitempty"`
	TermID   int

	SubjectCode string `json:",omitempty"`
	SubjectName string `json:",omitempty"`
	SubjectID   int

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

//Subject holds terms
type Subject struct {
	SubjectID   int
	TermID      int
	SubjectName string
	SubjectCode string
}
