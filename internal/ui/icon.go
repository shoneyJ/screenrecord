package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type ClickableIcon struct {
	widget.BaseWidget
	icon              fyne.Resource
	onTapped          func()
	onTappedSecondary func(*fyne.PointEvent)
}

func NewClickableIcon(icon fyne.Resource, tapped func(), tappedSecondary func(*fyne.PointEvent)) *ClickableIcon {
	c := &ClickableIcon{
		icon:              icon,
		onTapped:          tapped,
		onTappedSecondary: tappedSecondary,
	}
	c.ExtendBaseWidget(c)
	return c
}

func (c *ClickableIcon) CreateRenderer() fyne.WidgetRenderer {
	img := canvas.NewImageFromResource(c.icon)
	img.FillMode = canvas.ImageFillContain

	return &clickableIconRenderer{
		icon:  img,
		owner: c,
	}
}

func (c *ClickableIcon) Tapped(*fyne.PointEvent) {
	if c.onTapped != nil {
		c.onTapped()
	}
}

func (c *ClickableIcon) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonSecondary && c.onTappedSecondary != nil {
		c.onTappedSecondary(&fyne.PointEvent{
			AbsolutePosition: e.AbsolutePosition,
			Position:         e.Position,
		})
	}
}

func (c *ClickableIcon) MouseUp(*desktop.MouseEvent) {}

func (c *ClickableIcon) SetIcon(icon fyne.Resource) {
	c.icon = icon
	c.Refresh()
}

type clickableIconRenderer struct {
	icon  *canvas.Image
	owner *ClickableIcon
}

func (r *clickableIconRenderer) Layout(size fyne.Size) {
	r.icon.Resize(size)
}

func (r *clickableIconRenderer) MinSize() fyne.Size {
	return fyne.NewSize(24, 24)
}

func (r *clickableIconRenderer) Refresh() {
	r.icon.Resource = r.owner.icon
	r.icon.Refresh()
}

func (r *clickableIconRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.icon}
}

func (r *clickableIconRenderer) Destroy() {}

func (r *clickableIconRenderer) BackgroundColor() color.Color {
	return color.Transparent
}
