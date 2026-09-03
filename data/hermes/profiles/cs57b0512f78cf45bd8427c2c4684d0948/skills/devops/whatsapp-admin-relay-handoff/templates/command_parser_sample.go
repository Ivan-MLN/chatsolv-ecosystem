package main

import (
	"fmt"
	"strings"
)

type CommandType string

const (
	CmdNone   CommandType = "NONE"
	CmdAccept CommandType = "ACCEPT"
	CmdDone   CommandType = "DONE"
	CmdRelay  CommandType = "RELAY"
)

type ParsedCommand struct {
	Type           CommandType
	ConversationID string
	Text           string
}

func ParseRelayCommand(text string) ParsedCommand {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ParsedCommand{Type: CmdNone}
	}

	upper := strings.ToUpper(trimmed)
	if strings.HasPrefix(upper, "#ACC") {
		parts := strings.Fields(trimmed)
		convID := ""
		if len(parts) > 1 {
			convID = strings.TrimPrefix(parts[1], "#")
			convID = strings.TrimPrefix(convID, "CNV-")
			convID = strings.TrimPrefix(convID, "cnv-")
		}
		return ParsedCommand{
			Type:           CmdAccept,
			ConversationID: convID,
		}
	}

	if upper == "#DONE" || upper == "#CLOSE" {
		return ParsedCommand{
			Type: CmdDone,
		}
	}

	if strings.HasPrefix(trimmed, "#") {
		content := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
		if content != "" {
			return ParsedCommand{
				Type: CmdRelay,
				Text: content,
			}
		}
	}

	return ParsedCommand{Type: CmdNone}
}

func main() {
	sample := "#ACC CNV-1234abcd"
	cmd := ParseRelayCommand(sample)
	fmt.Printf("Command: %+v\n", cmd)
}
