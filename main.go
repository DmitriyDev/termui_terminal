package main

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	ssh "github.com/DmitriyDev/golibs/ssh/client"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
)

const rootID = "root"

const ServerIP = "xxx.xxx.xxx.xxx"
const ServerUser = "user"
const ServerPassword = "password"
const ServerPort = 22

func main() {

	connection := GetClient().Connect(ssh.CERT_PASSWORD)
	defer connection.Close()

	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	updateText := make(chan string)

	input, err := textinput.New(
		textinput.Label("Command:", cell.FgColor(cell.ColorNumber(33))),
		textinput.MaxWidthCells(100),
		textinput.Border(linestyle.Light),
		textinput.PlaceHolder("Enter command"),
	)
	if err != nil {
		panic(err)
	}

	txt, err := text.New()
	go writeResults(ctx, txt, updateText)

	execB, err := button.New("Execute", func() error {

		cmdString := input.ReadAndClear()
		updateText <- "Start : " + cmdString + "\n"
		err, out := connection.RunCmd(cmdString)
		if err != nil {
			updateText <- err.Error()
			return nil
		}

		out = strings.ReplaceAll(out, "\r", "")
		out = strings.TrimFunc(out, func(r rune) bool {
			return !unicode.IsGraphic(r)
		})

		updateText <- "Done -- \n" + out

		return nil
	},
		button.GlobalKey(keyboard.KeyEnter),
		button.FillColor(cell.ColorNumber(120)),
	)
	if err != nil {
		panic(err)
	}

	quitB, err := button.New("Quit", func() error {
		cancel()
		return nil
	},
		button.FillColor(cell.ColorNumber(212)),
		button.Key(keyboard.KeyEnter),
	)
	if err != nil {
		panic(err)
	}

	title := fmt.Sprintf("%s@%s:%d", ServerUser, ServerIP, ServerPort)

	c, err := container.New(
		t,
		container.Border(linestyle.Double),
		container.BorderTitle(title),
		container.SplitVertical(
			container.Left(
				container.Border(linestyle.Light),

				container.SplitHorizontal(
					container.Top(
						container.PlaceWidget(input),
					),
					container.Bottom(
						container.SplitVertical(
							container.Left(container.PlaceWidget(execB)),
							container.Right(container.PlaceWidget(quitB)),
						),
					),
				),
			),
			container.Right(
				container.Border(linestyle.Light),
				// container.BorderTitle("Result"),
				container.PlaceWidget(txt),
			),
		),
	)
	if err != nil {
		panic(err)
	}

	if err := termdash.Run(ctx, t, c, termdash.RedrawInterval(500*time.Millisecond)); err != nil {
		panic(err)
	}

}

func GetClient() *ssh.SSHClient {
	return &ssh.SSHClient{
		Ip:   ServerIP,
		User: ServerUser,
		Port: ServerPort,
		Cert: ServerPassword,
	}
}

func writeResults(ctx context.Context, t *text.Text, updateText <-chan string) {
	for {
		select {
		case tt := <-updateText:
			if err := t.Write(fmt.Sprintf("%s\n", tt)); err != nil {
				panic(err)
			}

		case <-ctx.Done():
			return
		}
	}
}
