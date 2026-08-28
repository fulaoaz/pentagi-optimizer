package models

import (
	"strings"

	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/locale"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"

	tea "github.com/charmbracelet/bubbletea"
)

// MockFormModel represents a placeholder screen for not-yet-migrated screens
type MockFormModel struct {
	*BaseScreen
	name        string
	title       string
	description string
}

// NewMockFormModel creates a new mock form model
func NewMockFormModel(
	c controller.Controller, s styles.Styles, w window.Window,
	name, title, description string,
) *MockFormModel {
	m := &MockFormModel{
		name:        name,
		title:       title,
		description: description,
	}

	// create base screen with this model as handler (no list handler needed)
	m.BaseScreen = NewBaseScreen(c, s, w, m, nil)

	return m
}

// BaseScreenHandler interface implementation

func (m *MockFormModel) BuildForm() tea.Cmd {
	// No form fields for mock screen
	m.SetFormFields([]FormField{})
	return nil
}

func (m *MockFormModel) GetFormTitle() string {
	return m.title
}

func (m *MockFormModel) GetFormDescription() string {
	return m.description
}

func (m *MockFormModel) GetFormName() string {
	return m.name
}

func (m *MockFormModel) GetFormSummary() string {
	return ""
}

func (m *MockFormModel) GetFormOverview() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(m.title))
	sections = append(sections, "")
	sections = append(sections, m.GetStyles().Paragraph.Render(m.description))
	sections = append(sections, "")

	sections = append(sections, m.GetStyles().Warning.Render(locale.MockScreenUnderDevelopment))
	sections = append(sections, "")
	sections = append(sections, locale.MockScreenAvailableLater)
	sections = append(sections, "")
	sections = append(sections, locale.MockScreenBackInstruction)

	return strings.Join(sections, "\n")
}

func (m *MockFormModel) GetCurrentConfiguration() string {
	return m.GetStyles().Info.Render(locale.MockScreenMigrationPending)
}

func (m *MockFormModel) IsConfigured() bool {
	return false // mock screens are never configured
}

func (m *MockFormModel) GetHelpContent() string {
	var sections []string

	sections = append(sections, m.GetStyles().Subtitle.Render(locale.MockScreenDevelopmentNotice))
	sections = append(sections, "")
	sections = append(sections, locale.MockScreenMigrationDescription)
	sections = append(sections, "")
	sections = append(sections, locale.MockScreenExpectedFeatures)
	sections = append(sections, locale.MockScreenModernForm)
	sections = append(sections, locale.MockScreenImprovedValidation)
	sections = append(sections, locale.MockScreenEnhancedUserExperience)
	sections = append(sections, "")
	sections = append(sections, locale.MockScreenCheckLater)

	return strings.Join(sections, "\n")
}

func (m *MockFormModel) HandleSave() error {
	// No save functionality for mock screen
	return nil
}

func (m *MockFormModel) HandleReset() {
	// No reset functionality for mock screen
}

func (m *MockFormModel) OnFieldChanged(fieldIndex int, oldValue, newValue string) {
	// No field change handling for mock screen
}

func (m *MockFormModel) GetFormFields() []FormField {
	return []FormField{}
}

func (m *MockFormModel) SetFormFields(fields []FormField) {
	// Ignore field setting for mock screen
}

// Override GetFormHotKeys to show only basic navigation
func (m *MockFormModel) GetFormHotKeys() []string {
	return []string{"enter"}
}

// Update method - handle only basic navigation
func (m *MockFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Return to previous screen
			return m, func() tea.Msg {
				return NavigationMsg{GoBack: true}
			}
		}
	}

	// delegate to base screen for common handling
	return m, m.BaseScreen.Update(msg)
}

// Compile-time interface validation
var _ BaseScreenModel = (*MockFormModel)(nil)
var _ BaseScreenHandler = (*MockFormModel)(nil)
