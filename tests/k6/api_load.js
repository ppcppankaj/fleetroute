/**
 * k6 Load Test — REST API  
 * Target: 10,000 concurrent virtual users, 5-minute sustained ramp
 * 
 * Run:  k6 run tests/k6/api_load.js -e BASE_URL=https://api.gpsgo.example.com
 */
import http from 'k6/http'
import { check, sleep, group } from 'k6'
import { Trend, Rate, Counter } from 'k6/metrics'

const BASE_URL   = __ENV.BASE_URL   || 'http://localhost:8000'
const TENANT_ID  = __ENV.TENANT_ID  || 'test-tenant-uuid'
const AUTH_TOKEN = __ENV.AUTH_TOKEN || 'test-jwt-token'

// Custom metrics
const vehicleListLatency = new Trend('vehicle_list_latency')
const alertListLatency   = new Trend('alert_list_latency')
const historyLatency     = new Trend('history_latency')
const errorRate          = new Rate('error_rate')
const requestCount       = new Counter('total_requests')

export const options = {
  stages: [
    { duration: '30s',   target: 100   },   // warm up
    { duration: '1m',    target: 1000  },   // ramp to 1k
    { duration: '2m',    target: 5000  },   // ramp to 5k
    { duration: '3m',    target: 10000 },   // peak: 10k concurrent
    { duration: '2m',    target: 10000 },   // sustain peak
    { duration: '1m',    target: 0     },   // wind down
  ],
  thresholds: {
    http_req_duration:    ['p(95)<500', 'p(99)<2000'],
    http_req_failed:      ['rate<0.01'],   // <1% errors
    error_rate:           ['rate<0.02'],
    vehicle_list_latency: ['p(95)<300'],
    alert_list_latency:   ['p(95)<300'],
    history_latency:      ['p(95)<800'],
  },
}

const headers = {
  'Content-Type':  'application/json',
  'Authorization': `Bearer ${AUTH_TOKEN}`,
  'X-Tenant-ID':   TENANT_ID,
}

export default function () {
  group('vehicles', () => {
    const r = http.get(`${BASE_URL}/api/v1/vehicles`, { headers })
    vehicleListLatency.add(r.timings.duration)
    requestCount.add(1)
    const ok = check(r, {
      'vehicles 200': (res) => res.status === 200,
      'vehicles has data': (res) => {
        try { return JSON.parse(res.body).data !== undefined } catch { return false }
      },
    })
    if (!ok) errorRate.add(1)
  })

  sleep(Math.random() * 0.5)

  group('alerts', () => {
    const r = http.get(`${BASE_URL}/api/v1/alerts?limit=20`, { headers })
    alertListLatency.add(r.timings.duration)
    requestCount.add(1)
    const ok = check(r, { 'alerts 200': (res) => res.status === 200 })
    if (!ok) errorRate.add(1)
  })

  sleep(Math.random() * 0.5)

  group('geofences', () => {
    const r = http.get(`${BASE_URL}/api/v1/geofences`, { headers })
    requestCount.add(1)
    check(r, { 'geofences 200': (res) => res.status === 200 })
  })

  sleep(Math.random() * 1)
}
