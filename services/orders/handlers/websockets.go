package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-dining-ordering/config"
	"golang-dining-ordering/pkg/responses"
	authDto "golang-dining-ordering/services/auth/dto"
	hndl "golang-dining-ordering/services/management/handlers"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/services"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// WebsocketHandler handles orders-related websocket requests.
type WebsocketHandler struct {
	svc        services.OrdersService
	upgrader   *websocket.Upgrader
	logger     *slog.Logger
	orderConns map[uuid.UUID]map[*websocket.Conn]bool
	mu         sync.Mutex
}

// NewWebsocketHandler creates a new Handler for orders websockets.
func NewWebsocketHandler(
	svc services.OrdersService,
	cfg *config.WebsocketConfig,
	logger *slog.Logger,
) *WebsocketHandler {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
		HandshakeTimeout:  time.Duration(cfg.HandshakeTimeout) * time.Second,
		ReadBufferSize:    cfg.ReadBufferSize,
		WriteBufferSize:   cfg.WriteBufferSize,
		WriteBufferPool:   nil,
		Subprotocols:      nil,
		Error:             nil,
		EnableCompression: false,
	}

	return &WebsocketHandler{
		svc:        svc,
		upgrader:   &upgrader,
		logger:     logger,
		orderConns: make(map[uuid.UUID]map[*websocket.Conn]bool),
		mu:         sync.Mutex{},
	}
}

// HandleOrderWebsocket handles websocket connections for ordering.
func (h *WebsocketHandler) HandleOrderWebsocket(c echo.Context) error {
	orderID, err := hndl.GetUUUIDFromParams(c, orderIDParamName)
	if err != nil {
		return err
	}

	user, err := hndl.GetUserFromContext(c, false)
	if err != nil {
		return err
	}

	conn, err := h.upgradeConnection(c)
	if err != nil {
		return err
	}

	defer func() { _ = conn.Close() }()

	h.joinOrder(orderID, conn)
	defer h.leaveOrder(orderID, conn)

	return h.readMessages(c, conn, orderID, user)
}

func (h *WebsocketHandler) upgradeConnection(c echo.Context) (*websocket.Conn, error) {
	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return nil, responses.JSONError(
			c,
			"failed to upgrade websocket",
			err,
			http.StatusInternalServerError,
		)
	}

	return conn, nil
}

func (h *WebsocketHandler) readMessages(
	c echo.Context,
	conn *websocket.Conn,
	orderID uuid.UUID,
	user *authDto.TokenClaimsDto,
) error {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			h.logger.Error("failed to read message", "error", err)

			return err
		}

		err = h.handleMessage(c, conn, orderID, user, msg)
		if err != nil {
			h.logger.Error("failed to handle message", "error", err)

			continue
		}
	}
}

func (h *WebsocketHandler) handleMessage(
	c echo.Context,
	conn *websocket.Conn,
	orderID uuid.UUID,
	user *authDto.TokenClaimsDto,
	msg []byte,
) error {
	var wsDto dto.WSReqMessage

	err := json.Unmarshal(msg, &wsDto)
	if err != nil {
		return h.sendMsg(conn, dto.MsgError, "failed to unmarshal message")
	}

	switch wsDto.Type {
	case dto.MsgAddItem:
		return h.handleAddItem(c.Request().Context(), conn, orderID, wsDto.Data)
	case dto.MsgDeleteItem:
		return h.handleDeleteItem(c.Request().Context(), conn, orderID, wsDto.Data)
	case dto.MsgUpdateOrder:
		return h.handleUpdateOrder(c.Request().Context(), conn, orderID, user, wsDto.Data)
	default:
		return h.sendMsg(conn, dto.MsgError, "unknown request type")
	}
}

func (h *WebsocketHandler) handleAddItem(
	ctx context.Context,
	conn *websocket.Conn,
	orderID uuid.UUID,
	data json.RawMessage,
) error {
	h.logger.Info("Order", "orderid", orderID)

	var reqDto dto.OrderItemRequestDto

	err := h.validateDto(data, &reqDto)
	if err != nil {
		h.logger.Error("dto validation failed", "error", err)
		_ = h.sendMsg(conn, dto.MsgError, err.Error())

		return err
	}

	respDto, err := h.svc.AddItemToOrder(ctx, orderID, reqDto.ItemID)
	if err != nil {
		h.logger.Error("failed to add item to order", "error", err)
		_ = h.sendMsg(conn, dto.MsgError, "failed to add item to order")

		return err
	}

	h.broadcastMessage(orderID, dto.MsgAddItem, respDto)

	return nil
}

func (h *WebsocketHandler) handleDeleteItem(
	ctx context.Context,
	conn *websocket.Conn,
	orderID uuid.UUID,
	data json.RawMessage,
) error {
	var reqDto dto.OrderItemRequestDto

	err := h.validateDto(data, &reqDto)
	if err != nil {
		h.logger.Error("dto validation failed", "error", err)
		_ = h.sendMsg(conn, dto.MsgError, err.Error())

		return err
	}

	respDto, err := h.svc.DeleteOrderItem(ctx, reqDto.ItemID, orderID)
	if err != nil {
		h.logger.Error("failed to delete item from an order", "error", err)
		_ = h.sendMsg(conn, dto.MsgError, "failed to delete item from an order")

		return err
	}

	h.broadcastMessage(orderID, dto.MsgDeleteItem, respDto)

	return nil
}

func (h *WebsocketHandler) handleUpdateOrder(
	ctx context.Context,
	conn *websocket.Conn,
	orderID uuid.UUID,
	user *authDto.TokenClaimsDto,
	data json.RawMessage,
) error {
	var reqDto dto.UpdateOrderReqDto

	reqDto.OrderID = orderID

	err := h.validateDto(data, &reqDto)
	if err != nil {
		h.logger.Error("dto validation failed", "error", err)
		_ = h.sendMsg(conn, dto.MsgError, err.Error())

		return err
	}

	respDto, err := h.svc.UpdateOrder(ctx, &reqDto, user)
	if err != nil {
		h.logger.Error("failed to update order", "error", err)
		_ = h.sendMsg(conn, dto.MsgError, "failed to update an order")

		return err
	}

	h.broadcastMessage(orderID, dto.MsgUpdateOrder, respDto)

	return nil
}

func (h *WebsocketHandler) joinOrder(orderID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.orderConns[orderID] == nil {
		h.orderConns[orderID] = make(map[*websocket.Conn]bool)
	}

	h.orderConns[orderID][conn] = true
}

func (h *WebsocketHandler) leaveOrder(orderID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients, ok := h.orderConns[orderID]
	if ok {
		delete(clients, conn)

		if len(clients) == 0 {
			delete(h.orderConns, orderID)
		}
	}
}

func (h *WebsocketHandler) broadcastMessage(
	orderID uuid.UUID,
	msgType dto.WSMessageType,
	data any,
) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.orderConns[orderID] {
		err := h.sendMsg(client, msgType, data)
		if err != nil {
			h.logger.Error("failed to send message to client", "error", err)

			closeErr := client.Close()
			if closeErr != nil {
				h.logger.Error("failed to close client", "error", err)
			}
		}
	}
}

func (h *WebsocketHandler) sendMsg(
	conn *websocket.Conn,
	msgType dto.WSMessageType,
	data any,
) error {
	errDto := &dto.WSRespMessage{
		Type: msgType,
		Data: data,
	}

	respJSON, err := json.Marshal(errDto)
	if err != nil {
		return fmt.Errorf("marshaling response to json: %w", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, respJSON)
	if err != nil {
		return fmt.Errorf("writing message to client: %w", err)
	}

	return nil
}

func (h *WebsocketHandler) validateDto(data json.RawMessage, dto any) error {
	err := json.Unmarshal(data, &dto)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	err = validator.New().Struct(dto)
	if err != nil {
		return fmt.Errorf("dto validation failed: %w", err)
	}

	return nil
}
