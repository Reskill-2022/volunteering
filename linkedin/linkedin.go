package linkedin

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rs/zerolog"

	"github.com/Reskill-2022/volunteering/config"
)

type (
	AccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	Service interface {
		GetProfile(authCode, redirectURI string) (*GetProfileOutput, error)
	}

	GetProfileInput struct {
		Email string
	}

	GetProfileOutput struct {
		Email         string
		Name          string
		Photo         string
		ProfileURL    string
		Location      string
		Phone         string
		HasExperience bool
	}

	UserPhone struct {
		Number string `json:"number"`
	}

	UserProfileResponse struct {
		Persons []struct {
			DisplayName  string      `json:"displayName"`
			PhoneNumbers []UserPhone `json:"phoneNumbers"`
			Location     string      `json:"location"`
			PhotoURL     string      `json:"photoUrl"`
			LinkedInURL  string      `json:"linkedInUrl"`
			Positions    struct {
				PositionHistory []struct {
					Title string `json:"title"`
				} `json:"positionHistory"`
			} `json:"positions"`
		} `json:"persons"`
	}

	lkd struct {
		logger       zerolog.Logger
		clientID     string
		clientSecret string
	}

	EmailResponse struct {
		Elements []struct {
			Handle        string `json:"handle"`
			HandleContent struct {
				EmailAddress string `json:"emailAddress"`
			} `json:"handle~"`
		}
	}

	ProfileResponse struct {
		LocalizedLastName  string `json:"localizedLastName"`
		LocalizedFirstName string `json:"localizedFirstName"`
		ProfilePicture     struct {
			DisplayImage string `json:"displayImage"`
		} `json:"profilePicture"`
	}

	PhotoResponse struct {
		ProfilePicture struct {
			DisplayImage struct {
				Elements []struct {
					Identifiers []struct {
						Identifier string `json:"identifier"`
					}
				} `json:"elements"`
			} `json:"displayImage~"`
		} `json:"profilePicture"`
	}
)

func New(logger zerolog.Logger, env config.Environment) Service {
	return &lkd{
		logger:       logger,
		clientID:     env[config.ClientID],
		clientSecret: env[config.ClientSecret],
	}
}

func (l *lkd) GetProfile(authCode, redirectURI string) (*GetProfileOutput, error) {
	return l.getProfile(authCode, redirectURI)
}

func (l *lkd) getProfile(authCode, redirectURI string) (*GetProfileOutput, error) {
	endpoint := "https://www.linkedin.com/oauth/v2/accessToken"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", authCode)
	data.Set("client_id", l.clientID)
	data.Set("client_secret", l.clientSecret)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		l.logger.Err(err).Msg("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to build request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		l.logger.Err(err).Msg("Failed to do request")
		return nil, fmt.Errorf("failed to get access token")
	}

	if resp.StatusCode != http.StatusOK {
		l.logger.Err(fmt.Errorf("expected status code 200, got %d", resp.StatusCode)).Msg("Request failed")
		_, err := io.Copy(os.Stdout, resp.Body)
		if err != nil {
			l.logger.Err(err).Msg("Failed to write response error")
		}
		return nil, fmt.Errorf("failed to get access token, not 200 ok")
	}
	defer resp.Body.Close()

	var payload AccessTokenResponse

	rawJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.logger.Err(err).Msg("Failed to read response body")
		return nil, fmt.Errorf("failed to read response body")
	}
	err = json.Unmarshal(rawJSON, &payload)
	if err != nil {
		l.logger.Err(err).Msg("Failed to unmarshal response body")
		// http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return nil, fmt.Errorf("failed to unmarshal response body")
	}

	email, err := getUserEmail(payload.AccessToken)
	if err != nil {
		return nil, err
	}

	fname, lname, picture, err := getUserProfile(payload.AccessToken)
	if err != nil {
		return nil, err
	}

	convPicture, err := getPhoto(picture, payload.AccessToken)
	if err != nil {
		l.logger.Debug().Msg(err.Error())
	}

	if convPicture != "" {
		picture = convPicture
	}

	return &GetProfileOutput{
		Email: email,
		Name:  fname + " " + lname,
		Photo: picture,
	}, nil
}

func getPhoto(urn, token string) (string, error) {
	endpoint := "https://api.linkedin.com/v2/me?projection=(id,profilePicture(displayImage~digitalmediaAsset:playableStreams))"

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build request")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to do request")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get access token, not ok")
	}
	defer resp.Body.Close()

	var payload PhotoResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("failed to unmarshal response body")
	}

	if len(payload.ProfilePicture.DisplayImage.Elements) == 0 {
		return "", nil
	}
	if len(payload.ProfilePicture.DisplayImage.Elements[0].Identifiers) == 0 {
		return "", nil
	}

	lenIdentifiers := len(payload.ProfilePicture.DisplayImage.Elements[0].Identifiers)
	// return last identifier
	return payload.ProfilePicture.DisplayImage.Elements[0].Identifiers[lenIdentifiers-1].Identifier, nil
}

func getUserProfile(token string) (string, string, string, error) {
	endpoint := "https://api.linkedin.com/v2/me"

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to build request")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to do request")
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("failed to get full user profile, not ok")
	}
	defer resp.Body.Close()

	var payload ProfileResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to unmarshal response body")
	}

	return payload.LocalizedFirstName, payload.LocalizedLastName, payload.ProfilePicture.DisplayImage, nil
}

func getUserEmail(token string) (string, error) {
	endpoint := "https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))"

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	rawJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	fmt.Println(string(rawJSON))

	var payload EmailResponse
	err = json.Unmarshal(rawJSON, &payload)
	if err != nil {
		return "", err
	}

	if len(payload.Elements) <= 0 {
		return "", fmt.Errorf("got empty email list")
	}

	return payload.Elements[0].HandleContent.EmailAddress, nil
}
