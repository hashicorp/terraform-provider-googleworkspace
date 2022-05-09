package googleworkspace

import (
	"fmt"
	"google.golang.org/api/googleapi"
	"time"
)

// The number of consistent responses we want before we consider the resource consistent
const numConsistent = 4
const Limit404s = 4

type consistencyCheck struct {
	currConsistent int
	etagChanges    int
	lastEtag       string
	resourceType   string
	// timeout should be set to the timeout of the action
	timeout time.Duration
	// num404s will count how many 404s we encounter before we
	// return it as not found rather than consider it eventual consistency
	num404s int
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

func (cc *consistencyCheck) is404(err error) error {
	gerr, ok := err.(*googleapi.Error)
	if !ok {
		return err
	}

	if gerr.Code == 404 && !(Limit404s == cc.num404s) {
		cc.num404s += 1
		return fmt.Errorf("timed out while waiting for %s to be inserted", cc.resourceType)
	}

	return err
}
