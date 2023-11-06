package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func NewCustomChoiceDialog(title string, message string, choices []string, callbacks []func(), win fyne.Window) dialog.Dialog {
	buttons := []fyne.CanvasObject{}

	for i, choice := range choices {
		localCallback := callbacks[i] // Capture loop variable
		button := widget.NewButton(choice, func() {
			localCallback()
			// Add your close dialog logic here if you want to close the dialog upon button click
		})
		buttons = append(buttons, button)
	}
	messageLabel := widget.NewLabel(dialogTextFormat(message))
	// Include the message label at the top of the dialog content
	allCanvasObjects := append([]fyne.CanvasObject{messageLabel}, buttons...)
	dialogContainer := container.NewVBox(allCanvasObjects...)

	return dialog.NewCustom(title, "No!", dialogContainer, win)
}
func NewCustomFieldDialog(title string, message string, fields *fyne.Container, callback func(), win fyne.Window) dialog.Dialog {
	messageLabel := widget.NewLabel(dialogTextFormat(message))
	dialogContainer := container.NewVBox(messageLabel, fields, widget.NewButton("Continue", func() { callback() }))
	return dialog.NewCustom(title, "Cancel", dialogContainer, win)
}
