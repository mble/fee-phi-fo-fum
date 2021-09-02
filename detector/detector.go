package detector

import (
	"errors"
	"math"
	"sync/atomic"
	"time"
)

// State represents the state of the failure detector
type State struct {
	History         *HeartbeatHistory
	LatestTimestamp float64
}

// FailureDetector represents a phi accural failure detector
// https://www.computer.org/csdl/proceedings-article/icdcsw/2006/25410046/12OmNro0I6u
type FailureDetector struct {
	Threshold                            float64
	MaxSamples                           int
	MinStdDevMilliseconds                int
	AcceptableHeartbeatPauseMilliseconds int
	FirstHeartbeatEstimateMilliseconds   int
	state                                atomic.Value
}

// NewFailureDetector returns a FailureDetector initialised with the config passed in
func NewFailureDetector(threshold float64, maxSamples, minStdDev, acceptableHeartbeatPause, firstHeartbeatEstimate int) (*FailureDetector, error) {
	if threshold <= 0.0 {
		return nil, errors.New("threshold must be > 0")
	}

	if maxSamples <= 0 {
		return nil, errors.New("maxSamples must be > 0")
	}

	if minStdDev <= 0 {
		return nil, errors.New("minStdDev must be > 0")
	}

	if acceptableHeartbeatPause < 0 {
		return nil, errors.New("acceptableHeartbeatPause must be > 0")
	}

	if firstHeartbeatEstimate <= 0 {
		return nil, errors.New("firstHeartbeatEstimate must be > 0")
	}

	fhb, err := firstHeartbeat(firstHeartbeatEstimate, maxSamples)
	if err != nil {
		return nil, err
	}

	var state atomic.Value
	state.Store(State{
		History:         fhb,
		LatestTimestamp: 0,
	})
	return &FailureDetector{
		Threshold:                            threshold,
		MaxSamples:                           maxSamples,
		MinStdDevMilliseconds:                minStdDev,
		AcceptableHeartbeatPauseMilliseconds: acceptableHeartbeatPause,
		FirstHeartbeatEstimateMilliseconds:   firstHeartbeatEstimate,
		state:                                state,
	}, nil
}

// Phi returns the phi value of the detector at the current timestamp
func (fd *FailureDetector) Phi() float64 {
	return fd.phi(getTimeMs())
}

func (fd *FailureDetector) ensureValidStdDev(stdDev float64) float64 {
	return math.Max(stdDev, float64(fd.MinStdDevMilliseconds))
}

func (fd *FailureDetector) phi(timestamp float64) float64 {
	lastState := fd.state.Load().(State)
	lastTimestamp := lastState.LatestTimestamp
	timeDiff := timestamp - lastTimestamp
	lastHistory := lastState.History
	mean := lastHistory.Mean()
	stdDev := fd.ensureValidStdDev(lastHistory.StdDev())

	return calcPhi(timeDiff, mean+float64(fd.AcceptableHeartbeatPauseMilliseconds), stdDev)
}

// IsAvailable returns true if the detector has determined availability, false otherwise
func (fd *FailureDetector) IsAvailable() bool {
	return fd.isAvailable(getTimeMs())
}

func (fd *FailureDetector) isAvailable(timestamp float64) bool {
	phi_val := fd.phi(timestamp)
	return phi_val < fd.Threshold
}

// Heartbeat either initialised the first heartbeat, or adds a new interval to the
// history if available
func (fd *FailureDetector) Heartbeat() error {
	var history *HeartbeatHistory
	timestamp := getTimeMs()
	oldState := fd.state.Load().(State)

	if oldState.LatestTimestamp == 0 {
		history, _ = firstHeartbeat(fd.FirstHeartbeatEstimateMilliseconds, fd.MaxSamples)
	} else {
		latestTimestamp := oldState.LatestTimestamp
		interval := timestamp - latestTimestamp

		history = oldState.History

		if fd.isAvailable(timestamp) {
			history.AddInterval(interval)
		}
	}

	newState := State{History: history, LatestTimestamp: timestamp}
	swapped := fd.state.CompareAndSwap(oldState, newState)
	if !swapped {
		fd.Heartbeat()
	}

	return nil
}

// firstHeartbeat bootstraps the history with 2 samples of a relatively high stddev from the estimate
func firstHeartbeat(firstHeartbeatEstimate, maxSamples int) (*HeartbeatHistory, error) {
	stdDev := firstHeartbeatEstimate / 4
	hh, err := NewHeartbeatHistory(maxSamples)

	if err != nil {
		return nil, err
	}

	hh.AddInterval(float64(firstHeartbeatEstimate) - float64(stdDev))
	hh.AddInterval(float64(firstHeartbeatEstimate) + float64(stdDev))
	return hh, nil
}

func getTimeMs() float64 {
	return float64(time.Now().UnixNano() / 1e6)
}

// calcPhi uses a logistic approximation of cumulative normal distribution
// following the method in https://github.com/akka/akka/issues/1821
// https://digitalcommons.odu.edu/cgi/viewcontent.cgi?referer=&httpsredir=1&article=1007&context=emse_fac_pubs
func calcPhi(timeDiff, mean, stdDev float64) float64 {
	y := (timeDiff - mean) / stdDev
	y1 := 0.07056
	y2 := 1.5976
	e := math.Exp(-y * (y2 + y1*y*y))

	// e can be Inf here due to overflow

	if timeDiff > mean {
		return -math.Log10(e / (1.0 + e))
	}

	return -math.Log(1.0 - 1.0/(1.0+e))
}
