package mocks

import (
	"context"

	"github.com/andychao217/callhome"
	ip2location "github.com/ip2location/ip2location-go/v9"
	mock "github.com/stretchr/testify/mock"
)

var _ callhome.LocationService = (*LocationService)(nil)

type LocationService struct {
	mock.Mock
}

func (_m *LocationService) GetLocation(ctx context.Context, ip string) (ip2location.IP2Locationrecord, error) {
	ret := _m.Called(ip)

	return ret.Get(0).(ip2location.IP2Locationrecord), ret.Error(1)
}

type mockConstructorTestingTNewLocationService interface {
	mock.TestingT
	Cleanup(func())
}

func NewLocationService(t mockConstructorTestingTNewLocationService) *LocationService {
	mock := &LocationService{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}
