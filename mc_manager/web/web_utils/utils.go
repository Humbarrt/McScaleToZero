package web_utils

import (
	"errors"
	"mc_manager/logging"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
)

func Render(ctx *gin.Context, status int, template templ.Component) error {
	ctx.Status(status)
	return template.Render(ctx.Request.Context(), ctx.Writer)
}

func IsAdmin(ctx *gin.Context) (bool, error) {
	isAdminAny, exists := ctx.Get("isAdmin")
	if !exists || isAdminAny == nil {
		logging.Error("Request was not checked for authentication", "exists", exists, "isAdmin", isAdminAny)
		return false, errors.New("Unauthorized")
	}
	isAdmin, ok := isAdminAny.(bool)
	if !ok {
		logging.Error("isAdmin is not a boolean", "isAdmin", isAdminAny)
		return false, errors.New("internal server error")
	}
	return isAdmin, nil
}

func RenderError(ctx *gin.Context, status int, message string) {
	ctx.Header("HX-Reswap", "innerHTML")
	ctx.String(status, message)
}
