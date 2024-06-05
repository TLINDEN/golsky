package main

import (
	"log"

	"github.com/golang/freetype/truetype"
	"github.com/tinne26/etxt"
	"golang.org/x/image/font"
)

var FontRenderer = LoadFonts("assets/fonts")

const (
	GameFont       string = "NotoSans-Regular"
	GameFontETXT   string = "Noto Sans"
	FontSizeBig    int    = 48
	FontSizeNormal int    = 24
	FontSizeSmall  int    = 12
)

type Texter struct {
	Renderer   *etxt.Renderer
	FontNormal *font.Face
	FontBig    *font.Face
	FontSmall  *font.Face
}

func LoadFonts(dir string) Texter {
	// load the font for use with ebitenui
	fontbytes, err := assetfs.ReadFile(dir + "/" + GameFont + ".ttf")
	if err != nil {
		log.Fatal(err)
	}

	gamefont, err := truetype.Parse(fontbytes)
	if err != nil {
		log.Fatal(err)
	}

	gameface := truetype.NewFace(gamefont, &truetype.Options{
		Size:    float64(FontSizeNormal),
		DPI:     72,
		Hinting: font.HintingFull,
	})

	biggameface := truetype.NewFace(gamefont, &truetype.Options{
		Size:    float64(FontSizeBig),
		DPI:     72,
		Hinting: font.HintingFull,
	})

	smallgameface := truetype.NewFace(gamefont, &truetype.Options{
		Size:    float64(FontSizeSmall),
		DPI:     72,
		Hinting: font.HintingFull,
	})

	// load the font for use with etxt
	fontlib := etxt.NewFontLibrary()
	_, _, err = fontlib.ParseEmbedDirFonts(dir, assetfs)
	if err != nil {
		log.Fatalf("Error while loading fonts: %s", err.Error())
	}

	/*
		err = fontlib.EachFont(
			func(fontName string, font *etxt.Font) error {
				fmt.Printf("font: %s\n", fontName)
				return nil
			})
		if err != nil {
			log.Fatal(err)
		}
	*/

	if !fontlib.HasFont(GameFontETXT) {
		log.Fatal("missing font: " + GameFontETXT)
	}

	err = fontlib.EachFont(checkMissingRunes)
	if err != nil {
		log.Fatal(err)
	}

	renderer := etxt.NewStdRenderer()

	glyphsCache := etxt.NewDefaultCache(10 * 1024 * 1024) // 10MB
	renderer.SetCacheHandler(glyphsCache.NewHandler())
	renderer.SetFont(fontlib.GetFont(GameFontETXT))

	return Texter{
		Renderer:   renderer,
		FontNormal: &gameface,
		FontBig:    &biggameface,
		FontSmall:  &smallgameface,
	}
}

// helper function used with FontLibrary.EachFont to make sure
// all loaded fonts contain the characters or alphabet we want
func checkMissingRunes(name string, font *etxt.Font) error {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const symbols = "0123456789 .,;:!?-()[]{}_&#@"

	missing, err := etxt.GetMissingRunes(font, letters+symbols)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		log.Fatalf("Font '%s' missing runes: %s", name, string(missing))
	}
	return nil
}
