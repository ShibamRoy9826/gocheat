package ui

import (
	"log"
	"os"

	"github.com/Achno/gocheat/config"
	cheatstyles "github.com/Achno/gocheat/styles"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var addItemAscii = `
▄▀█ █▀▄ █▀▄
█▀█ █▄▀ █▄▀`

type InputFormSpec struct {
	Title       string
	Desc        string
	PlaceHolder string
	TextInput   textinput.Model
}

// BuildInputItem renders a single form item
func BuildInputItem(formItem InputFormSpec) string {
	items := make([]string, 0)

	// Add title
	items = append(items, cheatstyles.Styles.Title.Render(formItem.Title), "")

	// Add description
	items = append(items, cheatstyles.Dimmed(formItem.Desc), "")

	// Render the text input with the placeholder
	items = append(items, formItem.TextInput.View(), "")

	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

// BuildInputMenu renders the entire form screen
func BuildInputMenu(formItems []InputFormSpec) string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		cheatstyles.Styles.Title.Render(addItemAscii), "",
		cheatstyles.Styles.SubText.Render("Add a keybind (The tag is optional)"), "",
		"\n",
		BuildInputItem(formItems[0]),
		BuildInputItem(formItems[1]),
	)
}

// FormScreen Model impliments tea.Model
type InputFormScreen struct {
	Forms      []InputFormSpec
	FocusIndex int
	CursorMode cursor.Mode
}

// Initialize the Input Form Screen with 2 forms
func InitInputFormScreen() InputFormScreen {
	model := InputFormScreen{
		Forms: make([]InputFormSpec, 2),
	}

	// Initialize styles and text of forms
	for i := range model.Forms {
		t := textinput.New()
		t.Cursor.Style = cheatstyles.Styles.Title
		t.TextStyle = cheatstyles.Styles.Title
		t.Placeholder = "Placeholder"
		t.CharLimit = 60
		t.Prompt = "➤ "

		switch i {
		case 0:
			t.Focus()
			model.Forms[i] = InputFormSpec{
				Title:       "Keybind",
				Desc:        "ex. New Alacritty instance: meta + i",
				PlaceHolder: "Placeholder",
				TextInput:   t,
			}
		case 1:
			t.CharLimit = 20
			model.Forms[i] = InputFormSpec{
				Title:       "Tag",
				Desc:        "ex. Rofi or Alacritty or Kitty or Kwin",
				PlaceHolder: "Placeholder",
				TextInput:   t,
			}
		}
	}

	return model
}

// Update function to handle user input and update the model
func (screen InputFormScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			screen.FocusIndex = (screen.FocusIndex + 1) % len(screen.Forms)
			cmds := make([]tea.Cmd, len(screen.Forms))
			for i := range screen.Forms {
				if i == screen.FocusIndex {
					cmds[i] = screen.Forms[i].TextInput.Focus()
					screen.Forms[i].TextInput.PromptStyle = cheatstyles.Styles.Title
					screen.Forms[i].TextInput.TextStyle = cheatstyles.Styles.Title
				} else {
					screen.Forms[i].TextInput.Blur()
					screen.Forms[i].TextInput.PromptStyle = cheatstyles.Styles.Success
					screen.Forms[i].TextInput.TextStyle = cheatstyles.Styles.Success
				}
			}
			return screen, tea.Batch(cmds...)

		case "enter":
			AddItemToList(screen)
			InitItems()
			ItemScreen := InitItemScreen()
			return ItemScreen, nil

		case "esc":
			ItemScreen := InitItemScreen()
			return ItemScreen, nil

		case "ctrl+c":
			// Handle form submission
			return screen, tea.Quit
		}
	}

	// Handle character input and blinking
	cmd := screen.updateInputs(msg)

	return screen, cmd
}

// Centralized input update function
func (screen *InputFormScreen) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(screen.Forms))
	for i := range screen.Forms {
		screen.Forms[i].TextInput, cmds[i] = screen.Forms[i].TextInput.Update(msg)
	}
	return tea.Batch(cmds...)
}

// View function to render the UI
func (screen InputFormScreen) View() string {
	width, height, _ := term.GetSize(int(os.Stdout.Fd()))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, BuildInputMenu(screen.Forms))
}

// Init function to initialize a blinking cursor
func (screen InputFormScreen) Init() tea.Cmd {
	return textinput.Blink
}

// Adds an item to the list depending on the values of the form and writes it to config.json
func AddItemToList(inputScreen InputFormScreen) error {

	// create the Item from the form
	item := Item{
		Title: inputScreen.Forms[0].TextInput.Value(),
		Tag:   inputScreen.Forms[1].TextInput.Value(),
	}

	items = append(items, item)

	// write the item to config.json
	wrapper := config.ItemWrapper{
		Title: inputScreen.Forms[0].TextInput.Value(),
		Tag:   inputScreen.Forms[1].TextInput.Value(),
	}

	config.GoCheatOptions.Items = append(config.GoCheatOptions.Items, wrapper)

	err := config.UpdateConfig()

	if err != nil {
		log.Fatalf("Failed writing item to config.json: %v", err)
	}

	return nil

}
