# golq: 2D locality queries in Go [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/aurelien-rainone/golq) [![Build Status](https://travis-ci.org/aurelien-rainone/golq.svg?branch=master)](https://travis-ci.org/aurelien-rainone/golq)

**golq is a 2D spatial database library**

This utility is a spatial database which stores objects each of which is
associated with a 2D point (a location in a 2D space). The points serve as
the *search key* for the associated object. It is intended to efficiently
answer **circle inclusion** queries, also known as *range queries*: basically
questions like:

>Which objects are within a radius R of the location L?

In this context, **efficiently means significantly faster** than the naive,
**brute force** ***O(n)*** testing of all known points. Additionally it is assumed that
the objects move along unpredictable paths, so that extensive preprocessing
(for example, constructing a Delaunay triangulation of the point set) may not
be practical.

The implementation is a **bin lattice**: a 2D rectangular array of brick-shaped
(rectangular parallelepipeds) regions of space. Each region is represented by
a pointer to a (possibly empty) doubly-linked list of objects. All of these
sub-bricks are the same size. All bricks are aligned with the global
coordinate axes.

## Credits

This library is loosely inspired by the C language lq utility in
[OpenSteer](https://github.com/meshula/OpenSteer).

## License

golq is open source software distributed in accordance with the MIT
License (http://www.opensource.org/licenses/mit-license.php), which says:

Copyright (c) 2016 Aurélien Rainone

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
