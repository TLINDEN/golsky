package main

import (
	"bytes"
	"log"
	"log/slog"
	"path"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

type ShaderRegistry map[string]*ebiten.Shader

var Shaders = LoadShaders("assets/shaders")

func LoadShaders(dir string) ShaderRegistry {
	shaders := ShaderRegistry{}

	entries, err := assetfs.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to read shaders dir %s: %s", dir, err)
	}

	for _, file := range entries {
		path := path.Join(dir, file.Name())
		fd, err := assetfs.Open(path)
		if err != nil {
			log.Fatalf("failed to open shader file %s: %s", file.Name(), err)
		}
		defer fd.Close()

		name := strings.TrimSuffix(file.Name(), ".kg")

		buf := new(bytes.Buffer)
		buf.ReadFrom(fd)

		shader, err := ebiten.NewShader([]byte(buf.Bytes()))
		if err != nil {
			log.Fatal(err)
		}

		shaders[name] = shader

		slog.Debug("loaded shader asset", "path", path)
	}

	return shaders
}
