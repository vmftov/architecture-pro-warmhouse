package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SensorValue mirrors the latest telemetry reading for a sensor.
// JSON tags assume the Python/Flask service returns snake_case.
type SensorValue struct {
	SensorID  int       `json:"sensor_id"`
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TelemetryService is an HTTP client for the telemetry microservice.
type TelemetryService struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewTelemetryService(baseURL string) *TelemetryService {
	return &TelemetryService{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *TelemetryService) GetSensorValues(ctx context.Context) ([]SensorValue, error) {
	u := s.BaseURL + "/api/v1/sensor-values"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("telemetry service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telemetry service returned status %d", resp.StatusCode)
	}
	var values []SensorValue
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return nil, fmt.Errorf("failed to decode telemetry response: %w", err)
	}
	return values, nil
}

func (s *TelemetryService) GetSensorValue(ctx context.Context, id int) (SensorValue, error) {
	u := fmt.Sprintf("%s/api/v1/sensor-values/%d", s.BaseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return SensorValue{}, err
	}
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return SensorValue{}, fmt.Errorf("telemetry service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return SensorValue{}, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return SensorValue{}, fmt.Errorf("telemetry service returned status %d", resp.StatusCode)
	}
	var value SensorValue
	if err := json.NewDecoder(resp.Body).Decode(&value); err != nil {
		return SensorValue{}, fmt.Errorf("failed to decode telemetry response: %w", err)
	}
	return value, nil
}

func (s *TelemetryService) DeleteSensorValue(ctx context.Context, id int) error {
	u := fmt.Sprintf("%s/api/v1/sensor-values/%d", s.BaseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("telemetry service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telemetry service returned status %d", resp.StatusCode)
	}
	return nil
}