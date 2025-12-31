package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/crypto-wallet/backend/internal/infrastructure/messaging"
)

// RatesWebSocketHandler handles WebSocket connections for real-time price updates.
type RatesWebSocketHandler struct {
	pubSubManager messaging.RedisPubSubManager
	logger        *slog.Logger
}

// NewRatesWebSocketHandler creates a new WebSocket handler for rates.
func NewRatesWebSocketHandler(pubSubManager messaging.RedisPubSubManager, logger *slog.Logger) *RatesWebSocketHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &RatesWebSocketHandler{
		pubSubManager: pubSubManager,
		logger:        logger,
	}
}

// Handle processes WebSocket connections.
func (h *RatesWebSocketHandler) Handle(c *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Send connection confirmation
	confirmMsg := map[string]interface{}{
		"event": "connected",
		"data": map[string]interface{}{
			"connection_id": "ws-" + time.Now().Format("20060102150405"),
			"server_time":   time.Now().UTC().Format(time.RFC3339),
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if err := c.WriteJSON(confirmMsg); err != nil {
		h.logger.Error("Failed to send connection confirmation", "error", err)
		return
	}

	// Handle incoming messages and subscriptions
	go h.handleIncomingMessages(ctx, c)

	// Keep connection alive and listen for close
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			h.logger.Info("WebSocket connection closed", "error", err)
			cancel()
			break
		}
	}
}

func (h *RatesWebSocketHandler) handleIncomingMessages(ctx context.Context, c *websocket.Conn) {
	subscribedChannels := make(map[string]bool)

	for {
		var msg map[string]interface{}
		if err := c.ReadJSON(&msg); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return
			}
			h.logger.Error("Failed to read WebSocket message", "error", err)
			continue
		}

		action, ok := msg["action"].(string)
		if !ok {
			continue
		}

		switch action {
		case "subscribe":
			channel, _ := msg["channel"].(string)
			if channel == "prices" {
				symbols, _ := msg["symbols"].([]interface{})
				symbolStrs := make([]string, 0, len(symbols))
				for _, s := range symbols {
					if str, ok := s.(string); ok {
						symbolStrs = append(symbolStrs, str)
					}
				}

				// Subscribe to price updates via Redis Pub/Sub
				err := h.pubSubManager.SubscribePriceUpdates(ctx, symbolStrs, func(ch string, payload []byte) error {
					// Forward message to WebSocket client
					var priceMsg map[string]interface{}
					if err := json.Unmarshal(payload, &priceMsg); err != nil {
						return err
					}
					return c.WriteJSON(priceMsg)
				})

				if err != nil {
					h.logger.Error("Failed to subscribe to prices", "error", err)
				} else {
					for _, sym := range symbolStrs {
						subscribedChannels["prices:"+sym] = true
					}
				}
			}

		case "ping":
			pong := map[string]interface{}{
				"event": "pong",
				"data": map[string]interface{}{
					"server_time": time.Now().UTC().Format(time.RFC3339),
				},
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
			if err := c.WriteJSON(pong); err != nil {
				h.logger.Error("Failed to send pong", "error", err)
			}
		}
	}
}
