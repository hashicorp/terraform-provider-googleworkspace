// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"testing"
	"time"
)

func TestConsistencyCheckReachedConsistency(t *testing.T) {
	// We'll test that there were 3 inserts
	numInserts := 3

	cc := consistencyCheck{
		timeout:        time.Duration(time.Minute * 5),
		currConsistent: 1,
		etagChanges:    1,
		lastEtag:       "12345",
	}

	// So far we've seen one etag and it's been consistent once
	if cc.reachedConsistency(numInserts) {
		t.Errorf("Failed: reached consistency (numInserts: %d, currConsistent: %d, etagChanges: %d, timeout: %d)", numInserts, cc.currConsistent, cc.etagChanges, int(cc.timeout.Minutes()))
	}

	// We only have 2 previous Etags, but we've been consistent for 3 minutes
	// We'll assume it's consistent and that one of the inserts already contained
	// and updated etag that we're missing
	cc.etagChanges = 2
	cc.currConsistent = 18

	if !cc.reachedConsistency(numInserts) {
		t.Errorf("Failed: did not reach consistency (numInserts: %d, currConsistent: %d, etagChanges: %d, timeout: %d)", numInserts, cc.currConsistent, cc.etagChanges, int(cc.timeout.Minutes()))
	}

	// We've seen all the inserts come through, but we haven't had 2 consistent tags yet
	cc.etagChanges = 3
	cc.currConsistent = 1

	if cc.reachedConsistency(numInserts) {
		t.Errorf("Failed: reached consistency (numInserts: %d, currConsistent: %d, etagChanges: %d, timeout: %d)", numInserts, cc.currConsistent, cc.etagChanges, int(cc.timeout.Minutes()))
	}

	// We've seen all the inserts come through, and it's been consistent 2 times
	cc.currConsistent = 2

	if !cc.reachedConsistency(numInserts) {
		t.Errorf("Failed: did not reach consistency (numInserts: %d, currConsistent: %d, etagChanges: %d, timeout: %d)", numInserts, cc.currConsistent, cc.etagChanges, int(cc.timeout.Minutes()))
	}
}

func TestConsistencyHandleNewEtag(t *testing.T) {
	cc := consistencyCheck{
		resourceType: "test",
	}

	cc.handleNewEtag("12345")
	if cc.currConsistent != 0 {
		t.Errorf("Failed ['12345']: new etag shows currConsistent > 0 (%d)", cc.currConsistent)
	}

	cc.handleNewEtag("abcde")
	if cc.lastEtag != "abcde" {
		t.Errorf("Failed ['abcde']: ends with incorrect lastEtag (%s)", cc.lastEtag)
	}

	cc.handleNewEtag("54321")
	if cc.etagChanges != 3 {
		t.Errorf("Failed ['abcde']: shows more/less etag changes (expected: %d, got: %d)", 3, cc.etagChanges)
	}
}
