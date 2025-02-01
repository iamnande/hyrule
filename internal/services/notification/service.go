package notification

import (
	"context"
	"errors"

	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/config"
)

var (
	ErrInvalidChannel = errors.New("invalid channel")
)

type Channel string

const (
	ChannelEmail Channel = "email"
)

type Template int

const (
	TemplateAccountVerification Template = iota
)

type Service struct {
	emailConfig config.Email
}

type Params struct {
	fx.In

	EmailConfig config.Email
}

func NewService(params Params) (*Service, error) {
	return &Service{
		emailConfig: params.EmailConfig,
	}, nil
}

type NotifyEmailInput struct {
	Recipient string

	Template Template
	Metadata map[string]string
}

func (service *Service) NotifyEmail(ctx context.Context, input *NotifyEmailInput) error {
	return nil
}
