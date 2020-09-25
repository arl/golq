package lq

import (
	"fmt"
	"log"
	"testing"
)

// create the test client proxy
func newEntity(id int) *ClientProxy {
	return NewClientProxy(id)
}

// map of ints, acting as a set of int
type idMap map[int]bool

func (m idMap) assertEmpty(t *testing.T) {
	if len(m) > 0 {
		t.Errorf("want empty ids map, got : %v", m)
	}
}

func (m idMap) assertContains(t *testing.T, id int) {
	if _, ok := m[id]; !ok {
		t.Errorf("want ids map contains id %d, didn't get it", id)
	}
}

func (m idMap) assertNotContains(t *testing.T, id int) {
	if _, ok := m[id]; ok {
		t.Errorf("want ids map not contains id %d, got it", id)
	}
}

func (m idMap) assertContainedIs(t *testing.T, id int, contains bool) {
	if contains {
		m.assertContains(t, id)
	} else {
		m.assertNotContains(t, id)
	}
}

// CallBackFunction that copies every found entity id into the provided idMap
func retrieveAllIds(clientObject interface{}, distanceSquared float64, clientQueryState interface{}) {
	m := clientQueryState.(idMap)
	m[clientObject.(int)] = true
}

// CallBackFunction that prints all entities, for debugging
func printAllEntities(clientObject interface{}, distanceSquared float64, clientQueryState interface{}) {
	id := clientObject.(int)
	log.Printf("printAllEntities: id:%+v %f\n", id, distanceSquared)
}

func TestAddObjectToDatabase(t *testing.T) {
	var flagtests = []struct {
		orgx, orgy float64
		szx, szy   float64
		divx, divy int
		ptx, pty   float64
		tested     string
	}{
		{0, 0, 10, 10, 5, 5, 5, 5, "inside super brick"},
		{0, 0, 10, 10, 5, 5, 0, 11, "outside super brick (top)"},
		{0, 0, 10, 10, 5, 5, -1, 0, "outside super brick (left)"},
		{0, 0, 10, 10, 5, 5, 11, 0, "outside super brick (right)"},
		{0, 0, 10, 10, 5, 5, 0, -1, "outside super brick (bottom)"},
	}

	for _, tt := range flagtests {

		t.Run(tt.tested, func(t *testing.T) {

			db := CreateDatabase(tt.orgx, tt.orgy, tt.szx, tt.szy, tt.divx, tt.divy)

			p1 := newEntity(1)
			db.UpdateForNewLocation(p1, tt.ptx, tt.pty)

			ids := idMap{}
			db.MapOverAllObjects(retrieveAllIds, ids)

			ids.assertContains(t, 1)
		})
	}
}

func TestRemoveObject(t *testing.T) {
	var flagtests = []struct {
		orgx, orgy float64
		szx, szy   float64
		divx, divy int
		ptx, pty   float64
		tested     string
	}{
		{0, 0, 10, 10, 5, 5, 5, 5, "inside super brick"},
		{0, 0, 10, 10, 5, 5, 0, 11, "outside super brick (top)"},
		{0, 0, 10, 10, 5, 5, -1, 0, "outside super brick (left)"},
		{0, 0, 10, 10, 5, 5, 11, 0, "outside super brick (right)"},
		{0, 0, 10, 10, 5, 5, 0, -1, "outside super brick (bottom)"},
	}

	for _, tt := range flagtests {

		t.Run(tt.tested, func(t *testing.T) {

			db := CreateDatabase(tt.orgx, tt.orgy, tt.szx, tt.szy, tt.divx, tt.divy)

			p1 := newEntity(1)
			db.UpdateForNewLocation(p1, tt.ptx, tt.pty)
			p1.RemoveFromBin()

			ids := idMap{}
			db.MapOverAllObjects(retrieveAllIds, ids)

			ids.assertNotContains(t, 1)
		})
	}
}

func TestRemoveAllObjects(t *testing.T) {
	var flagtests = []struct {
		orgx, orgy float64
		szx, szy   float64
		divx, divy int
		ptx, pty   float64
		tested     string
	}{
		{0, 0, 10, 10, 5, 5, 5, 5, "inside super brick"},
		{0, 0, 10, 10, 5, 5, -1, -1, "outside super brick"},
	}

	for _, tt := range flagtests {

		t.Run(tt.tested, func(t *testing.T) {

			db := CreateDatabase(tt.orgx, tt.orgy, tt.szx, tt.szy, tt.divx, tt.divy)

			p1, p2 := newEntity(1), newEntity(2)
			ids := idMap{}
			db.UpdateForNewLocation(p1, tt.ptx, tt.pty)
			db.UpdateForNewLocation(p2, tt.ptx, tt.pty)
			db.RemoveAllObjects()

			db.MapOverAllObjects(retrieveAllIds, ids)
			ids.assertEmpty(t)
		})
	}
}

func TestObjectLocality(t *testing.T) {
	var flagtests = []struct {
		p1x, p1y, p2x, p2y, p3x, p3y float64 // the 3 points in the db
		cx, cy                       float64 // search circle center
		cr                           float64 // search circle radius
		r1, r2, r3                   bool    // expected result for p1, p2 and p3
	}{
		{1, 1, 1, 2, 1, 3, 1, 1, 0.1, true, false, false},
		{1, 1, 1, 2, 1, 3, 1, 1, 1.1, true, true, false},
		{1, 1, 1, 2, 1, 3, 1, 1, 2.1, true, true, true},
		{1, 1, 1, 2, 1, 3, 1, 1, 10, true, true, true},
		{1, 1, 1, 2, 1, 3, -1, -1, 1, false, false, false},
		{1, 1, 1, 2, 1, 3, -1, -1, 3, true, false, false},
		{0, 0, 5, 5, 10, 10, 0, 0, 1.5, true, false, false},
		{0, 0, 5, 5, 10, 10, 1, 1, 1.5, true, false, false},
		{0, 0, 5, 5, 10, 10, 1, 1, 1.5, true, false, false},
		{0, 0, 5, 5, 10, 10, 4, 4, 1.5, false, true, false},
		{0, 0, 5, 5, 10, 10, 5, 5, 1.5, false, true, false},
		{0, 0, 5, 5, 10, 10, 6, 6, 1.5, false, true, false},
		{0, 0, 5, 5, 10, 10, 9, 9, 1.5, false, false, true},
		{0, 0, 5, 5, 10, 10, 10, 10, 1.5, false, false, true},
		{0, 0, 5, 5, 10, 10, 11, 11, 1.5, false, false, true},
		{1, 1, 1, 2, 1, 3, -1, -1, 0.1, false, false, false},
		{1, 1, 1, 2, 1, 3, -11, -1, 0.1, false, false, false},
		{1, 1, 1, 2, 1, 3, -11, -11, 0.1, false, false, false},
		{1, 1, 1, 2, 1, 3, -1, -11, 0.1, false, false, false},
	}
	for i, tt := range flagtests {

		t.Run(fmt.Sprintf("locality test %d", i), func(t *testing.T) {

			db := CreateDatabase(0, 0, 10, 10, 5, 5)

			p1 := newEntity(1)
			p2 := newEntity(2)
			p3 := newEntity(3)
			db.UpdateForNewLocation(p1, tt.p1x, tt.p1y)
			db.UpdateForNewLocation(p2, tt.p2x, tt.p2y)
			db.UpdateForNewLocation(p3, tt.p3x, tt.p3y)

			ids := idMap{}
			db.MapOverAllObjectsInLocality(tt.cx, tt.cy, tt.cr, retrieveAllIds, ids)
			ids.assertContainedIs(t, 1, tt.r1)
			ids.assertContainedIs(t, 2, tt.r2)
			ids.assertContainedIs(t, 3, tt.r3)
		})
	}
}

func TestBinRelinking(t *testing.T) {

	for i := range []int{1, 2, 3} {
		db := CreateDatabase(0, 0, 10, 10, 5, 5)
		p1 := newEntity(1)
		p2 := newEntity(2)
		p3 := newEntity(3)
		db.UpdateForNewLocation(p1, 5, 5)
		db.UpdateForNewLocation(p2, 5, 5)
		db.UpdateForNewLocation(p3, 5, 5)

		switch i {
		case 1:
			p1.RemoveFromBin()
		case 2:
			p2.RemoveFromBin()
		case 3:
			p3.RemoveFromBin()
		}

		ids := idMap{}
		db.MapOverAllObjectsInLocality(5, 5, 1, retrieveAllIds, ids)
		ids.assertContainedIs(t, 1, i != 1)
		ids.assertContainedIs(t, 2, i != 2)
		ids.assertContainedIs(t, 3, i != 3)
	}
}

func TestNearestNeighbor(t *testing.T) {
	var flagtests = []struct {
		p1x, p1y, p2x, p2y, p3x, p3y float64     // the 3 points in the db
		cx, cy                       float64     // search circle center
		cr                           float64     // search circle radius
		ignore                       interface{} // ignored object
		want                         interface{} // expected nearest object
	}{
		{1, 1, 1, 2, 1, 3, 1, 1, 0.1, nil, 1},
		{1, 1, 1, 2, 1, 3, 1, 1, 0.1, 1, nil},
		{1, 1, 1, 2, 1, 3, 1, 1, 1.1, 1, 2},
	}
	for i, tt := range flagtests {

		t.Run(fmt.Sprintf("nearest test %d", i), func(t *testing.T) {

			db := CreateDatabase(0, 0, 10, 10, 5, 5)

			p1 := newEntity(1)
			p2 := newEntity(2)
			p3 := newEntity(3)
			db.UpdateForNewLocation(p1, tt.p1x, tt.p1y)
			db.UpdateForNewLocation(p2, tt.p2x, tt.p2y)
			db.UpdateForNewLocation(p3, tt.p3x, tt.p3y)

			got := db.FindNearestNeighborWithinRadius(tt.cx, tt.cy, tt.cr, tt.ignore)
			if got != tt.want {
				t.Errorf("want nearest neighbour %v, got %v", tt.want, got)
			}
		})
	}
}
