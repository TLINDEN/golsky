## Various performance tests

Running with 1500x1500 grid 5k times

| Variation          | Description                                                                 | Duration |
|--------------------|-----------------------------------------------------------------------------|----------|
| perf-2dim          | uses 2d grid of bools, no tuning                                            | 00:03:14 |
| perf-2dim-pointers | use 2d grid of `Cell{Neighbors,NeighborCount}`s using pointers to neighbors | 00:03:35 |
| perf-1dim          | use 1d grid of bools, access using y*x, no further tuning                   | 00:03:24 |
