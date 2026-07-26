package spinner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yarlson/pin"
)

type Spinner struct {
	pin       *pin.Pin
	startTime time.Time

	mu      sync.RWMutex
	message string

	stopTimer chan struct{}
	timerDone chan struct{}

	cancelContext context.CancelFunc
	cancelSpinner context.CancelFunc

	stopOnce sync.Once
}

// New creates and starts a new spinner.
func New(prefix string, message string) *Spinner {
	p := pin.New(
		fmt.Sprintf("0.00s %s", message),
		pin.WithPrefix(prefix),
		pin.WithSeparator(":"),
		pin.WithSeparatorColor(pin.ColorWhite),
		pin.WithDoneSymbol('✔'),
		pin.WithDoneSymbolColor(pin.ColorGreen),
		pin.WithSpinnerColor(pin.ColorBlue),
		pin.WithTextColor(pin.ColorCyan),
	)

	ctx, cancelContext := context.WithCancel(context.Background())
	cancelSpinner := p.Start(ctx)

	s := &Spinner{
		pin:           p,
		startTime:     time.Now(),
		message:       message,
		stopTimer:     make(chan struct{}),
		timerDone:     make(chan struct{}),
		cancelContext: cancelContext,
		cancelSpinner: cancelSpinner,
	}

	go s.runTimer()

	return s
}

func (s *Spinner) runTimer() {
	defer close(s.timerDone)

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopTimer:
			return

		case <-ticker.C:
			s.mu.RLock()
			message := s.message
			s.mu.RUnlock()

			elapsed := time.Since(s.startTime).Seconds()

			s.pin.UpdateMessage(
				fmt.Sprintf("%.2fs %s", elapsed, message),
			)
		}
	}
}

// Update changes the text shown next to the spinner.
func (s *Spinner) Update(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// Stop finishes the spinner with a successful message.
func (s *Spinner) Stop(message string) {
	s.stopOnce.Do(func() {
		close(s.stopTimer)
		<-s.timerDone

		duration := time.Since(s.startTime).Seconds()

		s.pin.Stop(
			fmt.Sprintf("%s in %.2fs", message, duration),
		)

		s.cancelSpinner()
		s.cancelContext()
	})
}
