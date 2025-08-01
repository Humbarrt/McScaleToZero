package playtime

import (
	"mc_manager/logging"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func getTimezoneFromContext(ctx *gin.Context) (*time.Location, bool, error) {
	timezone := time.UTC
	timezoneString := ctx.Query("tz")
	var err error
	if timezoneString != "" {
		timezone, err = time.LoadLocation(timezoneString)
		if err != nil {
			logging.Error("Invalid timezone format", "error", err, "timezone", timezoneString)
			ctx.JSON(400, gin.H{"error": "Invalid timezone format"})
			return nil, false, err
		}
	}

	usingLocalTime := false
	usingLocalTimeString := ctx.Query("usingLocalTime")
	if usingLocalTimeString != "" {
		usingLocalTime, err = strconv.ParseBool(usingLocalTimeString)
		if err != nil {
			logging.Error("Invalid usingLocalTime format", "error", err, "usingLocalTime", usingLocalTimeString)
			ctx.JSON(400, gin.H{"error": "Invalid usingLocalTime format"})
			return nil, false, err
		}

		cookie := &http.Cookie{
			Name:     "usingLocalTime",
			Value:    strconv.FormatBool(usingLocalTime),
			Path:     "/",
			HttpOnly: true,
		}
		http.SetCookie(ctx.Writer, cookie)
	} else {
		localTimeCookie, err := ctx.Cookie("usingLocalTime")
		if err != nil && err != http.ErrNoCookie {
			logging.Error("Error getting usingLocalTime cookie", "error", err)
			ctx.JSON(500, gin.H{"error": "Error getting usingLocalTime cookie"})
			return nil, false, err
		}

		if localTimeCookie != "" {
			usingLocalTime, err = strconv.ParseBool(localTimeCookie)
			if err != nil {
				logging.Error("Invalid usingLocalTime cookie format", "error", err, "usingLocalTime", localTimeCookie)
				ctx.JSON(400, gin.H{"error": "Invalid usingLocalTime cookie format"})
				return nil, false, err
			}
		}
	}

	if !usingLocalTime {
		timezone = time.UTC
	}
	return timezone, usingLocalTime, nil
}
