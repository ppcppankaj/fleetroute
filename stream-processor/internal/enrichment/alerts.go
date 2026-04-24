package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ─── Rule Types ────────────────────────────────────────────────────────────────

type RuleCondition struct {
	Field    string  `json:"field"`    // speed | fuel_level | engine_rpm | ignition | sos | idle_minutes
	Operator string  `json:"op"`       // gt | lt | eq | ne | gte | lte | is_true | is_false
	Value    float64 `json:"value"`
}

type AlertRule struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	Name        string          `json:"name"`
	AlertType   string          `json:"alert_type"`
	Severity    string          `json:"severity"`
	Conditions  []RuleCondition `json:"conditions"` // ALL must be true (AND logic)
	SpeedLimit  int             `json:"speed_limit"` // convenience for overspeed rules
	Cooldown    int             `json:"cooldown_s"`  // seconds between repeated triggers
	Enabled     bool            `json:"enabled"`
}

// ─── Rules Engine ──────────────────────────────────────────────────────────────

type RulesEngine struct {
	pool   *pgxpool.Pool
	rdb    *redis.Client
	logger *zap.Logger
}

func NewRulesEngine(pool *pgxpool.Pool, rdb *redis.Client, logger *zap.Logger) *RulesEngine {
	return &RulesEngine{pool: pool, rdb: rdb, logger: logger}
}

// LoadRules fetches enabled rules for a tenant, using Redis cache (TTL 60s).
func (e *RulesEngine) LoadRules(ctx context.Context, tenantID string) ([]AlertRule, error) {
	cacheKey := fmt.Sprintf("gpsgo:rules:%s", tenantID)

	cached, err := e.rdb.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var rules []AlertRule
		if json.Unmarshal(cached, &rules) == nil {
			return rules, nil
		}
	}

	rows, err := e.pool.Query(ctx, `
		SELECT id, tenant_id, name, alert_type, severity, conditions, speed_limit,
		       COALESCE(cooldown_s, 300), enabled
		FROM alert_rules
		WHERE tenant_id = $1 AND enabled = true AND deleted_at IS NULL
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []AlertRule
	for rows.Next() {
		var r AlertRule
		var condJSON []byte
		if err := rows.Scan(&r.ID, &r.TenantID, &r.Name, &r.AlertType, &r.Severity,
			&condJSON, &r.SpeedLimit, &r.Cooldown, &r.Enabled); err != nil {
			continue
		}
		json.Unmarshal(condJSON, &r.Conditions)
		rules = append(rules, r)
	}

	if b, err := json.Marshal(rules); err == nil {
		e.rdb.Set(ctx, cacheKey, b, 60*time.Second)
	}
	return rules, nil
}

// Evaluate runs all tenant rules against the enriched record.
func (e *RulesEngine) Evaluate(ctx context.Context, rec *EnrichedRecord) {
	rules, err := e.LoadRules(ctx, rec.TenantID)
	if err != nil {
		e.logger.Error("failed to load rules", zap.Error(err))
		// Fall back to hardcoded defaults
		rules = defaultRules(rec.TenantID)
	}

	for _, rule := range rules {
		if e.isOnCooldown(ctx, rule.ID, rec.DeviceID) {
			continue
		}
		if e.matchesAll(rec, rule) {
			e.triggerAlert(ctx, rec, rule)
		}
	}

	// Always check built-in SOS (safety-critical, no rule needed)
	if rec.SOSEvent {
		e.insertAlert(ctx, rec, "sos", "critical", "SOS — Emergency button pressed", nil)
	}
}

// matchesAll returns true when ALL conditions in a rule are satisfied.
func (e *RulesEngine) matchesAll(rec *EnrichedRecord, rule AlertRule) bool {
	// Legacy speed_limit shortcut
	if rule.AlertType == "overspeed" && rule.SpeedLimit > 0 {
		return int(rec.Speed) > rule.SpeedLimit
	}
	if len(rule.Conditions) == 0 {
		return false
	}
	for _, cond := range rule.Conditions {
		if !evalCond(rec, cond) {
			return false
		}
	}
	return true
}

func evalCond(rec *EnrichedRecord, c RuleCondition) bool {
	var actual float64
	switch c.Field {
	case "speed":
		actual = float64(rec.Speed)
	case "fuel_level":
		actual = float64(rec.FuelLevel)
	case "engine_rpm":
		actual = float64(rec.EngineRPM)
	case "battery_level":
		actual = float64(rec.BatteryLevel)
	case "external_voltage":
		actual = rec.ExternalVoltage
	case "temperature_1":
		actual = rec.Temperature1
	case "ignition":
		if c.Operator == "is_true" {
			return rec.Ignition
		}
		return !rec.Ignition
	case "movement":
		if c.Operator == "is_true" {
			return rec.Movement
		}
		return !rec.Movement
	default:
		return false
	}
	switch c.Operator {
	case "gt", ">":
		return actual > c.Value
	case "lt", "<":
		return actual < c.Value
	case "gte", ">=":
		return actual >= c.Value
	case "lte", "<=":
		return actual <= c.Value
	case "eq", "=":
		return actual == c.Value
	case "ne", "!=":
		return actual != c.Value
	}
	return false
}

// isOnCooldown checks Redis for an existing cooldown flag.
func (e *RulesEngine) isOnCooldown(ctx context.Context, ruleID, deviceID string) bool {
	key := fmt.Sprintf("gpsgo:alert_cooldown:%s:%s", ruleID, deviceID)
	n, _ := e.rdb.Exists(ctx, key).Result()
	return n > 0
}

func (e *RulesEngine) setCooldown(ctx context.Context, ruleID, deviceID string, seconds int) {
	key := fmt.Sprintf("gpsgo:alert_cooldown:%s:%s", ruleID, deviceID)
	e.rdb.Set(ctx, key, 1, time.Duration(seconds)*time.Second)
}

func (e *RulesEngine) triggerAlert(ctx context.Context, rec *EnrichedRecord, rule AlertRule) {
	msg := fmt.Sprintf("[%s] %s", rule.Name, ruleMessage(rule, rec))
	extra := map[string]any{"rule_id": rule.ID}
	e.insertAlert(ctx, rec, rule.AlertType, rule.Severity, msg, extra)
	e.setCooldown(ctx, rule.ID, rec.DeviceID, rule.Cooldown)

	// Increment trigger count
	e.pool.Exec(ctx, `
		UPDATE alert_rules SET trigger_count = trigger_count+1, last_triggered=now()
		WHERE id=$1`, rule.ID)
}

func (e *RulesEngine) insertAlert(ctx context.Context, rec *EnrichedRecord,
	alertType, severity, message string, extra map[string]any) {

	extraJSON := "{}"
	if extra != nil {
		if b, err := json.Marshal(extra); err == nil {
			extraJSON = string(b)
		}
	}

	_, err := e.pool.Exec(ctx, `
		INSERT INTO alerts (tenant_id, device_id, alert_type, severity, message,
		                    lat, lng, speed, triggered_at, extra)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
		ON CONFLICT DO NOTHING`,
		rec.TenantID, rec.DeviceID, alertType, severity, message,
		rec.Lat, rec.Lng, rec.Speed, rec.Timestamp, extraJSON)
	if err != nil {
		e.logger.Error("insert alert", zap.String("type", alertType), zap.Error(err))
	} else {
		e.logger.Info("alert triggered",
			zap.String("type", alertType),
			zap.String("severity", severity),
			zap.String("device", rec.DeviceID))
	}
}

func ruleMessage(rule AlertRule, rec *EnrichedRecord) string {
	switch rule.AlertType {
	case "overspeed":
		return fmt.Sprintf("Speed %d km/h exceeds limit of %d km/h", rec.Speed, rule.SpeedLimit)
	case "fuel_theft":
		return fmt.Sprintf("Fuel level dropped to %d%%", rec.FuelLevel)
	case "idle":
		return "Engine idling without movement"
	case "power_cut":
		return "External power disconnected"
	case "harsh_braking":
		return "Harsh braking detected"
	default:
		return rule.Name + " condition triggered"
	}
}

// defaultRules provides safety-net rules when DB fails.
func defaultRules(tenantID string) []AlertRule {
	return []AlertRule{
		{
			ID: "builtin-overspeed", TenantID: tenantID,
			Name: "Overspeed (120 km/h)", AlertType: "overspeed",
			Severity: "warning", SpeedLimit: 120, Cooldown: 300, Enabled: true,
		},
	}
}
