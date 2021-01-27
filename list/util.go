package list

func (last *Pair) ncdr(car interface{}) *Pair {
	cdr := &Pair{Car: car}
	last.Cdr = cdr
	return cdr
}

type (
	carArgs struct {
		args     []interface{}
		cdrSlice []*Pair
	}

	listArgs struct {
		args     *Pair
		cdrSlice []*Pair
	}

	pairArgs struct {
		args     []*Pair
		cdrSlice []*Pair
	}

	cdrSlice []*Pair
)

func initCarArgs(lists []*Pair) (result carArgs, ok bool) {
	for _, p := range lists {
		if p == nil {
			ok = false
			return
		}
		result.args = append(result.args, p.Car)
		result.cdrSlice = append(result.cdrSlice, p.Cdr.(*Pair))
	}
	ok = true
	return
}

func (a *carArgs) next() bool {
	for index, p := range a.cdrSlice {
		if p == nil {
			return false
		}
		a.args[index] = p.Car
		a.cdrSlice[index] = p.Cdr.(*Pair)
	}
	return true
}

func initListArgs(lists []*Pair) (result listArgs, ok bool) {
	if len(lists) == 0 {
		ok = true
		return
	}
	p := lists[0]
	if p == nil {
		ok = false
		return
	}
	result.args = &Pair{Car: p.Car}
	last := result.args
	result.cdrSlice = []*Pair{p.Cdr.(*Pair)}
	for _, p = range lists[1:] {
		if p == nil {
			ok = false
			return
		}
		last = last.ncdr(p.Car)
		result.cdrSlice = append(result.cdrSlice, p.Cdr.(*Pair))
	}
	last.Cdr = (*Pair)(nil)
	ok = true
	return
}

func (a *listArgs) next() bool {
	p := a.cdrSlice[0]
	if p == nil {
		return false
	}
	a.args = &Pair{Car: p.Car}
	last := a.args
	a.cdrSlice[0] = p.Cdr.(*Pair)
	index := 1
	for _, p = range a.cdrSlice[1:] {
		if p == nil {
			return false
		}
		last = last.ncdr(p.Car)
		a.cdrSlice[index] = p.Cdr.(*Pair)
		index++
	}
	last.Cdr = (*Pair)(nil)
	return true
}

func initPairArgs(lists []*Pair) (result pairArgs, ok bool) {
	for _, p := range lists {
		if p == nil {
			ok = false
			return
		}
		result.args = append(result.args, p)
		result.cdrSlice = append(result.cdrSlice, p.Cdr.(*Pair))
	}
	ok = true
	return
}

func (a *pairArgs) next() bool {
	for index, p := range a.cdrSlice {
		if p == nil {
			return false
		}
		a.args[index] = p
		a.cdrSlice[index] = p.Cdr.(*Pair)
	}
	return true
}

func initCdrSlice(lists []*Pair) (result cdrSlice, ok bool) {
	for _, p := range lists {
		if p == nil {
			ok = false
			return
		}
		result = append(result, p)
	}
	ok = true
	return
}

func (cs *cdrSlice) next() bool {
	for index, p := range *cs {
		np := p.Cdr.(*Pair)
		if np == nil {
			return false
		}
		(*cs)[index] = np
	}
	return true
}
