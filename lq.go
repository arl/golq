// This utility is a spatial database which stores objects each of
// which is associated with a 3d point (a location in a 3d space).
// The points serve as the "search key" for the associated object.
// It is intended to efficiently answer "sphere inclusion" queries,
// also known as range queries: basically questions like:

//   Which objects are within a radius R of the location L?

// In this context, "efficiently" means significantly faster than the
// naive, brute force O(n) testing of all known points.  Additionally
// it is assumed that the objects move along unpredictable paths, so
// that extensive preprocessing (for example, constructing a Delaunay
// triangulation of the point set) may not be practical.

// The implementation is a "bin lattice": a 3d rectangular array of
// brick-shaped (rectangular parallelepipeds) regions of space.  Each
// region is represented by a pointer to a (possibly empty) doubly-
// linked list of objects.  All of these sub-bricks are the same
// size.  All bricks are aligned with the global coordinate axes.

// Terminology used here: the region of space associated with a bin
// is called a sub-brick.  The collection of all sub-bricks is called
// the super-brick.  The super-brick should be specified to surround
// the region of space in which (almost) all the key-points will
// exist.  If key-points move outside the super-brick everything will
// continue to work, but without the speed advantage provided by the
// spatial subdivision.  For more details about how to specify the
// super-brick's position, size and subdivisions see lqCreateDatabase
// below.

// Overview of usage: an application using this facility would first
// create a database with lqCreateDatabase.  For each client object
// the application wants to put in the database it creates a
// lqClientProxy and initializes it with lqInitClientProxy.  When a
// client object moves, the application calls lqUpdateForNewLocation.
// To perform a query MapOverAllObjectsInLocality is passed an
// application-supplied call-back function to be applied to all
// client objects in the locality.  See lqCallBackFunction below for
// more detail.  The lqFindNearestNeighborWithinRadius function can
// be used to find a single nearest neighbor using the database.

// Note that "locality query" is also known as neighborhood query,
// neighborhood search, near neighbor search, and range query.  For
// additional information on this and related topics see:
// http://www.red3d.com/cwr/boids/ips.html

// For some description and illustrations of this database in use,
// see this paper: http://www.red3d.com/cwr/papers/2000/pip.html
package lq

import "math"

// Database represents the spatial database.
//
// Typically one of these would be created (by a call to lq.CreateDatabase) for
// a given application.
type DB struct {
	// origin is the super-brick corner minimum coordinates
	originx, originy float64

	// length of the edges of the super-brick
	sizex, sizey float64

	// number of sub-brick divisions in each direction
	divx, divy int

	// pointer to an array of pointers, one for each bin
	//lqClientProxy **bins
	bins []*lqClientProxy

	// extra bin for "everything else" (points outside super-brick)
	other *lqClientProxy
}

//This structure is a proxy for (and contains a pointer to) a client
//(application) object in the spatial database.  One of these exists
//for each client object.  This might be included within the
//structure of a client object, or could be allocated separately.
type lqClientProxy struct {
	//previous object in this bin, or nil
	prev *lqClientProxy

	//next object in this bin, or nil
	next *lqClientProxy

	//bin ID (pointer to pointer to bin contents list)
	bin **lqClientProxy

	//client object interface
	object interface{}

	//the object's location ("key point") used for spatial sorting
	x, y float64
}

/* type for a pointer to a function used to map over client objects */
//func CallBackFunction(clientObject interface{}, distanceSquared float64, clientQueryState interface{})
type CallBackFunction func(interface{}, float64, interface{})

//Allocate and initialize an LQ database, return a pointer to it.
//The application needs to call this before using the LQ facility.
//The nine parameters define the properties of the "super-brick":
//(1) origin: coordinates of one corner of the super-brick, its
//minimum x, y and z extent.
//(2) size: the width, height and depth of the super-brick.
//(3) the number of subdivisions (sub-bricks) along each axis.
//This routine also allocates the bin array, and initialize its
//contents.
func CreateDatabase(originx, originy, sizex, sizey float64, divx, divy int) *DB {
	return &DB{
		originx: originx,
		originy: originy,
		sizex:   sizex,
		sizey:   sizey,
		divx:    divx,
		divy:    divy,
		bins:    make([]*lqClientProxy, divx*divy),
		other:   nil,
	}
}

//Deallocate the memory used by the LQ database
func DeleteDatabase(lq *DB) {
	lq.bins = nil
	lq = nil
}

// Determine index into linear bin array given 3D bin indices
func (lq *DB) binCoordsToBinIndex(ix, iy int) int {
	return ((ix * lq.divy) + iy)
}

//Find the bin ID for a location in space.  The location is given in
//terms of its XYZ coordinates.  The bin ID is a pointer to a pointer
//to the bin contents list.
func (lq *DB) BinForLocation(x, y float64) **lqClientProxy {
	var i, ix, iy int

	/* if point outside super-brick, return the "other" bin */
	if x < lq.originx {
		return &(lq.other)
	}
	if y < lq.originy {
		return &(lq.other)
	}
	if x >= lq.originx+lq.sizex {
		return &(lq.other)
	}
	if y >= lq.originy+lq.sizey {
		return &(lq.other)
	}

	/* if point inside super-brick, compute the bin coordinates */
	ix = (int)(((x - lq.originx) / float64(lq.sizex)) * float64(lq.divx))
	iy = (int)(((y - lq.originy) / float64(lq.sizey)) * float64(lq.divy))

	/* convert to linear bin number */
	i = lq.binCoordsToBinIndex(ix, iy)

	/* return pointer to that bin */
	return &(lq.bins[i])
}

// The application needs to call this once on each lqClientProxy at
// setup time to initialize its list pointers and associate the proxy
// with its client object.
func NewClientProxy(clientObject interface{}) *lqClientProxy {
	return &lqClientProxy{
		prev:   nil,
		next:   nil,
		bin:    nil,
		object: clientObject,
	}
}

//Adds a given client object to a given bin, linking it into the bin
//contents list.
func (object *lqClientProxy) AddToBin(bin **lqClientProxy) {
	/* if bin is currently empty */
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

	/* record bin ID in proxy object */
	object.bin = bin
}

//Removes a given client object from its current bin, unlinking it
//from the bin contents list.
func (object *lqClientProxy) RemoveFromBin() {
	/* adjust pointers if object is currently in a bin */
	if object.bin != nil {
		/* If this object is at the head of the list, move the bin
		   pointer to the next item in the list (might be nil). */
		if *(object.bin) == object {
			*(object.bin) = object.next
		}

		/* If there is a prev object, link its "next" pointer to the
		   object after this one. */
		if object.prev != nil {
			object.prev.next = object.next
		}

		/* If there is a next object, link its "prev" pointer to the
		   object before this one. */
		if object.next != nil {
			object.next.prev = object.prev
		}
	}

	/* Null out prev, next and bin pointers of this object. */
	object.prev = nil
	object.next = nil
	object.bin = nil
}

//Call for each client object every time its location changes.  For
//example, in an animation application, this would be called each
//frame for every moving object.

func (lq *DB) UpdateForNewLocation(object *lqClientProxy, x, y float64) {
	/* find bin for new location */
	newBin := lq.BinForLocation(x, y)

	/* store location in client object, for future reference */
	object.x = x
	object.y = y

	/* has object moved into a new bin? */
	if newBin != object.bin {
		object.RemoveFromBin()
		object.AddToBin(newBin)
	}
}

// Given a bin's list of client proxies, traverse the list and invoke
// the given lqCallBackFunction on each object that falls within the
// search radius.
func traverseBinClientObjectList(co *lqClientProxy, x, y, radiusSquared float64, fn CallBackFunction, state interface{}) {
	for co != nil {
		// compute distance (squared) from this client
		// object to given locality sphere's centerpoint
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

func (lq *DB) MapOverAllObjectsInLocalityClipped(x, y, radius float64,
	fn CallBackFunction,
	clientQueryState interface{},
	minBinX, minBinY, maxBinX, maxBinY int) {
	var (
		i, j                   int
		iindex, jindex, kindex int
		co                     *lqClientProxy
		bin                    **lqClientProxy
	)
	slab := lq.divy
	istart := minBinX * slab
	jstart := minBinY
	radiusSquared := radius * radius

	//#ifdef BOIDS_LQ_DEBUG
	//if (lqAnnoteEnable) drawBallGL (x, y, z, radius);
	//#endif

	/* loop for x bins across diameter of sphere */
	iindex = istart
	for i = minBinX; i <= maxBinX; i++ {
		/* loop for y bins across diameter of sphere */
		jindex = jstart
		for j = minBinY; j <= maxBinY; j++ {
			/* get current bin's client object list */
			bin = &lq.bins[iindex+jindex]
			co = *bin

			//#ifdef BOIDS_LQ_DEBUG
			//if (lqAnnoteEnable) drawBin (lq, bin);
			//#endif
			/* traverse current bin's client object list */
			traverseBinClientObjectList(co, x, y,
				radiusSquared,
				fn,
				clientQueryState)
			kindex += 1
		}
		iindex += slab
	}
}

//If the query region (sphere) extends outside of the "super-brick"
//we need to check for objects in the catch-all "other" bin which
//holds any object which are not inside the regular sub-bricks
func (lq *DB) MapOverAllOutsideObjects(
	x, y, radius float64,
	fn CallBackFunction,
	clientQueryState interface{}) {
	co := lq.other
	radiusSquared := radius * radius

	/* traverse the "other" bin's client object list */
	traverseBinClientObjectList(co, x, y,
		radiusSquared,
		fn,
		clientQueryState)
}

//Apply an application-specific function to all objects in a certain
//locality.  The locality is specified as a sphere with a given
//center and radius.  All objects whose location (key-point) is
//within this sphere are identified and the function is applied to
//them.  The application-supplied function takes three arguments:

//(1) a void* pointer to an lqClientProxy's "object".
//(2) the square of the distance from the center of the search
//locality sphere (x,y,z) to object's key-point.
//(3) a void* pointer to the caller-supplied "client query state"
//object -- typically NULL, but can be used to store state
//between calls to the lqCallBackFunction.

//This routine uses the LQ database to quickly reject any objects in
//bins which do not overlap with the sphere of interest.  Incremental
//calculation of index values is used to efficiently traverse the
//bins of interest.

func (lq *DB) MapOverAllObjectsInLocality(
	x, y, radius float64,
	fn CallBackFunction,
	clientQueryState interface{}) {
	partlyOut := false
	completelyOutside := (((x + radius) < lq.originx) || ((y + radius) < lq.originy) || ((x - radius) >= lq.originx+lq.sizex) || ((y - radius) >= lq.originy+lq.sizey))

	/* is the sphere completely outside the "super brick"? */
	if completelyOutside {
		lq.MapOverAllOutsideObjects(x, y, radius, fn,
			clientQueryState)
		return
	}

	/* compute min and max bin coordinates for each dimension */
	minBinX := (int)((((x - radius) - lq.originx) / lq.sizex) * float64(lq.divx))
	minBinY := (int)((((y - radius) - lq.originy) / lq.sizey) * float64(lq.divy))
	maxBinX := (int)((((x + radius) - lq.originx) / lq.sizex) * float64(lq.divx))
	maxBinY := (int)((((y + radius) - lq.originy) / lq.sizey) * float64(lq.divy))

	/* clip bin coordinates */
	if minBinX < 0 {
		partlyOut = true
		minBinX = 0
	}
	if minBinY < 0 {
		partlyOut = true
		minBinY = 0
	}
	if maxBinX >= lq.divx {
		partlyOut = true
		maxBinX = lq.divx - 1
	}
	if maxBinY >= lq.divy {
		partlyOut = true
		maxBinY = lq.divy - 1
	}

	/* map function over outside objects if necessary (if clipped) */
	if partlyOut {
		lq.MapOverAllOutsideObjects(x, y, radius, fn, clientQueryState)
	}

	/* map function over objects in bins */
	lq.MapOverAllObjectsInLocalityClipped(x, y,
		radius,
		fn,
		clientQueryState,
		minBinX, minBinY,
		maxBinX, maxBinY)
}

type lqFindNearestState struct {
	ignoreObject       interface{}
	nearestObject      interface{}
	minDistanceSquared float64
}

func lqFindNearestHelper(clientObject interface{}, distanceSquared float64, clientQueryState interface{}) {
	fns := clientQueryState.(*lqFindNearestState)

	/* do nothing if this is the "ignoreObject" */
	if fns.ignoreObject != clientObject {
		/* record this object if it is the nearest one so far */
		if fns.minDistanceSquared > distanceSquared {
			fns.nearestObject = clientObject
			fns.minDistanceSquared = distanceSquared
		}
	}
}

//Search the database to find the object whose key-point is nearest
//to a given location yet within a given radius.  That is, it finds
//the object (if any) within a given search sphere which is nearest
//to the sphere's center.  The ignoreObject argument can be used to
//exclude an object from consideration (or it can be NULL).  This is
//useful when looking for the nearest neighbor of an object in the
//database, since otherwise it would be its own nearest neighbor.
//The function returns a void* pointer to the nearest object, or
//NULL if none is found.
func (lq *DB) FindNearestNeighborWithinRadius(x, y, radius float64,
	ignoreObject interface{}) interface{} {
	/* initialize search state */
	var lqFNS lqFindNearestState
	lqFNS.nearestObject = nil
	lqFNS.ignoreObject = ignoreObject
	lqFNS.minDistanceSquared = math.MaxFloat64

	/* map search helper function over all objects within radius */
	lq.MapOverAllObjectsInLocality(x, y,
		radius,
		lqFindNearestHelper,
		&lqFNS)

	/* return nearest object found, if any */
	return lqFNS.nearestObject
}

func (proxy *lqClientProxy) MapOverAllObjectsInBin(
	fn CallBackFunction,
	clientQueryState interface{}) {
	/* walk down proxy list, applying call-back function to each one */
	for proxy != nil {
		fn(proxy.object, 0, clientQueryState)
		proxy = proxy.next
	}
}

//Apply a user-supplied function to all objects in the database,
//regardless of locality (cf lqMapOverAllObjectsInLocality)

func (lq *DB) MapOverAllObjects(fn CallBackFunction,
	clientQueryState interface{}) {
	bincount := lq.divx * lq.divy
	for i := 0; i < bincount; i++ {
		lq.bins[i].MapOverAllObjectsInBin(fn, clientQueryState)
	}
	lq.other.MapOverAllObjectsInBin(fn, clientQueryState)
}

func (bin *lqClientProxy) RemoveAllObjectsInBin() {
	for bin != nil {
		bin.RemoveFromBin()
	}
}

//Removes (all proxies for) all objects from all bins
func (lq *DB) RemoveAllObjects() {
	bincount := lq.divx * lq.divy
	for i := 0; i < bincount; i++ {
		lq.bins[i].RemoveAllObjectsInBin()
	}
	lq.other.RemoveAllObjectsInBin()
}
