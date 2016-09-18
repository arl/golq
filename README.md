# golq
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/aurelien-rainone/golq) [![Build Status](https://travis-ci.org/aurelien-rainone/golq.svg?branch=master)](https://travis-ci.org/aurelien-rainone/golq) [![Coverage Status](https://coveralls.io/repos/github/aurelien-rainone/golq/badge.svg?branch=master)](https://coveralls.io/github/aurelien-rainone/golq?branch=master)

**2D locality queries in Go**

This utility is a spatial database which stores objects each of which is
associated with a 2D point (a location in a 2D space). The points serve as
the *search key* for the associated object. It is intended to efficiently
answer **circle inclusion queries**, also known as **range queries**, basically
questions like:

>Which objects are within a radius R of the location L?

In this context, **efficiently means significantly faster** than the naive,
**brute force** ***O(n)*** testing of all known points. Additionally it is
assumed that the objects move along unpredictable paths, so that extensive
preprocessing (for example, constructing a Delaunay triangulation of the point
set) may not be practical.

The implementation is a **bin lattice**: a 2D rectangular array of brick-shaped
(rectangular parallelepipeds) regions of space. Each region is represented by
a pointer to a (possibly empty) doubly-linked list of objects. All of these
sub-bricks are the same size. All bricks are aligned with the global
coordinate axes.


## Credits

This library is loosely inspired by the C language lq utility in
[OpenSteer](https://github.com/meshula/OpenSteer).
