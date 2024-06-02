- add all other options like size etc

- changing options mid-game has no effect in most cases, even after a restart

- Statefile loading does not work correclty anymore. With larger grids
  everything  is empty.  With square  grids part  of the  grid is  cut
  off. Smaller grids load though
  
- Also  when loading a state  file, centering doesn't work  anymore, I
  think the geom calculation is overthrown by the parser func. So, put
  this calc into its own func and  always call. Or - as stated below -
  put it onto camera.go and call from Init().
  
- Zoom 0 on reset only  works when world<screen.  otherwise zoom would
  be negative So, on Init() memoize  centered camera position or add a
  Center() function  to camera.go.  Then on  reset calculate  the zoom
  level so that the world fits into the screen.
