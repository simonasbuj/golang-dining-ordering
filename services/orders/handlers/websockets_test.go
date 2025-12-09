package handlers

import (
	"golang-dining-ordering/config"
	"golang-dining-ordering/services/orders/services"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
)

type websocketsHandlerTestSuite struct {
	suite.Suite

	handler *WebsocketHandler
}

func (suite *websocketsHandlerTestSuite) SetupSuite() {
	mockOrdersRepo := NewMockOrdersRepo()
	svc := services.NewOrdersService(mockOrdersRepo)

	cfg := &config.WebsocketConfig{
		HandshakeTimeout: 5,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
	}
	noopHandler := slog.NewTextHandler(nil, nil)
	logger := slog.New(noopHandler)

	suite.handler = NewWebsocketHandler(svc, cfg, logger)
}

func TestWebsocketsHandlerTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(websocketsHandlerTestSuite))
}

func (suite *websocketsHandlerTestSuite) TestJoinOrder() {
	orderID := uuid.New()

	conn1 := &websocket.Conn{}

	suite.handler.joinOrder(orderID, conn1)

	conns, ok := suite.handler.orderConns[orderID]
	suite.Require().True(ok)
	suite.Require().True(conns[conn1])

	conn2 := &websocket.Conn{}
	suite.handler.joinOrder(orderID, conn2)
	suite.Len(suite.handler.orderConns[orderID], 2)
}

func (suite *websocketsHandlerTestSuite) TestLeaveOrder() {
	orderID := uuid.New()

	conn1 := &websocket.Conn{}
	conn2 := &websocket.Conn{}

	suite.handler.joinOrder(orderID, conn1)
	suite.handler.joinOrder(orderID, conn2)
	suite.Len(suite.handler.orderConns[orderID], 2)

	suite.handler.leaveOrder(orderID, conn1)
	conns, ok := suite.handler.orderConns[orderID]
	suite.True(ok)
	suite.Len(conns, 1)
	suite.False(conns[conn1])

	suite.handler.leaveOrder(orderID, conn2)
	_, ok = suite.handler.orderConns[orderID]
	suite.False(ok)

	// removing non existing connection should not panic
	suite.handler.leaveOrder(orderID, conn1)
}

func (suite *websocketsHandlerTestSuite) TestValidateDto() {
	type TestDto struct {
		Name string `json:"name" validate:"required"`
		Age  int    `json:"age"  validate:"gte=0,lte=120"`
	}

	tests := []struct {
		name         string
		body         string
		shouldError  bool
		errorMessage string
	}{
		{"valid json", `{"name":"sim","age":67}`, false, ""},
		{"invalid json", `{"name":"sim","age":}`, true, "failed to unmarshal message"},
		{"invalid dto in json", `{"name":"","age":200}`, true, "dto validation failed"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.handler.validateDto([]byte(tt.body), &TestDto{})
			if tt.shouldError {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tt.errorMessage)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
