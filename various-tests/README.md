## Various performance tests

Running with 1500x1500 grid 5k times

| Variation                    | Description                                                                 | Duration          |
|------------------------------|-----------------------------------------------------------------------------|-------------------|
| perf-2dim                    | uses 2d grid of bools, no tuning                                            | 00:03:14          |
| perf-2dim-pointers           | use 2d grid of `Cell{Neighbors,NeighborCount}`s using pointers to neighbors | 00:03:35/00:04:75 |
| perf-2dim-pointers-array     | same as above but array of neighbors instead of slice                       | 00:02:40          |
| perf-2dim-pointers-all-array | use arrays for everything, static 1500x1500                                 | infinite, aborted |
| perf-1dim                    | use 1d grid of bools, access using y*x, no further tuning                   | 00:03:24          |
| perf-ecs                     | use arche ecs, unusable                                                     | 00:14:51          |
