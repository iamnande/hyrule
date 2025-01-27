package partition

import (
	"errors"
	"slices"
	"strings"
)

const (
	Separator = "#"
)

var (
	ErrInvalidPartition = errors.New("invalid partition")
	ErrInvalidCategory  = errors.New("invalid category")
)

type Partition struct {
	ID       string
	Category Category
}

func Parse(input string) (Partition, error) {
	parts := strings.SplitN(input, Separator, 2)
	if len(parts) != 2 {
		return Partition{}, ErrInvalidPartition
	}
	if !slices.Contains(categories, Category(parts[0])) {
		return Partition{}, ErrInvalidCategory
	}
	return Partition{
		ID:       parts[1],
		Category: Category(parts[0]),
	}, nil
}

func (p Partition) String() string {
	return strings.Join([]string{string(p.Category), p.ID}, Separator)
}
