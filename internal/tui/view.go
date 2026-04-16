package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the entire TUI from current model state.
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing RecallKit…\n"
	}

	var sb strings.Builder

	// ── Header ───────────────────────────────────────────────────────────────
	header := headerStyle.Width(m.width).Render(
		" ◈ RecallKit  ·  " + m.sess.Name + "  ·  " + m.sess.Model,
	)
	sb.WriteString(header + "\n")

	// ── Chat viewport ────────────────────────────────────────────────────────
	sb.WriteString(m.viewport.View())
	sb.WriteString("\n")

	// ── Status line ──────────────────────────────────────────────────────────
	statusLine := m.renderStatus()
	sb.WriteString(statusLine + "\n")

	// ── Input box ────────────────────────────────────────────────────────────
	inputBox := inputBorderStyle.Width(m.width - 2).Render(m.input.View())
	sb.WriteString(inputBox)

	return sb.String()
}

// renderHistory builds the full chat history string shown in the viewport.
func (m Model) renderHistory() string {
	var sb strings.Builder

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			label := userStyle.Render("  You")
			sb.WriteString(label + "\n")
			sb.WriteString(indent(msg.Content, "  ") + "\n\n")

		case "assistant":
			label := assistantStyle.Render("  ◈ RecallKit")
			sb.WriteString(label + "\n")
			sb.WriteString(indent(msg.Content, "  ") + "\n\n")
		}
	}

	// Show in-progress streamed content with a cursor
	if m.streamBuf != "" {
		label := assistantStyle.Render("  ◈ RecallKit")
		sb.WriteString(label + "\n")
		sb.WriteString(indent(m.streamBuf, "  "))
		sb.WriteString(dimStyle.Render("█"))
		sb.WriteString("\n\n")
	}

	if m.status == statusThinking {
		sb.WriteString(dimStyle.Render("  ◈ thinking…") + "\n")
	}

	if m.status == statusError && m.err != nil {
		sb.WriteString(errorStyle.Render(fmt.Sprintf("  ✗ Error: %v", m.err)) + "\n")
	}

	return sb.String()
}

// renderStatus returns the one-line status bar shown above the input.
func (m Model) renderStatus() string {
	var parts []string

	switch m.status {
	case statusConnecting:
		parts = append(parts, dimStyle.Render("connecting to Ollama…"))
	case statusIdle:
		parts = append(parts, dimStyle.Render("Enter to send · Alt+Enter for newline · Ctrl+C to quit"))
	case statusThinking:
		parts = append(parts, dimStyle.Render("waiting for response…"))
	case statusStreaming:
		parts = append(parts, dimStyle.Render("streaming…"))
	case statusError:
		parts = append(parts, errorStyle.Render("error — see above"))
	}

	line := strings.Join(parts, "  ")
	// Pad to full width
	pad := m.width - lipgloss.Width(line)
	if pad > 0 {
		line += strings.Repeat(" ", pad)
	}
	return dimStyle.Render(line)
}

// indent prefixes every line of s with the given prefix string.
func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}
