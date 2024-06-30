#include "game.h"
#include "raylib.h"

int main(void) {
  Game *game = Init(800, 800, 10, 10, 8);

  while (!WindowShouldClose()) {
    Update(game);
    Draw(game);
  }

  CloseWindow();
  free(game);
  return 0;
}
