# WebSocket Protocol Specification

**Feature**: Real-Time Price Feed and Notifications
**Version**: 1.0.0
**Date**: 2025-10-14

---

## Overview

WebSocket protocol for real-time cryptocurrency price updates and user notifications. Uses Redis Pub/Sub for message distribution across multiple server instances.

---

## Connection

### Endpoint
```
ws://localhost:8080/ws
wss://api.cryptowallet.example.com/ws
```

### Authentication
```javascript
// JWT token in query parameter
const ws = new WebSocket('wss://api.cryptowallet.example.com/ws?token=JWT_TOKEN_HERE');
```

### Connection Lifecycle
1. Client initiates WebSocket connection with JWT token
2. Server validates token and establishes connection
3. Server sends connection confirmation
4. Client subscribes to channels
5. Server streams updates until disconnect

---

## Message Format

All messages use JSON format with consistent structure:

```json
{
  "event": "event_name",
  "data": {},
  "timestamp": "2025-10-14T10:30:00Z"
}
```

---

## Client → Server Messages

### Subscribe to Price Updates
```json
{
  "action": "subscribe",
  "channel": "prices",
  "symbols": ["BTC", "ETH", "SOL", "XLM"]
}
```

### Unsubscribe from Price Updates
```json
{
  "action": "unsubscribe",
  "channel": "prices",
  "symbols": ["BTC"]
}
```

### Subscribe to User Notifications
```json
{
  "action": "subscribe",
  "channel": "notifications",
  "user_id": "uuid"
}
```

### Heartbeat (Keep-Alive)
```json
{
  "action": "ping"
}
```

---

## Server → Client Messages

### Connection Established
```json
{
  "event": "connected",
  "data": {
    "connection_id": "uuid",
    "server_time": "2025-10-14T10:30:00Z"
  },
  "timestamp": "2025-10-14T10:30:00Z"
}
```

### Price Update (Every 5 seconds max)
```json
{
  "event": "price_update",
  "data": {
    "symbol": "BTC",
    "price_usd": "50000.00",
    "price_change_24h": "2.5",
    "volume_24h": "25000000000.00",
    "timestamp": "2025-10-14T10:30:00Z"
  },
  "timestamp": "2025-10-14T10:30:00Z"
}
```

### Batch Price Update
```json
{
  "event": "price_batch",
  "data": {
    "prices": [
      {"symbol": "BTC", "price_usd": "50000.00", "price_change_24h": "2.5"},
      {"symbol": "ETH", "price_usd": "3200.00", "price_change_24h": "-1.2"},
      {"symbol": "SOL", "price_usd": "110.00", "price_change_24h": "5.8"},
      {"symbol": "XLM", "price_usd": "0.33", "price_change_24h": "0.5"}
    ]
  },
  "timestamp": "2025-10-14T10:30:00Z"
}
```

### Transaction Notification
```json
{
  "event": "transaction_update",
  "data": {
    "transaction_id": "uuid",
    "type": "receive",
    "amount": "1.5",
    "symbol": "ETH",
    "status": "confirmed",
    "confirmations": 12,
    "from_address": "0x...",
    "tx_hash": "0x..."
  },
  "timestamp": "2025-10-14T10:30:00Z"
}
```

### Balance Update
```json
{
  "event": "balance_update",
  "data": {
    "wallet_id": "uuid",
    "balance": "5.25",
    "balance_usd": "168000.00",
    "symbol": "BTC"
  },
  "timestamp": "2025-10-14T10:30:00Z"
}
```

### System Notification
```json
{
  "event": "notification",
  "data": {
    "notification_id": "uuid",
    "type": "kyc_approved",
    "title": "KYC Verification Approved",
    "message": "Your identity verification has been approved",
    "priority": "high"
  },
  "timestamp": "2025-10-14T10:30:00Z"
}
```

### Heartbeat Response
```json
{
  "event": "pong",
  "data": {
    "server_time": "2025-10-14T10:30:00Z"
  },
  "timestamp": "2025-10-14T10:30:00Z"
}
```

### Error
```json
{
  "event": "error",
  "data": {
    "code": "INVALID_TOKEN",
    "message": "Authentication failed",
    "details": {}
  },
  "timestamp": "2025-10-14T10:30:00Z"
}
```

---

## Heartbeat Protocol

- Client sends `ping` every 30 seconds
- Server responds with `pong`
- Server closes connection if no ping received for 60 seconds
- Client reconnects if no pong received for 45 seconds

---

## Reconnection Strategy

### Exponential Backoff
```javascript
let reconnectDelay = 1000; // Start with 1 second
const maxDelay = 30000; // Max 30 seconds

function reconnect() {
  setTimeout(() => {
    connect();
    reconnectDelay = Math.min(reconnectDelay * 2, maxDelay);
  }, reconnectDelay);
}

ws.onclose = () => {
  reconnect();
};

ws.onopen = () => {
  reconnectDelay = 1000; // Reset on successful connection
};
```

### State Synchronization
After reconnection:
1. Resubscribe to all previous channels
2. Request current state (latest prices, pending notifications)
3. Resume normal operation

---

## Rate Limiting

- Max 100 subscribe/unsubscribe actions per minute per connection
- Max 60 ping messages per minute per connection
- Violation results in temporary suspension (1 minute)

---

## Backend Architecture

### Redis Pub/Sub Integration
```
Price Feed Service (CoinGecko WebSocket)
    ↓ publish
Redis Pub/Sub Channel: "prices:*"
    ↓ subscribe
Multiple Backend Servers
    ↓ WebSocket
Connected Clients
```

### Message Flow
1. Price feed service receives updates from CoinGecko
2. Publishes to Redis channel `prices:{symbol}`
3. All backend servers subscribed to Redis channels
4. Backend servers forward to connected WebSocket clients
5. Clients receive real-time updates

---

## Client Implementation Example

```typescript
class CryptoWalletWebSocket {
  private ws: WebSocket;
  private heartbeatInterval: NodeJS.Timeout;
  private reconnectDelay = 1000;

  constructor(private token: string, private url: string) {}

  connect() {
    this.ws = new WebSocket(`${this.url}?token=${this.token}`);

    this.ws.onopen = () => {
      console.log('Connected');
      this.startHeartbeat();
      this.reconnectDelay = 1000;
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };

    this.ws.onclose = () => {
      console.log('Disconnected');
      this.stopHeartbeat();
      this.reconnect();
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  subscribeToPrices(symbols: string[]) {
    this.send({
      action: 'subscribe',
      channel: 'prices',
      symbols
    });
  }

  private startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      this.send({ action: 'ping' });
    }, 30000);
  }

  private stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
    }
  }

  private send(data: any) {
    if (this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  private handleMessage(message: any) {
    switch (message.event) {
      case 'connected':
        console.log('Connection confirmed');
        break;
      case 'price_update':
        this.onPriceUpdate(message.data);
        break;
      case 'transaction_update':
        this.onTransactionUpdate(message.data);
        break;
      case 'error':
        console.error('Server error:', message.data);
        break;
    }
  }

  private reconnect() {
    setTimeout(() => {
      this.connect();
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000);
    }, this.reconnectDelay);
  }

  // Event handlers (override in implementation)
  onPriceUpdate(data: any) {}
  onTransactionUpdate(data: any) {}
}
```

---

## Security Considerations

1. **Authentication**: JWT token validated on connection
2. **Authorization**: Users only receive their own notifications
3. **Rate Limiting**: Prevent abuse and resource exhaustion
4. **Input Validation**: All client messages validated
5. **TLS/SSL**: Use WSS (WebSocket Secure) in production

---

## Performance Requirements

- **Latency**: <100ms from price source to client
- **Throughput**: Support 10,000 concurrent connections per server
- **Message Rate**: Up to 1,000 messages/second per server
- **Reconnection**: <5 seconds to re-establish connection
- **Update Frequency**: Max 1 update per 5 seconds per symbol (SC-004)

---

## Monitoring Metrics

- Active WebSocket connections
- Messages sent/received per second
- Average message delivery latency
- Reconnection rate
- Error rate by type
- Redis Pub/Sub lag
