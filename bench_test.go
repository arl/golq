package lq_test

import (
	"math/rand"
	"testing"

	lq "github.com/arl/golq"
)

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
func newProxyEntity(ent benchEntity) *lq.ClientProxy {
	return lq.NewClientProxy(ent)
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
		// within the search radius. Nevertheless the square root
		// calculation is not performed and nothing is done on
		// objects actually within search radius, so this time is
		// not measured. Anyway the complexity of brute force
		// doesn't lie there, but it is caused by the fact that
		// every distance to the search point has to computed.
		for _, ent := range ents {
			// compute the squared distance and assigns it to not let the
			// compiler optimize it away.
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
	db := lq.NewDB(orgx, orgy, szx, szy, divx, divy)
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
	db := lq.NewDB(orgx, orgy, szx, szy, divx, divy)
	for _, ent := range ents {
		db.UpdateForNewLocation(newProxyEntity(ent), ent.x, ent.y)
	}

	dummyCallback := func(clientObj interface{}, sqDist float64, queryState interface{}) {}
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
