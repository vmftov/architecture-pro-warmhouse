package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"smarthome/models"
	"smarthome/services"

	"github.com/gin-gonic/gin"
)

// SensorHandler composes sensor metadata (device service) with values (telemetry service).
type SensorHandler struct {
	Device    *services.DeviceService
	Telemetry *services.TelemetryService
}

func NewSensorHandler(device *services.DeviceService, telemetry *services.TelemetryService) *SensorHandler {
	return &SensorHandler{Device: device, Telemetry: telemetry}
}

func (h *SensorHandler) RegisterRoutes(router *gin.RouterGroup) {
	sensors := router.Group("/sensors")
	{
		sensors.GET("", h.GetSensors)
		sensors.GET("/:id", h.GetSensorByID)
		sensors.POST("", h.CreateSensor)
		sensors.PUT("/:id", h.UpdateSensor)
		sensors.DELETE("/:id", h.DeleteSensor)
		sensors.PATCH("/:id/value", h.UpdateSensorValue)
		sensors.GET("/temperature/:location", h.GetTemperatureByLocation)
	}
}

// toSensor maps device metadata into the monolith's public model.
func toSensor(d services.DeviceSensor) models.Sensor {
	return models.Sensor{
		ID:          d.SensorId,
		Name:        d.Name,
		Type:        models.SensorType(d.Type),
		Location:    d.Location,
		Unit:        d.Unit,
		Status:      d.Status,
		CreatedAt:   d.CreatedAt,
		LastUpdated: d.LastUpdated,
	}
}

// applyValue enriches a sensor with its latest telemetry reading.
func applyValue(s *models.Sensor, v services.SensorValue) {
	s.Value = v.Value
	if !v.Timestamp.IsZero() {
		s.LastUpdated = v.Timestamp
	}
}

// GetSensors handles GET /api/v1/sensors (optionally ?location=).
func (h *SensorHandler) GetSensors(c *gin.Context) {
	ctx := c.Request.Context()

	location := c.Query("location")
	var (
		devices []services.DeviceSensor
		err     error
	)
	if location != "" {
		devices, err = h.Device.GetSensorsByLocation(ctx, location)
	} else {
		devices, err = h.Device.GetSensors(ctx)
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "device service unavailable: " + err.Error()})
		return
	}

	// Fetch all values once and join in memory (avoids N calls to telemetry).
	valueByID := map[int]services.SensorValue{}
	if values, err := h.Telemetry.GetSensorValues(ctx); err == nil {
		for _, v := range values {
			valueByID[v.SensorID] = v
		}
	} else {
		log.Printf("telemetry unavailable, returning sensors without values: %v", err)
	}

	result := make([]models.Sensor, 0, len(devices))
	for _, d := range devices {
		s := toSensor(d)
		if v, ok := valueByID[d.SensorId]; ok {
			applyValue(&s, v)
		}
		result = append(result, s)
	}

	c.JSON(http.StatusOK, result)
}

// GetSensorByID handles GET /api/v1/sensors/:id.
func (h *SensorHandler) GetSensorByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}
	ctx := c.Request.Context()

	device, err := h.Device.GetSensorByID(ctx, id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	sensor := toSensor(device)
	if v, err := h.Telemetry.GetSensorValue(ctx, id); err == nil {
		applyValue(&sensor, v)
	} else if !errors.Is(err, services.ErrNotFound) {
		log.Printf("failed to fetch value for sensor %d: %v", id, err)
	}

	c.JSON(http.StatusOK, sensor)
}

// CreateSensor handles POST /api/v1/sensors.
func (h *SensorHandler) CreateSensor(c *gin.Context) {
	var in models.SensorCreate
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	device, err := h.Device.CreateSensor(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toSensor(device))
}

// UpdateSensor handles PUT /api/v1/sensors/:id.
func (h *SensorHandler) UpdateSensor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}
	var in models.SensorUpdate
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	device, err := h.Device.UpdateSensor(c.Request.Context(), id, in)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toSensor(device))
}

// DeleteSensor handles DELETE /api/v1/sensors/:id.
func (h *SensorHandler) DeleteSensor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}
	ctx := c.Request.Context()

	if err := h.Device.DeleteSensor(ctx, id); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	// Best-effort cleanup of the telemetry reading; a missing value is fine.
	if err := h.Telemetry.DeleteSensorValue(ctx, id); err != nil && !errors.Is(err, services.ErrNotFound) {
		log.Printf("failed to delete telemetry for sensor %d: %v", id, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sensor deleted successfully"})
}

// UpdateSensorValue handles PATCH /api/v1/sensors/:id/value.
// The value is forwarded to the device service, which is responsible for
// publishing it to telemetry (e.g. via Kafka).
func (h *SensorHandler) UpdateSensorValue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}
	var req struct {
		Value  float64 `json:"value"`
		Status string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Device.UpdateSensorValue(c.Request.Context(), id, req.Value, req.Status); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Sensor value updated successfully"})
}

// GetTemperatureByLocation handles GET /api/v1/sensors/temperature/:location.
// It finds temperature sensors in a location (device) and enriches them with
// their latest values (telemetry) — the cross-service join lives here in the facade.
func (h *SensorHandler) GetTemperatureByLocation(c *gin.Context) {
	location := c.Param("location")
	if location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Location is required"})
		return
	}
	ctx := c.Request.Context()

	devices, err := h.Device.GetSensorsByLocation(ctx, location)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	valueByID := map[int]services.SensorValue{}
	if values, err := h.Telemetry.GetSensorValues(ctx); err == nil {
		for _, v := range values {
			valueByID[v.SensorID] = v
		}
	}

	result := make([]models.Sensor, 0)
	for _, d := range devices {
		if models.SensorType(d.Type) != models.Temperature {
			continue
		}
		s := toSensor(d)
		if v, ok := valueByID[d.SensorId]; ok {
			applyValue(&s, v)
		}
		result = append(result, s)
	}

	c.JSON(http.StatusOK, gin.H{
		"location": location,
		"sensors":  result,
	})
}