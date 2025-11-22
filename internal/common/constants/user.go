package constants

type Role int

const (
	User Role = iota
	Admin
	Super
)

type Status int

func (r Role) String() string {
	switch r {
	case Admin:
		return "admin"
	case Super:
		return "super"
	default:
		return "user"
	}
}

const (
	Active Status = iota
	Disabled
	Deleted
)

func (s Status) String() string {
	switch s {
	case Disabled:
		return "disabled"
	case Deleted:
		return "deleted"
	default:
		return "active"
	}
}
