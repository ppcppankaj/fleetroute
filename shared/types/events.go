package types

import "time"

// ── Kafka Topic Constants ──────────────────────────────────────────────────
const (
	TopicGatewayRequest     = "fleet.gateway.request"
	TopicVehicleCreated     = "fleet.vehicle.created"
	TopicVehicleUpdated     = "fleet.vehicle.updated"
	TopicTripStarted        = "fleet.trip.started"
	TopicTripCompleted      = "fleet.trip.completed"
	TopicLocationUpdated    = "fleet.location.updated"
	TopicGeofenceBreach     = "fleet.geofence.breach"
	TopicAlertTriggered     = "fleet.alert.triggered"
	TopicAlertResolved      = "fleet.alert.resolved"
	TopicMaintenanceDue     = "fleet.maintenance.due"
	TopicFuelLogged         = "fleet.fuel.logged"
	TopicFuelTheftSuspected = "fleet.fuel.theft.suspected"
	TopicDriverAssigned     = "fleet.driver.assigned"
	TopicDeviceOffline      = "fleet.device.offline"
	TopicDeviceProvisioned  = "fleet.device.provisioned"
	TopicUserLogin          = "fleet.user.login"
	TopicUserAction         = "fleet.user.action"
	TopicTenantCreated      = "fleet.tenant.created"
	TopicSubscriptionUpdate = "fleet.subscription.updated"
	TopicInvoiceCreated     = "fleet.billing.invoice.created"
)

// ── Location / Tracking ───────────────────────────────────────────────────

type LocationUpdatedEvent struct {
	VehicleID      string    `json:"vehicle_id"`
	TenantID       string    `json:"tenant_id"`
	Lat            float64   `json:"lat"`
	Lng            float64   `json:"lng"`
	Speed          float64   `json:"speed"`
	Heading        float64   `json:"heading"`
	Altitude       float64   `json:"altitude"`
	Ignition       bool      `json:"ignition"`
	SignalStrength int       `json:"signal_strength"`
	EngineHours    float64   `json:"engine_hours"`
	TripID         string    `json:"trip_id,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

// ── Trips ─────────────────────────────────────────────────────────────────

type TripStartedEvent struct {
	TripID    string    `json:"trip_id"`
	VehicleID string    `json:"vehicle_id"`
	DriverID  string    `json:"driver_id,omitempty"`
	TenantID  string    `json:"tenant_id"`
	StartLat  float64   `json:"start_lat"`
	StartLng  float64   `json:"start_lng"`
	StartTime time.Time `json:"start_time"`
}

type TripCompletedEvent struct {
	TripID     string    `json:"trip_id"`
	VehicleID  string    `json:"vehicle_id"`
	DriverID   string    `json:"driver_id,omitempty"`
	TenantID   string    `json:"tenant_id"`
	DistanceKM float64   `json:"distance_km"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	FuelUsed   float64   `json:"fuel_used"`
	EndLat     float64   `json:"end_lat"`
	EndLng     float64   `json:"end_lng"`
}

// ── Geofencing ────────────────────────────────────────────────────────────

type GeofenceBreachEvent struct {
	ZoneID    string    `json:"zone_id"`
	ZoneName  string    `json:"zone_name"`
	VehicleID string    `json:"vehicle_id"`
	TenantID  string    `json:"tenant_id"`
	DriverID  string    `json:"driver_id,omitempty"`
	EventType string    `json:"event_type"` // ENTRY | EXIT | DWELL | SPEED
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Speed     float64   `json:"speed,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ── Alerts ────────────────────────────────────────────────────────────────

type AlertTriggeredEvent struct {
	AlertID   string    `json:"alert_id"`
	TenantID  string    `json:"tenant_id"`
	VehicleID string    `json:"vehicle_id"`
	DriverID  string    `json:"driver_id,omitempty"`
	Type      string    `json:"type"`
	Severity  string    `json:"severity"` // LOW | MEDIUM | HIGH | CRITICAL
	Message   string    `json:"message"`
	Metadata  any       `json:"metadata,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type AlertResolvedEvent struct {
	AlertID    string    `json:"alert_id"`
	TenantID   string    `json:"tenant_id"`
	ResolvedBy string    `json:"resolved_by"`
	ResolvedAt time.Time `json:"resolved_at"`
}

// ── Vehicles ──────────────────────────────────────────────────────────────

type VehicleCreatedEvent struct {
	VehicleID   string    `json:"vehicle_id"`
	TenantID    string    `json:"tenant_id"`
	PlateNumber string    `json:"plate_number"`
	Make        string    `json:"make"`
	Model       string    `json:"model"`
	Year        int       `json:"year"`
	FuelType    string    `json:"fuel_type"`
	CreatedAt   time.Time `json:"created_at"`
}

type VehicleUpdatedEvent struct {
	VehicleID string    `json:"vehicle_id"`
	TenantID  string    `json:"tenant_id"`
	Status    string    `json:"status,omitempty"`
	Odometer  float64   `json:"odometer,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ── Drivers ───────────────────────────────────────────────────────────────

type DriverAssignedEvent struct {
	DriverID  string    `json:"driver_id"`
	VehicleID string    `json:"vehicle_id"`
	TenantID  string    `json:"tenant_id"`
	AssignedAt time.Time `json:"assigned_at"`
}

// ── Maintenance ───────────────────────────────────────────────────────────

type MaintenanceDueEvent struct {
	TaskID    string    `json:"task_id"`
	VehicleID string    `json:"vehicle_id"`
	TenantID  string    `json:"tenant_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	DueAt     time.Time `json:"due_at"`
	Odometer  float64   `json:"odometer,omitempty"`
}

// ── Fuel ──────────────────────────────────────────────────────────────────

type FuelLoggedEvent struct {
	LogID     string    `json:"log_id"`
	VehicleID string    `json:"vehicle_id"`
	DriverID  string    `json:"driver_id,omitempty"`
	TenantID  string    `json:"tenant_id"`
	Liters    float64   `json:"liters"`
	TotalCost float64   `json:"total_cost"`
	Odometer  float64   `json:"odometer,omitempty"`
	LoggedAt  time.Time `json:"logged_at"`
}

type FuelTheftSuspectedEvent struct {
	VehicleID  string    `json:"vehicle_id"`
	TenantID   string    `json:"tenant_id"`
	LitersLost float64   `json:"liters_lost"`
	Lat        float64   `json:"lat,omitempty"`
	Lng        float64   `json:"lng,omitempty"`
	DetectedAt time.Time `json:"detected_at"`
}

// ── Devices ───────────────────────────────────────────────────────────────

type DeviceProvisionedEvent struct {
	DeviceID  string    `json:"device_id"`
	IMEI      string    `json:"imei"`
	TenantID  string    `json:"tenant_id"`
	VehicleID string    `json:"vehicle_id,omitempty"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
}

type DeviceOfflineEvent struct {
	DeviceID string    `json:"device_id"`
	IMEI     string    `json:"imei"`
	TenantID string    `json:"tenant_id"`
	LastSeen time.Time `json:"last_seen"`
}

// ── Users ─────────────────────────────────────────────────────────────────

type UserLoginEvent struct {
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Email     string    `json:"email"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Success   bool      `json:"success"`
	LoginAt   time.Time `json:"login_at"`
}

type UserActionEvent struct {
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	ResourceID string   `json:"resource_id,omitempty"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

// ── Tenants ───────────────────────────────────────────────────────────────

type TenantCreatedEvent struct {
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	PlanID    string    `json:"plan_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ── Billing ───────────────────────────────────────────────────────────────

type SubscriptionUpdatedEvent struct {
	TenantID string    `json:"tenant_id"`
	SubID    string    `json:"sub_id"`
	PlanID   string    `json:"plan_id"`
	Status   string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InvoiceCreatedEvent struct {
	InvoiceID string    `json:"invoice_id"`
	TenantID  string    `json:"tenant_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	DueDate   time.Time `json:"due_date"`
	CreatedAt time.Time `json:"created_at"`
}
