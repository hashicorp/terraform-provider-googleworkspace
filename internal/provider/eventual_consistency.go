package googleworkspace

import (
	"time"
)

// The number of consistent responses we want before we consider the resource consistent
const numConsistent = 4

type consistencyCheck struct {
	currConsistent int
	etagChanges    int
	lastEtag       string
	resourceType   string
	// timeout should be set to the timeout of the action
	timeout time.Duration
}

func (cc *consistencyCheck) reachedConsistency(numInserts int) bool {
	// Ideally we will get a min number of consistent tags and >= numInserts
	// However, there are cases where we'll have multiple Inserts, and the
	// initial changes were already consistent by the time the latter
	// inserts happened, thus once we start polling, those changes
	// wouldn't be counted. In which case, we have a maxConsistent
	// that we'll assume is good.

	// max consistent will be set to the timeout * 6 / 2 (or, every 10 seconds for half the timeout time)
	// so that it checks that at least the last half of responses were consistent
	maxConsistent := int(cc.timeout.Minutes()) * 6 / 2

	return (cc.currConsistent == numConsistent && cc.etagChanges >= numInserts) ||
		cc.currConsistent >= maxConsistent
}

func (cc *consistencyCheck) handleNewEtag(etag string) {
	cc.currConsistent = 0
	cc.lastEtag = etag
	cc.etagChanges += 1
}
