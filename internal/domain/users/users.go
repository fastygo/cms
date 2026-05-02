package users

type ID string
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted"
)

type User struct {
	ID          ID
	Login       string
	DisplayName string
	Email       string
	Status      Status
	Roles       []string
	Profile     AuthorProfile
}

type AuthorProfile struct {
	ID          ID
	Slug        string
	DisplayName string
	Bio         string
	AvatarURL   string
	WebsiteURL  string
}

func (u User) PublicAuthor() AuthorProfile {
	profile := u.Profile
	profile.ID = u.ID
	if profile.DisplayName == "" {
		profile.DisplayName = u.DisplayName
	}
	if profile.Slug == "" {
		profile.Slug = u.Login
	}
	return profile
}
