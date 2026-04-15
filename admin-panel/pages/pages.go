package pages

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/chartjs"
	tmplTypes "github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/action"
)

// ── Dashboard Page ─────────────────────────────────────────────────────────────

type platformStats struct {
	TotalDevices   int
	OnlineDevices  int
	TotalTenants   int
	MsgsPerSecond  float64
	RedisMemMB     float64
	ActiveConns    int
}

func DashboardHandler(db *sql.DB) context.Handler {
	return func(ctx *context.Context) {
		stats := queryPlatformStats(db)

		// Build stat boxes
		statBoxes := template.Default().Box().
			SetStyle("solid").
			SetGap("10").
			SetTheme("blue").
			GetContent()
		_ = statBoxes

		// Line chart for messages/second over last hour (60 data points)
		labels, msgsData := queryMsgsPerSecHistory(db)
		lineChart := chartjs.NewChart().
			SetID("msgs_per_sec_chart").
			SetType("line").
			SetLabels(labels).
			AddDataSet("Msgs/sec").
			DSData(msgsData).
			DSLineTension(0.3).
			DSBorderColor("rgba(60,141,188,0.8)").
			DSBackgroundColor("rgba(60,141,188,0.1)").
			DSFill(true).
			GetContent()

		// Build HTML page
		html := buildDashboardHTML(stats, lineChart)

		ctx.HTML(http.StatusOK, html)
	}
}

func queryPlatformStats(db *sql.DB) platformStats {
	var s platformStats
	_ = db.QueryRow(`SELECT COUNT(*) FROM devices WHERE deleted_at IS NULL`).Scan(&s.TotalDevices)
	_ = db.QueryRow(`SELECT COUNT(*) FROM devices WHERE status='active' AND last_seen_at > now() - interval '15 minutes' AND deleted_at IS NULL`).Scan(&s.OnlineDevices)
	_ = db.QueryRow(`SELECT COUNT(*) FROM tenants WHERE deleted_at IS NULL`).Scan(&s.TotalTenants)
	return s
}

func queryMsgsPerSecHistory(db *sql.DB) ([]string, []float64) {
	rows, err := db.Query(`
		SELECT date_trunc('minute', received_at) AS t, COUNT(*) 
		FROM avl_records 
		WHERE received_at > now() - interval '1 hour'
		GROUP BY t ORDER BY t
	`)
	if err != nil {
		logger.Error("queryMsgsPerSecHistory:", err)
		return demoLabels(), demoData()
	}
	defer rows.Close()
	var labels []string
	var data []float64
	for rows.Next() {
		var t time.Time
		var count int
		if err := rows.Scan(&t, &count); err != nil {
			continue
		}
		labels = append(labels, t.Format("15:04"))
		data = append(data, float64(count)/60.0)
	}
	if len(labels) == 0 {
		return demoLabels(), demoData()
	}
	return labels, data
}

func demoLabels() []string {
	var l []string
	now := time.Now().Truncate(time.Minute)
	for i := 59; i >= 0; i-- {
		l = append(l, now.Add(-time.Duration(i)*time.Minute).Format("15:04"))
	}
	return l
}
func demoData() []float64 {
	return make([]float64, 60)
}

func buildDashboardHTML(s platformStats, chart tmplTypes.HTML) string {
	return fmt.Sprintf(`
<div class="row">
  <div class="col-lg-3 col-xs-6">
    <div class="small-box bg-aqua">
      <div class="inner"><h3>%d</h3><p>Total Devices</p></div>
      <div class="icon"><i class="ion ion-ios-pulse-strong"></i></div>
    </div>
  </div>
  <div class="col-lg-3 col-xs-6">
    <div class="small-box bg-green">
      <div class="inner"><h3>%d</h3><p>Online Now</p></div>
      <div class="icon"><i class="ion ion-wifi"></i></div>
    </div>
  </div>
  <div class="col-lg-3 col-xs-6">
    <div class="small-box bg-yellow">
      <div class="inner"><h3>%d</h3><p>Tenants</p></div>
      <div class="icon"><i class="ion ion-ios-people"></i></div>
    </div>
  </div>
  <div class="col-lg-3 col-xs-6">
    <div class="small-box bg-red">
      <div class="inner"><h3>%.1f</h3><p>Msgs/sec</p></div>
      <div class="icon"><i class="ion ion-ios-speedometer"></i></div>
    </div>
  </div>
</div>
<div class="row">
  <div class="col-md-12">
    <div class="box box-info">
      <div class="box-header with-border">
        <h3 class="box-title">Messages Per Second (last 60 minutes)</h3>
        <div class="box-tools pull-right">
          <button type="button" class="btn btn-box-tool" onclick="location.reload()">
            <i class="fa fa-refresh"></i>
          </button>
        </div>
      </div>
      <div class="box-body">
        %s
      </div>
    </div>
  </div>
</div>
<script>setTimeout(()=>location.reload(), 5000)</script>
`, s.TotalDevices, s.OnlineDevices, s.TotalTenants, s.MsgsPerSecond, string(chart))
}

// ── Packet Inspector Page ──────────────────────────────────────────────────────

func PacketInspectorHandler(db *sql.DB) context.Handler {
	return func(ctx *context.Context) {
		imei := ctx.Query("imei")
		from := ctx.Query("from")
		to := ctx.Query("to")

		var rows []packetRow
		if imei != "" {
			rows = queryPackets(db, imei, from, to)
		}

		html := buildPacketInspectorHTML(imei, from, to, rows)
		ctx.HTML(http.StatusOK, html)
	}
}

type packetRow struct {
	ID          string
	ReceivedAt  time.Time
	SourceIP    string
	SourcePort  int
	Protocol    string
	SizeBytes   int
	CRCOK       bool
	ParseOK     bool
	RecordCount int
	ParseError  string
	RawHex      string
}

func queryPackets(db *sql.DB, imei, from, to string) []packetRow {
	q := `SELECT id, received_at, source_ip, source_port, protocol, packet_size_b, 
		         crc_ok, parse_ok, record_count, COALESCE(parse_error,''), COALESCE(raw_hex,'')
		  FROM packet_log WHERE imei = $1`
	args := []interface{}{imei}
	if from != "" {
		q += ` AND received_at >= $2`
		args = append(args, from)
	}
	if to != "" {
		q += fmt.Sprintf(` AND received_at <= $%d`, len(args)+1)
		args = append(args, to)
	}
	q += ` ORDER BY received_at DESC LIMIT 100`

	rows, err := db.Query(q, args...)
	if err != nil {
		logger.Error("queryPackets:", err)
		return nil
	}
	defer rows.Close()

	var result []packetRow
	for rows.Next() {
		var r packetRow
		_ = rows.Scan(&r.ID, &r.ReceivedAt, &r.SourceIP, &r.SourcePort, &r.Protocol,
			&r.SizeBytes, &r.CRCOK, &r.ParseOK, &r.RecordCount, &r.ParseError, &r.RawHex)
		result = append(result, r)
	}
	return result
}

func buildPacketInspectorHTML(imei, from, to string, rows []packetRow) string {
	rowsHTML := ""
	for _, r := range rows {
		crcBadge := `<span class="badge badge-success">OK</span>`
		if !r.CRCOK {
			crcBadge = `<span class="badge badge-danger">FAIL</span>`
		}
		parseBadge := `<span class="badge badge-success">OK</span>`
		if !r.ParseOK {
			parseBadge = `<span class="badge badge-danger">ERR</span>`
		}
		hexPreview := r.RawHex
		if len(hexPreview) > 80 {
			hexPreview = hexPreview[:80] + "…"
		}
		rowsHTML += fmt.Sprintf(`<tr>
			<td><small class="text-muted">%s</small></td>
			<td>%s:%d</td>
			<td><span class="badge badge-info">%s</span></td>
			<td>%d</td>
			<td>%s</td>
			<td>%s</td>
			<td>%d</td>
			<td><code style="font-size:10px;word-break:break-all">%s</code></td>
			<td><small class="text-danger">%s</small></td>
		</tr>`, r.ReceivedAt.Format("15:04:05.000"), r.SourceIP, r.SourcePort,
			r.Protocol, r.SizeBytes, crcBadge, parseBadge, r.RecordCount,
			hexPreview, r.ParseError)
	}

	return fmt.Sprintf(`
<div class="box box-primary">
  <div class="box-header"><h3 class="box-title">Raw Packet Inspector</h3></div>
  <div class="box-body">
    <form method="GET" class="form-inline" style="margin-bottom:16px">
      <input type="text" name="imei" value="%s" placeholder="IMEI (15 digits)" class="form-control" style="width:200px;margin-right:8px">
      <input type="datetime-local" name="from" value="%s" class="form-control" style="margin-right:8px">
      <input type="datetime-local" name="to" value="%s" class="form-control" style="margin-right:8px">
      <button type="submit" class="btn btn-primary">Inspect</button>
    </form>
    <div class="table-responsive">
      <table class="table table-bordered table-hover table-sm">
        <thead>
          <tr>
            <th>Time</th><th>Source</th><th>Protocol</th><th>Size</th>
            <th>CRC</th><th>Parse</th><th>Records</th><th>Raw Hex</th><th>Error</th>
          </tr>
        </thead>
        <tbody>%s</tbody>
      </table>
    </div>
    %s
  </div>
</div>
`, imei, from, to, rowsHTML,
		func() string {
			if len(rows) == 0 && imei != "" {
				return `<p class="text-muted text-center" style="padding:32px">No packets found for this IMEI in the selected range.</p>`
			}
			return ""
		}())
}

// ── Protocol Statistics Page ───────────────────────────────────────────────────

type protoStat struct {
	Protocol   string
	ConnCount  int
	PacketsPS  float64
	ErrorRate  float64
	CRCFails   int
	AvgSizeB   float64
}

func ProtocolStatsHandler(db *sql.DB) context.Handler {
	return func(ctx *context.Context) {
		stats := queryProtoStats(db)
		recentErrors := queryRecentParseErrors(db)
		ctx.HTML(http.StatusOK, buildProtoStatsHTML(stats, recentErrors))
	}
}

func queryProtoStats(db *sql.DB) []protoStat {
	rows, err := db.Query(`
		SELECT protocol,
		       COUNT(*) FILTER (WHERE received_at > now()-interval '1 minute') AS pps,
		       COUNT(*) FILTER (WHERE NOT parse_ok AND received_at > now()-interval '1 hour') AS errors,
		       COUNT(*) FILTER (WHERE NOT crc_ok AND received_at > now()-interval '1 hour') AS crc_fails,
		       AVG(packet_size_b) AS avg_size
		FROM packet_log
		WHERE received_at > now()-interval '1 hour'
		GROUP BY protocol
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []protoStat
	for rows.Next() {
		var s protoStat
		_ = rows.Scan(&s.Protocol, &s.PacketsPS, &s.ErrorRate, &s.CRCFails, &s.AvgSizeB)
		result = append(result, s)
	}
	return result
}

func queryRecentParseErrors(db *sql.DB) []packetRow {
	rows, err := db.Query(`
		SELECT id, received_at, source_ip, source_port, protocol, packet_size_b,
		       crc_ok, parse_ok, record_count, COALESCE(parse_error,''), COALESCE(LEFT(raw_hex,120),'')
		FROM packet_log
		WHERE NOT parse_ok AND received_at > now()-interval '24 hours'
		ORDER BY received_at DESC LIMIT 50
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []packetRow
	for rows.Next() {
		var r packetRow
		_ = rows.Scan(&r.ID, &r.ReceivedAt, &r.SourceIP, &r.SourcePort, &r.Protocol,
			&r.SizeBytes, &r.CRCOK, &r.ParseOK, &r.RecordCount, &r.ParseError, &r.RawHex)
		result = append(result, r)
	}
	return result
}

func buildProtoStatsHTML(stats []protoStat, errors []packetRow) string {
	cards := ""
	for _, s := range stats {
		cards += fmt.Sprintf(`
		<div class="col-md-3 col-sm-6">
		  <div class="info-box"><span class="info-box-icon bg-aqua"><i class="fa fa-satellite-dish"></i></span>
		    <div class="info-box-content">
		      <span class="info-box-text">%s</span>
		      <span class="info-box-number">%.1f pkt/s</span>
		      <div class="progress"><div class="progress-bar" style="width:%.0f%%"></div></div>
		      <span class="progress-description">CRC failures: %d | Avg size: %.0f B</span>
		    </div>
		  </div>
		</div>`, s.Protocol, s.PacketsPS, s.ErrorRate*100, s.CRCFails, s.AvgSizeB)
	}

	errorRows := ""
	for _, r := range errors {
		errorRows += fmt.Sprintf(`<tr>
			<td>%s</td><td>%s</td><td>%s:%d</td>
			<td><code style="font-size:10px">%s</code></td>
			<td><small class="text-danger">%s</small></td>
		</tr>`, r.ReceivedAt.Format("15:04:05"), r.Protocol, r.SourceIP, r.SourcePort,
			func() string {
				if len(r.RawHex) > 60 {
					return r.RawHex[:60] + "…"
				}
				return r.RawHex
			}(), r.ParseError)
	}

	return fmt.Sprintf(`
<div class="row">%s</div>
<div class="box box-danger" style="margin-top:16px">
  <div class="box-header"><h3 class="box-title">Recent Parse Errors (last 24h)</h3></div>
  <div class="box-body table-responsive">
    <table class="table table-sm table-bordered">
      <thead><tr><th>Time</th><th>Protocol</th><th>Source</th><th>Raw Hex</th><th>Error</th></tr></thead>
      <tbody>%s</tbody>
    </table>
  </div>
</div>
<script>setTimeout(()=>location.reload(), 10000)</script>
`, cards, errorRows)
}

// ── NATS Monitor Page ──────────────────────────────────────────────────────────

type streamInfo struct {
	Name      string `json:"name"`
	Messages  uint64 `json:"num_msgs"`
	Bytes     uint64 `json:"num_bytes"`
	Consumers int    `json:"num_consumers"`
}

func NatsMonitorHandler() context.Handler {
	natsMonitorURL := "http://localhost:8222/jsz?accounts=true&consumers=true&config=true"

	return func(ctx *context.Context) {
		streams := fetchNatsStreams(natsMonitorURL)
		ctx.HTML(http.StatusOK, buildNatsHTML(streams))
	}
}

func fetchNatsStreams(url string) []streamInfo {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return demoStreams()
	}
	defer resp.Body.Close()

	var payload struct {
		Streams []streamInfo `json:"streams"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return demoStreams()
	}
	return payload.Streams
}

func demoStreams() []streamInfo {
	return []streamInfo{
		{Name: "GPS_RAW", Messages: 0, Bytes: 0, Consumers: 1},
		{Name: "GPS_ENRICHED", Messages: 0, Bytes: 0, Consumers: 2},
		{Name: "ALERTS", Messages: 0, Bytes: 0, Consumers: 1},
		{Name: "NOTIFICATIONS", Messages: 0, Bytes: 0, Consumers: 1},
		{Name: "REPORTS", Messages: 0, Bytes: 0, Consumers: 1},
		{Name: "COMMANDS", Messages: 0, Bytes: 0, Consumers: 1},
	}
}

func buildNatsHTML(streams []streamInfo) string {
	rows := ""
	for _, s := range streams {
		sizeMB := float64(s.Bytes) / 1024 / 1024
		rows += fmt.Sprintf(`<tr>
			<td><strong>%s</strong></td>
			<td>%d</td>
			<td>%d</td>
			<td>%.2f MB</td>
		</tr>`, s.Name, s.Messages, s.Consumers, sizeMB)
	}
	return fmt.Sprintf(`
<div class="box box-success">
  <div class="box-header"><h3 class="box-title">JetStream Stream Monitor</h3>
    <div class="box-tools pull-right">
      <button onclick="location.reload()" class="btn btn-box-tool"><i class="fa fa-refresh"></i></button>
    </div>
  </div>
  <div class="box-body table-responsive">
    <table class="table table-bordered table-hover">
      <thead><tr><th>Stream</th><th>Messages</th><th>Consumers</th><th>Storage</th></tr></thead>
      <tbody>%s</tbody>
    </table>
  </div>
</div>
<script>setTimeout(()=>location.reload(),5000)</script>
`, rows)
}

// ── Live Map Page ─────────────────────────────────────────────────────────────

func LiveMapHandler(db *sql.DB) context.Handler {
	return func(ctx *context.Context) {
		devices := queryLiveDevices(db)
		devicesJSON, _ := json.Marshal(devices)
		ctx.HTML(http.StatusOK, buildLiveMapHTML(string(devicesJSON)))
	}
}

type liveDevice struct {
	ID       string  `json:"id"`
	IMEI     string  `json:"imei"`
	Name     string  `json:"name"`
	Tenant   string  `json:"tenant"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Speed    float64 `json:"speed"`
	LastSeen string  `json:"last_seen"`
}

func queryLiveDevices(db *sql.DB) []liveDevice {
	rows, err := db.Query(`
		SELECT d.id::text, d.imei, COALESCE(d.name,''), t.name,
		       COALESCE(a.lat,0), COALESCE(a.lng,0), COALESCE(a.speed,0),
		       COALESCE(d.last_seen_at::text,'')
		FROM devices d
		JOIN tenants t ON t.id = d.tenant_id
		LEFT JOIN LATERAL (
			SELECT lat, lng, speed FROM avl_records
			WHERE device_id = d.id ORDER BY timestamp DESC LIMIT 1
		) a ON true
		WHERE d.deleted_at IS NULL AND d.last_seen_at > now()-interval '24 hours'
		LIMIT 5000
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []liveDevice
	for rows.Next() {
		var d liveDevice
		_ = rows.Scan(&d.ID, &d.IMEI, &d.Name, &d.Tenant, &d.Lat, &d.Lng, &d.Speed, &d.LastSeen)
		result = append(result, d)
	}
	return result
}

func buildLiveMapHTML(devicesJSON string) string {
	return fmt.Sprintf(`
<link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css"/>
<script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
<div id="admin-live-map" style="height:70vh;width:100%%;border-radius:6px"></div>
<script>
const devices = %s;
const map = L.map('admin-live-map').setView([20.5937,78.9629],5);
L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',{
  attribution:'© OpenStreetMap contributors',maxZoom:18
}).addTo(map);
devices.forEach(d=>{
  if(!d.lat||!d.lng)return;
  const color = d.speed>5?'#10b981':d.speed>0?'#f59e0b':'#6b7280';
  const circle = L.circleMarker([d.lat,d.lng],{radius:6,fillColor:color,color:'#fff',weight:1,fillOpacity:0.9});
  circle.bindPopup('<b>'+d.imei+'</b><br>'+d.name+'<br>Tenant: '+d.tenant+'<br>Speed: '+d.speed+' km/h<br>Last: '+d.last_seen);
  circle.addTo(map);
});
</script>
`, devicesJSON)
}

// Ensure action import is used
var _ = action.Ajax
