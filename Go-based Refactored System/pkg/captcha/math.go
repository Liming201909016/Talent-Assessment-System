package captcha

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
)

// Simple math captcha image generator (bitmap). Returns base64 PNG and the answer string.
// 与 Java 端 KaptchaTextCreator(math) 效果对齐：生成形如 "3+5 =?" 的题目，答案为数字字符串。
func GenerateMath() (img string, answer string, err error) {
	a := rand.Intn(10)
	b := rand.Intn(10)
	op := []string{"+", "-", "*"}[rand.Intn(3)]
	var ans int
	switch op {
	case "+":
		ans = a + b
	case "-":
		ans = a - b
	case "*":
		ans = a * b
	}
	text := fmt.Sprintf("%d %s %d =?", a, op, b)
	pngB64, err := renderText(text)
	if err != nil {
		return "", "", err
	}
	return pngB64, fmt.Sprintf("%d", ans), nil
}

// 极简渲染：白底+简单像素字体。前端只需要 base64 可显示即可。
// 为了不引入 TrueType 依赖，使用 5x7 点阵简易数字。
var glyphs = map[rune][7]uint8{
	'0': {0x1E, 0x33, 0x37, 0x3B, 0x33, 0x33, 0x1E},
	'1': {0x0C, 0x1C, 0x0C, 0x0C, 0x0C, 0x0C, 0x1E},
	'2': {0x1E, 0x33, 0x30, 0x1C, 0x06, 0x33, 0x3F},
	'3': {0x1E, 0x33, 0x30, 0x1C, 0x30, 0x33, 0x1E},
	'4': {0x38, 0x3C, 0x36, 0x33, 0x7F, 0x30, 0x30},
	'5': {0x3F, 0x03, 0x1F, 0x30, 0x30, 0x33, 0x1E},
	'6': {0x1C, 0x06, 0x03, 0x1F, 0x33, 0x33, 0x1E},
	'7': {0x3F, 0x33, 0x30, 0x18, 0x0C, 0x0C, 0x0C},
	'8': {0x1E, 0x33, 0x33, 0x1E, 0x33, 0x33, 0x1E},
	'9': {0x1E, 0x33, 0x33, 0x3E, 0x30, 0x18, 0x0E},
	'+': {0x00, 0x0C, 0x0C, 0x3F, 0x0C, 0x0C, 0x00},
	'-': {0x00, 0x00, 0x00, 0x3F, 0x00, 0x00, 0x00},
	'*': {0x00, 0x12, 0x0C, 0x3F, 0x0C, 0x12, 0x00},
	'=': {0x00, 0x3F, 0x00, 0x00, 0x3F, 0x00, 0x00},
	'?': {0x1E, 0x33, 0x30, 0x18, 0x0C, 0x00, 0x0C},
	' ': {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
}

func renderText(text string) (string, error) {
	const w, h = 120, 40
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	x := 6
	fg := color.RGBA{R: 30, G: 60, B: 120, A: 255}
	for _, ch := range text {
		g, ok := glyphs[ch]
		if !ok {
			x += 6
			continue
		}
		for row := 0; row < 7; row++ {
			bits := g[row]
			for col := 0; col < 6; col++ {
				if bits&(1<<col) != 0 {
					for dy := 0; dy < 3; dy++ {
						for dx := 0; dx < 3; dx++ {
							px := x + col*3 + dx
							py := 6 + row*3 + dy
							if px < w && py < h {
								img.Set(px, py, fg)
							}
						}
					}
				}
			}
		}
		x += 20
	}
	// 添加简单噪点
	for i := 0; i < 60; i++ {
		img.Set(rand.Intn(w), rand.Intn(h), color.RGBA{R: uint8(rand.Intn(200)), G: uint8(rand.Intn(200)), B: uint8(rand.Intn(200)), A: 255})
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
