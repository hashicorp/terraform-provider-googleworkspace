package googleworkspace

import (
	"fmt"
	"time"
)

// The number of consistent responses we want before we consider the resource consistent
const numConsistent = 3

type consistencyCheck struct {
	currConsistent int
	previousEtags  []string
	resourceType   string
	// timeout should be set to the timeout of the action
	timeout time.Duration
}

func (cc *consistencyCheck) inPreviousEtags(etag string) bool {
	return stringInSlice(cc.previousEtags, etag)
}

func (cc *consistencyCheck) isConsistentWithLastEtag(etag string) bool {
	return cc.previousEtags[len(cc.previousEtags)-1] == etag
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

	return (cc.currConsistent == numConsistent && len(cc.previousEtags) >= numInserts) ||
		cc.currConsistent == maxConsistent
}

func (cc *consistencyCheck) handleFirstRun(etag string) error {
	if len(cc.previousEtags) == 0 {
		cc.previousEtags = append(cc.previousEtags, etag)
		cc.currConsistent = 0
		return fmt.Errorf("timed out while waiting for %s to be inserted (%d/%d consistent etags)", cc.resourceType, cc.currConsistent, numConsistent)
	}

	return nil
}

func (cc *consistencyCheck) checkChangedEtags(numInserts int, etag string) error {
	if len(cc.previousEtags) >= numInserts {
		// We've seen the number of changes we're expecting,
		// check if this etag is consistent with the last
		if cc.isConsistentWithLastEtag(etag) {
			// We got another consistent tag
			cc.currConsistent += 1
		} else if !cc.inPreviousEtags(etag) {
			// on create, this is possible since the etag changes with every request
			cc.currConsistent = 0
			cc.previousEtags = append(cc.previousEtags, etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be inserted (%d/%d consistent etags)", cc.resourceType, cc.currConsistent, numConsistent)
	}

	return nil
}

func (cc *consistencyCheck) handleConsistency(etag string) {
	if !cc.isConsistentWithLastEtag(etag) && !cc.inPreviousEtags(etag) {
		// A new Etag! Working our way toward our new etags:numInserts ratio
		cc.currConsistent = 0
		cc.previousEtags = append(cc.previousEtags, etag)
	}

	if cc.isConsistentWithLastEtag(etag) {
		// We have a consistent tag, but haven't hit our ration of etags:numInserts, we're counting these just
		// in case we need to check maxConsistent (see reachedConsistency code)
		cc.currConsistent += 1
	}
}
