package lq

import (
	"fmt"
	"log"
	"math/rand"
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

func TestDeleteDatabase(t *testing.T) {
	db := CreateDatabase(0, 0, 1, 1, 1, 1)
	DeleteDatabase(db)
	if db.bins != nil {
		t.Error("db.bins was non-nil when it should have been nil")
	}
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

// Benchmarks

const (
	seed = 436784 // rng seed
)

// global dummy variable to avoid optimization (see uses)
var dummy float64

type benchEntity struct {
	ID   int
	x, y float64
}

// create the bench client proxy
func newProxyEntity(ent benchEntity) *ClientProxy {
	return NewClientProxy(ent)
}

func randomNEntities(b *testing.B, src rand.Source, numPoints int) []benchEntity {
	rnd := rand.New(src)
	ents := make([]benchEntity, numPoints)
	for i := 0; i < numPoints; i++ {
		x, y := 10*rnd.Float64(), 10*rnd.Float64()
		ents[i] = benchEntity{i, x, y}
	}
	return ents
}

func benchmarkBruteForce(b *testing.B, numPts int) {
	var (
		src rand.Source
		rng *rand.Rand
	)
	src = rand.NewSource(seed)
	rng = rand.New(src)

	for n := 0; n < b.N; n++ {
		ents := randomNEntities(b, src, numPts)

		// generate random query point
		x, y := 10*rng.Float64(), 10*rng.Float64()

		// Brute force looks at all entities and would perform
		// something (keep the nearest, accumulate into a slice,
		// pass to a callback) with the ones wich distance is
		// within the search raidus. Nevertheless the square root
		// calculation is not performed and nothing is done on
		// objects actually within search radius, so this time is
		// not measured. Anyway the complexity of brute force
		// doesn't lie there, but it is caused by the fact that
		// every distance to the search point has to computed.
		for _, ent := range ents {
			// set to dummy variable to not let the compiler
			// optimize this statement out.
			dummy = (x-ent.x)*(x-ent.x) + (y-ent.y)*(y-ent.y)
		}
	}
}

// Brute force benchmarks

func BenchmarkBruteForce10(b *testing.B) {
	benchmarkBruteForce(b, 10)
}

func BenchmarkBruteForce50(b *testing.B) {
	benchmarkBruteForce(b, 50)
}

func BenchmarkBruteForce100(b *testing.B) {
	benchmarkBruteForce(b, 100)
}

func BenchmarkBruteForce200(b *testing.B) {
	benchmarkBruteForce(b, 200)
}

func BenchmarkBruteForce500(b *testing.B) {
	benchmarkBruteForce(b, 500)
}

func BenchmarkBruteForce1000(b *testing.B) {
	benchmarkBruteForce(b, 1000)
}

// NearestNeighbour benchmarks

func benchmarkNearestNeighbourLq(b *testing.B, numPts int, radius float64) {
	// superbrick settings
	orgx, orgy := 0.0, 0.0
	szx, szy := 10.0, 10.0
	divx, divy := 10, 10

	src := rand.NewSource(seed)
	rng := rand.New(src)

	// create and fill the database
	ents := randomNEntities(b, src, numPts)
	db := CreateDatabase(orgx, orgy, szx, szy, divx, divy)
	for _, ent := range ents {
		db.UpdateForNewLocation(newProxyEntity(ent), ent.x, ent.y)
	}

	for n := 0; n < b.N; n++ {
		// generate random query point
		x, y := 10*rng.Float64(), 10*rng.Float64()
		db.FindNearestNeighborWithinRadius(x, y, radius, nil)
	}
}

func BenchmarkNearestNeighbourLq10Radius2(b *testing.B) {
	benchmarkNearestNeighbourLq(b, 10, 2)
}

func BenchmarkNearestNeighbourLq50Radius2(b *testing.B) {
	benchmarkNearestNeighbourLq(b, 50, 2)
}

func BenchmarkNearestNeighbourLq100Radius2(b *testing.B) {
	benchmarkNearestNeighbourLq(b, 100, 2)
}

func BenchmarkNearestNeighbourLq200Radius2(b *testing.B) {
	benchmarkNearestNeighbourLq(b, 200, 2)
}

func BenchmarkNearestNeighbourLq500Radius2(b *testing.B) {
	benchmarkNearestNeighbourLq(b, 500, 2)
}

func BenchmarkNearestNeighbourLq1000Radius2(b *testing.B) {
	benchmarkNearestNeighbourLq(b, 1000, 2)
}

func BenchmarkNearestNeighbourLq10Radius4(b *testing.B) {
	benchmarkNearestNeighbourLq(b, 10, 4)
}

func BenchmarkNearestNeighbourLq50Radius4(b *testing.B) {
	benchmarkNearestNeighbourLq(b, 50, 4)
}

func BenchmarkNearestNeighbourLq100Radius4(b *testing.B) {
	benchmarkNearestNeighbourLq(b, 100, 4)
}

func BenchmarkNearestNeighbourLq200Radius4(b *testing.B) {
	benchmarkNearestNeighbourLq(b, 200, 4)
}

func BenchmarkNearestNeighbourLq500Radius4(b *testing.B) {
	benchmarkNearestNeighbourLq(b, 500, 4)
}

func BenchmarkNearestNeighbourLq1000Radius4(b *testing.B) {
	benchmarkNearestNeighbourLq(b, 1000, 4)
}

// ObjectsInLocality benchmarks

func benchmarkObjectsInLocalityLq(b *testing.B, numPts int, radius float64) {
	// superbrick settings
	orgx, orgy := 0.0, 0.0
	szx, szy := 10.0, 10.0
	divx, divy := 10, 10

	src := rand.NewSource(seed)
	rng := rand.New(src)

	// create and fill the database
	ents := randomNEntities(b, src, numPts)
	db := CreateDatabase(orgx, orgy, szx, szy, divx, divy)
	for _, ent := range ents {
		db.UpdateForNewLocation(newProxyEntity(ent), ent.x, ent.y)
	}

	dummyCallback := func(clientObj interface{}, distSquare float64, queryState interface{}) {}
	for n := 0; n < b.N; n++ {
		// generate random query point
		x, y := 10*rng.Float64(), 10*rng.Float64()
		db.MapOverAllObjectsInLocality(x, y, radius, dummyCallback, nil)
	}
}

func BenchmarkObjectsInLocalityLq10Radius2(b *testing.B) {
	benchmarkObjectsInLocalityLq(b, 10, 2)
}

func BenchmarkObjectsInLocalityLq50Radius2(b *testing.B) {
	benchmarkObjectsInLocalityLq(b, 50, 2)
}

func BenchmarkObjectsInLocalityLq100Radius2(b *testing.B) {
	benchmarkObjectsInLocalityLq(b, 100, 2)
}

func BenchmarkObjectsInLocalityLq200Radius2(b *testing.B) {
	benchmarkObjectsInLocalityLq(b, 200, 2)
}

func BenchmarkObjectsInLocalityLq500Radius2(b *testing.B) {
	benchmarkObjectsInLocalityLq(b, 500, 2)
}

func BenchmarkObjectsInLocalityLq1000Radius2(b *testing.B) {
	benchmarkObjectsInLocalityLq(b, 1000, 2)
}

func BenchmarkObjectsInLocalityLq10Radius4(b *testing.B) {
	benchmarkObjectsInLocalityLq(b, 10, 4)
}

func BenchmarkObjectsInLocalityLq50Radius4(b *testing.B) {
	benchmarkObjectsInLocalityLq(b, 50, 4)
}

func BenchmarkObjectsInLocalityLq100Radius4(b *testing.B) {
	benchmarkObjectsInLocalityLq(b, 100, 4)
}

func BenchmarkObjectsInLocalityLq200Radius4(b *testing.B) {
	benchmarkObjectsInLocalityLq(b, 200, 4)
}

func BenchmarkObjectsInLocalityLq500Radius4(b *testing.B) {
	benchmarkObjectsInLocalityLq(b, 500, 4)
}

func BenchmarkObjectsInLocalityLq1000Radius4(b *testing.B) {
	benchmarkObjectsInLocalityLq(b, 1000, 4)
}
