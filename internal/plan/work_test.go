package plan

import "testing"

func TestNextWorkFeature_fileOrderWhenNoPriority(t *testing.T) {
	plans := []Plan{
		{ID: 1, Description: "a", Tested: false},
		{ID: 2, Description: "b", Tested: false},
	}
	id, _, _, _ := NextWorkFeature(plans)
	if id != 1 {
		t.Fatalf("got id %d want 1", id)
	}
}

func TestNextWorkFeature_skipsTestedAndDeferred(t *testing.T) {
	plans := []Plan{
		{ID: 1, Description: "done", Tested: true},
		{ID: 2, Description: "later", Deferred: true},
		{ID: 3, Description: "now", Tested: false},
	}
	id, _, _, _ := NextWorkFeature(plans)
	if id != 3 {
		t.Fatalf("got id %d want 3", id)
	}
}

func TestNextWorkFeature_higherPriorityFirst(t *testing.T) {
	plans := []Plan{
		{ID: 1, Description: "low", Priority: 1, Tested: false},
		{ID: 2, Description: "high", Priority: 10, Tested: false},
	}
	id, _, _, _ := NextWorkFeature(plans)
	if id != 2 {
		t.Fatalf("got id %d want 2", id)
	}
}

func TestNextWorkFeature_tieBreaksByFileOrder(t *testing.T) {
	plans := []Plan{
		{ID: 1, Description: "first", Priority: 5, Tested: false},
		{ID: 2, Description: "second", Priority: 5, Tested: false},
	}
	id, _, _, _ := NextWorkFeature(plans)
	if id != 1 {
		t.Fatalf("got id %d want 1", id)
	}
}

func TestFileUsesPriority(t *testing.T) {
	if FileUsesPriority([]Plan{{ID: 1}}) {
		t.Fatal("expected false")
	}
	if !FileUsesPriority([]Plan{{ID: 1, Priority: 1}}) {
		t.Fatal("expected true")
	}
}
