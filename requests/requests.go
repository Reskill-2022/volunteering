package requests

type (
	CreateUserRequest struct {
		AuthCode    string `json:"code"`
		RedirectURI string `json:"redirect_uri"`
	}

	UpdateUserRequest struct {
		State              string   `json:"state"`
		Organization       string   `json:"organization"`
		YearsOfExperience  string   `json:"years_of_experience"`
		VolunteerAreas     []string `json:"volunteer_areas"`
		IsUnderrepresented *bool    `json:"is_underrepresented,omitempty"`
		IsConvicted        *bool    `json:"is_convicted,omitempty"`
	}
)
