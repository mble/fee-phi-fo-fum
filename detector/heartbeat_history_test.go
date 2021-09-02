package detector_test

import (
	"reflect"
	"testing"

	"github.com/mble/fee-phi-fo-fum/detector"
)

func TestHeartBeatHistoryMinSamples(t *testing.T) {
	_, err := detector.NewHeartbeatHistory(0)
	if err == nil {
		t.Errorf("expected err, got nil")
	}

	expectedErr := "sample capacity must be > 0"
	if err.Error() != expectedErr {
		t.Errorf("expected: %v, got: %v", expectedErr, err.Error())
	}
}

func TestHeartbeatHistoryAddInterval(t *testing.T) {
	hh, err := detector.NewHeartbeatHistory(3)
	if err != nil {
		t.Fatal(err)
	}

	hh.AddInterval(1)
	hh.AddInterval(2)
	hh.AddInterval(3)

	expected := []float64{1, 2, 3}
	if !reflect.DeepEqual(hh.Intervals, expected) {
		t.Errorf("expected: %v, got: %v", expected, hh.Intervals)
	}
}

func TestHeartbeatHistoryMean(t *testing.T) {
	hh, err := detector.NewHeartbeatHistory(3)
	if err != nil {
		t.Fatal(err)
	}

	hh.AddInterval(1)
	hh.AddInterval(2)
	hh.AddInterval(3)

	expected := 2.0
	if hh.Mean() != expected {
		t.Errorf("expected: %v, got: %v", expected, hh.Mean())
	}
}

func TestHeartbeatHistoryStdDev(t *testing.T) {
	hh, err := detector.NewHeartbeatHistory(3)
	if err != nil {
		t.Fatal(err)
	}

	hh.AddInterval(1)
	hh.AddInterval(2)
	hh.AddInterval(3)

	expected := 0.8164965809277263
	if hh.StdDev() != expected {
		t.Errorf("expected: %v, got: %v", expected, hh.StdDev())
	}
}

func TestHeartbeatHistoryVariance(t *testing.T) {
	hh, err := detector.NewHeartbeatHistory(3)
	if err != nil {
		t.Fatal(err)
	}

	hh.AddInterval(1)
	hh.AddInterval(2)
	hh.AddInterval(3)

	expected := 0.666666666666667
	if hh.Variance() != expected {
		t.Errorf("expected: %v, got: %v", expected, hh.Variance())
	}
}

func TestHeartbeatHistoryDropOldestInterval(t *testing.T) {
	hh, err := detector.NewHeartbeatHistory(3)
	if err != nil {
		t.Fatal(err)
	}

	hh.AddInterval(1)
	hh.AddInterval(2)
	hh.AddInterval(3)
	hh.AddInterval(4)

	expected := []float64{2, 3, 4}
	if !reflect.DeepEqual(hh.Intervals, expected) {
		t.Errorf("expected: %v, got: %v", expected, hh.Intervals)
	}
}
