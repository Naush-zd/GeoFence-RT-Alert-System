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
		{name: "missing name", input: withGeofence(valid, func(input *models.GeofenceRequest) { input.Name = " " })},
		{name: "too few points", input: withGeofence(valid, func(input *models.GeofenceRequest) { input.Coordinates = input.Coordinates[:3] })},
		{name: "invalid category", input: withGeofence(valid, func(input *models.GeofenceRequest) { input.Category = "unknown" })},
		{name: "invalid latitude", input: withGeofence(valid, func(input *models.GeofenceRequest) { input.Coordinates[1][0] = 91 })},
		{name: "open polygon", input: withGeofence(valid, func(input *models.GeofenceRequest) { input.Coordinates[len(input.Coordinates)-1] = []float64{0, 0} })},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateGeofence(test.input)
			if test.valid && err != nil {
				t.Fatalf("validateGeofence() error = %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("validateGeofence() expected an error")
			}
		})
	}
}

func TestValidateVehicle(t *testing.T) {
	valid := models.VehicleRequest{
		VehicleNumber: "KA-01-AB-1234",
		DriverName:    "John Doe",
		VehicleType:   "truck",
		Phone:         "+1234567890",
	}

	if err := validateVehicle(valid); err != nil {
		t.Fatalf("validateVehicle() error = %v", err)
	}

	invalid := valid
	invalid.Phone = ""
	if err := validateVehicle(invalid); err == nil {
		t.Fatal("validateVehicle() expected an error for missing phone")
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
