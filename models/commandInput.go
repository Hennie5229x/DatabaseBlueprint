package models

type CommandInput struct {
	RawArgs     []string
	Arguments   []string
	Flags       map[string]bool
	StringFlags map[string]string
}
