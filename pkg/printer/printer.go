package printer

import (
	"fmt"

	"httplogmonitor/pkg/config"
)

// Printer sends all the messages to STDOUT
type Printer struct {
	config *config.Config
}

// New gives a new Printer instance
func New(cfg *config.Config) *Printer {
	return &Printer{
		config: cfg,
	}
}

// Start sends all the messages received on the passed channel to STDOUT
// if the incoming message is the verbose more and the mode is not on - do nothing
func (p *Printer) Start(fmsgCh <-chan Formatter) {
	for fm := range fmsgCh {
		if fm.Verbose() && !p.config.Verbose {
			continue
		}
		printFormattedMessage(fm)
	}
}

func printFormattedMessage(msg Formatter) {
	fmt.Println(msg.Format())
}
