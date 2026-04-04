package tui

import (
	"context"

	"github.com/RecallKit/recallkit/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// cmdPing checks Ollama is reachable.
func cmdPing(client *engine.OllamaClient) tea.Cmd {
	return func() tea.Msg {
		return pingResultMsg{err: client.Ping()}
	}
}

// cmdStartStream launches an Ollama streaming request and returns a
// nextTokenFnMsg so the update loop can begin pulling tokens one at a time.
func cmdStartStream(
	client *engine.OllamaClient,
	model string,
	history []engine.Message,
) tea.Cmd {
	return func() tea.Msg {
		tokenCh := make(chan string, 64)
		errCh := make(chan error, 1)
		client.StreamChat(context.Background(), model, history, tokenCh, errCh)
		return nextTokenFnMsg(makeTokenPuller(tokenCh, errCh))
	}
}

// makeTokenPuller returns a Cmd that reads exactly one item from the stream
// channels. On success it returns a tokenMsg that embeds the *next* puller,
// so the update loop can keep scheduling without holding any shared state.
func makeTokenPuller(tokenCh <-chan string, errCh <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case token, ok := <-tokenCh:
			if !ok {
				select {
				case err := <-errCh:
					return streamErrMsg{err: err}
				default:
					return streamDoneMsg{}
				}
			}
			return tokenMsg{
				token:    token,
				nextPull: makeTokenPuller(tokenCh, errCh),
			}
		case err := <-errCh:
			return streamErrMsg{err: err}
		}
	}
}
