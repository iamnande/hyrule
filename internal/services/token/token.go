package token

import (
	"time"
)

type Token struct {
	access  string
	refresh string
	expires time.Time
	eol     time.Time
}

func (token *Token) Access() string {
	return token.access
}

func (token *Token) Refresh() string {
	return token.refresh
}

func (token *Token) Expires() time.Time {
	return token.expires
}

func (token *Token) EOL() time.Time {
	return token.eol
}
