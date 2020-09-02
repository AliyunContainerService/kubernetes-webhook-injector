package openapi

import (
	"strings"
)

type ErrorMessage struct {
	Raw       string
	ErrorCode string
	Recommend string
	RequestID string
	Message   string
}

func ParseErrorMessage(s string) (msg ErrorMessage) {
	msg.Raw = s
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		tks := strings.Split(line, ":")
		switch tks[0] {
		case "ErrorCode":
			msg.ErrorCode = line
		case "Recommend":
			msg.Recommend = line
		case "RequestId":
			msg.RequestID = line
		case "Message":
			msg.Message = line
		}
	}
	return
}
