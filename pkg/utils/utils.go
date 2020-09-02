package utils

//Course represents the Course infrmation
type Course struct {
	SectionNumber          string
	CourseNumber           int
	Title                  string
	EnrollmentActual       int
	EnrollmentMaximum      int
	WaitlistCapacity       int
	WaitlistActual         int
	WaitlistSeatsAvailable int
	CourseDescription      string
}
