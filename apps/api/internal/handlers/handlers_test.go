package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"github.com/yourusername/geofencing-platform/apps/api/internal/services"
)

func TestValidateGeofence(t *testing.T) {
	valid := models.GeofenceRequest{
		Name:     "Delivery zone",
		Category: "delivery_zone",
		Coordinates: [][]float64{
			{37.7749, -122.4194},
			{37.7849, -122.4194},
			{37.7849, -122.4094},
			{37.7749, -122.4094},
			{37.7749, -122.4194},
		},
	}

	tests := []struct {
		name  string
		input models.GeofenceRequest
		valid bool
	}{
		{name: "valid polygon", input: valid, valid: true},
		{name: "too few points", input: withGeofence(valid, func(input *models.GeofenceRequest) { input.Coordinates = input.Coordinates[:3] })},
		{name: "invalid latitude", input: withGeofence(valid, func(input *models.GeofenceRequest) { input.Coordinates[1][0] = 91 })},
		{name: "open polygon", input: withGeofence(valid, func(input *models.GeofenceRequest) { input.Coordinates[len(input.Coordinates)-1] = []float64{0, 0} })},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := services.ValidateGeofence(test.input)
			if test.valid && err != nil {
				t.Fatalf("validateGeofence() error = %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("validateGeofence() expected an error")
			}
		})
	}
}

func TestDecodeJSONValidatesRequiredVehicleFields(t *testing.T) {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"vehicle_number":"KA-01-AB-1234","driver_name":"John Doe","vehicle_type":"truck"}`))

	var input models.VehicleRequest
	if err := decodeJSON(context, &input); err == nil {
		t.Fatal("decodeJSON() expected a validation error for missing phone")
	}
}

func TestDecodeJSONAcceptsValidGeofenceBindings(t *testing.T) {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"zone","coordinates":[[37,-122],[38,-122],[38,-121],[37,-122]],"category":"delivery_zone"}`))

	var input models.GeofenceRequest
	if err := decodeJSON(context, &input); err != nil {
		t.Fatalf("decodeJSON() rejected valid geofence: %v", err)
	}
}

func TestDecodeJSONValidatesGeofenceBindings(t *testing.T) {
	validCoordinates := `[[37,-122],[38,-122],[38,-121],[37,-122]]`
	tests := []struct {
		name string
		body string
	}{
		{name: "missing name", body: `{"coordinates":` + validCoordinates + `,"category":"delivery_zone"}`},
		{name: "invalid category", body: `{"name":"zone","coordinates":` + validCoordinates + `,"category":"unknown"}`},
		{name: "too few points", body: `{"name":"zone","coordinates":[[37,-122],[38,-122],[37,-122]],"category":"delivery_zone"}`},
		{name: "malformed coordinate pair", body: `{"name":"zone","coordinates":[[37,-122],[38,-122],[38,-121],[37]],"category":"delivery_zone"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))

			var input models.GeofenceRequest
			if err := decodeJSON(context, &input); err == nil {
				t.Fatal("decodeJSON() expected a binding validation error")
			}
		})
	}
}

func TestDecodeJSONValidatesLocationAndAlertBindings(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		destination any
		wantError   bool
	}{
		{
			name:        "zero coordinates are valid",
			body:        `{"vehicle_id":"00000000-0000-0000-0000-000000000001","latitude":0,"longitude":0,"timestamp":"2026-11-11T00:00:00Z"}`,
			destination: &models.LocationRequest{},
		},
		{
			name:        "latitude is out of range",
			body:        `{"vehicle_id":"00000000-0000-0000-0000-000000000001","latitude":91,"longitude":0,"timestamp":"2026-11-11T00:00:00Z"}`,
			destination: &models.LocationRequest{},
			wantError:   true,
		},
		{
			name:        "valid alert event type",
			body:        `{"geofence_id":"00000000-0000-0000-0000-000000000001","vehicle_id":"","event_type":"both"}`,
			destination: &models.AlertConfigRequest{},
		},
		{
			name:        "invalid alert event type",
			body:        `{"geofence_id":"00000000-0000-0000-0000-000000000001","vehicle_id":"","event_type":"sideways"}`,
			destination: &models.AlertConfigRequest{},
			wantError:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))

			err := decodeJSON(context, test.destination)
			if (err != nil) != test.wantError {
				t.Fatalf("decodeJSON() error = %v, wantError %t", err, test.wantError)
			}
		})
	}
}

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"zone","unexpected":true}`))

	var input struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(context, &input); err == nil {
		t.Fatal("decodeJSON() expected an unknown-field error")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("decodeJSON() changed response status to %d", recorder.Code)
	}
}

func TestWriteJSONIncludesExecutionTime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	writeJSON(context, http.StatusOK, map[string]any{"status": "ok"}, time.Now())

	if recorder.Code != http.StatusOK {
		t.Fatalf("writeJSON() status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var response map[string]any
	if err := json.NewDecoder(bytes.NewReader(recorder.Body.Bytes())).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := response["time_ns"]; !ok {
		t.Fatal("writeJSON() response does not include time_ns")
	}
}

func withGeofence(input models.GeofenceRequest, mutate func(*models.GeofenceRequest)) models.GeofenceRequest {
	copy := input
	copy.Coordinates = make([][]float64, len(input.Coordinates))
	for index, point := range input.Coordinates {
		copy.Coordinates[index] = append([]float64(nil), point...)
	}
	mutate(&copy)
	return copy
}
