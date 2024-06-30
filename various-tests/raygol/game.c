#include "game.h"
#include <stdio.h>

Game *Init(int width, int height, int gridwidth, int gridheight, int density) {
  struct Game *game = malloc(sizeof(struct Game));

  game->ScreenWidth = width;
  game->ScreenHeight = height;
  game->Cellsize = width / gridwidth;
  game->Width = gridwidth;
  game->Height = gridheight;

  InitWindow(width, height, "golsky");
  SetTargetFPS(60);

  game->Grid = NewGrid(gridwidth, gridheight, density);

  return game;
}

void Update(Game *game) {
  if (IsKeyDown(KEY_Q)) {
    game->Done = true;
    exit(0);
  }
}

void Draw(Game *game) {
  BeginDrawing();

  ClearBackground(RAYWHITE);

  for (int y = 0; y < game->Width; y++) {
    for (int x = 0; x < game->Height; x++) {
      if (game->Grid->Data[y][x] == 1) {
        DrawRectangle(x * game->Cellsize, y * game->Cellsize, game->Cellsize,
                      game->Cellsize, GREEN);
      } else {
        DrawRectangle(x * game->Cellsize, y * game->Cellsize, game->Cellsize,
                      game->Cellsize, RAYWHITE);
      }
    }
  }

  DrawText("TEST", game->ScreenWidth / 2, 10, 20, RED);

  EndDrawing();
}
