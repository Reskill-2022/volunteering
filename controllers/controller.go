package controllers

import (
	"github.com/Reskill-2022/volunteering/errors"
	"github.com/labstack/echo/v4"
)

func (u *UserController) HandleError(c echo.Context, err error, code int) error {
	if code < 100 {
		code = 500
	}

	if code >= 500 {
		u.logger.Err(err).Msg("internal error")
		return c.JSON(code, map[string]interface{}{
			"error": "Internal Server Error. Something Bad Happened!",
		})
	}

	msg := err.Error()
	zErr, ok := err.(errors.Error)
	if ok {
		msg = zErr.Message()
	}

	return c.JSON(code, map[string]interface{}{
		"error": msg,
	})
}

func HandleSuccess(c echo.Context, data interface{}, code int) error {
	return c.JSON(code, map[string]interface{}{
		"payload": data,
	})
}
