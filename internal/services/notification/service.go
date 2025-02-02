package notification

import (
	"context"
	"errors"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/jordan-wright/email"
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

func (template Template) String() string {
	switch template {
	case TemplateAccountVerification:
		return "account-verification"
	default:
		return "unknown"
	}
}

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
	var (
		span = sentry.StartSpan(ctx, "services:notification:NotifyEmail")
	)
	defer span.Finish()
	span.SetTag("request.template", input.Template.String())

	// send verification email
	// TODO: replace with an event produced to the notification outbox
	// TODO: event outbox?
	emailHost := service.emailConfig.Host
	emailPort := service.emailConfig.Port
	emailEndpoint := fmt.Sprintf("%s:%d", emailHost, emailPort)
	verificationEmail := email.NewEmail()
	verificationEmail.To = []string{input.Recipient}
	verificationEmail.From = "Hyrule <noreply@hyrule.com>"
	verificationEmail.Subject = "[MHQ] Welcome to Hyrule!"

	// TODO: validation based on template required metadata
	userName := input.Metadata["user.name"]
	userPlan := input.Metadata["plan.name"]
	verificationURL := input.Metadata["verify.url"]

	switch input.Template {
	case TemplateAccountVerification:
		verificationEmail.HTML = []byte(fmt.Sprintf(`<h1>Welcome to Hyrule!</h1>
	    <p>Thank you %s for signing up for our %s plan on Hyrule. We're excited to have you on board!</p>
        <p>Please verify your account by clicking here: <a href="%s">Verify Account</a></p>
		<p>Thanks,</p>
		<p>The Hyrule Team</p>
	`, userName, userPlan, verificationURL))
	default:
		return ErrInvalidChannel
	}

	return verificationEmail.Send(emailEndpoint, nil)
}
