/* TODO(anton2920): this is far from being efficient. Serves only demonstration purposes.
 * Things to consider before shipping:
 * - remove []byte("string") to remove allocations;
 * - change os.Stdout.Write([]byte(fmt.Sprintf())) to fmt.Printf() after making sure it does not buffer before output;
 */

/* NOTE(anton2920): useful resources:
 * http://www.linusakesson.net/programming/tty/
 * http://www.braun-home.net/michael/info/misc/VT100_commands.htm
 * https://espterm.github.io/docs/VT100%20escape%20codes.html
 */

package main

import (
	"fmt"
	"os"
)

type CurrencyID int

const (
	USD = iota
	EUR
	GBP
	RUB
)

var CurrencyID2Name = [...]string{
	USD: "USD",
	EUR: "EUR",
	GBP: "GBP",
	RUB: "RUB",
}

var Rates = [...]float64{
	USD: 0.31,
	EUR: 0.28,
	GBP: 0.24,
	RUB: 27.44,
}

const ESC = "\033"

const (
	UpArrow   = ESC + "[A"
	DownArrow = ESC + "[B"

	Enter = "\r"
)

type Color string

const (
	ColorBgGreen = ESC + "[42m"
	ColorBgReset = ESC + "[49m"
)

func BeginColor(color Color) {
	os.Stdout.Write([]byte(color))
}

func EndColor() {
	os.Stdout.Write([]byte(ColorBgReset))
}

func ClearScreen() {
	os.Stdout.Write([]byte(ESC + "[J"))
}

func PrintMenu(items []string, pos int) {
	os.Stdout.Write([]byte("Select currency:\r\n"))

	for i := 0; i < len(items); i++ {
		if i > 0 {
			os.Stdout.Write([]byte("\r\n"))
		}

		if i == pos {
			BeginColor(ColorBgGreen)
		}

		os.Stdout.Write([]byte("\t"))
		os.Stdout.Write([]byte(items[i]))

		if i == pos {
			EndColor()
		}
	}

	os.Stdout.Write([]byte(fmt.Sprintf(ESC+"[%dA\r", len(items))))
}

func PrintExchangeRate(idx int) {
	byn := 1000 * Rates[idx]
	os.Stdout.Write([]byte(fmt.Sprintf("1000 BYN => %f %s\r\n", byn, CurrencyID2Name[idx])))
}

func App() error {
	var pos int

	items := make([]string, len(Rates))
	for i := 0; i < len(Rates); i++ {
		items[i] = fmt.Sprintf("%s (%f)", CurrencyID2Name[i], Rates[i])
	}

	for {
		PrintMenu(items, pos)

		buffer := make([]byte, 10)
		n, err := os.Stdin.Read(buffer)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %v", err)
		}
		buffer = buffer[:n]

		if buffer[0] == 'q' {
			break
		}
		if string(buffer[:len(Enter)]) == Enter {
			PrintExchangeRate(pos)
			break
		}
		switch string(buffer[:len(UpArrow)]) {
		case UpArrow:
			pos--
		case DownArrow:
			pos++
		}
		if pos < 0 {
			pos += len(items)
		} else if pos >= len(items) {
			pos -= len(items)
		}
	}

	ClearScreen()
	return nil
}
