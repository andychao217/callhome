package callhome_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/andychao217/callhome"
	"github.com/andychao217/callhome/mocks"
	"github.com/andychao217/callhome/timescale"
	repoMocks "github.com/andychao217/callhome/timescale/mocks"
	"github.com/ip2location/ip2location-go/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetrieve(t *testing.T) {
	ctx := context.TODO()
	t.Run("failed repo save", func(t *testing.T) {
		timescaleRepo := repoMocks.NewTelemetryRepo(t)
		svc := callhome.New(timescaleRepo, nil)
		timescaleRepo.On("RetrieveAll", ctx, callhome.PageMetadata{}).Return(callhome.TelemetryPage{}, timescale.ErrSaveEvent)
		_, err := svc.Retrieve(ctx, callhome.PageMetadata{}, callhome.TelemetryFilters{})
		assert.NotNil(t, err)
		assert.Equal(t, timescale.ErrSaveEvent, err)
	})
	t.Run("success", func(t *testing.T) {
		timescaleRepo := repoMocks.NewTelemetryRepo(t)
		svc := callhome.New(timescaleRepo, nil)
		timescaleRepo.On("RetrieveAll", ctx, callhome.PageMetadata{}).Return(callhome.TelemetryPage{}, nil)
		_, err := svc.Retrieve(ctx, callhome.PageMetadata{}, callhome.TelemetryFilters{})
		assert.Nil(t, err)
	})
}

func TestSave(t *testing.T) {
	ctx := context.TODO()
	t.Run("error obtaining location", func(t *testing.T) {
		timescaleRepo := repoMocks.NewTelemetryRepo(t)
		locMock := mocks.NewLocationService(t)
		locMock.On("GetLocation", "").Return(ip2location.IP2Locationrecord{}, fmt.Errorf("error getting loc"))
		svc := callhome.New(timescaleRepo, locMock)
		err := svc.Save(ctx, callhome.Telemetry{})
		assert.NotNil(t, err)
	})
	t.Run("error saving to timescale", func(t *testing.T) {
		timescaleRepo := repoMocks.NewTelemetryRepo(t)
		locMock := mocks.NewLocationService(t)
		locMock.On("GetLocation", "").Return(ip2location.IP2Locationrecord{
			Latitude:     1.2,
			Longitude:    30,
			Country_long: "SomeCountry",
			City:         "someCity",
		}, nil)
		timescaleRepo.On("Save", ctx, mock.AnythingOfType("callhome.Telemetry")).Return(timescale.ErrSaveEvent)
		svc := callhome.New(timescaleRepo, locMock)
		err := svc.Save(ctx, callhome.Telemetry{})
		assert.NotNil(t, err)
		assert.Equal(t, timescale.ErrSaveEvent, err)
	})
	t.Run("successful save", func(t *testing.T) {
		timescaleRepo := repoMocks.NewTelemetryRepo(t)
		locMock := mocks.NewLocationService(t)
		locMock.On("GetLocation", "").Return(ip2location.IP2Locationrecord{
			Latitude:     1.2,
			Longitude:    30,
			Country_long: "SomeCountry",
			City:         "someCity",
		}, nil)
		timescaleRepo.On("Save", ctx, mock.AnythingOfType("callhome.Telemetry")).Return(nil)
		svc := callhome.New(timescaleRepo, locMock)
		err := svc.Save(ctx, callhome.Telemetry{})
		assert.Nil(t, err)
	})
	t.Run("successful update", func(t *testing.T) {
		timescaleRepo := repoMocks.NewTelemetryRepo(t)
		locMock := mocks.NewLocationService(t)
		locMock.On("GetLocation", "").Return(ip2location.IP2Locationrecord{
			Latitude:     1.2,
			Longitude:    30,
			Country_long: "SomeCountry",
			City:         "someCity",
		}, nil)
		timescaleRepo.On("Save", ctx, mock.AnythingOfType("callhome.Telemetry")).Return(nil)
		svc := callhome.New(timescaleRepo, locMock)
		err := svc.Save(ctx, callhome.Telemetry{})
		assert.Nil(t, err)
	})
}
