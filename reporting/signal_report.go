package reporting

import (
	"fmt"
	"go-backtesting/strategy"
	"os"
	"text/tabwriter"
	"time"
)

// PrintAllSignals prints a table of all entry signals.
func PrintAllSignals(signals []strategy.EntrySignal) {
	if len(signals) == 0 {
		fmt.Println("No signals were generated.")
		return
	}

	fmt.Println("\n--- All Generated Signals ---")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Time\tPrice\tDirection\t")
	fmt.Fprintln(w, "----\t-----\t---------\t")

	for _, s := range signals {
		fmt.Fprintf(w, "%s\t%.2f\t%s\t\n", s.Time.Format(time.RFC3339), s.Price, s.Direction)
	}
	w.Flush()
}
