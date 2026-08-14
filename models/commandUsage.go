package models

type CommandUsage struct {
	Arguments []UsageItem
	Flags     []Flag
	Examples  []string
}

type UsageItem struct {
	Name        string
	Description string
}

type Flag struct {
	Key         string
	Names       []string
	Description string
}
