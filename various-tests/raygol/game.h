#ifndef _HAVE_GAME_H
#define _HAVE_GAME_H

#include "grid.h"
#include "raylib.h"
#include <stdlib.h>

typedef struct Game {
  // Camera2D Camera;
  int ScreenWidth;
  int ScreenHeight;
  int Cellsize;

  // Grid dimensions
  int Width;
  int Height;
  bool Done;
  Grid *Grid;
} Game;

Game *Init(int width, int height, int gridwidth, int gridheight, int density);
void Update(Game *game);
void Draw(Game *game);

#endif
