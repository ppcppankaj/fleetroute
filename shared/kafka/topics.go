package kafka

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
