package utils

import (
	"fmt"
	"mc_manager/model"
	"strings"
)

func EventsFormatter(events []model.EB_EventType) string {
	var swapEventsString strings.Builder
	for _, event := range events {
		fmt.Fprintf(&swapEventsString, ",sse:%d", event)
	}
	return swapEventsString.String()
}
