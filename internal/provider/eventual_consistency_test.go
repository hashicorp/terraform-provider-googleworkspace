package googleworkspace

import (
	"reflect"
	"testing"
	"time"
)

func TestConsistencyCheckInPreviousEtags(t *testing.T) {
	cc := consistencyCheck{
		previousEtags: []string{
			"12345",
			"abcde",
			"10987",
			"zyxwv",
		},
	}

	if cc.inPreviousEtags("a1b2c3") {
		t.Errorf("Failed ['a1b2c3']: result ('a1b2c3') was found in previous etags (%s)", cc.previousEtags)
	}

	if !cc.inPreviousEtags("12345") {
		t.Errorf("Failed ['12345']: result ('12345') was not found in previous etags (%s)", cc.previousEtags)
	}
}

func TestConsistencyCheckIsConsistentWithLastEtag(t *testing.T) {
	cc := consistencyCheck{
		previousEtags: []string{
			"12345",
			"abcde",
			"10987",
			"zyxwv",
		},
	}

	if cc.isConsistentWithLastEtag("a1b2c3") {
		t.Errorf("Failed ['a1b2c3']: result ('a1b2c3') was consistent with last etag (%s)", cc.previousEtags[3])
	}

	if cc.isConsistentWithLastEtag("12345") {
		t.Errorf("Failed ['12345']: result ('12345') was consistent with last etag (%s)", cc.previousEtags[3])
	}

	if !cc.isConsistentWithLastEtag("zyxwv") {
		t.Errorf("Failed ['zyxwv']: result ('zyxwv') was not consistent with last etag (%s)", cc.previousEtags[3])
	}
}

func TestConsistencyCheckReachedConsistency(t *testing.T) {
	// We'll test that there were 3 inserts
	numInserts := 3

	cc := consistencyCheck{
		timeout:        time.Duration(time.Minute * 5),
		currConsistent: 1,
		previousEtags: []string{
			"12345",
		},
	}

	// So far we've seen one etag and it's been consistent once
	if cc.reachedConsistency(numInserts) {
		t.Errorf("Failed: reached consistency (numInserts: %d, currConsistent: %d, len(etags): %d, timeout: %d)", numInserts, cc.currConsistent, len(cc.previousEtags), int(cc.timeout.Minutes()))
	}

	// We only have 2 previous Etags, but we've been consistent for 3 minutes
	// We'll assume it's consistent and that one of the inserts already contained
	// and updated etag that we're missing
	cc.previousEtags = append(cc.previousEtags, "abcde")
	cc.currConsistent = 18

	if !cc.reachedConsistency(numInserts) {
		t.Errorf("Failed: did not reach consistency (numInserts: %d, currConsistent: %d, len(etags): %d, timeout: %d)", numInserts, cc.currConsistent, len(cc.previousEtags), int(cc.timeout.Minutes()))
	}

	// We've seen all the inserts come through, but we haven't had 3 consistent tags yet
	cc.previousEtags = append(cc.previousEtags, "10987")
	cc.currConsistent = 1

	if cc.reachedConsistency(numInserts) {
		t.Errorf("Failed: reached consistency (numInserts: %d, currConsistent: %d, len(etags): %d, timeout: %d)", numInserts, cc.currConsistent, len(cc.previousEtags), int(cc.timeout.Minutes()))
	}

	// We've seen all the inserts come through, but and it's been consistent 3 times
	cc.currConsistent = 3

	if !cc.reachedConsistency(numInserts) {
		t.Errorf("Failed: did not reach consistency (numInserts: %d, currConsistent: %d, len(etags): %d, timeout: %d)", numInserts, cc.currConsistent, len(cc.previousEtags), int(cc.timeout.Minutes()))
	}
}

func TestConsistencyCheckFirstRun(t *testing.T) {
	cc := consistencyCheck{
		resourceType: "test",
	}

	err := cc.handleFirstRun("12345")
	if err == nil {
		t.Errorf("Failed ['12345']: did not return an error on first run (%s)", cc.previousEtags)
	}

	err = cc.handleFirstRun("abcde")
	if err != nil {
		t.Errorf("Failed ['abcde']: returned an error on first run (%s)", cc.previousEtags)
	}
}

func TestConsistencyCheckCheckChangedEtags(t *testing.T) {
	// We'll test that there were 3 inserts
	numInserts := 2

	cc := consistencyCheck{
		resourceType:   "test",
		currConsistent: 0,
		previousEtags: []string{
			"12345",
		},
	}

	// We haven't hit every insert yet, we should pass through this
	err := cc.checkChangedEtags(numInserts, "abcde")
	if err != nil {
		t.Errorf("Failed ['abcde']: previousEtags (%+v) does not include all inserts (%d) yet, but we received an error", cc.previousEtags, numInserts)
	}

	expectedEtags := []string{"12345"}
	expectedConsistent := 0

	if !reflect.DeepEqual(expectedEtags, cc.previousEtags) || expectedConsistent != cc.currConsistent {
		t.Errorf("Failed ['abcde']: previousEtags (%+v) or currConsistent (%d) did not match expected (%+v, %d)", cc.previousEtags, cc.currConsistent, expectedEtags, expectedConsistent)
	}

	cc.previousEtags = append(cc.previousEtags, "abcde")

	// We've seen all inserts, but receive a brand new etag, we should append it to previousEtags and restart currConsistent
	err = cc.checkChangedEtags(numInserts, "zyxwv")
	if err == nil {
		t.Errorf("Failed ['zyxwv']: previousEtags (%+v) received new etag (%s), but we did not receive an error", cc.previousEtags, "zyxwv")
	}

	expectedEtags = []string{"12345", "abcde", "zyxwv"}
	expectedConsistent = 0

	if !reflect.DeepEqual(expectedEtags, cc.previousEtags) || expectedConsistent != cc.currConsistent {
		t.Errorf("Failed ['zyxwv']: previousEtags (%+v) or currConsistent (%d) did not match expected (%+v, %d)", cc.previousEtags, cc.currConsistent, expectedEtags, expectedConsistent)
	}

	// We've seen all inserts, and receive an etag consistent with the last one, update currConsistent
	err = cc.checkChangedEtags(numInserts, "zyxwv")
	if err == nil {
		t.Errorf("Failed ['zyxwv']: previousEtags (%+v) received a consistent etag (%s), but we did not receive an error", cc.previousEtags, "zyxwv")
	}

	expectedConsistent = 1

	if !reflect.DeepEqual(expectedEtags, cc.previousEtags) || expectedConsistent != cc.currConsistent {
		t.Errorf("Failed ['zyxwv']: previousEtags (%+v) or currConsistent (%d) did not match expected (%+v, %d)", cc.previousEtags, cc.currConsistent, expectedEtags, expectedConsistent)
	}

	// We've seen all inserts, and receive an etag that's not new, but not the last, we just need the error
	err = cc.checkChangedEtags(numInserts, "12345")
	if err == nil {
		t.Errorf("Failed ['12345']: previousEtags (%+v) received an etag (%s) its seen previously, but not last, but we did not receive an error", cc.previousEtags, "zyxwv")
	}

	if !reflect.DeepEqual(expectedEtags, cc.previousEtags) || expectedConsistent != cc.currConsistent {
		t.Errorf("Failed ['12345']: previousEtags (%+v) or currConsistent (%d) did not match expected (%+v, %d)", cc.previousEtags, cc.currConsistent, expectedEtags, expectedConsistent)
	}
}

func TestConsistencyCheckConsistency(t *testing.T) {
	cc := consistencyCheck{
		resourceType:   "test",
		currConsistent: 0,
		previousEtags: []string{
			"12345",
		},
	}

	// A brand new etag should be appended to previousEtags
	cc.handleConsistency("abcde")

	expectedEtags := []string{"12345", "abcde"}
	expectedConsistent := 0

	if !reflect.DeepEqual(expectedEtags, cc.previousEtags) || expectedConsistent != cc.currConsistent {
		t.Errorf("Failed ['abcde']: previousEtags (%+v) or currConsistent (%d) did not match expected (%+v, %d)", cc.previousEtags, cc.currConsistent, expectedEtags, expectedConsistent)
	}

	// An old etag, but not consistent with the last, we should see no updates
	cc.handleConsistency("12345")

	if !reflect.DeepEqual(expectedEtags, cc.previousEtags) || expectedConsistent != cc.currConsistent {
		t.Errorf("Failed ['abcde']: previousEtags (%+v) or currConsistent (%d) did not match expected (%+v, %d)", cc.previousEtags, cc.currConsistent, expectedEtags, expectedConsistent)
	}

	// A consistent etag should update consistent by 1
	cc.handleConsistency("abcde")

	expectedConsistent = 1

	if !reflect.DeepEqual(expectedEtags, cc.previousEtags) || expectedConsistent != cc.currConsistent {
		t.Errorf("Failed ['abcde']: previousEtags (%+v) or currConsistent (%d) did not match expected (%+v, %d)", cc.previousEtags, cc.currConsistent, expectedEtags, expectedConsistent)
	}

	// A consistent etag should update consistent by 1
	cc.handleConsistency("abcde")

	expectedConsistent = 2

	if !reflect.DeepEqual(expectedEtags, cc.previousEtags) || expectedConsistent != cc.currConsistent {
		t.Errorf("Failed ['abcde']: previousEtags (%+v) or currConsistent (%d) did not match expected (%+v, %d)", cc.previousEtags, cc.currConsistent, expectedEtags, expectedConsistent)
	}
}
