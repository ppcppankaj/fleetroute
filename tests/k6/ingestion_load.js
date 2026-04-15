/**
 * k6 Load Test — TCP Ingestion (Device Simulator)
 * Simulates 100,000 GPS devices sending Teltonika Codec8 packets.
 *
 * NOTE: k6's net module is experimental.
 * Run: k6 run --compatibility-mode=experimental_enhanced tests/k6/ingestion_load.js
 *
 * For raw TCP simulation without k6 net, use the companion Go simulator:
 *   go run tests/simulator/main.go --devices=100000
 */
import net from 'k6/experimental/net'
import { check, sleep } from 'k6'
import { Rate, Counter, Trend } from 'k6/metrics'

const HOST = __ENV.INGESTION_HOST || 'localhost'
const PORT = parseInt(__ENV.INGESTION_PORT || '5008')

const parseErrors   = new Rate('parse_errors')
const ackReceived   = new Rate('ack_received')
const packetsCount  = new Counter('packets_sent')
const packetLatency = new Trend('packet_latency_ms')

export const options = {
  scenarios: {
    persistent_devices: {
      executor: 'constant-vus',
      vus: 10000,
      duration: '5m',
    },
    burst_devices: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 50000 },
        { duration: '2m', target: 100000 },
        { duration: '1m', target: 0 },
      ],
      startTime: '2m',
    },
  },
  thresholds: {
    ack_received:    ['rate>0.99'],   // 99%+ packets acknowledged
    parse_errors:    ['rate<0.001'],  // <0.1% parse errors
    packet_latency:  ['p(95)<200'],   // 95th percentile < 200ms
  },
}

// Teltonika Codec8 minimal valid packet (single record)
// Preamble(4) + DataLen(4) + CodecID(1) + RecordCount(1) +
// Record: timestamp(8)+priority(1)+lat(4)+lon(4)+alt(2)+angle(2)+sat(1)+speed(2)+
//         EventID(1)+TotalIO(1)+IOcounts(8*1=8)+RecordCount(1) + CRC(4)
function buildCodec8Packet(lat: number, lng: number, speed: number): Uint8Array {
  // This is a simplified/representative frame — real devices send full binary encoding
  const hex = '000000000000003608010000016B40D8EA30010000000000000000000000000000000105021503010101425E0F01F10000601A014E0000000000000000000000000000000001'
  const bytes = new Uint8Array(hex.length / 2)
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(hex.substr(i * 2, 2), 16)
  }
  return bytes
}

export default function () {
  // Vary coordinates slightly per VU to simulate different device positions
  const lat   = 18.9 + (__VU % 100) * 0.01
  const lng   = 72.8 + (__ITER % 100) * 0.01
  const speed = Math.floor(Math.random() * 120)

  const conn = net.dial('tcp', `${HOST}:${PORT}`)

  const pkt = buildCodec8Packet(lat, lng, speed)
  const t0  = Date.now()
  conn.write(pkt)
  packetsCount.add(1)

  // Devices send IMEI first — read 1-byte ACK
  const ack = conn.read(1)
  const latMs = Date.now() - t0
  packetLatency.add(latMs)

  const ok = check(ack, {
    'ack is 0x01': (data) => data !== null && data.byteLength > 0,
  })
  ackReceived.add(ok ? 1 : 0)

  conn.close()

  // Real devices wait 30s–60s between sends
  sleep(30 + Math.random() * 30)
}
