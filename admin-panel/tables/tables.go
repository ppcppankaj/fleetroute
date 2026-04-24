package tables

import (
	"database/sql"
	"time"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/action"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

// GetGenerators returns all GoAdmin table generators.
func GetGenerators(pool *sql.DB) table.GeneratorList {
	return table.GeneratorList{
		"devices":           DeviceTable(pool),
		"tenants":           TenantTable(pool),
		"alert_rules":       AlertRulesTable(pool),
		"alerts_history":    AlertsHistoryTable(pool),
		"report_jobs":       ReportJobsTable(pool),
		"firmware_registry": FirmwareRegistryTable(pool),
		"admin_users":       AdminUsersTable(pool),
		"audit_log":         AuditLogTable(pool),
		"system_config":     SystemConfigTable(pool),
	}
}

// ── Device Registry ──────────────────────────────────────────────────────────

func DeviceTable(pool *sql.DB) table.Generator {
	return func(ctx *context.Context) table.Table {
		t := table.NewDefaultTable(table.DefaultConfigWithDriverAndConnection(
			db.DriverPostgresql, "default",
		))

		info := t.GetInfo().
			SetSortField("created_at").
			SetDefaultPageSize(50)

		info.AddField("ID", "id", db.UUID).FieldFilterable()
		info.AddField("IMEI", "imei", db.Varchar).FieldFilterable().FieldSortable()
		info.AddField("Name", "name", db.Varchar).FieldFilterable()
		info.AddField("Protocol", "protocol", db.Varchar).FieldFilterable().
			FieldDisplay(func(value types.FieldModel) interface{} {
				return `<span class="badge badge-primary">` + value.Value + `</span>`
			})
		info.AddField("Firmware", "firmware_version", db.Varchar)
		info.AddField("Tenant", "tenant_id", db.UUID).FieldFilterable()
		info.AddField("Last Seen", "last_seen_at", db.Timestamptz).FieldSortable().
			FieldDisplay(func(value types.FieldModel) interface{} {
				if value.Value == "" || value.Value == "null" {
					return `<span class="text-muted">Never</span>`
				}
				t, err := time.Parse(time.RFC3339, value.Value)
				if err != nil {
					return value.Value
				}
				diff := time.Since(t)
				color := "success"
				if diff > 15*time.Minute {
					color = "warning"
				}
				if diff > 1*time.Hour {
					color = "danger"
				}
				return `<span class="badge badge-` + color + `">` + t.Format("02 Jan 15:04") + `</span>`
			})
		info.AddField("Status", "status", db.Varchar).
			FieldDisplay(func(value types.FieldModel) interface{} {
				color := "success"
				if value.Value != "active" {
					color = "warning"
				}
				return `<span class="badge badge-` + color + `">` + value.Value + `</span>`
			})
		info.AddField("Created", "created_at", db.Timestamptz).FieldSortable()

		info.SetTable("devices").SetTitle("Device Registry").SetDescription("All registered GPS devices across all tenants")

		info.AddActionButton("Force Disconnect", action.Ajax("force_disconnect",
			func(ctx *context.Context) (success bool, msg string, data interface{}) {
				// POST to ingestion service connection registry
				return true, "Force disconnect queued", ""
			}))
		info.AddActionButton("View Packets", action.PopUpWithIframe(
			"view_packets",
			"Packet Inspector",
			action.IframeData{Src: "/admin/custom/packet-inspector?imei={{.Id}}"},
			"950px",
			"600px",
		))

		// Form for editing devices
		formList := t.GetForm()
		formList.AddField("IMEI", "imei", db.Varchar, form.Text).FieldMust()
		formList.AddField("Name", "name", db.Varchar, form.Text)
		formList.AddField("Protocol", "protocol", db.Varchar, form.SelectSingle).
			FieldOptions(types.FieldOptions{
				{Text: "Teltonika", Value: "teltonika"},
				{Text: "GT06", Value: "gt06"},
				{Text: "JT808", Value: "jt808"},
				{Text: "AIS140", Value: "ais140"},
				{Text: "TK103", Value: "tk103"},
			})
		formList.AddField("Status", "status", db.Varchar, form.SelectSingle).
			FieldOptions(types.FieldOptions{
				{Text: "Active", Value: "active"},
				{Text: "Inactive", Value: "inactive"},
				{Text: "Suspended", Value: "suspended"},
			})
		formList.AddField("Firmware Version", "firmware_version", db.Varchar, form.Text)
		formList.SetTable("devices").SetTitle("Device")

		return t
	}
}

// ── Tenant Management ─────────────────────────────────────────────────────────

func TenantTable(pool *sql.DB) table.Generator {
	return func(ctx *context.Context) table.Table {
		t := table.NewDefaultTable(table.DefaultConfigWithDriverAndConnection(
			db.DriverPostgresql, "default",
		))

		info := t.GetInfo().SetSortField("created_at").SetDefaultPageSize(25)
		info.AddField("ID", "id", db.UUID)
		info.AddField("Company", "name", db.Varchar).FieldFilterable().FieldSortable()
		info.AddField("Email", "contact_email", db.Varchar).FieldFilterable()
		info.AddField("Plan", "plan_tier", db.Varchar).
			FieldDisplay(func(value types.FieldModel) interface{} {
				colors := map[string]string{"starter": "info", "professional": "primary", "enterprise": "success"}
				c := colors[value.Value]
				if c == "" {
					c = "secondary"
				}
				return `<span class="badge badge-` + c + `">` + value.Value + `</span>`
			})
		info.AddField("Device Limit", "device_limit", db.Int)
		info.AddField("Status", "status", db.Varchar).
			FieldDisplay(func(value types.FieldModel) interface{} {
				c := "success"
				if value.Value != "active" {
					c = "danger"
				}
				return `<span class="badge badge-` + c + `">` + value.Value + `</span>`
			})
		info.AddField("Created", "created_at", db.Timestamptz).FieldSortable()
		info.SetTable("tenants").SetTitle("Tenant Management").SetDescription("All platform tenants")

		formList := t.GetForm()
		formList.AddField("Company Name", "name", db.Varchar, form.Text).FieldMust()
		formList.AddField("Contact Email", "contact_email", db.Varchar, form.Email).FieldMust()
		formList.AddField("Country", "country", db.Varchar, form.Text)
		formList.AddField("Plan Tier", "plan_tier", db.Varchar, form.SelectSingle).
			FieldOptions(types.FieldOptions{
				{Text: "Starter", Value: "starter"},
				{Text: "Professional", Value: "professional"},
				{Text: "Enterprise", Value: "enterprise"},
			})
		formList.AddField("Device Limit", "device_limit", db.Int, form.Number)
		formList.AddField("Status", "status", db.Varchar, form.SelectSingle).
			FieldOptions(types.FieldOptions{
				{Text: "Active", Value: "active"},
				{Text: "Suspended", Value: "suspended"},
			})
		formList.SetTable("tenants").SetTitle("Tenant")

		return t
	}
}

// ── Alert Rules ───────────────────────────────────────────────────────────────

func AlertRulesTable(pool *sql.DB) table.Generator {
	return func(ctx *context.Context) table.Table {
		t := table.NewDefaultTable(table.DefaultConfigWithDriverAndConnection(
			db.DriverPostgresql, "default",
		))
		info := t.GetInfo().SetSortField("created_at")
		info.AddField("ID", "id", db.UUID)
		info.AddField("Tenant", "tenant_id", db.UUID).FieldFilterable()
		info.AddField("Name", "name", db.Varchar).FieldFilterable()
		info.AddField("Template", "template_id", db.Varchar).FieldFilterable()
		info.AddField("Severity", "severity", db.Varchar).
			FieldDisplay(func(value types.FieldModel) interface{} {
				colors := map[string]string{"info": "info", "warning": "warning", "critical": "danger"}
				c := colors[value.Value]
				if c == "" {
					c = "secondary"
				}
				return `<span class="badge badge-` + c + `">` + value.Value + `</span>`
			})
		info.AddField("Cooldown (s)", "cooldown_s", db.Int)
		info.AddField("Enabled", "enabled", db.Bool).
			FieldDisplay(func(value types.FieldModel) interface{} {
				if value.Value == "1" || value.Value == "true" {
					return `<i class="fa fa-check-circle text-success"></i>`
				}
				return `<i class="fa fa-times-circle text-danger"></i>`
			})
		info.AddField("Triggers", "trigger_count", db.Int).FieldSortable()
		info.AddField("Last Triggered", "last_triggered", db.Timestamptz)
		info.SetTable("alert_rules").SetTitle("Alert Rules").SetDescription("Tenant alert rule definitions")
		return t
	}
}

// ── Alerts History ───────────────────────────────────────────────────────────

func AlertsHistoryTable(pool *sql.DB) table.Generator {
	return func(ctx *context.Context) table.Table {
		t := table.NewDefaultTable(table.DefaultConfigWithDriverAndConnection(
			db.DriverPostgresql, "default",
		))
		info := t.GetInfo().SetSortField("triggered_at").SetDefaultPageSize(100)
		info.AddField("ID", "id", db.UUID)
		info.AddField("Tenant", "tenant_id", db.UUID).FieldFilterable()
		info.AddField("Device", "device_id", db.UUID).FieldFilterable()
		info.AddField("Type", "alert_type", db.Varchar).FieldFilterable()
		info.AddField("Severity", "severity", db.Varchar).FieldFilterable().
			FieldDisplay(func(value types.FieldModel) interface{} {
				colors := map[string]string{"info": "info", "warning": "warning", "critical": "danger"}
				c := colors[value.Value]
				if c == "" {
					c = "secondary"
				}
				return `<span class="badge badge-` + c + `">` + value.Value + `</span>`
			})
		info.AddField("Message", "message", db.Varchar)
		info.AddField("Triggered", "triggered_at", db.Timestamptz).FieldSortable()
		info.AddField("Acknowledged", "acknowledged_at", db.Timestamptz)
		info.SetTable("alerts").SetTitle("Alert History").SetDescription("Cross-tenant alert history")
		return t
	}
}

// ── Report Jobs ───────────────────────────────────────────────────────────────

func ReportJobsTable(pool *sql.DB) table.Generator {
	return func(ctx *context.Context) table.Table {
		t := table.NewDefaultTable(table.DefaultConfigWithDriverAndConnection(
			db.DriverPostgresql, "default",
		))
		info := t.GetInfo().SetSortField("created_at").SetDefaultPageSize(50)
		info.AddField("ID", "id", db.UUID)
		info.AddField("Tenant", "tenant_id", db.UUID).FieldFilterable()
		info.AddField("Type", "report_type", db.Varchar).FieldFilterable()
		info.AddField("Title", "title", db.Varchar)
		info.AddField("Format", "format", db.Varchar)
		info.AddField("Status", "status", db.Varchar).FieldFilterable().
			FieldDisplay(func(value types.FieldModel) interface{} {
				colors := map[string]string{
					"pending":    "warning",
					"processing": "info",
					"completed":  "success",
					"failed":     "danger",
				}
				c := colors[value.Value]
				if c == "" {
					c = "secondary"
				}
				return `<span class="badge badge-` + c + `">` + value.Value + `</span>`
			})
		info.AddField("Progress", "progress_pct", db.Int).
			FieldDisplay(func(value types.FieldModel) interface{} {
				return `<div class="progress" style="height:6px;min-width:80px">
					<div class="progress-bar" style="width:` + value.Value + `%"></div>
				</div>`
			})
		info.AddField("Size", "output_size_b", db.Int)
		info.AddField("Requested", "created_at", db.Timestamptz).FieldSortable()
		info.AddField("Completed", "completed_at", db.Timestamptz)
		info.SetTable("report_jobs").SetTitle("Report Jobs").SetDescription("All report generation jobs across tenants")
		return t
	}
}

// ── Firmware Registry ─────────────────────────────────────────────────────────

func FirmwareRegistryTable(pool *sql.DB) table.Generator {
	return func(ctx *context.Context) table.Table {
		t := table.NewDefaultTable(table.DefaultConfigWithDriverAndConnection(
			db.DriverPostgresql, "default",
		))
		info := t.GetInfo()
		info.AddField("ID", "id", db.UUID)
		info.AddField("Family", "family", db.Varchar).FieldFilterable().FieldSortable()
		info.AddField("Firmware", "firmware_version", db.Varchar).FieldFilterable()
		info.AddField("Codec Support", "codec_support", db.Varchar)
		info.AddField("IO Elements", "io_element_count", db.Int)
		info.AddField("Notes", "notes", db.Varchar)
		info.SetTable("device_firmware_registry").SetTitle("Firmware Registry").SetDescription("Teltonika device family IO element support matrix")
		formList := t.GetForm()
		formList.AddField("Device Family", "family", db.Varchar, form.Text).FieldMust()
		formList.AddField("Firmware Version", "firmware_version", db.Varchar, form.Text).FieldMust()
		formList.AddField("Codec Support", "codec_support", db.Varchar, form.Text)
		formList.AddField("Notes", "notes", db.Varchar, form.TextArea)
		formList.SetTable("device_firmware_registry").SetTitle("Firmware Entry")
		return t
	}
}

// ── Admin Users ───────────────────────────────────────────────────────────────

func AdminUsersTable(pool *sql.DB) table.Generator {
	return func(ctx *context.Context) table.Table {
		t := table.NewDefaultTable(table.DefaultConfigWithDriverAndConnection(
			db.DriverPostgresql, "default",
		))
		info := t.GetInfo()
		info.AddField("ID", "id", db.UUID)
		info.AddField("Username", "username", db.Varchar)
		info.AddField("Email", "email", db.Varchar)
		info.AddField("Role", "role", db.Varchar)
		info.AddField("Last Login", "last_login_at", db.Timestamptz)
		info.AddField("Created", "created_at", db.Timestamptz)
		info.SetTable("goadmin_users").SetTitle("Admin Users").SetDescription("GoAdmin panel user accounts (not tenant users)")
		return t
	}
}

// ── Audit Log ─────────────────────────────────────────────────────────────────

func AuditLogTable(pool *sql.DB) table.Generator {
	return func(ctx *context.Context) table.Table {
		t := table.NewDefaultTable(table.DefaultConfigWithDriverAndConnection(
			db.DriverPostgresql, "default",
		))
		info := t.GetInfo().SetSortField("created_at").SetDefaultPageSize(100)
		info.AddField("ID", "id", db.UUID)
		info.AddField("User", "user_email", db.Varchar).FieldFilterable()
		info.AddField("Action", "action", db.Varchar).FieldFilterable()
		info.AddField("Entity Type", "entity_type", db.Varchar).FieldFilterable()
		info.AddField("Entity ID", "entity_id", db.Varchar).FieldFilterable()
		info.AddField("IP Address", "ip_address", db.Varchar)
		info.AddField("Timestamp", "created_at", db.Timestamptz).FieldSortable()
		info.SetTable("audit_log").SetTitle("Audit Log").SetDescription("All mutations performed through GoAdmin")
		info.HideDeleteButton()
		info.HideNewButton()
		info.HideEditButton()
		return t
	}
}

// ── System Configuration ─────────────────────────────────────────────────────

func SystemConfigTable(pool *sql.DB) table.Generator {
	return func(ctx *context.Context) table.Table {
		t := table.NewDefaultTable(table.DefaultConfigWithDriverAndConnection(
			db.DriverPostgresql, "default",
		))
		info := t.GetInfo()
		info.AddField("Key", "key", db.Varchar).FieldFilterable()
		info.AddField("Value", "value", db.Varchar)
		info.AddField("Description", "description", db.Varchar)
		info.AddField("Updated At", "updated_at", db.Timestamptz)
		info.SetTable("system_config").SetTitle("System Configuration").SetDescription("Platform-wide configuration (DB-backed, hot reload)")
		formList := t.GetForm()
		formList.AddField("Key", "key", db.Varchar, form.Text).FieldMust()
		formList.AddField("Value", "value", db.Varchar, form.TextArea).FieldMust()
		formList.AddField("Description", "description", db.Varchar, form.Text)
		formList.SetTable("system_config").SetTitle("Config Entry")
		return t
	}
}
