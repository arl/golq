// Package lq implements a spatial database which stores objects each of which
// is associated with a 2D point (a location in a 2D space). The points serve
// as the "search key" for the associated object. It is intended to efficiently
// answer "circle inclusion" queries, also known as "range queries": basically
// questions like:
//
// Which objects are within a radius R of the location L?
//
// In this context, "efficiently" means significantly faster than the naive,
// brute force O(n) testing of all known points. Additionally it is assumed that
// the objects move along unpredictable paths, so that extensive preprocessing
// (for example, constructing a Delaunay triangulation of the point set) may not
// be practical.
//
// The implementation is a "bin lattice": a 2D rectangular array of brick-shaped
// (rectangles) regions of space. Each region is represented by a pointer to a
// (possibly empty) doubly-linked list of objects. All of these sub-bricks are
// the same size. All bricks are aligned with the global coordinate axes.
//
// Terminology used here: the region of space associated with a bin is called a
// sub-brick. The collection of all sub-bricks is called the super-brick. The
// super-brick should be specified to surround the region of space in which
// (almost) all the key-points will exist. If key-points move outside the
// super-brick everything will continue to work, but without the speed advantage
// provided by the spatial subdivision. For more details about how to specify
// the super-brick's position, size and subdivisions see NewDB below.
//
// Overview of usage: an application using this facility would first create a
// database with:
//  db := NewDBabase().
// For each client object the application wants to put in the database it
// creates a ClientProxy with :
//  p := NewClientProxy(clientObj).
// When a client object moves, the application calls :
//  p.UpdateForNewLocation()
// To perform a query, DB.MapOverAllObjectsInLocality is passed an
// application-supplied ObjCallback function to be applied to all client
// objects in the locality. See ObjCallback below for more detail.
//  func myObjCallback (clientObj interface{}, sqDist float64, queryState interface{}) {
//      // do something with clientObj...
//  }
//  DB.MapOverAllObjectsInLocality(x, y, radius, myObjCallback, nil)
// The DB.FindNearestNeighborWithinRadius function can be used to find a single
// nearest neighbor using the database. Note that "locality query" is also
// known as neighborhood query, neighborhood search, near neighbor search, and
// range query.
//
// Author: Aurélien Rainone
//
// Based on original work of: Craig Reynolds
package lq

import "math"

// DB represents the spatial database.
//
// Typically one of these would be created (by a call to DB.NewDB)
// for a given application.
type DB struct {
	// originx and originy are the super-brick corner minimum coordinates
	originx, originy float64

	// length of the edges of the super-brick
	sizex, sizey float64

	// number of sub-brick divisions in each direction
	divx, divy int

	// slice of proxy pointers, one for each bin
	bins []*ClientProxy

	// extra bin for "everything else" (points outside super-brick)
	other *ClientProxy
}

// NewDB creates and returns a new golq database object.
//
// The application needs to call this before using the lq facility. The six
// parameters define the properties of the "super-brick":
//     - origin: coordinates of one corner of the super-brick, its minimum x
//               and y extent.
//     - size: the width and height of the super-brick.
//     - the number of subdivisions (sub-bricks) along each axis.
// This routine also allocates the bin array.
func NewDB(originx, originy, sizex, sizey float64, divx, divy int) *DB {
	return &DB{
		originx: originx,
		originy: originy,
		sizex:   sizex,
		sizey:   sizey,
		divx:    divx,
		divy:    divy,
		bins:    make([]*ClientProxy, divx*divy),
	}
}

// Determine index into linear bin array given 2D bin indices
func (db *DB) binCoordsToBinIndex(ix, iy int) int {
	return ix*db.divy + iy
}

// Find the bin ID for a location in space. The location is given in
// terms of its XY coordinates. The bin ID is a pointer to a pointer
// to the bin contents list.
func (db *DB) binForLocation(x, y float64) **ClientProxy {
	// if point outside super-brick, return the "other" bin
	if x < db.originx {
		return &(db.other)
	}
	if y < db.originy {
		return &(db.other)
	}
	if x >= db.originx+db.sizex {
		return &(db.other)
	}
	if y >= db.originy+db.sizey {
		return &(db.other)
	}

	// if point inside super-brick, compute the bin coordinates
	ix := int((x - db.originx) / db.sizex * float64(db.divx))
	iy := int((y - db.originy) / db.sizey * float64(db.divy))

	// convert to linear bin number
	i := db.binCoordsToBinIndex(ix, iy)

	// return pointer to that bin
	return &(db.bins[i])
}

// UpdateForNewLocation updates a proxy object position in the database.
//
// It should be called for each client object every time its location
// changes. For example, in an animation application, this would be called
// each frame for every moving object.
func (db *DB) UpdateForNewLocation(object *ClientProxy, x, y float64) {
	// find bin for new location
	newBin := db.binForLocation(x, y)

	// store location in client object, for future reference
	object.x = x
	object.y = y

	// has object still in the same bin?
	if newBin == object.bin {
		return
	}

	object.RemoveFromBin()
	object.AddToBin(newBin)
}

// ObjCallback is the type of the user-supplied function used to map over
// client objects.
//
// An instance of ObjCallback takes three arguments:
//
//    - an empty interface corresponding to a ClientProxy's "object".
//    - the square of the distance from the center of the search locality
//      circle (x,y) to object's key-point.
//    - an empty interface corresponding to the caller-supplied "client query
//      state" object, typically nil, but can be used to store state between
//      calls to the ObjCallback.
type ObjCallback func(clientObj interface{}, distSquare float64, queryState interface{})

// MapOverAllObjects applies a user-supplied function to all objects in the
// database, regardless of locality (see DB.MapOverAllObjectsInLocality)
func (db *DB) MapOverAllObjects(fn ObjCallback, queryState interface{}) {
	bincount := db.divx * db.divy
	for i := 0; i < bincount; i++ {
		db.bins[i].mapOverAllObjectsInBin(fn, queryState)
	}
	db.other.mapOverAllObjectsInBin(fn, queryState)
}

// RemoveAllObjects removes (all proxies for) all objects from all bins.
func (db *DB) RemoveAllObjects() {
	removeAllObjectsInBin := func(pbin **ClientProxy) {
		for *pbin != nil {
			(*pbin).RemoveFromBin()
		}
	}

	for i := range db.bins {
		removeAllObjectsInBin(&(db.bins[i]))
	}

	if db.other != nil {
		removeAllObjectsInBin(&db.other)
	}
}

// This subroutine of MapOverAllObjectsInLocality efficiently traverses a
// subset of bins specified by max and min bin coordinates.
func (db *DB) mapOverAllObjectsInLocalityClipped(x, y, radius float64,
	fn ObjCallback,
	queryState interface{},
	minBinX, minBinY, maxBinX, maxBinY int) {

	var iindex, jindex int

	radiusSquared := radius * radius

	// loop for x bins across diameter of circle
	iindex = minBinX * db.divy
	for i := minBinX; i <= maxBinX; i++ {
		// loop for y bins across diameter of circle
		jindex = minBinY
		for j := minBinY; j <= maxBinY; j++ {
			// traverse current bin's client object list
			traverseBinClientObjectList(
				db.bins[iindex+jindex],
				x, y,
				radiusSquared,
				fn,
				queryState)
			jindex++
		}
		iindex += db.divy
	}
}

// If the query region (sphere) extends outside of the "super-brick"
// we need to check for objects in the catch-all "other" bin which
// holds any object which are not inside the regular sub-bricks
func (db *DB) mapOverAllOutsideObjects(
	x, y, radius float64,
	fn ObjCallback,
	queryState interface{}) {
	co := db.other
	radiusSquared := radius * radius

	// traverse the "other" bin's client object list
	traverseBinClientObjectList(co, x, y, radiusSquared, fn, queryState)
}

// MapOverAllObjectsInLocality applies an application-specific ObjCallback
// to all objects in a certain locality.
//
// The locality is specified as a circle with a given center and radius. All
// objects whose location (key-point) is within this circle are identified and
// the fn ObjCallback function is applied to them.
// This routine uses the "lq" database to quickly reject any objects in bins
// which do not overlap with the circle of interest. Incremental calculation of
// index values is used to efficiently traverse the bins of interest.
func (db *DB) MapOverAllObjectsInLocality(
	x, y, radius float64,
	fn ObjCallback,
	queryState interface{}) {
	partlyOut := false
	completelyOutside := x+radius < db.originx || y+radius < db.originy ||
		x-radius >= db.originx+db.sizex || y-radius >= db.originy+db.sizey

	// is the circle completely outside the "super brick"?
	if completelyOutside {
		db.mapOverAllOutsideObjects(x, y, radius, fn,
			queryState)
		return
	}

	// compute min and max bin coordinates for each dimension
	minBinX := int(float64(db.divx) * (x - radius - db.originx) / db.sizex)
	minBinY := int(float64(db.divy) * (y - radius - db.originy) / db.sizey)
	maxBinX := int(float64(db.divx) * (x + radius - db.originx) / db.sizex)
	maxBinY := int(float64(db.divy) * (y + radius - db.originy) / db.sizey)

	// clip bin coordinates
	if minBinX < 0 {
		partlyOut = true
		minBinX = 0
	}
	if minBinY < 0 {
		partlyOut = true
		minBinY = 0
	}
	if maxBinX >= db.divx {
		partlyOut = true
		maxBinX = db.divx - 1
	}
	if maxBinY >= db.divy {
		partlyOut = true
		maxBinY = db.divy - 1
	}

	// map function over outside objects if necessary (if clipped)
	if partlyOut {
		db.mapOverAllOutsideObjects(x, y, radius, fn, queryState)
	}

	// map function over objects in bins
	db.mapOverAllObjectsInLocalityClipped(x, y,
		radius,
		fn,
		queryState,
		minBinX, minBinY,
		maxBinX, maxBinY)
}

type findNearestState struct {
	ignoreObject       interface{}
	nearestObject      interface{}
	minDistanceSquared float64
}

func findNearestHelper(clientObj interface{}, distanceSquared float64, queryState interface{}) {
	fns := queryState.(*findNearestState)

	if fns.ignoreObject == clientObj {
		// do nothing if this is the "ignoreObject"
		return
	}

	// record this object if it is the nearest one so far
	if fns.minDistanceSquared > distanceSquared {
		fns.nearestObject = clientObj
		fns.minDistanceSquared = distanceSquared
	}
}

// FindNearestNeighborWithinRadius searches the database to find the object
// whose key-point is nearest to a given location yet within a given radius.
//
// That is, it finds the object (if any) within a given search circle which is
// nearest to the circle's center. The ignoreObject argument can be used to
// exclude an object from consideration (or it can be nil). This is useful when
// looking for the nearest neighbor of an object in the database, since
// otherwise it would be its own nearest neighbor.
// The function returns an interface to the nearest object, or nil if none is
// found.
func (db *DB) FindNearestNeighborWithinRadius(x, y, radius float64, ignoreObject interface{}) interface{} {
	// initialize search state
	fns := findNearestState{
		ignoreObject:       ignoreObject,
		minDistanceSquared: math.MaxFloat64,
	}

	// map search helper function over all objects within radius
	db.MapOverAllObjectsInLocality(x, y, radius, findNearestHelper, &fns)

	// return nearest object found, if any
	return fns.nearestObject
}

// ClientProxy is a proxy for a client (application) object in the spatial
// database.
//
// One of these exists for each client object. This might be included within
// the structure of a client object, or could be allocated separately.
type ClientProxy struct {
	// previous object in this bin, or nil
	prev *ClientProxy

	// next object in this bin, or nil
	next *ClientProxy

	// bin ID (pointer to pointer to bin contents list)
	bin **ClientProxy

	// client object interface
	object interface{}

	// the object's location ("key point") used for spatial sorting
	x, y float64
}

// NewClientProxy creates a new client object proxy.
//
// The application needs to call this once on each ClientProxy at
// setup time to initialize its list pointers and associate the proxy
// with its client object.
func NewClientProxy(clientObj interface{}) *ClientProxy {
	return &ClientProxy{object: clientObj}
}

// AddToBin adds a given client object to a given bin, linking it into the
// bin contents list.
func (cp *ClientProxy) AddToBin(bin **ClientProxy) {
	// if bin is currently empty
	if *bin == nil {
		cp.prev = nil
		cp.next = nil
		*bin = cp
	} else {
		cp.prev = nil
		cp.next = *bin
		(*bin).prev = cp
		*bin = cp
	}

	// record bin ID in proxy object
	cp.bin = bin
}

// RemoveFromBin removes a given client object from its current bin, unlinking
// it from the bin contents list.
func (cp *ClientProxy) RemoveFromBin() {
	// adjust pointers if object is currently in a bin
	if cp.bin != nil {
		// If this object is at the head of the list, move the bin
		//  pointer to the next item in the list (might be nil).
		if *(cp.bin) == cp {
			*(cp.bin) = cp.next
		}

		// If there is a prev object, link its "next" pointer to the
		// object after this one.
		if cp.prev != nil {
			cp.prev.next = cp.next
		}

		// If there is a next object, link its "prev" pointer to the
		// object before this one.
		if cp.next != nil {
			cp.next.prev = cp.prev
		}
	}

	// Null out prev, next and bin pointers of this object.
	cp.prev = nil
	cp.next = nil
	cp.bin = nil
}

// Given a bin's list of client proxies, traverse the list and invoke
// the given ObjCallback on each object that falls within the
// search radius.
func traverseBinClientObjectList(cp *ClientProxy, x, y, radiusSquared float64, fn ObjCallback, state interface{}) {
	for cp != nil {
		// compute distance (squared) from this client
		// object to given locality circle's centerpoint
		distanceSquared := (x-cp.x)*(x-cp.x) + (y-cp.y)*(y-cp.y)

		// apply function if client object within sphere
		if distanceSquared < radiusSquared {
			fn(cp.object, distanceSquared, state)
		}

		// consider next client object in bin list
		cp = cp.next
	}
}

func (cp *ClientProxy) mapOverAllObjectsInBin(fn ObjCallback, queryState interface{}) {
	// walk down proxy list, applying call-back function to each one
	for cp != nil {
		fn(cp.object, 0, queryState)
		cp = cp.next
	}
}
