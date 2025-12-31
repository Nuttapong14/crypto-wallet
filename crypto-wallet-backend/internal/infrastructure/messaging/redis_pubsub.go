package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Channel patterns for Redis Pub/Sub
	PriceUpdateChannelPrefix = "prices:"  // prices:BTC, prices:ETH, etc.
	PriceBatchChannel        = "prices:batch"
	TransactionChannel       = "transactions"
	BalanceUpdateChannel     = "balance:updates"
	NotificationChannel      = "notifications"

	// Default configuration
	defaultPublishTimeout    = 5 * time.Second
	defaultSubscribeTimeout  = 30 * time.Second
	defaultReconnectDelay    = 1 * time.Second
	defaultMaxReconnectDelay = 30 * time.Second
)

var (
	ErrNilRedisClient   = errors.New("redis pubsub: redis client is not configured")
	ErrPublishFailed    = errors.New("redis pubsub: failed to publish message")
	ErrSubscribeFailed  = errors.New("redis pubsub: failed to subscribe to channel")
	ErrInvalidMessage   = errors.New("redis pubsub: invalid message format")
	ErrChannelClosed    = errors.New("redis pubsub: channel closed")
)

// Message represents a generic message structure for Pub/Sub.
type Message struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// PriceUpdateMessage represents a price update message.
type PriceUpdateMessage struct {
	Symbol          string  `json:"symbol"`
	PriceUSD        string  `json:"price_usd"`
	PriceChange24h  string  `json:"price_change_24h"`
	Volume24h       string  `json:"volume_24h,omitempty"`
	Timestamp       string  `json:"timestamp"`
}

// MessageHandler is a callback function that processes messages from subscribed channels.
type MessageHandler func(channel string, message []byte) error

// RedisPubSubManager manages Redis Pub/Sub operations.
type RedisPubSubManager interface {
	// Publish publishes a message to a channel.
	Publish(ctx context.Context, channel string, message interface{}) error

	// PublishPrice publishes a price update to a specific symbol channel.
	PublishPrice(ctx context.Context, symbol string, priceData *PriceUpdateMessage) error

	// PublishBatchPrices publishes multiple price updates at once.
	PublishBatchPrices(ctx context.Context, prices []*PriceUpdateMessage) error

	// Subscribe subscribes to a channel with a message handler.
	Subscribe(ctx context.Context, channel string, handler MessageHandler) error

	// SubscribePriceUpdates subscribes to price updates for specific symbols.
	SubscribePriceUpdates(ctx context.Context, symbols []string, handler MessageHandler) error

	// SubscribePattern subscribes to channels matching a pattern.
	SubscribePattern(ctx context.Context, pattern string, handler MessageHandler) error

	// Unsubscribe unsubscribes from channels.
	Unsubscribe(ctx context.Context, channels ...string) error

	// Close closes all subscriptions and releases resources.
	Close() error

	// GetSubscribedChannels returns list of currently subscribed channels.
	GetSubscribedChannels() []string
}

// redisPubSubManagerImpl implements RedisPubSubManager.
type redisPubSubManagerImpl struct {
	client         *redis.Client
	logger         *slog.Logger
	pubsub         *redis.PubSub
	subscriptions  map[string]MessageHandler
	publishTimeout time.Duration
	stopCh         chan struct{}
}

// RedisPubSubConfig holds configuration for Redis Pub/Sub manager.
type RedisPubSubConfig struct {
	RedisClient    *redis.Client
	Logger         *slog.Logger
	PublishTimeout time.Duration
}

// NewRedisPubSubManager creates a new Redis Pub/Sub manager.
func NewRedisPubSubManager(config RedisPubSubConfig) (RedisPubSubManager, error) {
	if config.RedisClient == nil {
		return nil, ErrNilRedisClient
	}

	if config.PublishTimeout == 0 {
		config.PublishTimeout = defaultPublishTimeout
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	manager := &redisPubSubManagerImpl{
		client:         config.RedisClient,
		logger:         config.Logger,
		subscriptions:  make(map[string]MessageHandler),
		publishTimeout: config.PublishTimeout,
		stopCh:         make(chan struct{}),
	}

	return manager, nil
}

// Publish publishes a message to a channel.
func (m *redisPubSubManagerImpl) Publish(ctx context.Context, channel string, message interface{}) error {
	if m.client == nil {
		return ErrNilRedisClient
	}

	// Serialize message to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	// Create context with timeout
	pubCtx, cancel := context.WithTimeout(ctx, m.publishTimeout)
	defer cancel()

	// Publish to Redis
	err = m.client.Publish(pubCtx, channel, payload).Err()
	if err != nil {
		m.logger.Error("Failed to publish message",
			"channel", channel,
			"error", err)
		return fmt.Errorf("%w: %v", ErrPublishFailed, err)
	}

	m.logger.Debug("Published message to channel",
		"channel", channel,
		"payload_size", len(payload))

	return nil
}

// PublishPrice publishes a price update to a symbol-specific channel.
func (m *redisPubSubManagerImpl) PublishPrice(ctx context.Context, symbol string, priceData *PriceUpdateMessage) error {
	channel := PriceUpdateChannelPrefix + symbol

	message := Message{
		Event: "price_update",
		Data: map[string]interface{}{
			"symbol":           priceData.Symbol,
			"price_usd":        priceData.PriceUSD,
			"price_change_24h": priceData.PriceChange24h,
			"volume_24h":       priceData.Volume24h,
			"timestamp":        priceData.Timestamp,
		},
		Timestamp: time.Now().UTC(),
	}

	return m.Publish(ctx, channel, message)
}

// PublishBatchPrices publishes multiple price updates to the batch channel.
func (m *redisPubSubManagerImpl) PublishBatchPrices(ctx context.Context, prices []*PriceUpdateMessage) error {
	if len(prices) == 0 {
		return nil
	}

	priceList := make([]map[string]interface{}, 0, len(prices))
	for _, price := range prices {
		priceList = append(priceList, map[string]interface{}{
			"symbol":           price.Symbol,
			"price_usd":        price.PriceUSD,
			"price_change_24h": price.PriceChange24h,
			"volume_24h":       price.Volume24h,
		})
	}

	message := Message{
		Event: "price_batch",
		Data: map[string]interface{}{
			"prices": priceList,
		},
		Timestamp: time.Now().UTC(),
	}

	return m.Publish(ctx, PriceBatchChannel, message)
}

// Subscribe subscribes to a channel with a message handler.
func (m *redisPubSubManagerImpl) Subscribe(ctx context.Context, channel string, handler MessageHandler) error {
	if m.client == nil {
		return ErrNilRedisClient
	}

	// Initialize PubSub if not already done
	if m.pubsub == nil {
		m.pubsub = m.client.Subscribe(ctx)
	}

	// Subscribe to the channel
	err := m.pubsub.Subscribe(ctx, channel)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSubscribeFailed, err)
	}

	// Store handler
	m.subscriptions[channel] = handler

	m.logger.Info("Subscribed to channel", "channel", channel)

	// Start message processing in background if not already running
	go m.processMessages(ctx)

	return nil
}

// SubscribePriceUpdates subscribes to price updates for specific symbols.
func (m *redisPubSubManagerImpl) SubscribePriceUpdates(ctx context.Context, symbols []string, handler MessageHandler) error {
	for _, symbol := range symbols {
		channel := PriceUpdateChannelPrefix + symbol
		if err := m.Subscribe(ctx, channel, handler); err != nil {
			return err
		}
	}
	return nil
}

// SubscribePattern subscribes to channels matching a pattern.
func (m *redisPubSubManagerImpl) SubscribePattern(ctx context.Context, pattern string, handler MessageHandler) error {
	if m.client == nil {
		return ErrNilRedisClient
	}

	// Initialize PubSub if not already done
	if m.pubsub == nil {
		m.pubsub = m.client.Subscribe(ctx)
	}

	// Subscribe to pattern
	err := m.pubsub.PSubscribe(ctx, pattern)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSubscribeFailed, err)
	}

	// Store handler with pattern
	m.subscriptions[pattern] = handler

	m.logger.Info("Subscribed to pattern", "pattern", pattern)

	// Start message processing in background if not already running
	go m.processMessages(ctx)

	return nil
}

// Unsubscribe unsubscribes from channels.
func (m *redisPubSubManagerImpl) Unsubscribe(ctx context.Context, channels ...string) error {
	if m.pubsub == nil {
		return nil
	}

	err := m.pubsub.Unsubscribe(ctx, channels...)
	if err != nil {
		return fmt.Errorf("unsubscribe failed: %w", err)
	}

	// Remove handlers
	for _, ch := range channels {
		delete(m.subscriptions, ch)
	}

	m.logger.Info("Unsubscribed from channels", "channels", channels)

	return nil
}

// Close closes all subscriptions and releases resources.
func (m *redisPubSubManagerImpl) Close() error {
	close(m.stopCh)

	if m.pubsub != nil {
		return m.pubsub.Close()
	}

	return nil
}

// GetSubscribedChannels returns list of currently subscribed channels.
func (m *redisPubSubManagerImpl) GetSubscribedChannels() []string {
	channels := make([]string, 0, len(m.subscriptions))
	for ch := range m.subscriptions {
		channels = append(channels, ch)
	}
	return channels
}

// processMessages processes incoming messages from subscribed channels.
func (m *redisPubSubManagerImpl) processMessages(ctx context.Context) {
	if m.pubsub == nil {
		return
	}

	ch := m.pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Stopping message processing due to context cancellation")
			return
		case <-m.stopCh:
			m.logger.Info("Stopping message processing due to close signal")
			return
		case msg, ok := <-ch:
			if !ok {
				m.logger.Warn("Subscription channel closed")
				return
			}

			// Find handler for this channel
			handler, exists := m.subscriptions[msg.Channel]
			if !exists {
				// Try pattern matching
				for pattern, h := range m.subscriptions {
					if matchPattern(pattern, msg.Channel) {
						handler = h
						break
					}
				}
			}

			if handler == nil {
				m.logger.Warn("No handler found for channel", "channel", msg.Channel)
				continue
			}

			// Process message with handler
			if err := handler(msg.Channel, []byte(msg.Payload)); err != nil {
				m.logger.Error("Failed to process message",
					"channel", msg.Channel,
					"error", err)
			}
		}
	}
}

// matchPattern checks if a channel matches a subscription pattern.
func matchPattern(pattern, channel string) bool {
	// Simple wildcard pattern matching for Redis patterns
	// Redis uses * for wildcards, e.g., "prices:*"
	if pattern == channel {
		return true
	}

	// Check for wildcard pattern
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(channel) >= len(prefix) && channel[:len(prefix)] == prefix
	}

	return false
}
