#! /usr/bin/env python
# -*- coding: utf-8 -*-

from __future__ import print_function

import sys
import re

from collections import defaultdict
from itertools import cycle

import matplotlib.pyplot as plt

def main():
    if len(sys.argv) < 2:
        print ("need a benchmark to plot")
        sys.exit(1)

    pat = re.compile(r'Benchmark(\w*)Lq(\d*)Radius(\d*)-(?:\d)*(?:\s)*(\w*)(?:\s)*(\w*)')
    patbf = re.compile(r'BenchmarkBruteForce(\d*)-(?:\d)*(?:\s)*(\w*)(?:\s)*(\w*)')
    lines = [line.rstrip('\n') for line in open(sys.argv[1])]

    bf = [list(), list()]
    lq = defaultdict(lambda: [[], []])

    for line in lines:
        if 'BruteForce' in line:
            v = patbf.split(line)
            bf[0].append(v[1])
            bf[1].append(v[3])
        else:
            v = pat.split(line)
            if len(v) == 7:
                title = '{} radius {}'.format(v[1], v[3])
                lq[title][0].append(v[2])
                lq[title][1].append(v[5])


    legend = []

    plt.subplot(111, axisbg='darkslategray')
    plt.xlabel('elements in search space')
    plt.ylabel('time(ns)')

    # brute force
    plt.loglog(bf[0], bf[1], linewidth=2)
    legend.append('Brute Force')

    # lq benchmarks
    for label, lqb in lq.iteritems():
        plt.loglog(lqb[0], lqb[1], linewidth=2)
        legend.append(label)

    plt.legend(legend, loc='upper left')

    plt.grid(True)
    plt.savefig("benchmark.png")
    plt.show()

if __name__ == "__main__":
    main()
