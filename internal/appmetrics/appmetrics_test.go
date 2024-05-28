package appmetrics

import (
	"testing"
)

// TestSetUpMetric tests the SetUpMetric function
func TestSetUpMetric(t *testing.T) {
	value := 10.0
	UpMetric.Set(value)
	// Assert that the metric value is set correctly
	if UpMetric.Get() != value {
		t.Errorf("Expected metric value to be %f, but got %f", value, UpMetric.Get())
	}
}

// TestIncApiCallCounter tests the IncApiCallCounter function
func TestIncApiCallCounter(t *testing.T) {
	// Assert that the api call counter is incremented
	IncApiCallCounter()

	if ApiCallCounter.Get() != 1 {
		t.Errorf("Expected api call counter to be 1, but got %d", ApiCallCounter.Get())
	}
}

// TestIncApiErrorsCounter tests the IncApiErrorsCounter function
func TestIncApiErrorsCounter(t *testing.T) {
	// Assert that the api errors counter is incremented
	IncApiErrorsCounter()
	if ApiErrorsCounter.Get() != 1 {
		t.Errorf("Expected api errors counter to be 1, but got %d", ApiErrorsCounter.Get())
	}
}
