/**
 * k6 Load Test — Report Generation Under Load
 * Validates async report queue behaviour under concurrent submit + poll.
 *
 * Run: k6 run tests/k6/report_load.js -e BASE_URL=https://api.gpsgo.example.com
 */
import http from 'k6/http'
import { check, sleep } from 'k6'
import { Rate, Trend, Counter } from 'k6/metrics'

const BASE_URL   = __ENV.BASE_URL   || 'http://localhost:8000'
const AUTH_TOKEN = __ENV.AUTH_TOKEN || 'test-jwt-token'
const TENANT_ID  = __ENV.TENANT_ID  || 'test-tenant-uuid'

const submitLatency   = new Trend('report_submit_latency')
const pollLatency     = new Trend('report_poll_latency')
const completionRate  = new Rate('report_completion_rate')
const submitErrors    = new Rate('report_submit_errors')
const reportsCreated  = new Counter('reports_created')

export const options = {
  scenarios: {
    report_submitters: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50  },
        { duration: '2m',  target: 200 },
        { duration: '1m',  target: 200 },
        { duration: '30s', target: 0   },
      ],
    },
  },
  thresholds: {
    report_submit_errors:  ['rate<0.05'],   // <5% submit failures
    report_completion_rate: ['rate>0.80'],  // >80% complete within poll window
    report_submit_latency:  ['p(95)<500'],
  },
}

const headers = {
  'Content-Type':  'application/json',
  'Authorization': `Bearer ${AUTH_TOKEN}`,
  'X-Tenant-ID':   TENANT_ID,
}

const REPORT_TYPES = ['trip', 'fuel', 'driver_behavior', 'idle', 'overspeed', 'geofence_violations']

export default function () {
  const reportType = REPORT_TYPES[__VU % REPORT_TYPES.length]
  const toDate   = new Date().toISOString().slice(0, 10)
  const fromDate = new Date(Date.now() - 7 * 86400_000).toISOString().slice(0, 10)

  // Submit report job
  const submitStart = Date.now()
  const submitResp = http.post(
    `${BASE_URL}/api/v1/reports`,
    JSON.stringify({
      report_type: reportType,
      format:      'csv',
      parameters:  { from: fromDate, to: toDate },
    }),
    { headers }
  )
  submitLatency.add(Date.now() - submitStart)
  reportsCreated.add(1)

  const submitted = check(submitResp, {
    'submit 201': (r) => r.status === 201,
    'has job id': (r) => {
      try { return !!JSON.parse(r.body).data?.id } catch { return false }
    },
  })

  if (!submitted) { submitErrors.add(1); return }

  let jobID: string
  try {
    jobID = JSON.parse(submitResp.body).data.id
  } catch { submitErrors.add(1); return }

  // Poll for completion (up to 30 seconds)
  let completed = false
  for (let attempt = 0; attempt < 10; attempt++) {
    sleep(3)
    const pollStart = Date.now()
    const pollResp = http.get(`${BASE_URL}/api/v1/reports/${jobID}`, { headers })
    pollLatency.add(Date.now() - pollStart)

    check(pollResp, { 'poll 200': (r) => r.status === 200 })

    try {
      const job = JSON.parse(pollResp.body).data
      if (job.status === 'completed') {
        completed = true
        check(pollResp, { 'has output url': () => !!job.output_url })
        break
      }
      if (job.status === 'failed') break
    } catch { break }
  }

  completionRate.add(completed ? 1 : 0)
  sleep(1)
}
