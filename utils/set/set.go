package set

var null = struct{}{}

type Set map[interface{}]struct{}

func New(a ...interface{}) (set Set) {
	set = Set{}
	for _, v := range a {
		set[v] = null
	}
	return
}

func (set Set) Len() int {
	return len(set)
}

func (set Set) List() (list []interface{}) {
	for v := range set {
		list = append(list, v)
	}
	return
}

func (set Set) Add(a ...interface{}) {
	for _, v := range a {
		set[v] = null
	}
}

func (set Set) Delete(v interface{}) {
	delete(set, v)
}

func (set Set) Contains(v interface{}) (ok bool) {
	_, ok = set[v]
	return
}
