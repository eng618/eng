package config

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eng618/eng/utils/config"
	"github.com/eng618/eng/utils/log"
)

// UI states
const (
	stateList       = iota // Main list view
	stateConfirm           // Confirmation dialog
	stateDetailView        // Detailed view of a proxy
)

var (
	// Styles for the list and title
	titleStyle = lipgloss.NewStyle().
			MarginLeft(2).
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"}).
			MarginTop(1)

	statusMessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
			Bold(true).
			MarginTop(1).
			MarginLeft(2)

	// Confirmation dialog styles
	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	dialogTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Align(lipgloss.Center)

	confirmButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#F25D94")).
			Padding(0, 3).
			Margin(1, 3, 0, 0).
			Bold(true)

	cancelButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#3c3836")).
			Padding(0, 3).
			Margin(1, 0, 0, 0).
			Bold(true)

	// Detailed view styles
	detailTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#874BFD")).
			Padding(0, 1).
			Bold(true).
			Width(76).
			Align(lipgloss.Center)

	detailLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true).
			PaddingLeft(2)

	detailValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"}).
			PaddingLeft(2)

	detailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			Width(80)
)

// Implement a custom item for the list
type proxyItem struct {
	title   string
	value   string
	noProxy string
	enabled bool
	index   int
}

func (i proxyItem) Title() string {
	status := " "
	if i.enabled {
		status = "✓"
	}
	return fmt.Sprintf("[%s] %s", status, i.title)
}

func (i proxyItem) Description() string {
	desc := i.value
	if i.noProxy != "" {
		desc += fmt.Sprintf(" (No Proxy: %s)", i.noProxy)
	}
	return desc
}

func (i proxyItem) FilterValue() string {
	return i.title
}

// Custom keymap for proxy actions
type proxyKeymap struct {
	enable      key.Binding
	disable     key.Binding
	add         key.Binding
	edit        key.Binding
	delete      key.Binding
	viewDetails key.Binding
}

// ProxyUIModel represents the state of the proxy management UI
type ProxyUIModel struct {
	list            list.Model
	proxies         []config.ProxyConfig
	statusMessage   string
	showingHelp     bool
	quitting        bool
	selectedForEdit int // Index of the proxy selected for editing
	
	// UI state management
	currentState   int            // Current UI state (list, confirm, detail)
	confirmMessage string         // Message to show in confirmation dialog
	confirmAction  func() error   // Action to perform if confirmed
	cancelAction   func()         // Action to perform if canceled
	detailIndex    int            // Index of proxy to show details for
	width          int            // Current terminal width
	height         int            // Current terminal height
}

// Initialize the model
func NewProxyUIModel() ProxyUIModel {
	// Custom key bindings
	keys := proxyKeymap{
		enable: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "enable"),
		),
		disable: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "disable"),
		),
		add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		edit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit"),
		),
		delete: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "delete"),
		),
		viewDetails: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "view details"),
		),
	}

	// Create the default delegate with custom keys
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.SetHeight(2)
	delegate.SetSpacing(1)
	
	// Create the list model
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Proxy Configurations"
	l.Styles.Title = titleStyle
	l.Help.ShowAll = true
	
	// Add our custom key bindings to the list
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.enable,
			keys.disable,
			keys.add,
			keys.edit,
			keys.delete,
			keys.viewDetails,
		}
	}
	
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.enable,
			keys.disable,
			keys.add,
			keys.delete,
			keys.viewDetails,
		}
	}
	
	// Load proxies from config (using silent version)
	proxies, _ := uiGetProxyConfigs()
	
	// Create the model
	m := ProxyUIModel{
		list:          l,
		proxies:       proxies,
		statusMessage: "Select a proxy configuration or add a new one",
		showingHelp:   false,
		quitting:      false,
		currentState:  stateList,
		selectedForEdit: -1,
		detailIndex:   -1,
	}
	
	// Update the list with proxy items
	m.updateProxyItems()
	
	return m
}

// DeleteProxy removes a proxy configuration at the specified index
func DeleteProxy(index int, proxies []config.ProxyConfig) error {
	if index < 0 || index >= len(proxies) {
		return fmt.Errorf("proxy index out of range")
	}
	
	// If we're deleting the active proxy, disable all proxies first
	if proxies[index].Enabled {
		if err := config.DisableAllProxies(); err != nil {
			return err
		}
	}
	
	// Remove the proxy from the slice
	newProxies := append(proxies[:index], proxies[index+1:]...)
	
	// Save the updated proxy configurations
	if err := config.SaveProxyConfigs(newProxies); err != nil {
		return err
	}
	
	return nil
}

// Convert ProxyConfigs to list items
func (m *ProxyUIModel) updateProxyItems() {
	var items []list.Item
	for i, p := range m.proxies {
		items = append(items, proxyItem{
			title:   p.Title,
			value:   p.Value,
			noProxy: p.NoProxy,
			enabled: p.Enabled,
			index:   i,
		})
	}
	m.list.SetItems(items)
}

// Reload proxies from config
func (m *ProxyUIModel) reloadProxies() {
	proxies, _ := uiGetProxyConfigs()
	m.proxies = proxies
	m.updateProxyItems()
}

// Show confirmation dialog
func (m *ProxyUIModel) showConfirmation(message string, action func() error, cancel func()) {
	m.currentState = stateConfirm
	m.confirmMessage = message
	m.confirmAction = action
	m.cancelAction = cancel
}

// Show detailed view for a proxy
func (m *ProxyUIModel) showDetailView(index int) {
	if index >= 0 && index < len(m.proxies) {
		m.detailIndex = index
		m.currentState = stateDetailView
	}
}

// Tea init function
func (m ProxyUIModel) Init() tea.Cmd {
	return nil
}

// Tea update function - handle messages/events
func (m ProxyUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := titleStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		
	case tea.KeyMsg:
		// Handle key events based on current state
		switch m.currentState {
		case stateList:
			// Main list view keys
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c", "q"))):
				m.quitting = true
				return m, tea.Quit
				
			case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
				if i, ok := m.list.SelectedItem().(proxyItem); ok {
					_, err := uiEnableProxy(i.index, m.proxies)
					if err != nil {
						m.statusMessage = fmt.Sprintf("Error: %v", err)
					} else {
						m.statusMessage = fmt.Sprintf("Proxy '%s' enabled", i.title)
						m.reloadProxies()
					}
				}
				
			case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
				// Show confirmation for disabling all proxies
				m.currentState = stateConfirm
				m.confirmMessage = "Are you sure you want to disable all proxies?"
				m.confirmAction = func() error {
					err := uiDisableAllProxies()
					if err == nil {
						m.statusMessage = "All proxies disabled"
						m.reloadProxies()
					}
					return err
				}
				m.cancelAction = func() {
					m.statusMessage = "Disable operation cancelled"
				}
				return m, nil
				
			case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
				m.statusMessage = "Adding new proxy..."
				return m, tea.Quit
				
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if i, ok := m.list.SelectedItem().(proxyItem); ok {
					m.selectedForEdit = i.index
					m.statusMessage = fmt.Sprintf("Editing proxy '%s'...", i.title)
					return m, tea.Quit
				}
				
			case key.Matches(msg, key.NewBinding(key.WithKeys("x"))):
				if i, ok := m.list.SelectedItem().(proxyItem); ok {
					// Show confirmation for deleting a proxy
					proxyIndex := i.index
					proxyTitle := i.title
					
					m.currentState = stateConfirm
					m.confirmMessage = fmt.Sprintf("Are you sure you want to delete proxy '%s'?", proxyTitle)
					m.confirmAction = func() error {
						err := uiDeleteProxy(proxyIndex, m.proxies)
						if err == nil {
							m.statusMessage = fmt.Sprintf("Proxy '%s' deleted", proxyTitle)
							m.reloadProxies()
						}
						return err
					}
					m.cancelAction = func() {
						m.statusMessage = "Delete operation cancelled"
					}
					return m, nil
				}
				
			case key.Matches(msg, key.NewBinding(key.WithKeys("v"))):
				if i, ok := m.list.SelectedItem().(proxyItem); ok {
					if i.index >= 0 && i.index < len(m.proxies) {
						m.detailIndex = i.index
						m.currentState = stateDetailView
						return m, nil
					}
				}
			}
			
		case stateConfirm:
			// Confirmation dialog keys
			switch msg.String() {
			case "y", "Y":
				// Confirmed action
				if m.confirmAction != nil {
					err := m.confirmAction()
					if err != nil {
						m.statusMessage = fmt.Sprintf("Error: %v", err)
					}
				}
				m.currentState = stateList
				return m, nil
				
			case "n", "N", "esc", "q", "ctrl+c":
				// Cancelled action
				if m.cancelAction != nil {
					m.cancelAction()
				}
				m.currentState = stateList
				return m, nil
			}
			
		case stateDetailView:
			// Detailed view keys - any key returns to list view
			m.currentState = stateList
			return m, nil
		}
	}

	// Only pass messages to the list when in list state
	if m.currentState == stateList {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
	
	return m, nil
}

// Render confirmation dialog
func (m ProxyUIModel) renderConfirmationView() string {
	// Center the dialog in the terminal
	dialogWidth := 60
	horizontalPad := max(0, (m.width-dialogWidth)/2)
	verticalPad := max(0, (m.height-10)/2)
	
	// Build vertical padding
	verticalPadding := strings.Repeat("\n", verticalPad)
	
	// Build the dialog content
	dialogContent := m.confirmMessage + "\n\n"
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		confirmButtonStyle.Render("Yes (Y)"),
		cancelButtonStyle.Render("No (N)"),
	)
	dialogContent += buttons
	
	// Style the dialog box
	styledDialog := dialogBoxStyle.Width(dialogWidth).Render(dialogContent)
	
	// Add horizontal padding with spaces
	horizontalPadStr := strings.Repeat(" ", horizontalPad)
	paddedDialog := strings.ReplaceAll(styledDialog, "\n", "\n"+horizontalPadStr)
	paddedDialog = horizontalPadStr + paddedDialog
	
	return verticalPadding + paddedDialog
}

// Render detailed view of a proxy
func (m ProxyUIModel) renderDetailView() string {
	if m.detailIndex < 0 || m.detailIndex >= len(m.proxies) {
		return "No proxy selected for detail view"
	}
	
	proxy := m.proxies[m.detailIndex]
	
	// Build the detail content
	var content strings.Builder
	
	// Title
	content.WriteString(detailTitleStyle.Render("Proxy Details") + "\n\n")
	
	// Fields
	content.WriteString(detailLabelStyle.Render("Title:"))
	content.WriteString("\n" + detailValueStyle.Render(proxy.Title) + "\n\n")
	
	content.WriteString(detailLabelStyle.Render("Proxy URL:"))
	content.WriteString("\n" + detailValueStyle.Render(proxy.Value) + "\n\n")
	
	content.WriteString(detailLabelStyle.Render("Status:"))
	status := "Disabled"
	if proxy.Enabled {
		status = "Enabled"
	}
	content.WriteString("\n" + detailValueStyle.Render(status) + "\n\n")
	
	content.WriteString(detailLabelStyle.Render("Custom No Proxy Settings:"))
	noProxyValue := proxy.NoProxy
	if noProxyValue == "" {
		noProxyValue = "(None - using system defaults)"
	}
	content.WriteString("\n" + detailValueStyle.Render(noProxyValue) + "\n\n")
	
	content.WriteString(detailLabelStyle.Render("Effective No Proxy Values:"))
	effectiveNoProxy := "localhost,127.0.0.1,::1,.local"
	if proxy.NoProxy != "" {
		effectiveNoProxy += "," + proxy.NoProxy
	}
	content.WriteString("\n" + detailValueStyle.Render(effectiveNoProxy) + "\n\n")
	
	content.WriteString("\n" + infoStyle.Render("Press any key to return to list"))
	
	// Style the detail box
	styledDetail := detailBoxStyle.Render(content.String())
	
	// Center in the terminal
	horizontalPad := max(0, (m.width-detailBoxStyle.GetWidth())/2)
	verticalPad := max(0, (m.height-strings.Count(styledDetail, "\n")-10)/2)
	
	// Build vertical padding
	verticalPadding := strings.Repeat("\n", verticalPad)
	
	// Add horizontal padding with spaces
	horizontalPadStr := strings.Repeat(" ", horizontalPad)
	paddedDetail := strings.ReplaceAll(styledDetail, "\n", "\n"+horizontalPadStr)
	paddedDetail = horizontalPadStr + paddedDetail
	
	return verticalPadding + paddedDetail
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Tea view function
func (m ProxyUIModel) View() string {
	if m.quitting {
		return ""
	}
	
	// Render based on current state
	switch m.currentState {
	case stateConfirm:
		return m.renderConfirmationView()
		
	case stateDetailView:
		return m.renderDetailView()
		
	default: // stateList
		// Combine the list view with our status message
		content := strings.Builder{}
		content.WriteString(m.list.View())
		
		if m.statusMessage != "" {
			content.WriteString("\n")
			content.WriteString(statusMessageStyle.Render(m.statusMessage))
		}
		
		return content.String()
	}
}

// Function to start the Bubble Tea UI
func StartProxyUI() {
	model := NewProxyUIModel()
	model.selectedForEdit = -1 // Initialize to -1 (no proxy selected for edit)
	
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Error("Error running proxy UI: %v", err)
	}
	
	// Handle post-UI actions
	if model.statusMessage == "Adding new proxy..." {
		// Add a new proxy configuration (using silent version)
		uiAddOrUpdateProxy()
	} else if model.selectedForEdit >= 0 && model.selectedForEdit < len(model.proxies) {
		// Edit an existing proxy configuration (using silent version)
		uiEditProxy(model.selectedForEdit)
	}
}

// Silent UI-specific versions of proxy operations

// uiGetProxyConfigs gets proxy configs with logging suppressed
func uiGetProxyConfigs() ([]config.ProxyConfig, int) {
	config.SilentMode = true
	defer func() { config.SilentMode = false }()
	return config.GetProxyConfigs()
}

// uiEnableProxy enables a proxy with logging suppressed
func uiEnableProxy(index int, proxies []config.ProxyConfig) ([]config.ProxyConfig, error) {
	config.SilentMode = true
	defer func() { config.SilentMode = false }()
	return config.EnableProxy(index, proxies)
}

// uiDisableAllProxies disables all proxies with logging suppressed
func uiDisableAllProxies() error {
	config.SilentMode = true
	defer func() { config.SilentMode = false }()
	return config.DisableAllProxies()
}

// uiAddOrUpdateProxy adds/updates a proxy with logging suppressed
func uiAddOrUpdateProxy() ([]config.ProxyConfig, int) {
	config.SilentMode = true
	defer func() { config.SilentMode = false }()
	return config.AddOrUpdateProxy()
}

// uiEditProxy edits a proxy with logging suppressed
func uiEditProxy(index int) ([]config.ProxyConfig, int) {
	config.SilentMode = true
	defer func() { config.SilentMode = false }()
	return config.EditProxy(index)
}

// uiDeleteProxy deletes a proxy with logging suppressed
func uiDeleteProxy(index int, proxies []config.ProxyConfig) error {
	config.SilentMode = true
	defer func() { config.SilentMode = false }()
	return DeleteProxy(index, proxies)
}
