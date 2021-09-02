package main

import (
	"log"
	"time"

	"github.com/mble/fee-phi-fo-fum/detector"
)

func main() {
	// Initialise a detector with our chosen configuration
	// This is the "suspicion threshold". 8 is pretty standard
	// (https://github.com/apache/cassandra/blob/trunk/src/java/org/apache/cassandra/gms/FailureDetector.java#L78-L81)
	threshold := 8.0
	// Number of samples to store in the history for calculating mean and stddev of intervals
	maxSamples := 10
	// Minimum stddev of the normal distribution of heartbeat arrival. Too low might be too sensitive for normal deviations
	minStdDev := 10
	// The acceptable pause in milliseconds between successful heartbeats.
	// Ideally this should be > 0 in real-world conditions to handle stuff like temporary network drops, latency spikes etc
	acceptableHeartbeatPause := 0
	// Used to bootstrap the detector with samples, used with a high stddev to deal with the unknown.
	firstHeartbeatEstimate := 100

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)

	detector, err := detector.NewFailureDetector(threshold, maxSamples, minStdDev, acceptableHeartbeatPause, firstHeartbeatEstimate)
	if err != nil {
		log.Fatal(err)
	}

	// The detector can be used in a loop, periodically checking the status of some node/resource R.
	// When a successful response is recorded, write a heartbeat.
	// When an unsuccessful response is recorded, skip the heartbeat.
	// Check the availability each period.

	// errs are ignored for this example.
	detector.Heartbeat()

	time.Sleep(100 * time.Millisecond)
	log.Printf("available?: %t\n", detector.IsAvailable())
	detector.Heartbeat()

	time.Sleep(100 * time.Millisecond)
	log.Printf("available?: %t\n", detector.IsAvailable())
	// oops, not reachable!

	time.Sleep(200 * time.Millisecond)
	log.Printf("available?: %t\n", detector.IsAvailable())

	// and we're back!
	detector.Heartbeat()
	time.Sleep(100 * time.Millisecond)
	log.Printf("available?: %t\n", detector.IsAvailable())
}
