package main

import (
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"golang.org/x/image/font"
)

func NewMenuButton(
	text string,
	face font.Face,
	action func(args *widget.ButtonClickedEventArgs)) *widget.Button {

	buttonImage, _ := LoadButtonImage()

	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position:  widget.RowLayoutPositionCenter,
				Stretch:   true,
				MaxWidth:  200,
				MaxHeight: 100,
			}),
		),

		widget.ButtonOpts.Image(buttonImage),

		widget.ButtonOpts.Text(text, face, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
		}),

		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:   5,
			Right:  5,
			Top:    5,
			Bottom: 5,
		}),

		widget.ButtonOpts.ClickedHandler(action),
	)
}

func NewCheckbox(
	text string,
	face font.Face,
	action func(args *widget.CheckboxChangedEventArgs)) *widget.LabeledCheckbox {

	checkboxImage, _ := LoadCheckboxImage()
	buttonImage, _ := LoadButtonImage()

	return widget.NewLabeledCheckbox(
		widget.LabeledCheckboxOpts.CheckboxOpts(
			widget.CheckboxOpts.ButtonOpts(widget.ButtonOpts.Image(buttonImage)),
			widget.CheckboxOpts.Image(checkboxImage),
			widget.CheckboxOpts.StateChangedHandler(action),
		),
		widget.LabeledCheckboxOpts.LabelOpts(
			widget.LabelOpts.Text(text, face,
				&widget.LabelColor{
					Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
				}),
		),
	)
}

type RowContainer struct {
	Root *widget.Container
	Row  *widget.Container
}

func (container *RowContainer) AddChild(child widget.PreferredSizeLocateableWidget) {
	container.Row.AddChild(child)
}

func (container *RowContainer) Container() *widget.Container {
	return container.Root
}

// set arg to false if no background needed
func NewRowContainer() *RowContainer {
	uiContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	rowContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(20)),
			widget.RowLayoutOpts.Spacing(0),
		)),
	)

	uiContainer.AddChild(rowContainer)

	return &RowContainer{
		Root: uiContainer,
		Row:  rowContainer,
	}
}

func LoadButtonImage() (*widget.ButtonImage, error) {
	idle := image.NewNineSlice(Assets["button-9slice2"], [3]int{3, 3, 3}, [3]int{3, 3, 3})
	hover := image.NewNineSlice(Assets["button-9slice3"], [3]int{3, 3, 3}, [3]int{3, 3, 3})
	pressed := image.NewNineSlice(Assets["button-9slice1"], [3]int{3, 3, 3}, [3]int{3, 3, 3})

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}, nil
}

func LoadCheckboxImage() (*widget.CheckboxGraphicImage, error) {
	unchecked := &widget.ButtonImageImage{
		Idle:     Assets["checkbox-9slice2"],
		Disabled: Assets["checkbox-9slice2"],
	}

	checked := &widget.ButtonImageImage{
		Idle:     Assets["checkbox-9slice1"],
		Disabled: Assets["checkbox-9slice1"],
	}

	return &widget.CheckboxGraphicImage{
		Checked:   checked,
		Unchecked: unchecked,
		Greyed:    unchecked,
	}, nil
}
