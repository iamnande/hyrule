package invite

import (
	"context"

	"github.com/getsentry/sentry-go"
	securityprofiles "github.com/iamnande/hyrule/internal/repositories/security-profiles"
	"github.com/segmentio/ksuid"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/config"
)

type SecurityProfileRepository interface {
	SetVerified(
		ctx context.Context,
		request *securityprofiles.SetVerifiedRequest,
	) error
	DescribeSecurityProfile(
		ctx context.Context,
		request *securityprofiles.DescribeSecurityProfileRequest,
	) (*securityprofiles.Record, error)
}

type Service struct {
	config                    config.Invite
	securityProfileRepository SecurityProfileRepository
}

type Params struct {
	fx.In

	Config                    config.Invite
	SecurityProfileRepository SecurityProfileRepository
}

func NewService(params Params) (*Service, error) {
	return &Service{
		config:                    params.Config,
		securityProfileRepository: params.SecurityProfileRepository,
	}, nil
}

type AcceptInviteRequest struct {
	Token ksuid.KSUID
}

func (service *Service) AcceptInvite(ctx context.Context, input *AcceptInviteRequest) error {
	var (
		span = sentry.StartSpan(ctx, "services:invite:AcceptInvite")
	)
	defer span.Finish()
	span.SetTag("request.token", input.Token.String())

	// if err := service.securityProfileRepository.SetVerified(span.Context(), &securityprofiles.SetVerifiedRequest{
	// 	Partition: input.Partition(),
	// 	Verified:  true,
	// }); err != nil {
	// 	return err
	// }

	return nil
}
