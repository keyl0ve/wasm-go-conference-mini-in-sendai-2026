//go:build js && wasm

package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"strconv"
	"strings"
	"syscall/js"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func processImage(this js.Value, args []js.Value) interface{} {
	defer func() {
		if r := recover(); r != nil {
			js.Global().Get("console").Call("error", "WASM panic:", r)
		}
	}()

	if len(args) < 4 {
		return resultOrError("", "arguments: base64, text, fontSize, colorHex required")
	}

	base64Input := args[0].String()
	text := args[1].String()
	fontSize := args[2].Int()
	colorHex := args[3].String()

	if fontSize <= 0 || fontSize > 500 {
		fontSize = 48
	}

	raw, err := base64.StdEncoding.DecodeString(base64Input)
	if err != nil {
		return resultOrError("", "base64 decode: "+err.Error())
	}

	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return resultOrError("", "image decode: "+err.Error())
	}

	// Support JPEG input; re-encode as JPEG. If other format, still try to encode as JPEG.
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	textColor, err := parseHexColor(colorHex)
	if err != nil {
		textColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}

	ttf, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return resultOrError("", "font parse: "+err.Error())
	}

	face, err := opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    float64(fontSize),
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return resultOrError("", "font face: "+err.Error())
	}
	defer face.Close()

	drawer := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(textColor),
		Face: face,
	}

	// Measure text and center horizontally; baseline at vertical center
	adv := drawer.MeasureString(text)
	width := bounds.Dx()
	height := bounds.Dy()
	x := (width - adv.Round()) / 2
	if x < 0 {
		x = 0
	}
	// Baseline: vertical center + (ascent/2) so the text is roughly centered
	metrics := face.Metrics()
	ascent := metrics.Ascent.Ceil()
	y := height/2 + ascent/2
	drawer.Dot = fixed.P(x, y)
	drawer.DrawString(text)

	var out bytes.Buffer
	if err := jpeg.Encode(&out, rgba, &jpeg.Options{Quality: 90}); err != nil {
		return resultOrError("", "jpeg encode: "+err.Error())
	}

	result := base64.StdEncoding.EncodeToString(out.Bytes())
	return resultOrError(result, "")
}

func resultOrError(result, errMsg string) js.Value {
	obj := js.Global().Get("Object").New()
	obj.Set("result", result)
	obj.Set("error", errMsg)
	return obj
}

// parseHexColor parses "#rrggbb" or "rrggbb" into color.RGBA.
func parseHexColor(s string) (color.RGBA, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.RGBA{}, nil
	}
	r, _ := strconv.ParseUint(s[0:2], 16, 8)
	g, _ := strconv.ParseUint(s[2:4], 16, 8)
	b, _ := strconv.ParseUint(s[4:6], 16, 8)
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}

func main() {
	js.Global().Set("processImage", js.FuncOf(processImage))
	select {}
}
