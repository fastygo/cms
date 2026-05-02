package settings

type Key string

type Definition struct {
	Key         Key
	Label       string
	Type        string
	Public      bool
	Default     any
	Description string
}

type Value struct {
	Key    Key
	Value  any
	Public bool
}
