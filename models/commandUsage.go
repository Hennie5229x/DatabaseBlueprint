package models

type CommandUsage struct {
	Arguments []UsageItem
	Flags     []UsageItem
	Examples  []string
}

type UsageItem struct {
	Name        string
	Description string
}
