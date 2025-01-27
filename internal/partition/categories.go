package partition

type Category string

const (
	CategoryUser Category = "user"
)

var (
	categories = []Category{
		CategoryUser,
	}
)
