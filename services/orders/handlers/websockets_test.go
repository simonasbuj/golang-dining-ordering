package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang-dining-ordering/config"
	authDto "golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/orders/services"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
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

	buf := &bytes.Buffer{}
	noopHandler := slog.NewTextHandler(buf, nil)
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

func (suite *websocketsHandlerTestSuite) TestHandleUpdateOrder_Success() {
	data := json.RawMessage(`{"status":"locked"}`)

	err := suite.handler.handleUpdateOrder(
		context.Background(),
		&websocket.Conn{},
		testOrderID,
		&authDto.TokenClaimsDto{},
		data,
	)
	suite.Require().NoError(err)
}

func (suite *websocketsHandlerTestSuite) TestHandleDeleteItemdateOrder_Success() {
	data := json.RawMessage(fmt.Sprintf(`{"item_id":"%s"}`, testOrderItemID))

	err := suite.handler.handleDeleteItem(
		context.Background(),
		&websocket.Conn{},
		testOrderID,
		data,
	)
	suite.Require().NoError(err)
}

func (suite *websocketsHandlerTestSuite) TestHandleAddItem_Success() {
	data := json.RawMessage(fmt.Sprintf(`{"item_id":"%s"}`, testOrderItemID))

	err := suite.handler.handleAddItem(
		context.Background(),
		&websocket.Conn{},
		testOrderID,
		data,
	)
	suite.Require().NoError(err)
}

func (suite *websocketsHandlerTestSuite) TesthandleMessage_Success() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	tests := []struct {
		name string
		msg  string
	}{
		{"update order", `{"type": "update_order", "data": {"status": "locked"}}`},
		{
			"add item to order",
			fmt.Sprintf(`{"type": "add_item", "data": {"item_id": "%s"}}`, testOrderItemID),
		},
		{
			"delete item from order",
			fmt.Sprintf(`{"type": "delete_item", "data": {"item_id": "%s"}}`, testOrderItemID),
		},
	}

	for _, tt := range tests {
		msg := []byte(tt.msg)

		err := suite.handler.handleMessage(
			c,
			&websocket.Conn{},
			testOrderID,
			&authDto.TokenClaimsDto{},
			msg,
		)
		suite.Require().NoError(err)
	}
}

func (suite *websocketsHandlerTestSuite) TestHandleOrderWebsocket_BadRequest() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	tests := []struct {
		name          string
		orderID       string
		expectedError string
	}{
		{"invalid order id param", "invalid-id", "parsing uuid from params"},
		{
			"cant establish connection",
			testOrderID.String(),
			"'upgrade' token not found in 'Connection'",
		},
	}

	for _, tt := range tests {
		c.SetParamNames(orderIDParamName)
		c.SetParamValues(tt.orderID)

		err := suite.handler.HandleOrderWebsocket(c)
		suite.Require().Error(err)
		suite.Contains(err.Error(), tt.expectedError)
	}
}
