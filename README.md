golq : 2D Locality Queries
====

[![Build Status](https://github.com/arl/golq/workflows/Tests/badge.svg)](https://github.com/arl/golq/actions)
[![codecov](https://codecov.io/gh/arl/golq/branch/main/graph/badge.svg)](https://codecov.io/gh/arl/golq)
[![pkg.go.dev](https://pkg.go.dev/badge/github.com/arl/golq.svg)](https://pkg.go.dev/github.com/arl/golq)

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
(rectangles) regions of space. Each region is represented by a pointer to a
(possibly empty) doubly-linked list of objects. All of these sub-bricks are the
same size. All bricks are aligned with the global coordinate axes.


Usage example
-------------

**TODO**

Benchmarks
----------

![benchmark image](https://github.com/arl/golq/blob/readme-stuff/benchmarks.png)

This *logarithmic scale* plot shows the numbers obtained in the following
benchmarks. The *brute-force method* computes squared distances between every
object and the randomly chosen search query point. It doesn't even include the
additional time that would be taken to sort them in order to extract the
nearest (or K nearest).

```
BenchmarkBruteForce10-2                          2000000               819 ns/op
BenchmarkBruteForce50-2                           500000              3570 ns/op
BenchmarkBruteForce100-2                          200000              6958 ns/op
BenchmarkBruteForce200-2                          100000             13564 ns/op
BenchmarkBruteForce500-2                           50000             33037 ns/op
BenchmarkBruteForce1000-2                          20000             66053 ns/op
BenchmarkNearestNeighbourLq10Radius2-2           3000000               546 ns/op
BenchmarkNearestNeighbourLq50Radius2-2           2000000               788 ns/op
BenchmarkNearestNeighbourLq100Radius2-2          1000000              1025 ns/op
BenchmarkNearestNeighbourLq200Radius2-2          1000000              1434 ns/op
BenchmarkNearestNeighbourLq500Radius2-2           500000              2431 ns/op
BenchmarkNearestNeighbourLq1000Radius2-2          300000              4242 ns/op
BenchmarkNearestNeighbourLq10Radius4-2           2000000               987 ns/op
BenchmarkNearestNeighbourLq50Radius4-2           1000000              1480 ns/op
BenchmarkNearestNeighbourLq100Radius4-2          1000000              1985 ns/op
BenchmarkNearestNeighbourLq200Radius4-2           500000              2974 ns/op
BenchmarkNearestNeighbourLq500Radius4-2           300000              5427 ns/op
BenchmarkNearestNeighbourLq1000Radius4-2          200000              9927 ns/op
BenchmarkObjectsInLocalityLq10Radius2-2          3000000               410 ns/op
BenchmarkObjectsInLocalityLq50Radius2-2          3000000               570 ns/op
BenchmarkObjectsInLocalityLq100Radius2-2         2000000               731 ns/op
BenchmarkObjectsInLocalityLq200Radius2-2         1000000              1051 ns/op
BenchmarkObjectsInLocalityLq500Radius2-2         1000000              1854 ns/op
BenchmarkObjectsInLocalityLq1000Radius2-2         500000              3418 ns/op
BenchmarkObjectsInLocalityLq10Radius4-2          2000000               808 ns/op
BenchmarkObjectsInLocalityLq50Radius4-2          1000000              1102 ns/op
BenchmarkObjectsInLocalityLq100Radius4-2         1000000              1440 ns/op
BenchmarkObjectsInLocalityLq200Radius4-2         1000000              2169 ns/op
BenchmarkObjectsInLocalityLq500Radius4-2          300000              4000 ns/op
BenchmarkObjectsInLocalityLq1000Radius4-2         200000              7881 ns/op
PASS
ok      github.com/arl/golq        53.587s
```

Credits
-------

This library is loosely inspired by the C language lq utility in
[OpenSteer](https://github.com/meshula/OpenSteer).


License
-------

- [MIT License](LICENSE)
