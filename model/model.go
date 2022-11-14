package model

import "time"

type User struct {
	// Basic
	Email     string `json:"email" firestore:"email"`
	Name      string `json:"name" firestore:"name"`
	Phone     string `json:"phone" firestore:"phone"`
	FirstName string `json:"first_name" firestore:"first_name"`
	LastName  string `json:"last_name" firestore:"last_name"`
	Photo     string `json:"photo" firestore:"photo"`

	// Extras
	State             string `json:"state" firestore:"state"`
	Organization      string `json:"organization" firestore:"organization"`
	YearsOfExperience string `json:"years_of_experience" firestore:"years_of_experience"`
	VolunteerAreas    string `json:"volunteer_areas" firestore:"volunteer_areas"`
	VolunteerMeans    string `json:"volunteer_means" firestore:"volunteer_means"`
	Convicted         bool   `json:"convicted" firestore:"convicted"`
	Representation    string `json:"representation" firestore:"representation"`
	ProvidedName      string `json:"provided_name" firestore:"provided_name"`

	Enrolled  bool      `json:"enrolled" firestore:"enrolled"`
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
}
