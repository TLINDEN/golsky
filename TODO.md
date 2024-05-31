- add all other options like size etc

- Clear screen problem:
  - it works when hitting the K key, immediately
  - its being turned off correctly when entering menu and on when leaving it
  - but  regardless of the setting,  after turning it off,  the engine
    seems to run a couple of  ticks with the old setting before switching
    scenes
  - looks like a race condition
  - obviously  with K there  are more loops before  actually switching
    scenes, which doesn't happen with ESC
