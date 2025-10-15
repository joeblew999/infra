package tui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/mattn/go-runewidth"

	runtimeui "github.com/joeblew999/infra/core/pkg/runtime/ui"
	"github.com/joeblew999/infra/core/pkg/runtime/ui/live"
	"github.com/joeblew999/infra/core/pkg/runtime/ui/render"
)

// session coordinates an interactive Ultraviolet-driven TUI instance.
type session struct {
	term  *uv.Terminal
	tmpl  templateExecutor
	store *live.Store
	page  string
	ctx   context.Context

	snapshot runtimeui.Snapshot

	width  int
	height int

	inputMode    bool
	inputPurpose string
	inputBuffer  []rune
	statusLine   string
}

// templateExecutor captures the subset of template.Template we rely on so the
// session logic remains decoupled from the concrete type and easy to test.
type templateExecutor interface {
	ExecuteTemplate(w io.Writer, name string, data any) error
}

const processActionTimeout = 8 * time.Second

func newSession(ctx context.Context, term *uv.Terminal, tmpl templateExecutor, store *live.Store, page string) *session {
	return &session{
		term:  term,
		tmpl:  tmpl,
		store: store,
		page:  page,
		ctx:   ctx,
	}
}

func (s *session) run(ctx context.Context) error {
	defer s.term.ExitAltScreen()
	s.ctx = ctx

	w, h, err := s.term.GetSize()
	if err != nil {
		return fmt.Errorf("terminal size: %w", err)
	}
	s.width, s.height = w, h
	_ = s.term.Resize(w, h)

	updates, cancel := s.store.Subscribe()
	defer cancel()

	initial, ok := <-updates
	if !ok {
		return nil
	}
	s.snapshot = initial

	if err := s.render(render.NewViewModel("Core Runtime", s.snapshot, s.page, true)); err != nil {
		return err
	}

	events := s.term.Events()
	eventCtx, cancelEvents := context.WithCancel(ctx)
	defer cancelEvents()

	errCh := make(chan error, 1)
	go func() {
		err := s.term.StreamEvents(eventCtx, events)
		if err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case snapshot, ok := <-updates:
			if !ok {
				return nil
			}
			s.snapshot = snapshot
			if err := s.render(render.NewViewModel("Core Runtime", s.snapshot, s.page, true)); err != nil {
				return err
			}
		case err := <-errCh:
			if err != nil {
				return err
			}
			return nil
		case ev, ok := <-events:
			if !ok {
				return nil
			}
			switch tev := ev.(type) {
			case uv.WindowSizeEvent:
				s.width, s.height = tev.Width, tev.Height
				_ = s.term.Resize(s.width, s.height)
				if err := s.render(render.NewViewModel("Core Runtime", s.snapshot, s.page, true)); err != nil {
					return err
				}
			case uv.KeyPressEvent:
				done, err := s.handleKey(tev)
				if err != nil {
					return err
				}
				if err := s.render(render.NewViewModel("Core Runtime", s.snapshot, s.page, true)); err != nil {
					return err
				}
				if done {
					return nil
				}
			}
		}
	}
}

func (s *session) handleKey(ev uv.KeyPressEvent) (bool, error) {
	if s.inputMode {
		switch {
		case ev.MatchStrings("enter"):
			s.handleInputConfirm()
			return false, nil
		case ev.MatchStrings("esc", "escape"):
			s.statusLine = "input cancelled"
			s.inputBuffer = s.inputBuffer[:0]
			s.inputMode = false
			s.inputPurpose = ""
			return false, nil
		case ev.MatchStrings("backspace"):
			if len(s.inputBuffer) > 0 {
				s.inputBuffer = s.inputBuffer[:len(s.inputBuffer)-1]
			}
			return false, nil
		}
		key := ev.Key()
		if key.Text != "" {
			for _, r := range key.Text {
				s.inputBuffer = append(s.inputBuffer, r)
			}
		}
		return false, nil
	}

	if ev.MatchStrings("q", "ctrl+c", "ctrl+d") {
		return true, nil
	}

	if ev.MatchStrings("left", "h") {
		s.navigate(-1)
		return false, nil
	}
	if ev.MatchStrings("right", "l") {
		s.navigate(1)
		return false, nil
	}

	if ev.MatchStrings("s") {
		s.promptScale()
		return false, nil
	}

	if ev.MatchStrings("p") {
		s.performProcessAction("start")
		return false, nil
	}
	if ev.MatchStrings("o") {
		s.performProcessAction("stop")
		return false, nil
	}
	if ev.MatchStrings("r") {
		s.performProcessAction("restart")
		return false, nil
	}

	if ev.MatchStrings("a") {
		s.inputMode = true
		s.inputBuffer = s.inputBuffer[:0]
		s.inputPurpose = "event"
		s.statusLine = "enter message (Enter to submit, Esc to cancel)"
		return false, nil
	}

	if key := ev.Key(); key.Text != "" {
		if index, err := strconv.Atoi(key.Text); err == nil {
			s.jumpTo(index - 1)
			return false, nil
		}
	}

	return false, nil
}

func (s *session) navigate(delta int) {
	nav := s.snapshot.Navigation
	if len(nav) == 0 {
		return
	}
	currentIndex := 0
	for i, item := range nav {
		if item.Route == s.page {
			currentIndex = i
			break
		}
	}
	next := (currentIndex + len(nav) + delta) % len(nav)
	s.page = nav[next].Route
	s.statusLine = fmt.Sprintf("page: %s", s.page)
}

func (s *session) jumpTo(index int) {
	nav := s.snapshot.Navigation
	if index < 0 || index >= len(nav) {
		return
	}
	s.page = nav[index].Route
	s.statusLine = fmt.Sprintf("page: %s", s.page)
}

func (s *session) render(vm render.ViewModel) error {
	s.page = vm.CurrentPage

	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, "tui", vm); err != nil {
		return fmt.Errorf("render tui template: %w", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	lines = append(lines, "", s.instructions(vm))

	if s.inputMode {
		lines = append(lines, fmt.Sprintf("Enter message: %s_", string(s.inputBuffer)))
	} else if s.statusLine != "" {
		lines = append(lines, fmt.Sprintf("Status: %s", s.statusLine))
	}

	if len(lines) > s.height {
		lines = lines[:s.height]
		if len(lines) > 0 {
			lines[len(lines)-1] = "(output truncated -- increase terminal height)"
		}
	}

	_ = s.term.Resize(s.width, s.height)
	s.term.Fill(nil)

	for y, line := range lines {
		if y >= s.height {
			break
		}
		x := 0
		for _, r := range line {
			width := runewidth.RuneWidth(r)
			if width <= 0 {
				width = 1
			}
			cell := uv.Cell{Content: string(r), Width: width}
			s.term.SetCell(x, y, &cell)
			x += width
		}
		for ; x < s.width; x++ {
			s.term.SetCell(x, y, nil)
		}
	}

	return s.term.Display()
}

func (s *session) instructions(vm render.ViewModel) string {
	builder := strings.Builder{}
	builder.WriteString("[Left/Right] navigate  [a] add event  [q] quit")
	if vm.Live {
		if vm.CurrentProcess != nil && vm.CurrentProcess.Scalable {
			builder.WriteString("  [s] scale")
		}
		builder.WriteString("  [p] start  [o] stop  [r] restart")
	}
	if count := len(vm.Navigation); count > 0 {
		builder.WriteString("  [1-")
		builder.WriteString(strconv.Itoa(count))
		builder.WriteString("] jump")
	}
	builder.WriteString("  page=")
	builder.WriteString(vm.CurrentPage)
	builder.WriteString("  events:")
	builder.WriteString(strconv.Itoa(len(vm.Snapshot.Events)))
	builder.WriteString("  generated=")
	builder.WriteString(vm.Generated)
	return builder.String()
}

func (s *session) currentProcessDetail() (string, runtimeui.ProcessDetail, bool) {
	if s.snapshot.Processes == nil {
		return "", runtimeui.ProcessDetail{}, false
	}
	if !strings.HasPrefix(s.page, "service/") {
		return "", runtimeui.ProcessDetail{}, false
	}
	id := strings.TrimPrefix(s.page, "service/")
	detail, ok := s.snapshot.Processes[id]
	if !ok {
		return "", runtimeui.ProcessDetail{}, false
	}
	if detail.Runtime.ID == "" {
		detail.Runtime.ID = id
	}
	return id, detail, true
}

func (s *session) performProcessAction(action string) {
	if s.store == nil {
		s.statusLine = "live mode required for process control"
		return
	}
	_, detail, ok := s.currentProcessDetail()
	if !ok {
		s.statusLine = "open a service detail page to control that process"
		return
	}
	name := strings.TrimSpace(detail.Runtime.ID)
	if name == "" {
		s.statusLine = "process identifier unavailable"
		return
	}
	port := s.store.ComposePort()
	if port <= 0 {
		s.statusLine = "process compose port unavailable"
		return
	}
	baseCtx := s.ctx
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	ctx, cancel := context.WithTimeout(baseCtx, processActionTimeout)
	defer cancel()

	var err error
	switch action {
	case "start":
		err = s.store.StartProcess(ctx, port, name)
	case "stop":
		err = s.store.StopProcess(ctx, port, name)
	case "restart":
		err = s.store.RestartProcess(ctx, port, name)
	default:
		err = fmt.Errorf("unknown action %s", action)
	}
	if err != nil {
		s.statusLine = fmt.Sprintf("%s %s failed: %v", action, name, err)
		return
	}
	s.statusLine = fmt.Sprintf("%s %s requested", action, name)
}

func (s *session) promptScale() {
	_, detail, ok := s.currentProcessDetail()
	if !ok {
		s.statusLine = "open a service detail page to scale"
		return
	}
	if !detail.Scalable {
		s.statusLine = fmt.Sprintf("%s can only be scaled via infrastructure", detail.Runtime.ID)
		return
	}
	value := fmt.Sprintf("%d", detail.Runtime.Replicas)
	s.inputMode = true
	s.inputPurpose = "scale"
	s.inputBuffer = append([]rune{}, []rune(value)...)
	s.statusLine = fmt.Sprintf("scale %s to count (Enter to submit)", detail.Runtime.ID)
}

func (s *session) handleInputConfirm() {
	value := strings.TrimSpace(string(s.inputBuffer))
	s.inputBuffer = s.inputBuffer[:0]
	purpose := s.inputPurpose
	s.inputPurpose = ""
	s.inputMode = false

	switch purpose {
	case "event":
		if value == "" {
			value = "manual event (tui)"
		}
		s.store.AppendEvent(value)
		s.statusLine = fmt.Sprintf("event submitted: %s", value)
	case "scale":
		s.performScale(value)
	default:
		s.statusLine = "input submitted"
	}
}

func (s *session) performScale(input string) {
	count, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || count < 0 {
		s.statusLine = fmt.Sprintf("invalid scale value %q", input)
		return
	}
	if s.store == nil {
		s.statusLine = "live mode required for process control"
		return
	}
	_, detail, ok := s.currentProcessDetail()
	if !ok {
		s.statusLine = "open a service detail page to scale"
		return
	}
	if !detail.Scalable {
		s.statusLine = fmt.Sprintf("%s cannot be scaled locally", detail.Runtime.ID)
		return
	}
	name := strings.TrimSpace(detail.Runtime.ID)
	if name == "" {
		s.statusLine = "process identifier unavailable"
		return
	}
	port := s.store.ComposePort()
	if port <= 0 {
		s.statusLine = "process compose port unavailable"
		return
	}
	baseCtx := s.ctx
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	ctx, cancel := context.WithTimeout(baseCtx, processActionTimeout)
	defer cancel()
	if err := s.store.ScaleProcess(ctx, port, name, count); err != nil {
		s.statusLine = fmt.Sprintf("scale %s failed: %v", name, err)
		return
	}
	s.statusLine = fmt.Sprintf("scale %s to %d requested", name, count)
}
