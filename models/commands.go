package models

type Commands struct {
	Name        string
	Description string
	Usage       CommandUsage
	Run         func(args []string)
	Category    CommandType

	SubCommands []Commands
}
