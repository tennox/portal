package receiverui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
	"www.github.com/ZinoKader/portal/tools"
	"www.github.com/ZinoKader/portal/ui"
)

type uiState int

// ui state flows from the top down
const (
	showEstablishing uiState = iota
	showReceivingProgress
	showFinished
	showError
)

type receiverUIModel struct {
	state                   uiState
	receivedFiles           []string
	payloadSize             int64
	decompressedPayloadSize int64
	spinner                 spinner.Model
	progressBar             progress.Model
	errorMessage            string
}

func NewReceiverUI() *tea.Program {
	s := spinner.NewModel()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.SPINNER_COLOR))
	m := receiverUIModel{
		spinner:     s,
		progressBar: ui.ProgressBar,
	}
	return tea.NewProgram(m)
}

func (receiverUIModel) Init() tea.Cmd {
	return spinner.Tick
}

func (m receiverUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case ui.FileInfoMsg:
		m.state = showReceivingProgress
		m.payloadSize = msg.Bytes
		return m, nil

	case ui.ProgressMsg:
		m.state = showReceivingProgress
		cmd := m.progressBar.SetPercent(float64(msg.Progress))
		return m, cmd

	case ui.FinishedMsg:
		m.state = showFinished
		m.receivedFiles = msg.Files
		m.decompressedPayloadSize = msg.PayloadSize
		cmd := m.progressBar.SetPercent(1.0)
		return m, cmd

	case ui.ErrorMsg:
		m.state = showError
		m.errorMessage = msg.Message
		return m, nil

	case tea.KeyMsg:
		if tools.Contains(ui.QuitKeys, strings.ToLower(msg.String())) {
			return m, tea.Quit
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.progressBar.Width = msg.Width - 2*ui.PADDING - 4
		if m.progressBar.Width > ui.MAX_WIDTH {
			m.progressBar.Width = ui.MAX_WIDTH
		}
		return m, nil

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progressBar.Update(msg)
		m.progressBar = progressModel.(progress.Model)
		return m, cmd

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m receiverUIModel) View() string {

	switch m.state {

	case showEstablishing:
		return "\n" +
			ui.PadText + ui.InfoStyle(fmt.Sprintf("%s Establishing connection with sender", m.spinner.View())) + "\n\n"

	case showReceivingProgress:
		payloadSize := ui.BoldText(tools.ByteCountSI(m.payloadSize))
		receivingText := fmt.Sprintf("Receiving files (total size %s)", payloadSize)
		return "\n" +
			ui.PadText + ui.InfoStyle(receivingText) + "\n\n" +
			ui.PadText + m.progressBar.View() + "\n\n" +
			ui.PadText + ui.QuitCommandsHelpText + "\n\n"

	case showFinished:
		payloadSize := ui.BoldText(tools.ByteCountSI(m.payloadSize))
		indentedWrappedFiles := indent.String(fmt.Sprintf("Received: %s", wordwrap.String(ui.ItalicText(ui.TopLevelFilesText(m.receivedFiles)), ui.MAX_WIDTH)), ui.PADDING)
		finishedText := fmt.Sprintf("File transfer completed! Received %d files (%s decompressed)\n\n%s", len(m.receivedFiles), payloadSize, indentedWrappedFiles)
		return "\n" +
			ui.PadText + ui.InfoStyle(finishedText) + "\n\n" +
			ui.PadText + m.progressBar.View() + "\n\n" +
			ui.PadText + ui.QuitCommandsHelpText + "\n\n"

	case showError:
		return m.errorMessage

	default:
		return ""
	}
}
