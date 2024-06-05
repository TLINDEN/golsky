package main

import (
	"embed"
	"image"
	_ "image/png"
	"io/fs"
	"log"
	"path"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// Maps image name to image data
type AssetRegistry map[string]*ebiten.Image

// A helper to pass the registry easier around
type assetData struct {
	Registry AssetRegistry
}

//go:embed assets/sprites/*.png assets/fonts/*.ttf assets/shaders/*.kg
var assetfs embed.FS

// Called at build time, creates the global asset and animation registries
var Assets = LoadImages("assets/sprites")

// load pngs and json files
func LoadImages(dir string) AssetRegistry {
	Registry := AssetRegistry{}

	// we use embed.FS to iterate over all files in ./assets/
	entries, err := assetfs.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to read assets dir %s: %s", dir, err)
	}

	for _, imagefile := range entries {
		path := path.Join(dir, imagefile.Name())

		fd, err := assetfs.Open(path)
		if err != nil {
			log.Fatalf("failed to open file %s: %s", imagefile.Name(), err)
		}
		defer fd.Close()

		switch {
		case strings.HasSuffix(path, ".png"):
			name, image := ReadImage(imagefile, fd)
			Registry[name] = image
		}
	}

	return Registry
}

func ReadImage(imagefile fs.DirEntry, fd fs.File) (string, *ebiten.Image) {
	name := strings.TrimSuffix(imagefile.Name(), ".png")

	img, _, err := image.Decode(fd)
	if err != nil {
		log.Fatalf("failed to decode image %s: %s", imagefile.Name(), err)
	}

	image := ebiten.NewImageFromImage(img)

	return name, image
}
