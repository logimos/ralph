package plan

// NextWorkFeature returns the untested, non-deferred feature Ralph should treat as current work.
// Higher priority values are chosen first; when priority ties or all priorities are zero, file order wins.
func NextWorkFeature(plans []Plan) (id int, steps int, description string, category string) {
	type cand struct {
		idx int
		p   Plan
	}
	var c []cand
	for i, p := range plans {
		if p.Tested || p.Deferred {
			continue
		}
		c = append(c, cand{idx: i, p: p})
	}
	if len(c) == 0 {
		return 0, 0, "", ""
	}

	best := c[0]
	for _, x := range c[1:] {
		if betterWorkCandidate(x.p, x.idx, best.p, best.idx) {
			best = x
		}
	}
	p := best.p
	return p.ID, len(p.Steps), p.Description, p.Category
}

func betterWorkCandidate(p Plan, idx int, q Plan, qidx int) bool {
	if p.Priority != q.Priority {
		return p.Priority > q.Priority
	}
	return idx < qidx
}

// FileUsesPriority returns true if any plan row sets priority != 0 (prompt / UX hint).
func FileUsesPriority(plans []Plan) bool {
	for _, p := range plans {
		if p.Priority != 0 {
			return true
		}
	}
	return false
}
