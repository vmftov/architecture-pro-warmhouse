package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"smarthome/models"
)

// ErrNotFound signals a 404 from a downstream service.
var ErrNotFound = errors.New("resource not found")

// DeviceSensor mirrors the sensor metadata owned by the device service.
// Note: the device service does NOT hold the runtime value — that lives in telemetry.
// JSON tags assume ASP.NET Core default camelCase serialization.
type DeviceSensor struct {
	SensorId    int       `json:"sensor_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Location    string    `json:"location"`
	Unit        string    `json:"unit"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_updated"`
}

// DeviceService is an HTTP client for the device microservice.
type DeviceService struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewDeviceService(baseURL string) *DeviceService {
	return &DeviceService{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// do performs a JSON request and decodes the response into out (if provided).
func (s *DeviceService) do(ctx context.Context, method, u string, body, out interface{}) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("device service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("device service returned status %d", resp.StatusCode)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("failed to decode device response: %w", err)
		}
	}
	return nil
}

func (s *DeviceService) GetSensors(ctx context.Context) ([]DeviceSensor, error) {
	var sensors []DeviceSensor
	err := s.do(ctx, http.MethodGet, s.BaseURL+"/api/v1/sensors", nil, &sensors)
	return sensors, err
}

func (s *DeviceService) GetSensorsByLocation(ctx context.Context, location string) ([]DeviceSensor, error) {
	u := fmt.Sprintf("%s/api/v1/sensors?location=%s", s.BaseURL, url.QueryEscape(location))
	var sensors []DeviceSensor
	err := s.do(ctx, http.MethodGet, u, nil, &sensors)
	return sensors, err
}

func (s *DeviceService) GetSensorByID(ctx context.Context, id int) (DeviceSensor, error) {
	u := fmt.Sprintf("%s/api/v1/sensors/%d", s.BaseURL, id)
	var sensor DeviceSensor
	err := s.do(ctx, http.MethodGet, u, nil, &sensor)
	return sensor, err
}

func (s *DeviceService) CreateSensor(ctx context.Context, in models.SensorCreate) (DeviceSensor, error) {
	payload := map[string]interface{}{
		"name":     in.Name,
		"type":     string(in.Type),
		"location": in.Location,
		"unit":     in.Unit,
	}
	var sensor DeviceSensor
	err := s.do(ctx, http.MethodPost, s.BaseURL+"/api/v1/sensors", payload, &sensor)
	return sensor, err
}

func (s *DeviceService) UpdateSensor(ctx context.Context, id int, in models.SensorUpdate) (DeviceSensor, error) {
	// Only forward metadata fields; the runtime value goes through UpdateSensorValue.
	payload := map[string]interface{}{}
	if in.Name != "" {
		payload["name"] = in.Name
	}
	if in.Type != "" {
		payload["type"] = string(in.Type)
	}
	if in.Location != "" {
		payload["location"] = in.Location
	}
	if in.Unit != "" {
		payload["unit"] = in.Unit
	}
	if in.Status != "" {
		payload["status"] = in.Status
	}

	u := fmt.Sprintf("%s/api/v1/sensors/%d", s.BaseURL, id)
	var sensor DeviceSensor
	err := s.do(ctx, http.MethodPut, u, payload, &sensor)
	return sensor, err
}

func (s *DeviceService) UpdateSensorValue(ctx context.Context, id int, value float64, status string) error {
	payload := map[string]interface{}{"value": value}
	if status != "" {
		payload["status"] = status
	}
	u := fmt.Sprintf("%s/api/v1/sensors/%d/value", s.BaseURL, id)
	return s.do(ctx, http.MethodPatch, u, payload, nil)
}

func (s *DeviceService) DeleteSensor(ctx context.Context, id int) error {
	u := fmt.Sprintf("%s/api/v1/sensors/%d", s.BaseURL, id)
	return s.do(ctx, http.MethodDelete, u, nil, nil)
}