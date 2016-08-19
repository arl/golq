package lq

import "testing"

// create our mock entity for test, it's just an int
func newEntity(id int) *ClientProxy {
	return NewClientProxy(id)
}

// convenience types
type (
	idMap  map[int]bool // map of ints, acting as a set of int
	idList []int        // slice of ints
)

func (m idMap) assertContains(t *testing.T, id int) {
	if _, ok := m[1]; !ok {
		t.Errorf("ids map was expected to contain id:", id)
	}
}

// CallBackFunction that copies every found entity id into the provided idMap or idList
func retrieveAllIds(clientObject interface{}, distanceSquared float64, clientQueryState interface{}) {
	switch clientQueryState.(type) {
	case idList:
		s := clientQueryState.(idList)
		s = append(s, clientObject.(int))
	case idMap:
		m := clientQueryState.(idMap)
		m[clientObject.(int)] = true
	}
}

func TestAddObjectToSuperBrick(t *testing.T) {
	db := CreateDatabase(0, 0, 10, 10, 5, 5)

	p1 := newEntity(1)
	db.UpdateForNewLocation(p1, 2, 0)

	ids := idMap{}
	db.MapOverAllObjects(retrieveAllIds, ids)
	ids.assertContains(t, 1)
}

func TestAddObjectOutOfSuperBrick(t *testing.T) {
	db := CreateDatabase(0, 0, 10, 10, 5, 5)

	p1 := newEntity(1)
	db.UpdateForNewLocation(p1, 11, 0)

	ids := idMap{}
	db.MapOverAllObjects(retrieveAllIds, ids)
	ids.assertContains(t, 1)
}
