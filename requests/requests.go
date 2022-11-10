package requests

type (
	CreateUserRequest struct {
		AuthCode    string `json:"code"`
		RedirectURI string `json:"redirect_uri"`
	}

	UpdateUserRequest struct {
		State             string   `json:"state"`
		Organization      string   `json:"organization"`
		YearsOfExperience string   `json:"years_of_experience"`
		VolunteerAreas    []string `json:"volunteer_areas,omitempty"`
		VolunteerMeans    []string `json:"volunteer_means,omitempty"`
		WillJoinDirectory *bool    `json:"will_join_directory,omitempty"`
		SelfSummary       string   `json:"self_summary"`
		Representation    string   `json:"representation"`
		ProvidedName      string   `json:"provided_name"`
	}
)
