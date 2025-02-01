package partition

type Category string

const (
	CategoryUser   Category = "user"
	CategoryInvite Category = "invite"
)

var (
	categories = []Category{
		CategoryUser,
		CategoryInvite,
	}
)
