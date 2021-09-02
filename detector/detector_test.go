package detector_test

import (
	"testing"
	"time"

	"github.com/mble/fee-phi-fo-fum/detector"
)

func TestPhi(t *testing.T) {
	fd, err := detector.NewFailureDetector(8, 200, 500, 0, 500)
	if err != nil {
		t.Fatal(err)
	}

	if fd.Phi() == 0 {
		t.Error("expected non zero value for phi")
	}
}

func TestHeartbeat(t *testing.T) {
	fd, err := detector.NewFailureDetector(8, 200, 500, 0, 500)
	if err != nil {
		t.Fatal(err)
	}

	err = fd.Heartbeat()
	if err != nil {
		t.Fatal(err)
	}

	err = fd.Heartbeat()
	if err != nil {
		t.Fatal(err)
	}

	if fd.Phi() == 0 {
		t.Error("expected non zero value for phi")
	}
}

func TestSuccessfulHeartbeats(t *testing.T) {
	fd, err := detector.NewFailureDetector(8, 3, 10, 0, 10)
	if err != nil {
		t.Error(err)
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := fd.Heartbeat()
				if err != nil {
					t.Log(err)
					return
				}
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)
	ticker.Stop()
	done <- true

	if !fd.IsAvailable() {
		t.Errorf("expected detector to assert available")
	}
}

func TestSuccessfulHeartbeatsWithPause(t *testing.T) {
	fd, err := detector.NewFailureDetector(8, 100, 1, 10, 10)
	if err != nil {
		t.Error(err)
	}

	err = fd.Heartbeat()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond)

	err = fd.Heartbeat()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond)

	err = fd.Heartbeat()
	if err != nil {
		t.Fatal(err)
	}

	if !fd.IsAvailable() {
		t.Errorf("expected detector to assert available")
	}

	time.Sleep(20 * time.Millisecond)

	if !fd.IsAvailable() {
		t.Errorf("expected detector to assert available")
	}
}

func TestMissedHeartbeats(t *testing.T) {
	fd, err := detector.NewFailureDetector(8, 100, 1, 0, 10)
	if err != nil {
		t.Error(err)
	}

	err = fd.Heartbeat()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond)

	err = fd.Heartbeat()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond)

	err = fd.Heartbeat()
	if err != nil {
		t.Fatal(err)
	}

	if !fd.IsAvailable() {
		t.Errorf("expected detector to assert available")
	}

	time.Sleep(20 * time.Millisecond)

	if fd.IsAvailable() {
		t.Errorf("expected detector to assert unavailable")
	}
}
