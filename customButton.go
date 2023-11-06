package main

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type customButton struct {
	widget.BaseWidget
	text               string
	onTapped           func()
	isHovered, focused bool
	tapAnim            *fyne.Animation
	background         *canvas.Rectangle
}

func NewCustomButton(text string, icon fyne.Resource, tapped func()) *customButton {
	b := &customButton{text: text, onTapped: tapped}
	b.ExtendBaseWidget(b)
	return b
}

func (b *customButton) Tapped(*fyne.PointEvent) {
	b.onTapped()
}
func ToNRGBA(c color.Color) (r, g, b, a int) {
	// We use UnmultiplyAlpha with RGBA, RGBA64, and unrecognized implementations of Color.
	// It works for all Colors whose RGBA() method is implemented according to spec, but is only necessary for those.
	// Only RGBA and RGBA64 have components which are already premultiplied.
	switch col := c.(type) {
	// NRGBA and NRGBA64 are not premultiplied
	case color.NRGBA:
		r = int(col.R)
		g = int(col.G)
		b = int(col.B)
		a = int(col.A)
	case *color.NRGBA:
		r = int(col.R)
		g = int(col.G)
		b = int(col.B)
		a = int(col.A)
	case color.NRGBA64:
		r = int(col.R) >> 8
		g = int(col.G) >> 8
		b = int(col.B) >> 8
		a = int(col.A) >> 8
	case *color.NRGBA64:
		r = int(col.R) >> 8
		g = int(col.G) >> 8
		b = int(col.B) >> 8
		a = int(col.A) >> 8
	// Gray and Gray16 have no alpha component
	case *color.Gray:
		r = int(col.Y)
		g = int(col.Y)
		b = int(col.Y)
		a = 0xff
	case color.Gray:
		r = int(col.Y)
		g = int(col.Y)
		b = int(col.Y)
		a = 0xff
	case *color.Gray16:
		r = int(col.Y) >> 8
		g = int(col.Y) >> 8
		b = int(col.Y) >> 8
		a = 0xff
	case color.Gray16:
		r = int(col.Y) >> 8
		g = int(col.Y) >> 8
		b = int(col.Y) >> 8
		a = 0xff
	// Alpha and Alpha16 contain only an alpha component.
	case color.Alpha:
		r = 0xff
		g = 0xff
		b = 0xff
		a = int(col.A)
	case *color.Alpha:
		r = 0xff
		g = 0xff
		b = 0xff
		a = int(col.A)
	case color.Alpha16:
		r = 0xff
		g = 0xff
		b = 0xff
		a = int(col.A) >> 8
	case *color.Alpha16:
		r = 0xff
		g = 0xff
		b = 0xff
		a = int(col.A) >> 8
	default: // RGBA, RGBA64, and unknown implementations of Color
		r, g, b, a = unmultiplyAlpha(c)
	}
	return
}
func unmultiplyAlpha(c color.Color) (r, g, b, a int) {
	red, green, blue, alpha := c.RGBA()
	if alpha != 0 && alpha != 0xffff {
		red = (red * 0xffff) / alpha
		green = (green * 0xffff) / alpha
		blue = (blue * 0xffff) / alpha
	}
	// Convert from range 0-65535 to range 0-255
	r = int(red >> 8)
	g = int(green >> 8)
	b = int(blue >> 8)
	a = int(alpha >> 8)
	return
}
func newButtonTapAnimation(bg *canvas.Rectangle, w fyne.Widget) *fyne.Animation {
	return fyne.NewAnimation(canvas.DurationStandard, func(done float32) {
		mid := w.Size().Width / 2
		size := mid * done
		bg.Resize(fyne.NewSize(size*2, w.Size().Height))
		bg.Move(fyne.NewPos(mid-size, 0))

		r, g, bb, a := ToNRGBA(theme.PressedColor())
		aa := uint8(a)
		fade := aa - uint8(float32(aa)*done)
		bg.FillColor = &color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(bb), A: fade}
		canvas.Refresh(bg)
	})
}

func (b *customButton) CreateRenderer() fyne.WidgetRenderer {
	label := widget.NewLabel(b.text)
	b.background = canvas.NewRectangle(theme.ButtonColor())
	tapBG := canvas.NewRectangle(color.Transparent)
	b.tapAnim = newButtonTapAnimation(tapBG, b)
	b.tapAnim.Curve = fyne.AnimationEaseOut
	return &customButtonRenderer{button: b, label: label, background: b.background,
		tapBG:   tapBG,
		objects: []fyne.CanvasObject{b.background, tapBG, label}}
}

type customButtonRenderer struct {
	button     *customButton
	background *canvas.Rectangle
	tapBG      *canvas.Rectangle
	label      *widget.Label
	objects    []fyne.CanvasObject
}

func (b *customButton) MouseIn(*desktop.MouseEvent) {
	b.isHovered = true
	b.Refresh()
}

func (b *customButton) MouseOut() {
	b.isHovered = false
	b.Refresh()
}
func (b *customButton) MouseMoved(*desktop.MouseEvent) {}
func (b *customButton) Dragged(*fyne.PointEvent)       {}
func (b *customButton) TapAnimation() {
	if b.tapAnim == nil {
		return
	}
	b.tapAnim.Stop()
	b.tapAnim.Start()
}
func blendColor(under, over color.Color) color.Color {
	// This alpha blends with the over operator, and accounts for RGBA() returning alpha-premultiplied values
	dstR, dstG, dstB, dstA := under.RGBA()
	srcR, srcG, srcB, srcA := over.RGBA()

	srcAlpha := float32(srcA) / 0xFFFF
	dstAlpha := float32(dstA) / 0xFFFF

	outAlpha := srcAlpha + dstAlpha*(1-srcAlpha)
	outR := srcR + uint32(float32(dstR)*(1-srcAlpha))
	outG := srcG + uint32(float32(dstG)*(1-srcAlpha))
	outB := srcB + uint32(float32(dstB)*(1-srcAlpha))
	// We create an RGBA64 here because the color components are already alpha-premultiplied 16-bit values (they're just stored in uint32s).
	return color.RGBA64{R: uint16(outR), G: uint16(outG), B: uint16(outB), A: uint16(outAlpha * 0xFFFF)}

}
func (b *customButton) buttonColor() color.Color {
	switch {
	case b.focused:
		bg := theme.ButtonColor()
		return blendColor(bg, theme.FocusColor())
	case b.isHovered:
		return theme.HoverColor()
	default:
		return theme.ButtonColor()
	}
}
func (b *customButton) applyButtonTheme() {
	if b.background == nil {
		return
	}

	b.background.FillColor = b.buttonColor()
	b.background.Refresh()
}

func (r *customButtonRenderer) applyTheme() {
	r.button.applyButtonTheme()
}
func (b *customButtonRenderer) BackgroundColor() color.Color {
	if b.button.isHovered {
		return theme.HoverColor()
	}
	return theme.ButtonColor()
}
func (b *customButtonRenderer) Layout(size fyne.Size) {
	b.background.Resize(size)
	iconSize := fyne.NewSize(size.Height, size.Height)

	labelSize := fyne.NewSize(size.Width-iconSize.Width, size.Height)
	b.label.Resize(labelSize)
}

func (b *customButtonRenderer) MinSize() fyne.Size {
	split := strings.Split(b.button.text, "\n")
	if len(split) > 0 {
		return fyne.NewSize(598, float32(len(split))*25) // set the size you want here
	} else {
		return fyne.NewSize(598, 100)
	}
}

func (b *customButtonRenderer) Refresh() {
	b.applyTheme()
	b.background.Refresh()
	canvas.Refresh(b.button)
}

func (b *customButtonRenderer) Objects() []fyne.CanvasObject {
	return b.objects
}

func (b *customButtonRenderer) Destroy() {
}
