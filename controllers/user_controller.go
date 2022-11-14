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
			return u.HandleError(c, errors.New("Responses already recorded. You have applied!", 400), http.StatusBadRequest)
		}

		{
			if requestBody.State == "" {
				return u.HandleError(c, errors.New("Missing Field! State is required", 400), http.StatusBadRequest)
			}
			update.State = requestBody.State

			if requestBody.Organization == "" {
				return u.HandleError(c, errors.New("Missing Field! Organization is required", 400), http.StatusBadRequest)
			}
			update.Organization = requestBody.Organization

			if requestBody.YearsOfExperience == "" {
				return u.HandleError(c, errors.New("Missing Field! Years of Experience is required", 400), http.StatusBadRequest)
			}
			update.YearsOfExperience = requestBody.YearsOfExperience

			if requestBody.VolunteerAreas == nil {
				return u.HandleError(c, errors.New("Missing Field! Volunteer Areas is required", 400), http.StatusBadRequest)
			}
			update.VolunteerAreas = strings.Join(requestBody.VolunteerAreas, ",")

			if requestBody.VolunteerMeans == nil {
				return u.HandleError(c, errors.New("Missing Field! Volunteer Means is required", 400), http.StatusBadRequest)
			}
			update.VolunteerMeans = strings.Join(requestBody.VolunteerMeans, ",")

			if requestBody.Convicted == nil {
				return u.HandleError(c, errors.New("Missing Field! Convicted is required", 400), http.StatusBadRequest)
			}
			update.Convicted = *requestBody.Convicted

			// if requestBody.WillJoinDirectory != nil {
			// 	update.WillJoinDirectory = *requestBody.WillJoinDirectory
			// }

			// if requestBody.SelfSummary != "" {
			// 	update.SelfSummary = requestBody.SelfSummary
			// }

			if requestBody.Representation == "" {
				return u.HandleError(c, errors.New("Missing Field! Representation is required", 400), http.StatusBadRequest)
			}
			update.Representation = requestBody.Representation

			if requestBody.ProvidedName == "" {
				return u.HandleError(c, errors.New("Missing Field! Name is required", 400), http.StatusBadRequest)
			}
			update.ProvidedName = requestBody.ProvidedName
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
