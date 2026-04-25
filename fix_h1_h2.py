import re

with open('api-service/internal/handler/handlers.go', 'r') as f:
    content = f.read()

# H1 fixes
# 1. replace vals, _ := rows.Values() in ReportHandler
def fix_vals_ignore(match):
    return """vals, err := rows.Values()
		if err != nil {
			h.logger.Error("scan row", zap.Error(err))
			continue
		}"""

content = re.sub(r'vals,\s*_\s*:=\s*rows\.Values\(\)', fix_vals_ignore, content)

# 2. Add rows.Err() after each for rows.Next() { ... }
# We can find the end of the loop and insert the check.
def fix_rows_err(match):
    var_name = match.group(1)
    return f"""	}}
	if err := rows.Err(); err != nil {{
		h.logger.Error("rows error", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}}
	if {var_name} == nil {{
		{var_name} = []map[string]any{{}}
	}}
	respondOK(c, {var_name})"""

content = re.sub(r'\t\}\n\trespondOK\(c, (vs|gs|alerts|trips|items|ds)\)', fix_rows_err, content)

# H2 fixes
# GeofenceHandler:
content = re.sub(r'func \(h \*GeofenceHandler\) Get\(c \*gin\.Context\)\s*\{\s*respondOK\(c, gin\.H\{\}\)\s*\}', r'''func (h *GeofenceHandler) Get(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	id := c.Param("id")
	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT id, name, shape_type, ST_AsGeoJSON(geometry)::text, created_at
		 FROM geofences WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		id, tenantID,
	)
	if err != nil {
		h.logger.Error("geofence get query", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()
	if !rows.Next() {
		respondError(c, http.StatusNotFound, "geofence not found")
		return
	}
	vals, err := rows.Values()
	if err != nil {
		h.logger.Error("geofence get scan", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	respondOK(c, map[string]any{"id": vals[0], "name": vals[1], "shape_type": vals[2], "geometry": vals[3], "created_at": vals[4]})
}''', content)

content = re.sub(r'func \(h \*GeofenceHandler\) Update\(c \*gin\.Context\)\s*\{\s*c\.JSON\(http\.StatusNoContent, nil\)\s*\}', r'''func (h *GeofenceHandler) Update(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	var body struct {
		Name      string `json:"name" binding:"required"`
		ShapeType string `json:"shape_type" binding:"required"`
		GeoJSON   string `json:"geojson" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	tag, err := h.pool.Exec(c.Request.Context(),
		`UPDATE geofences SET name=$1, shape_type=$2, geometry=ST_GeomFromGeoJSON($3)
		 WHERE id=$4 AND tenant_id=$5 AND deleted_at IS NULL`,
		body.Name, body.ShapeType, body.GeoJSON, c.Param("id"), tenantID,
	)
	if err != nil {
		h.logger.Error("geofence update", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	if tag.RowsAffected() == 0 {
		respondError(c, http.StatusNotFound, "geofence not found")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}''', content)

content = re.sub(r'func \(h \*GeofenceHandler\) Delete\(c \*gin\.Context\)\s*\{\s*c\.JSON\(http\.StatusNoContent, nil\)\s*\}', r'''func (h *GeofenceHandler) Delete(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	tag, err := h.pool.Exec(c.Request.Context(),
		`UPDATE geofences SET deleted_at=now() WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		c.Param("id"), tenantID,
	)
	if err != nil {
		h.logger.Error("geofence delete", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	if tag.RowsAffected() == 0 {
		respondError(c, http.StatusNotFound, "geofence not found")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}''', content)

content = re.sub(r'func \(h \*GeofenceHandler\) Events\(c \*gin\.Context\)\s*\{\s*respondOK\(c, \[\]any\{\}\)\s*\}', r'''func (h *GeofenceHandler) Events(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	geofenceID := c.Param("id")
	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT id, device_id, event_type, occurred_at
		 FROM geofence_events WHERE geofence_id=$1 AND tenant_id=$2 ORDER BY occurred_at DESC LIMIT 100`,
		geofenceID, tenantID,
	)
	if err != nil {
		h.logger.Error("geofence events query", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()
	var events []map[string]any
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			h.logger.Error("geofence events scan", zap.Error(err))
			continue
		}
		events = append(events, map[string]any{"id": vals[0], "device_id": vals[1], "event_type": vals[2], "occurred_at": vals[3]})
	}
	if err := rows.Err(); err != nil {
		h.logger.Error("geofence events rows error", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	if events == nil {
		events = []map[string]any{}
	}
	respondOK(c, events)
}''', content)

content = re.sub(r'func \(h \*VehicleHandler\) Update\(c \*gin\.Context\)\s*\{\s*c\.JSON\(http\.StatusNoContent, nil\)\s*\}', r'''func (h *VehicleHandler) Update(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	id := c.Param("id")
	var body struct {
		Registration string  `json:"registration" binding:"required"`
		Make         string  `json:"make"`
		Model        string  `json:"model"`
		Year         int     `json:"year"`
		DeviceID     *string `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	tag, err := h.pool.Exec(c.Request.Context(),
		`UPDATE vehicles SET registration=$1, make=$2, model=$3, year=$4, device_id=$5
		 WHERE id=$6 AND tenant_id=$7 AND deleted_at IS NULL`,
		body.Registration, body.Make, body.Model, body.Year, body.DeviceID, id, tenantID,
	)
	if err != nil {
		h.logger.Error("vehicle update", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	if tag.RowsAffected() == 0 {
		respondError(c, http.StatusNotFound, "vehicle not found")
		return
	}
	c.JSON(http.StatusNoContent, nil)
}''', content)

content = re.sub(r'func \(h \*DriverHandler\) Create\(c \*gin\.Context\)\s*\{\s*respondCreated\(c, gin\.H\{"id": "new"\}\)\s*\}', r'''func (h *DriverHandler) Create(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	var body struct {
		Name          string `json:"name" binding:"required"`
		LicenseNumber string `json:"license_number"`
		RfidUid       string `json:"rfid_uid"`
		Phone         string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	var id string
	err := h.pool.QueryRow(c.Request.Context(),
		`INSERT INTO drivers (tenant_id, name, license_number, rfid_uid, phone)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		tenantID, body.Name, body.LicenseNumber, body.RfidUid, body.Phone,
	).Scan(&id)
	if err != nil || id == "" {
		h.logger.Error("driver create", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	respondCreated(c, gin.H{"id": id})
}''', content)

content = re.sub(r'func \(h \*DriverHandler\) Get\(c \*gin\.Context\)\s*\{\s*respondOK\(c, gin\.H\{\}\)\s*\}', r'''func (h *DriverHandler) Get(c *gin.Context) {
	tenantID := pkgauth.TenantID(c)
	id := c.Param("id")
	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT id, name, license_number, rfid_uid, phone, created_at
		 FROM drivers WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		id, tenantID,
	)
	if err != nil {
		h.logger.Error("driver get query", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()
	if !rows.Next() {
		respondError(c, http.StatusNotFound, "driver not found")
		return
	}
	vals, err := rows.Values()
	if err != nil {
		h.logger.Error("driver get scan", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	respondOK(c, map[string]any{"id": vals[0], "name": vals[1], "license_number": vals[2], "rfid_uid": vals[3], "phone": vals[4], "created_at": vals[5]})
}''', content)

content = re.sub(r'respondError\(c, http\.StatusInternalServerError, "failed to compute score"\)', r'respondError(c, http.StatusInternalServerError, "database error")', content)

with open('api-service/internal/handler/handlers.go', 'w') as f:
    f.write(content)
