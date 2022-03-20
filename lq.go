// Package lq implements a spatial database which stores objects each of which
// is associated with a 2D point (a location in a 2D space). The points serve as
// the "search key" for the associated object. It is intended to efficiently
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
// Overview of usage: an application using this facility to perform locality
// queries over objects of type myStruct would first create a database with:
//  db := NewDB[myObject]()
// Then, call Attach for each objects to attach to the database. Attach returns
// a 'proxy' object, which is a link between the user object and its
// representation in the locality database.
//  p := db.Attach(obj)
// When a client object moves, the application calls Update with the new
// location. Update is a method of the lq.Proxy object, that's why the the proxy
// object is generally kept within the user object, though it can be managed
// separately:
//  db.Update(123, 456)
// To perform a query, DB.ForEachWithinRadius is passed a user function which
// will be called for all client objects in the locality. See Func below for
// more detail.
//  func myFunc(obj T, sqDist float64) {
//      // do something with obj
//  }
//  DB.ForEachWithinRadius(x, y, radius, myFunc, nil)
// The DB.FindNearestInRadius function can be used to find a single nearest
// neighbor using the database. Note that "locality query" is also known as
// neighborhood query, neighborhood search, near neighbor search, and range
// query.
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
type DB[T comparable] struct {
	xorg, yorg float64 // xorg and yorg are the super-brick corner minimum coordinates
	szx, szy   float64 // length of the edges of the super-brick
	xdiv, ydiv int     // number of sub-brick divisions in each direction

	// Actual bins, allocated in a 1D slice (use coordsToIndex to go from bin
	// coordinates to index in this slice).
	bins []*Proxy[T]

	// Extra bin for "everything else" (points outside super-brick).
	other *Proxy[T]
}

// NewDB creates a new database, allocates the bin array, and returns the DB
// object.
//
// The six parameters define the properties of the 'super-brick':
//  - xorg/yorg: x/y coordinates of one corner of the super-brick, its minimum x
//    and y extent.
//  - xsize/ysize: the width and height of the super-brick.
//  - xdiv/ydiv: the number of subdivisions (sub-bricks) along each axis.
func NewDB[T comparable](xorg, yorg, xsize, ysize float64, xdiv, divy int) *DB[T] {
	return &DB[T]{
		xorg: xorg,
		yorg: yorg,
		szx:  xsize,
		szy:  ysize,
		xdiv: xdiv,
		ydiv: divy,
		bins: make([]*Proxy[T], xdiv*divy),
	}
}

// Attach attaches a new object to the database and returns a proxy object.
func (db *DB[T]) Attach(t T, x, y float64) *Proxy[T] {
	obj := &Proxy[T]{object: t}
	db.Update(obj, x, y)
	return obj
}

// Detach detaches the given proxy object from the database.
func (db *DB[T]) Detach(obj *Proxy[T]) {
	obj.removeFromBin()
	return
}

// Update updates the location of a proxy object in the database.
//
// It should be called for each client object every time its location changes.
// For example, in an animation application, this would be called each frame for
// every moving object.
func (db *DB[T]) Update(obj *Proxy[T], x, y float64) {
	// find bin for new location
	newBin := db.binForLocation(x, y)

	// Store location in client object, for future reference.
	obj.x = x
	obj.y = y

	// Has object's changed bin?
	if newBin != obj.bin {
		obj.removeFromBin()
		obj.addToBin(newBin)
	}
}

// coordsToIndex determines the index into linear bin array given 2D bin
// indices
func (db *DB[T]) coordsToIndex(ix, iy int) int {
	return ix*db.ydiv + iy
}

// Find the bin ID for a location in space. The location is given in
// terms of its XY coordinates. The bin ID is a pointer to a pointer
// to the bin contents list.
func (db *DB[T]) binForLocation(x, y float64) **Proxy[T] {
	// If point is outside the super-brick, return the 'other' bin.
	if x < db.xorg {
		return &(db.other)
	}
	if y < db.yorg {
		return &(db.other)
	}
	if x >= db.xorg+db.szx {
		return &(db.other)
	}
	if y >= db.yorg+db.szy {
		return &(db.other)
	}

	// Point is inside the super brik, compute the bin coordinates and return that bin.
	ix := int((x - db.xorg) / db.szx * float64(db.xdiv))
	iy := int((y - db.yorg) / db.szy * float64(db.ydiv))
	return &(db.bins[db.coordsToIndex(ix, iy)])
}

// Func is the function called, for each proxy object, when iterating over a set
// of proxies. Func gets called with the object in question and the squared
// distance from the center of the search locality circle (x,y) to the object's
// key-point (when applicable).
type Func[T any] func(obj T, sqDist float64)

// ForEachObject applies a user-supplied function to all objects in the
// database, regardless of locality (see DB.ForEachWithinRadius). Since there's
// no search locality, the squared distance argument to f is undefined.
func (db *DB[T]) ForEachObject(f Func[T]) {
	for i := range db.bins {
		db.bins[i].traverseBin(f)
	}
	db.other.traverseBin(f)
}

// DetachAll detaches all proxy objects from the database.
func (db *DB[T]) DetachAll() {
	for i := range db.bins {
		pbin := &(db.bins[i])
		for *pbin != nil {
			(*pbin).removeFromBin()
		}
	}

	if db.other != nil {
		pbin := &(db.other)
		for *pbin != nil {
			(*pbin).removeFromBin()
		}
	}
}

// This subroutine of ForEachWithinRadius efficiently traverses a
// subset of bins specified by max and min bin coordinates.
func (db *DB[T]) forEachInRadiusClipped(x, y, radius float64, f Func[T], xmin, ymin, xmax, ymax int) {
	sqRadius := radius * radius

	// Loop for x bins across diameter of circle.
	idx := xmin * db.ydiv
	for i := xmin; i <= xmax; i++ {
		// Loop for y bins across diameter of circle.
		jdx := ymin
		for j := ymin; j <= ymax; j++ {
			// Traverse current bin's client object list.
			traverseBinWithinRadius(db.bins[idx+jdx], x, y, sqRadius, f)
			jdx++
		}
		idx += db.ydiv
	}
}

// If the query region (sphere) extends outside of the "super-brick"
// we need to check for objects in the catch-all "other" bin which
// holds any object which are not inside the regular sub-bricks
func (db *DB[T]) forEachObjectOutside(x, y, radius float64, f Func[T]) {
	// traverse the "other" bin's client object list
	traverseBinWithinRadius(db.other, x, y, radius*radius, f)
}

// ForEachWithinRadius applies an application-specific ObjectFunc to all objects
// in a certain locality.
//
// The locality is specified as a circle with a given center and radius. All
// objects whose location (key-point) is within this circle are identified and
// the f ObjectFunc function is applied to them. This method uses the lq
// database to quickly reject any objects in bins which do not overlap with the
// circle of interest. Incremental calculation of index values is used to
// efficiently traverse the bins of interest.
func (db *DB[T]) ForEachWithinRadius(x, y, radius float64, f Func[T]) {
	partlyOut := false
	completelyOutside := x+radius < db.xorg ||
		y+radius < db.yorg ||
		x-radius >= db.xorg+db.szx ||
		y-radius >= db.yorg+db.szy

	// Is the circle completely outside the "super brick"?
	if completelyOutside {
		db.forEachObjectOutside(x, y, radius, f)
	}

	// compute min and max bin coordinates for each dimension
	minBinX := int(float64(db.xdiv) * (x - radius - db.xorg) / db.szx)
	minBinY := int(float64(db.ydiv) * (y - radius - db.yorg) / db.szy)
	maxBinX := int(float64(db.xdiv) * (x + radius - db.xorg) / db.szx)
	maxBinY := int(float64(db.ydiv) * (y + radius - db.yorg) / db.szy)

	// clip bin coordinates
	if minBinX < 0 {
		partlyOut = true
		minBinX = 0
	}
	if minBinY < 0 {
		partlyOut = true
		minBinY = 0
	}
	if maxBinX >= db.xdiv {
		partlyOut = true
		maxBinX = db.xdiv - 1
	}
	if maxBinY >= db.ydiv {
		partlyOut = true
		maxBinY = db.ydiv - 1
	}

	// Map function over outside objects if necessary (if clipped)
	if partlyOut {
		db.forEachObjectOutside(x, y, radius, f)
	}

	// Map function over objects in bins
	db.forEachInRadiusClipped(x, y, radius, f, minBinX, minBinY, maxBinX, maxBinY)
}

// FindNearestInRadius searches the database to find the object whose key-point
// is nearest to a given location yet within a given radius.
//
// That is, it finds the object (if any) within a given search circle which is
// nearest to the circle's center. The ignored argument can be used to exclude
// an object from consideration. This is useful when looking for the nearest
// neighbor of an object in the database, since otherwise it would be its own
// nearest neighbor. The function returns the nearest object and true, or if
// there was no object with the provided radius, it returns the zero value of T,
// and false.
func (db *DB[T]) FindNearestInRadius(x, y, radius float64, ignored T) (T, bool) {
	nearest := *new(T)
	minSqDist := math.MaxFloat64
	found := false

	// Map search helper function over all objects within radius.
	db.ForEachWithinRadius(x, y, radius, func(obj T, sqDist float64) {
		if ignored == obj {
			return
		}

		if sqDist < minSqDist {
			// Update nearest
			nearest = obj
			minSqDist = sqDist
			found = true
		}
	})

	return nearest, found
}

// Proxy is a proxy for a client (application) object in the spatial database.
//
// One of these should be created for each client object. This might be included
// within the structure of a client object, or could be allocated separately.
type Proxy[T any] struct {
	// Previous/next objects in this bin, or nil.
	prev, next *Proxy[T]

	// Bin (pointer to pointer to bin contents list).
	bin **Proxy[T]

	// Client object interface.
	object T

	// Object's location ("key point") used for spatial sorting.
	x, y float64
}

// addToBin adds a given client object to a given bin, linking it into the bin
// contents list.
func (cp *Proxy[T]) addToBin(bin **Proxy[T]) {
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

	cp.bin = bin
}

// removeFromBin removes a given client object from its current bin, unlinking
// it from the bin contents list.
func (cp *Proxy[T]) removeFromBin() {
	// Adjust pointers if object is currently in a bin
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
// the given ObjectFunc on each object that falls within the
// search radius.
func traverseBinWithinRadius[T comparable](cp *Proxy[T], x, y, sqRadius float64, fn Func[T]) {
	for cp != nil {
		// compute distance (squared) from this client
		// object to given locality circle's centerpoint
		sqDist := (x-cp.x)*(x-cp.x) + (y-cp.y)*(y-cp.y)

		// apply function if client object within sphere
		if sqDist < sqRadius {
			fn(cp.object, sqDist)
		}

		// consider next client object in bin list
		cp = cp.next
	}
}

func (cp *Proxy[T]) traverseBin(fn Func[T]) {
	// Walk down proxy list, applying call-back function to each one.
	for cp != nil {
		fn(cp.object, 0)
		cp = cp.next
	}
}
