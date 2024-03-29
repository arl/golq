package lq

import (
	"fmt"
	"log"
	"testing"
)

type set[K comparable] map[K]struct{}

type idset set[int]

func (m idset) assertEmpty(t *testing.T) {
	t.Helper()

	if len(m) > 0 {
		t.Errorf("got len(idset) = %d, want empty", len(m))
	}
}

func (m idset) assertContains(t *testing.T, id int) {
	t.Helper()

	if _, ok := m[id]; !ok {
		t.Errorf("idset doesn't not contain id=%d but should", id)
	}
}

func (m idset) assertNotContains(t *testing.T, id int) {
	t.Helper()

	if _, ok := m[id]; ok {
		t.Errorf("idset contains id=%d but shouldn't", id)
	}
}

func (m idset) assertIsContained(t *testing.T, id int, contains bool) {
	t.Helper()

	if contains {
		m.assertContains(t, id)
	} else {
		m.assertNotContains(t, id)
	}
}

// storeID is a CallBackFunction that stores the entity ID into the set.
func (m idset) storeID(id int, sqDist float64) {
	m[id] = struct{}{}
}

// printEntity is a CallBackFunction that prints the entity.
func printEntity(id int, sqDist float64) {
	log.Printf("printAllEntities: id:%+v %f\n", id, sqDist)
}

func TestAddObjectToDatabase(t *testing.T) {
	var tests = []struct {
		orgx, orgy float64
		szx, szy   float64
		divx, divy int
		ptx, pty   float64
		name       string
	}{
		{0, 0, 10, 10, 5, 5, 5, 5, "inside super brick"},
		{0, 0, 10, 10, 5, 5, 0, 11, "outside super brick (top)"},
		{0, 0, 10, 10, 5, 5, -1, 0, "outside super brick (left)"},
		{0, 0, 10, 10, 5, 5, 11, 0, "outside super brick (right)"},
		{0, 0, 10, 10, 5, 5, 0, -1, "outside super brick (bottom)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewDB[int](tt.orgx, tt.orgy, tt.szx, tt.szy, tt.divx, tt.divy)

			db.Attach(1, tt.ptx, tt.pty)

			ids := make(idset)
			db.ForEachObject(ids.storeID)

			ids.assertContains(t, 1)
		})
	}
}

func TestRemoveObject(t *testing.T) {
	var tests = []struct {
		orgx, orgy float64
		szx, szy   float64
		divx, divy int
		ptx, pty   float64
		name       string
	}{
		{0, 0, 10, 10, 5, 5, 5, 5, "inside super brick"},
		{0, 0, 10, 10, 5, 5, 0, 11, "outside super brick (top)"},
		{0, 0, 10, 10, 5, 5, -1, 0, "outside super brick (left)"},
		{0, 0, 10, 10, 5, 5, 11, 0, "outside super brick (right)"},
		{0, 0, 10, 10, 5, 5, 0, -1, "outside super brick (bottom)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewDB[int](tt.orgx, tt.orgy, tt.szx, tt.szy, tt.divx, tt.divy)

			p1 := db.Attach(1, tt.ptx, tt.pty)
			p1.removeFromBin()

			ids := make(idset)
			db.ForEachObject(ids.storeID)

			ids.assertNotContains(t, 1)
		})
	}
}

func TestRemoveAllObjects(t *testing.T) {
	var tests = []struct {
		orgx, orgy float64
		szx, szy   float64
		divx, divy int
		ptx, pty   float64
		name       string
	}{
		{0, 0, 10, 10, 5, 5, 5, 5, "inside super brick"},
		{0, 0, 10, 10, 5, 5, -1, -1, "outside super brick"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := NewDB[int](tt.orgx, tt.orgy, tt.szx, tt.szy, tt.divx, tt.divy)

			db.Attach(1, tt.ptx, tt.pty)
			db.Attach(2, tt.ptx, tt.pty)
			db.DetachAll()

			ids := make(idset)
			db.ForEachObject(ids.storeID)
			ids.assertEmpty(t)
		})
	}
}

func TestObjectLocality(t *testing.T) {
	var tests = []struct {
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
	for i, tt := range tests {
		t.Run(fmt.Sprintf("locality test %d", i), func(t *testing.T) {
			db := NewDB[int](0, 0, 10, 10, 5, 5)

			db.Attach(1, tt.p1x, tt.p1y)
			db.Attach(2, tt.p2x, tt.p2y)
			db.Attach(3, tt.p3x, tt.p3y)

			ids := make(idset)
			db.ForEachWithinRadius(tt.cx, tt.cy, tt.cr, ids.storeID)

			ids.assertIsContained(t, 1, tt.r1)
			ids.assertIsContained(t, 2, tt.r2)
			ids.assertIsContained(t, 3, tt.r3)
		})
	}
}

func TestBinRelinking(t *testing.T) {
	for i := range []int{1, 2, 3} {
		db := NewDB[int](0, 0, 10, 10, 5, 5)

		p1 := db.Attach(1, 5, 5)
		p2 := db.Attach(2, 5, 5)
		p3 := db.Attach(3, 5, 5)

		switch i {
		case 1:
			p1.removeFromBin()
		case 2:
			p2.removeFromBin()
		case 3:
			p3.removeFromBin()
		}

		ids := make(idset)
		db.ForEachWithinRadius(5, 5, 1, ids.storeID)
		ids.assertIsContained(t, 1, i != 1)
		ids.assertIsContained(t, 2, i != 2)
		ids.assertIsContained(t, 3, i != 3)
	}
}

func TestNearestNeighbor(t *testing.T) {
	var tests = []struct {
		p1x, p1y, p2x, p2y, p3x, p3y float64 // the 3 points in the db
		cx, cy                       float64 // search circle center
		cr                           float64 // search circle radius
		ignore                       int     // ignored
		want                         int     // expected nearest object
		wantFound                    bool    // expected value of found
	}{
		{1, 1, 1, 2, 1, 3, 1, 1, 0.1, 1, 0, false},
		{1, 1, 1, 2, 1, 3, 1, 1, 1.1, 1, 2, true},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("nearest test %d", i), func(t *testing.T) {
			db := NewDB[int](0, 0, 10, 10, 5, 5)

			db.Attach(1, tt.p1x, tt.p1y)
			db.Attach(2, tt.p2x, tt.p2y)
			db.Attach(3, tt.p3x, tt.p3y)

			got, found := db.FindNearestInRadius(tt.cx, tt.cy, tt.cr, tt.ignore)
			if found != tt.wantFound {
				t.Errorf("found = %t, wantFound = %t", found, tt.wantFound)
			}
			if got != tt.want {
				t.Errorf("got nearest neighbour = %v, want %v", got, tt.want)
			}
		})
	}
}
