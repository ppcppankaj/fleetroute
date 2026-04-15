/**
 * k6 Load Test — WebSocket Live Tracking
 * Simulates 100,000 concurrent device subscriptions receiving real-time updates.
 *
 * Run: k6 run tests/k6/websocket_load.js -e WS_URL=wss://api.gpsgo.example.com/ws
 */
import ws from 'k6/ws'
import { check } from 'k6'
import { Gauge, Rate, Counter } from 'k6/metrics'

const WS_URL     = __ENV.WS_URL     || 'ws://localhost:8081/ws'
const AUTH_TOKEN = __ENV.AUTH_TOKEN || 'test-jwt-token'
const TENANT_ID  = __ENV.TENANT_ID  || 'test-tenant-uuid'

const activeConnections   = new Gauge('ws_active_connections')
const messagesReceived    = new Counter('ws_messages_received')
const connectionErrors    = new Rate('ws_connection_error_rate')
const messageLatency      = new Gauge('ws_last_message_latency_ms')

export const options = {
  stages: [
    { duration: '30s',  target: 1000   },
    { duration: '1m',   target: 10000  },
    { duration: '2m',   target: 50000  },
    { duration: '3m',   target: 100000 },
    { duration: '2m',   target: 100000 },
    { duration: '1m',   target: 0      },
  ],
  thresholds: {
    ws_connection_error_rate: ['rate<0.01'],
    ws_messages_received:     ['count>0'],
  },
}

export default function () {
  const url    = `${WS_URL}?token=${AUTH_TOKEN}&tenant_id=${TENANT_ID}`
  const params = { tags: { name: 'ws-tracking' } }

  const opened = ws.connect(url, params, (socket) => {
    activeConnections.add(1)

    socket.on('open', () => {
      // Subscribe to all devices for this tenant
      socket.send(JSON.stringify({ type: 'subscribe', filter: 'all' }))
    })

    socket.on('message', (data) => {
      messagesReceived.add(1)
      try {
        const msg = JSON.parse(data)
        if (msg.ts) {
          messageLatency.add(Date.now() - new Date(msg.ts).getTime())
        }
      } catch (_) {}
    })

    socket.on('error', () => {
      connectionErrors.add(1)
    })

    // Hold connection for 30–90 seconds (realistic session duration)
    socket.setTimeout(() => {
      socket.close()
    }, 30000 + Math.random() * 60000)
  })

  check(opened, { 'ws connected': (r) => r && r.status === 101 })
  activeConnections.add(-1)
}
