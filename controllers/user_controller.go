package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/Reskill-2022/volunteering/errors"
	"github.com/Reskill-2022/volunteering/linkedin"
	"github.com/Reskill-2022/volunteering/model"
	"github.com/Reskill-2022/volunteering/repository"
	"github.com/Reskill-2022/volunteering/requests"
)

type UserController struct {
	logger zerolog.Logger
}

func NewUserController(logger zerolog.Logger) *UserController {
	return &UserController{logger}
}

func (u *UserController) CreateUser(userCreator repository.UserCreator, service linkedin.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var requestBody requests.CreateUserRequest

		err := json.NewDecoder(c.Request().Body).Decode(&requestBody)
		if err != nil {
			return u.HandleError(c, errors.New("Invalid JSON Request Body", 400), http.StatusBadRequest)
		}

		authCode := requestBody.AuthCode
		if authCode == "" {
			return u.HandleError(c, errors.New("Auth Code is required", 400), http.StatusBadRequest)
		}
		redirectURI := requestBody.RedirectURI
		if redirectURI == "" {
			return u.HandleError(c, errors.New("Redirect URI is required", 400), http.StatusBadRequest)
		}

		fmt.Printf("Auth Code: %s, Redirect URI: %s", authCode, redirectURI)

		profile, err := service.GetProfile(authCode, redirectURI)
		if err != nil {
			u.logger.Err(err).Msg("Error getting profile")
			return u.HandleError(c, errors.New("Failed to Validate LinkedIn Profile", 400), http.StatusBadRequest)
		}

		// do validations
		if profile.Name == "" {
			return u.HandleError(c, errors.New("Invalid Profile. Found No Name", 400), http.StatusBadRequest)
		}

		if profile.Photo == "" {
			return u.HandleError(c, errors.New("Invalid Profile. Please Set Your Profile Picture on LinkedIn", 400), http.StatusBadRequest)
		}

		firstname, lastname := u.splitNames(profile.Name)

		data := model.User{
			Email:     profile.Email,
			Name:      profile.Name,
			FirstName: firstname,
			LastName:  lastname,
			Phone:     profile.Phone,
			Photo:     profile.Photo,
			CreatedAt: time.Now().UTC(),
		}

		user, err := userCreator.CreateUser(ctx, data)
		if err != nil {
			return u.HandleError(c, err, errors.CodeFrom(err))
		}

		return HandleSuccess(c, user, http.StatusCreated)
	}
}

func (u *UserController) UpdateUser(userGetter repository.UserGetter, userUpdater repository.UserUpdater) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var requestBody requests.UpdateUserRequest

		err := json.NewDecoder(c.Request().Body).Decode(&requestBody)
		if err != nil {
			return u.HandleError(c, err, http.StatusBadRequest)
		}

		update, err := userGetter.GetUser(ctx, c.Param("email"))
		if err != nil {
			return u.HandleError(c, err, errors.CodeFrom(err))
		}
		if update.Enrolled {
			return u.HandleError(c, errors.New("User Already Enrolled", 400), http.StatusBadRequest)
		}

		{
			if requestBody.State == "" {
				return u.HandleError(c, errors.New("Missing Fields! LinkedIn URL is required", 400), http.StatusBadRequest)
			}
			update.State = requestBody.State

			if requestBody.Organization != "" {
				update.Organization = requestBody.Organization
			}

			if requestBody.YearsOfExperience != "" {
				update.YearsOfExperience = requestBody.YearsOfExperience
			}

			if len(requestBody.VolunteerAreas) == 0 {
				return u.HandleError(c, errors.New("Missing Fields! Volunteer Areas is required", 400), http.StatusBadRequest)
			}
			update.VolunteerAreas = strings.Join(requestBody.VolunteerAreas, ",")

			if requestBody.IsUnderrepresented != nil {
				update.IsUnderrepresented = *requestBody.IsUnderrepresented
			}

			if requestBody.IsConvicted != nil {
				update.IsConvicted = *requestBody.IsConvicted
			}
		}

		update.Enrolled = true
		user, err := userUpdater.UpdateUser(ctx, *update)
		if err != nil {
			return u.HandleError(c, err, errors.CodeFrom(err))
		}

		return HandleSuccess(c, user, http.StatusOK)
	}
}

func (u *UserController) GetUser(userGetter repository.UserGetter) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		userEmail := c.Param("email")
		if userEmail == "" {
			return u.HandleError(c, errors.New(" Email is required", 400), http.StatusBadRequest)
		}

		user, err := userGetter.GetUser(ctx, userEmail)
		if err != nil {
			return u.HandleError(c, err, errors.CodeFrom(err))
		}

		return HandleSuccess(c, user, http.StatusOK)
	}
}

func (u *UserController) splitNames(name string) (string, string) {
	names := strings.Split(name, " ")
	if len(names) == 1 {
		return names[0], ""
	}
	return names[0], names[len(names)-1]
}
