#ifndef _HAVE_GRID_H
#define _HAVE_GRID_H

#include "raylib.h"
#include <stdio.h>
#include <stdlib.h>

typedef struct Grid {
  int Width;
  int Height;
  int Density;
  int **Data;
} Grid;

Grid *NewGrid(int width, int height, int density);
void FillRandom(Grid *grid);

#endif
