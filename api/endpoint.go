package api

import (
	"context"

	"github.com/andychao217/callhome"
	"github.com/go-kit/kit/endpoint"
)

func saveEndpoint(svc callhome.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(saveTelemetryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		tel := callhome.Telemetry{
			Service:     req.Service,
			IpAddress:   req.IpAddress,
			Version:     req.Version,
			ServiceTime: req.LastSeen,
		}
		if err := svc.Save(ctx, tel); err != nil {
			return nil, err
		}
		res := saveTelemetryRes{
			created: true,
		}
		return res, nil
	}
}

func retrieveEndpoint(svc callhome.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(listTelemetryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		pm := callhome.PageMetadata{
			Offset: req.offset,
			Limit:  req.limit,
		}
		filter := callhome.TelemetryFilters{
			From:    req.from,
			To:      req.to,
			Country: req.country,
			City:    req.city,
			Version: req.version,
			Service: req.service,
		}
		tm, err := svc.Retrieve(ctx, pm, filter)
		if err != nil {
			return nil, err
		}
		res := telemetryPageRes{
			pageRes: pageRes{
				Total:  tm.Total,
				Offset: tm.Offset,
				Limit:  tm.Limit,
			},
			Telemetry: tm.Telemetry,
		}
		return res, nil
	}
}

func retrieveSummaryEndpoint(svc callhome.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(listTelemetryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		filter := callhome.TelemetryFilters{
			From:    req.from,
			To:      req.to,
			Country: req.country,
			City:    req.city,
			Version: req.version,
			Service: req.service,
		}
		summary, err := svc.RetrieveSummary(ctx, filter)
		if err != nil {
			return nil, err
		}
		return telemetrySummaryRes{
			Countries:        summary.Countries,
			Cities:           summary.Cities,
			Services:         summary.Services,
			Versions:         summary.Versions,
			TotalDeployments: summary.TotalDeployments,
		}, nil
	}
}

func serveUI(svc callhome.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(listTelemetryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		filter := callhome.TelemetryFilters{
			From:    req.from,
			To:      req.to,
			Country: req.country,
			City:    req.city,
			Version: req.version,
			Service: req.service,
		}
		res, err := svc.ServeUI(ctx, filter)
		return uiRes{
			html: res,
		}, err
	}
}
