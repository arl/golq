// This utility is a spatial database which stores objects each of which is
// associated with a 2d point (a location in a 2d space). The points serve as
// the "search key" for the associated object. It is intended to efficiently
// answer "circle inclusion" queries, also known as range queries: basically
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
// The implementation is a "bin lattice": a 2d rectangular array of brick-shaped
// (rectangular parallelepipeds) regions of space. Each region is represented by
// a pointer to a (possibly empty) doubly-linked list of objects. All of these
// sub-bricks are the same size. All bricks are aligned with the global
// coordinate axes.
//
// Terminology used here: the region of space associated with a bin is called a
// sub-brick. The collection of all sub-bricks is called the super-brick. The
// super-brick should be specified to surround the region of space in which
// (almost) all the key-points will exist. If key-points move outside the
// super-brick everything will continue to work, but without the speed advantage
// provided by the spatial subdivision. For more details about how to specify
// the super-brick's position, size and subdivisions see CreateDatabase below.
//
// Overview of usage: an application using this facility would first create a
// database with CreateDatabase. For each client object the application wants to
// put in the database it creates a ClientProxy with NewClientProxy. When a
// client object moves, the application calls ClientProxy.UpdateForNewLocation.
// To perform a query, DB.MapOverAllObjectsInLocality is passed an
// application-supplied call-back function to be applied to all client objects
// in the locality. See CallBackFunction below for more detail. The
// DB.FindNearestNeighborWithinRadius function can be used to find a single
// nearest neighbor using the database. Note that "locality query" is also
// known as neighborhood query, neighborhood search, near neighbor search, and
// range query.

package lq

import "math"

// DB represents the spatial database.
//
// Typically one of these would be created (by a call to DB.CreateDatabase)
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

// CallBackFunction is the type of the function used to map over client
// objects.
type CallBackFunction func(clientObject interface{}, distanceSquared float64, clientQueryState interface{})

// CreateDatabase initializes and returns an lq database.
//
// The application needs to call this before using the lq facility. The six
// parameters define the properties of the "super-brick":
//     - origin: coordinates of one corner of the super-brick, its minimum x
//               and y extent.
//     - size: the width and height of the super-brick.
//     - the number of subdivisions (sub-bricks) along each axis.
// This routine also allocates the bin array.
func CreateDatabase(originx, originy, sizex, sizey float64, divx, divy int) *DB {
	return &DB{
		originx: originx,
		originy: originy,
		sizex:   sizex,
		sizey:   sizey,
		divx:    divx,
		divy:    divy,
		bins:    make([]*ClientProxy, divx*divy),
		other:   nil,
	}
}

// DeleteDatabase unreferences the memory used by the lq database
func DeleteDatabase(db *DB) {
	db.bins = nil
	db = nil
}

// Determine index into linear bin array given 2D bin indices
func (db *DB) binCoordsToBinIndex(ix, iy int) int {
	return ((ix * db.divy) + iy)
}

// Find the bin ID for a location in space. The location is given in
// terms of its XY coordinates. The bin ID is a pointer to a pointer
// to the bin contents list.
func (db *DB) binForLocation(x, y float64) **ClientProxy {
	var i, ix, iy int

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
	ix = (int)(((x - db.originx) / float64(db.sizex)) * float64(db.divx))
	iy = (int)(((y - db.originy) / float64(db.sizey)) * float64(db.divy))

	// convert to linear bin number
	i = db.binCoordsToBinIndex(ix, iy)

	// return pointer to that bin
	return &(db.bins[i])
}

// NewClientProxy creates a new client object proxy.
//
// The application needs to call this once on each ClientProxy at
// setup time to initialize its list pointers and associate the proxy
// with its client object.
func NewClientProxy(clientObject interface{}) *ClientProxy {
	return &ClientProxy{
		prev:   nil,
		next:   nil,
		bin:    nil,
		object: clientObject,
	}
}

// AddToBin adds a given client object to a given bin, linking it into the
// bin contents list.
func (object *ClientProxy) AddToBin(bin **ClientProxy) {
	// if bin is currently empty
	if *bin == nil {
		object.prev = nil
		object.next = nil
		*bin = object
	} else {
		object.prev = nil
		object.next = *bin
		(*bin).prev = object
		*bin = object
	}

	// record bin ID in proxy object
	object.bin = bin
}

// RemoveFromBin removes a given client object from its current bin, unlinking
// it from the bin contents list.
func (object *ClientProxy) RemoveFromBin() {
	// adjust pointers if object is currently in a bin
	if object.bin != nil {
		// If this object is at the head of the list, move the bin
		//  pointer to the next item in the list (might be nil).
		if *(object.bin) == object {
			*(object.bin) = object.next
		}

		// If there is a prev object, link its "next" pointer to the
		// object after this one.
		if object.prev != nil {
			object.prev.next = object.next
		}

		// If there is a next object, link its "prev" pointer to the
		// object before this one.
		if object.next != nil {
			object.next.prev = object.prev
		}
	}

	// Null out prev, next and bin pointers of this object.
	object.prev = nil
	object.next = nil
	object.bin = nil
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

	// has object moved into a new bin?
	if newBin != object.bin {
		object.RemoveFromBin()
		object.AddToBin(newBin)
	}
}

// Given a bin's list of client proxies, traverse the list and invoke
// the given CallBackFunction on each object that falls within the
// search radius.
func traverseBinClientObjectList(co *ClientProxy, x, y, radiusSquared float64, fn CallBackFunction, state interface{}) {
	for co != nil {
		// compute distance (squared) from this client
		// object to given locality circle's centerpoint
		dx := x - co.x
		dy := y - co.y
		distanceSquared := (dx * dx) + (dy * dy)

		// apply function if client object within sphere
		if distanceSquared < radiusSquared {
			fn(co.object, distanceSquared, state)
		}

		// consider next client object in bin list
		co = co.next
	}
}

// This subroutine of MapOverAllObjectsInLocality efficiently traverses of
// subset of bins specified by max and min bin coordinates.
func (db *DB) mapOverAllObjectsInLocalityClipped(x, y, radius float64,
	fn CallBackFunction,
	clientQueryState interface{},
	minBinX, minBinY, maxBinX, maxBinY int) {

	var (
		iindex, jindex int
		co             *ClientProxy
		bin            **ClientProxy
	)

	radiusSquared := radius * radius

	// loop for x bins across diameter of circle
	iindex = minBinX * db.divy
	for i := minBinX; i <= maxBinX; i++ {
		// loop for y bins across diameter of circle
		jindex = minBinY
		for j := minBinY; j <= maxBinY; j++ {
			// get current bin's client object list
			bin = &db.bins[iindex+jindex]
			co = *bin

			// traverse current bin's client object list
			traverseBinClientObjectList(co, x, y,
				radiusSquared,
				fn,
				clientQueryState)
			jindex += 1
		}
		iindex += db.divy
	}
}

// If the query region (sphere) extends outside of the "super-brick"
// we need to check for objects in the catch-all "other" bin which
// holds any object which are not inside the regular sub-bricks
func (db *DB) mapOverAllOutsideObjects(
	x, y, radius float64,
	fn CallBackFunction,
	clientQueryState interface{}) {
	co := db.other
	radiusSquared := radius * radius

	// traverse the "other" bin's client object list
	traverseBinClientObjectList(co, x, y,
		radiusSquared,
		fn,
		clientQueryState)
}

// MapOverAllObjectsInLocality applies an application-specific function to all
// objects in a certain locality.
//
// The locality is specified as a circle with a given center and radius. All
// objects whose location (key-point) is within this circle are identified and
// the function is applied to them. The application-supplied function takes
// three arguments:
//
//    - an interface to a ClientProxy's "object".
//    - the square of the distance from the center of the search
//      locality circle (x,y) to object's key-point.
//    - an interface to the caller-supplied "client query state" object,
//      typically nil, but can be used to store state between calls to the
//      CallBackFunction.
//
// This routine uses the LQ database to quickly reject any objects in bins which
// do not overlap with the circle of interest. Incremental calculation of index
// values is used to efficiently traverse the bins of interest.
func (db *DB) MapOverAllObjectsInLocality(
	x, y, radius float64,
	fn CallBackFunction,
	clientQueryState interface{}) {
	partlyOut := false
	completelyOutside := (((x + radius) < db.originx) || ((y + radius) < db.originy) || ((x - radius) >= db.originx+db.sizex) || ((y - radius) >= db.originy+db.sizey))

	// is the circle completely outside the "super brick"?
	if completelyOutside {
		db.mapOverAllOutsideObjects(x, y, radius, fn,
			clientQueryState)
		return
	}

	// compute min and max bin coordinates for each dimension
	minBinX := (int)((((x - radius) - db.originx) / db.sizex) * float64(db.divx))
	minBinY := (int)((((y - radius) - db.originy) / db.sizey) * float64(db.divy))
	maxBinX := (int)((((x + radius) - db.originx) / db.sizex) * float64(db.divx))
	maxBinY := (int)((((y + radius) - db.originy) / db.sizey) * float64(db.divy))

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
		db.mapOverAllOutsideObjects(x, y, radius, fn, clientQueryState)
	}

	// map function over objects in bins
	db.mapOverAllObjectsInLocalityClipped(x, y,
		radius,
		fn,
		clientQueryState,
		minBinX, minBinY,
		maxBinX, maxBinY)
}

type findNearestState struct {
	ignoreObject       interface{}
	nearestObject      interface{}
	minDistanceSquared float64
}

func findNearestHelper(clientObject interface{}, distanceSquared float64, clientQueryState interface{}) {
	fns := clientQueryState.(*findNearestState)

	// do nothing if this is the "ignoreObject"
	if fns.ignoreObject != clientObject {
		// record this object if it is the nearest one so far
		if fns.minDistanceSquared > distanceSquared {
			fns.nearestObject = clientObject
			fns.minDistanceSquared = distanceSquared
		}
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
func (db *DB) FindNearestNeighborWithinRadius(x, y, radius float64,
	ignoreObject interface{}) interface{} {
	// initialize search state
	var fns findNearestState
	fns.nearestObject = nil
	fns.ignoreObject = ignoreObject
	fns.minDistanceSquared = math.MaxFloat64

	// map search helper function over all objects within radius
	db.MapOverAllObjectsInLocality(x, y,
		radius,
		findNearestHelper,
		&fns)

	// return nearest object found, if any
	return fns.nearestObject
}

func (proxy *ClientProxy) mapOverAllObjectsInBin(
	fn CallBackFunction,
	clientQueryState interface{}) {
	// walk down proxy list, applying call-back function to each one
	for proxy != nil {
		fn(proxy.object, 0, clientQueryState)
		proxy = proxy.next
	}
}

// MapOverAllObjects applies a user-supplied function to all objects in the
// database, regardless of locality (see DB.MapOverAllObjectsInLocality)
func (db *DB) MapOverAllObjects(fn CallBackFunction,
	clientQueryState interface{}) {
	bincount := db.divx * db.divy
	for i := 0; i < bincount; i++ {
		db.bins[i].mapOverAllObjectsInBin(fn, clientQueryState)
	}
	db.other.mapOverAllObjectsInBin(fn, clientQueryState)
}

func removeAllObjectsInBin(pbin **ClientProxy) {
	for {
		bin := *pbin
		if bin != nil {
			bin.RemoveFromBin()
		} else {
			break
		}
	}
}

// RemoveAllObjects removes (all proxies for) all objects from all bins.
func (db *DB) RemoveAllObjects() {
	bincount := db.divx * db.divy
	for i := 0; i < bincount; i++ {
		removeAllObjectsInBin(&(db.bins[i]))
	}

	if db.other != nil {
		removeAllObjectsInBin(&db.other)
	}
}
