package partition

type Category string

const (
	CategoryUser    Category = "user"
	CategoryInvite  Category = "invite"
	CategoryAccount Category = "account"
)

var (
	categories = []Category{
		CategoryUser,
		CategoryInvite,
		CategoryAccount,
	}
)
