package wails

import (
	"context"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/getAlby/hub/api"
	"github.com/getAlby/hub/logger"
)

type mockRouterAPI struct {
	api.API
	paidInvoices []string
}

func (m *mockRouterAPI) SendPayment(ctx context.Context, invoice string, amountMsat *uint64, metadata map[string]interface{}, fromAppId *uint) (*api.SendPaymentResponse, error) {
	m.paidInvoices = append(m.paidInvoices, invoice)
	return &api.SendPaymentResponse{}, nil
}

func (m *mockRouterAPI) LookupSwap(swapId string) (*api.LookupSwapResponse, error) {
	return &api.LookupSwapResponse{}, nil
}

func (m *mockRouterAPI) ListSwaps() (*api.ListSwapsResponse, error) {
	return &api.ListSwapsResponse{}, nil
}

func newTestWailsApp() (*WailsApp, *mockRouterAPI) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockAPI := &mockRouterAPI{}
	return &WailsApp{ctx: context.Background(), api: mockAPI}, mockAPI
}

func TestWailsRequestRouter_PaymentRouteNotMatchedInsideOtherRoute(t *testing.T) {
	for _, method := range []string{"GET", "POST"} {
		t.Run(method, func(t *testing.T) {
			app, mockAPI := newTestWailsApp()
			route := "/api/swaps//api/payments/lnbc1test"

			response := app.WailsRequestRouter(route, method, "")

			assert.Empty(t, mockAPI.paidInvoices)
			assert.Equal(t, "Unhandled route: "+method+" "+route, response.Error)
		})
	}
}

func TestWailsRequestRouter_PaymentRoute(t *testing.T) {
	app, mockAPI := newTestWailsApp()

	response := app.WailsRequestRouter("/api/payments/lnbc1test", "POST", "")

	assert.Equal(t, []string{"lnbc1test"}, mockAPI.paidInvoices)
	assert.Empty(t, response.Error)
}

func TestWailsRequestRouter_PaymentRouteRequiresPost(t *testing.T) {
	app, mockAPI := newTestWailsApp()
	route := "/api/payments/lnbc1test"

	response := app.WailsRequestRouter(route, "GET", "")

	assert.Empty(t, mockAPI.paidInvoices)
	assert.Equal(t, "Unhandled route: GET "+route, response.Error)
}
