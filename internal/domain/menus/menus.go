package menus

type ID string
type ItemID string
type Location string

type Menu struct {
	ID       ID
	Name     string
	Location Location
	Items    []Item
}

type Item struct {
	ID       ItemID
	Label    string
	URL      string
	Kind     string
	TargetID string
	Children []Item
}
