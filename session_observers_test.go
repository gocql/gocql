// +build all cassandra

package gocql

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

type CounterConnectObserver struct {
	count int32
}

func (o *CounterConnectObserver) ObserveConnect(e ObservedConnect) {
	atomic.AddInt32(&o.count, 1)
}

func (o *CounterConnectObserver) ObserveDisconnect(e ObservedDisconnect) {
	atomic.AddInt32(&o.count, -1)
}

// Verify if all connection creation and closing
// are observable
func TestSessionObserveConnectAndDisconnect(t *testing.T) {
	obs := CounterConnectObserver{}
	cluster := createCluster()
	cluster.ConnectObserver = &obs

	// We expect to see some connections being opened...
	if !assert.Equal(t, int32(0), obs.count, "Observed connections must start at zero.") {
		return
	}
	session, _ := cluster.CreateSession()
	if !assert.NotEqual(t, int32(0), atomic.LoadInt32(&obs.count), "No connections opening were observed.") {
		return
	}

	// ...and being closed.
	session.Close()
	assert.Equal(t, int32(0), atomic.LoadInt32(&obs.count), fmt.Sprintf("[%d] connections still open.", obs.count))
}
