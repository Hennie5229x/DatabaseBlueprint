package models

type Commands struct {
	Name        string
	Description string
	Run         func(args []string)
	Category    CommandType

	SubCommands []Commands
}
