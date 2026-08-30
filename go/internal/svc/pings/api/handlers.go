package api

import (
	"context"

	transporterrors "github.com/iamnande/hyrule/go/internal/lib/rest/transport/errors"
	"github.com/iamnande/hyrule/go/internal/lib/rest/transport/validation"
	"github.com/iamnande/hyrule/go/internal/svc/pings/domain"
)

type pingRecorder interface {
	Record(ctx context.Context, name string, kind domain.Kind) (domain.Ping, error)
	List(ctx context.Context) ([]domain.Ping, error)
}

type Handlers struct {
	service pingRecorder
}

func NewHandlers(service pingRecorder) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) RecordPing(ctx context.Context, request RecordPingRequestObject) (RecordPingResponseObject, error) {
	if request.Body == nil || request.Body.Name == "" {
		return RecordPing400JSONResponse{toBadRequest(transporterrors.NewBadRequestError(ctx, transporterrors.InvalidAttribute{
			Path:   "name",
			Rule:   validation.Required,
			Reason: "is required",
			Source: transporterrors.AttributeSourceBody,
		}))}, nil
	}
	if !request.Body.Kind.Valid() {
		return RecordPing400JSONResponse{toBadRequest(transporterrors.NewBadRequestError(ctx, transporterrors.InvalidAttribute{
			Path:   "kind",
			Rule:   validation.Unsupported,
			Reason: "must be one of host, service, app",
			Source: transporterrors.AttributeSourceBody,
		}))}, nil
	}

	ping, err := h.service.Record(ctx, request.Body.Name, domain.Kind(request.Body.Kind))
	if err != nil {
		return RecordPing500JSONResponse{toInternalError(transporterrors.NewInternalServerError(ctx, err))}, nil
	}
	return RecordPing200JSONResponse(toAPIPing(ping)), nil
}

func (h *Handlers) ListPings(ctx context.Context, _ ListPingsRequestObject) (ListPingsResponseObject, error) {
	pings, err := h.service.List(ctx)
	if err != nil {
		return ListPings500JSONResponse{toInternalError(transporterrors.NewInternalServerError(ctx, err))}, nil
	}
	result := make(ListPings200JSONResponse, len(pings))
	for i, ping := range pings {
		result[i] = toAPIPing(ping)
	}
	return result, nil
}

func toAPIPing(ping domain.Ping) Ping {
	return Ping{
		Name:        ping.Name,
		Kind:        Kind(ping.Kind),
		FirstSeenAt: ping.FirstSeenAt,
		LastSeenAt:  ping.LastSeenAt,
		State:       State(ping.State),
	}
}

func toBadRequest(err *transporterrors.BaseError) BadRequestJSONResponse {
	return BadRequestJSONResponse(toAPIError(err))
}

func toInternalError(err *transporterrors.BaseError) InternalServerErrorJSONResponse {
	return InternalServerErrorJSONResponse(toAPIError(err))
}

func toAPIError(err *transporterrors.BaseError) Error {
	apiErr := Error{
		Code:        err.Code.String(),
		Name:        err.Name,
		Message:     err.Message,
		OperationId: err.OperationID,
	}
	if len(err.InvalidAttributes) == 0 {
		return apiErr
	}
	attrs := make([]InvalidAttribute, len(err.InvalidAttributes))
	for i, attr := range err.InvalidAttributes {
		attrs[i] = InvalidAttribute{
			Path:   attr.Path,
			Reason: attr.Reason,
			Rule:   string(attr.Rule),
			Source: InvalidAttributeSource(attr.Source),
		}
	}
	apiErr.InvalidAttributes = &attrs
	return apiErr
}
