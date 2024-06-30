#include "grid.h"

Grid *NewGrid(int width, int height, int density) {
  Grid *grid = malloc(sizeof(struct Grid));
  grid->Width = width;
  grid->Height = height;
  grid->Density = density;

  grid->Data = malloc(height * sizeof(int *));
  for (int y = 0; y < grid->Height; y++) {
    grid->Data[y] = malloc(width * sizeof(int *));
  }

  FillRandom(grid);

  return grid;
}

void FillRandom(Grid *grid) {
  int r;
  for (int y = 0; y < grid->Width; y++) {
    for (int x = 0; x < grid->Height; x++) {
      r = GetRandomValue(0, grid->Density);
      if (r == 1)
        grid->Data[y][x] = r;
    }
  }
}
