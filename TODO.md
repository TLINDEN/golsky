- add all other options like size etc
- add gif export
- add toolbar (not working yet, see branch trackui)
- only draw visible part of the world
- print current mode to the bottom like pause, insert and mark
- add https://www.ibiblio.org/lifepatterns/october1970.html
- history: dont count age but do calc to get index to age tile based on cell age
- maybe pre calc neighbors as 8 slice of pointers to neighboring cells to faster do the count
  see various-tests/perf-2dim-pointers/: it's NOT faster :(
- https://mattnakama.com/blog/go-branchless-coding/
- add performance measurements, see:
  DrawTriangles: https://github.com/TLINDEN/testgol
  WritePixels:   https://github.com/TLINDEN/testgol/tree/wrpixels
https://www.tasnimzotder.com/blog/optimizing-game-of-life-algorithm

- Speed
  https://conwaylife.com/forums/viewtopic.php?f=7&t=3237
  
- Patterns:

A   Catagolue   textcensus   of,  say,   period-2   oscillators   from
non-symmetrical soups can be found at

https://catagolue.hatsya.com/textcensus/b3s23/C1/xp2

The URL is made by just adding the prefix "text" to the word "census",
in any URL linked to from a Catagolue census page such as this one:

https://catagolue.hatsya.com/census/b3s23/C1

Format:
https://conwaylife.com/wiki/Apgcode


Collections:

https://conwaylife.com/wiki/Pattern_of_the_Year
https://www.ibiblio.org/lifepatterns/
https://entropymine.com/jason/life/
https://github.com/Matthias-Merzenich/jslife-moving
https://conwaylife.com/ref/mniemiec/lifepage.htm
https://conwaylife.com/wiki/Spaceship ff.
