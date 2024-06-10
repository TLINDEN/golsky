package main

import (
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func NewMenuButton(
	text string,
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

		widget.ButtonOpts.Text(text, *FontRenderer.FontSmall, &widget.ButtonTextColor{
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

func NewToolbarButton(
	icon *ebiten.Image,
	action func(args *widget.ButtonClickedEventArgs)) *widget.Container {

	buttonImage, _ := LoadButtonImage()

	iconContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
		})),
	)

	button := widget.NewButton(
		widget.ButtonOpts.Image(buttonImage),
		widget.ButtonOpts.ClickedHandler(action),
	)

	iconContainer.AddChild(button)

	iconContainer.AddChild(widget.NewGraphic(widget.GraphicOpts.Image(icon)))

	return iconContainer
}

func NewCheckbox(
	text string,
	initialvalue bool,
	action func(args *widget.CheckboxChangedEventArgs)) *widget.LabeledCheckbox {

	checkboxImage, _ := LoadCheckboxImage()
	buttonImage, _ := LoadButtonImage()

	var state widget.WidgetState
	if initialvalue {
		state = widget.WidgetChecked
	}

	return widget.NewLabeledCheckbox(
		widget.LabeledCheckboxOpts.CheckboxOpts(
			widget.CheckboxOpts.ButtonOpts(
				widget.ButtonOpts.Image(buttonImage),
			),
			widget.CheckboxOpts.Image(checkboxImage),
			widget.CheckboxOpts.StateChangedHandler(action),
			widget.CheckboxOpts.InitialState(state),
		),
		widget.LabeledCheckboxOpts.LabelOpts(
			widget.LabelOpts.Text(text, *FontRenderer.FontSmall,
				&widget.LabelColor{
					Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
				}),
		),
	)
}

func NewSeparator(padding int) widget.PreferredSizeLocateableWidget {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Top:    padding,
				Bottom: 0,
			}))),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(
				widget.RowLayoutData{Stretch: true})))
	return c
}

type ListEntry struct {
	id   int
	Name string
}

func NewCombobox(items []string, selected string,
	action func(args *widget.ListComboButtonEntrySelectedEventArgs)) *widget.ListComboButton {
	buttonImage, _ := LoadButtonImage()

	entries := make([]any, 0, len(items))
	idxselected := 0
	for i, item := range items {
		entries = append(entries, ListEntry{i, item})
		if items[i] == selected {
			idxselected = i
		}
	}

	comboBox := widget.NewListComboButton(
		widget.ListComboButtonOpts.SelectComboButtonOpts(
			widget.SelectComboButtonOpts.ComboButtonOpts(
				//Set the max height of the dropdown list
				widget.ComboButtonOpts.MaxContentHeight(150),
				//Set the parameters for the primary displayed button
				widget.ComboButtonOpts.ButtonOpts(
					widget.ButtonOpts.Image(buttonImage),
					widget.ButtonOpts.TextPadding(widget.NewInsetsSimple(5)),
					widget.ButtonOpts.Text("", *FontRenderer.FontSmall, &widget.ButtonTextColor{
						Idle:     color.White,
						Disabled: color.White,
					}),
					widget.ButtonOpts.WidgetOpts(
						//Set how wide the button should be
						widget.WidgetOpts.MinSize(50, 0),
						//Set the combobox's position
						widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
							HorizontalPosition: widget.AnchorLayoutPositionCenter,
							VerticalPosition:   widget.AnchorLayoutPositionCenter,
						})),
				),
			),
		),
		widget.ListComboButtonOpts.ListOpts(
			//Set how wide the dropdown list should be
			widget.ListOpts.ContainerOpts(
				widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(50, 0)),
			),
			//Set the entries in the list
			widget.ListOpts.Entries(entries),
			widget.ListOpts.ScrollContainerOpts(
				//Set the background images/color for the dropdown list
				widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
					Idle:     image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
					Disabled: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
					Mask:     image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
				}),
			),
			widget.ListOpts.SliderOpts(
				//Set the background images/color for the background of the slider track
				widget.SliderOpts.Images(&widget.SliderTrackImage{
					Idle:  image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
					Hover: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
				}, buttonImage),
				widget.SliderOpts.MinHandleSize(5),
				//Set how wide the track should be
				widget.SliderOpts.TrackPadding(widget.NewInsetsSimple(2))),
			//Set the font for the list options
			widget.ListOpts.EntryFontFace(*FontRenderer.FontSmall),
			//Set the colors for the list
			widget.ListOpts.EntryColor(&widget.ListEntryColor{
				Selected:                   color.NRGBA{254, 255, 255, 255},
				Unselected:                 color.NRGBA{254, 255, 255, 255},
				SelectedBackground:         HexColor2RGBA(THEMES["standard"].life),
				SelectedFocusedBackground:  HexColor2RGBA(THEMES["standard"].old),
				FocusedBackground:          HexColor2RGBA(THEMES["standard"].old),
				DisabledUnselected:         HexColor2RGBA(THEMES["standard"].grid),
				DisabledSelected:           HexColor2RGBA(THEMES["standard"].grid),
				DisabledSelectedBackground: HexColor2RGBA(THEMES["standard"].grid),
			}),
			//Padding for each entry
			widget.ListOpts.EntryTextPadding(widget.NewInsetsSimple(5)),
		),
		//Define how the entry is displayed
		widget.ListComboButtonOpts.EntryLabelFunc(
			func(e any) string {
				//Button Label function, visible if not open
				return e.(ListEntry).Name
			},
			func(e any) string {
				//List Label function, visible items if open
				return e.(ListEntry).Name
			}),
		//Callback when a new entry is selected
		widget.ListComboButtonOpts.EntrySelectedHandler(action),
	)

	//Select the middle entry -- optional
	comboBox.SetSelectedEntry(entries[idxselected])

	return comboBox
}

func NewLabel(text string) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(text, *FontRenderer.FontSmall, color.White),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

}

/////////////// containers

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

// setup a top level toolbar container
func NewTopRowContainer(title string) *RowContainer {
	buttonImageHover := image.NewNineSlice(Assets["button-9slice3"], [3]int{3, 3, 3}, [3]int{3, 3, 3})

	uiContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	rowContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(8)),
			widget.RowLayoutOpts.Spacing(0),
		)),
		widget.ContainerOpts.BackgroundImage(buttonImageHover),
	)

	uiContainer.AddChild(rowContainer)

	return &RowContainer{
		Root: uiContainer,
		Row:  rowContainer,
	}
}

// set arg to false if no background needed
func NewRowContainer(title string) *RowContainer {
	buttonImageHover := image.NewNineSlice(Assets["button-9slice3"], [3]int{3, 3, 3}, [3]int{3, 3, 3})

	uiContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	titleLabel := widget.NewText(
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.TextOpts.Text(title, *FontRenderer.FontNormal, color.NRGBA{0xdf, 0xf4, 0xff, 0xff}))

	rowContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(8)),
			widget.RowLayoutOpts.Spacing(0),
		)),
		widget.ContainerOpts.BackgroundImage(buttonImageHover),
	)

	rowContainer.AddChild(titleLabel)

	uiContainer.AddChild(rowContainer)

	return &RowContainer{
		Root: uiContainer,
		Row:  rowContainer,
	}
}

func NewColumnContainer() *widget.Container {
	colcontainer := widget.NewContainer(
		widget.ContainerOpts.Layout(
			widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(2),
				widget.GridLayoutOpts.Spacing(5, 0),
			),
		),
	)

	return colcontainer
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

func LoadComboLabelImage() *widget.ButtonImageImage {
	return &widget.ButtonImageImage{
		Idle:     Assets["checkbox-9slice2"],
		Disabled: Assets["checkbox-9slice2"],
	}
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
