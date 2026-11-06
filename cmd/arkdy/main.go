package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"unicode"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/utkkat/arkdy/internal/core"
	"github.com/utkkat/arkdy/internal/transport"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, input io.ReadCloser, output, errorOutput io.Writer) error {
	defer input.Close()
	flags := flag.NewFlagSet("arkdy", flag.ContinueOnError)
	flags.SetOutput(errorOutput)
	listen := flags.String("listen", "/ip4/127.0.0.1/tcp/0", "адрес прослушивания libp2p")
	initialPeer := flags.String("peer", "", "полный адрес собеседника /ip4/.../tcp/.../p2p/...")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("неожиданные аргументы: %s", strings.Join(flags.Args(), " "))
	}
	client, err := core.NewClient(ctx, transport.HostConfig{ListenAddrs: []string{*listen}})
	if err != nil {
		return err
	}
	defer client.Close()
	fmt.Fprintln(output, "PeerID:", client.ID())
	printAddresses(output, client)
	fmt.Fprintln(output, "Локальный чат: ключи и переписка сохраняются только до выхода.")
	printHelp(output)
	var target peer.ID
	known := make(map[peer.ID]bool)
	if *initialPeer != "" {
		target, err = client.Connect(ctx, *initialPeer)
		if err != nil {
			return err
		}
		known[target] = true
		fmt.Fprintln(output, "Собеседник:", target)
	}
	readCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	lines := make(chan string)
	readErrors := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(input)
		scanner.Buffer(make([]byte, 4096), 64*1024)
		defer func() {
			readErrors <- scanner.Err()
			close(lines)
		}()
		for scanner.Scan() {
			select {
			case lines <- scanner.Text():
			case <-readCtx.Done():
				return
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-client.Messages():
			known[msg.From] = true
			if target == "" {
				target = msg.From
				fmt.Fprintln(output, "Собеседник:", target)
			}
			fmt.Fprintf(output, "[%s] %s\n", msg.From, terminalText(msg.Text))
		case line, ok := <-lines:
			if !ok {
				return <-readErrors
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			if strings.HasPrefix(line, "/") {
				command, arg, _ := strings.Cut(line, " ")
				arg = strings.TrimSpace(arg)
				switch command {
				case "/quit", "/exit":
					return nil
				case "/help":
					printHelp(output)
				case "/addr":
					printAddresses(output, client)
				case "/peers":
					for id := range known {
						marker := ""
						if id == target {
							marker = " (выбран)"
						}
						fmt.Fprintln(output, id.String()+marker)
					}
				case "/connect":
					id, err := client.Connect(ctx, arg)
					if err != nil {
						fmt.Fprintln(errorOutput, "Ошибка подключения:", err)
						continue
					}
					target = id
					known[id] = true
					fmt.Fprintln(output, "Собеседник:", id)
				case "/to":
					id, err := peer.Decode(arg)
					if err != nil || !known[id] {
						fmt.Fprintln(errorOutput, "Неизвестный PeerID; используйте /peers или /connect.")
						continue
					}
					target = id
					fmt.Fprintln(output, "Собеседник:", id)
				default:
					fmt.Fprintln(errorOutput, "Неизвестная команда. /help — список команд.")
				}
				continue
			}
			if target == "" {
				fmt.Fprintln(errorOutput, "Сначала выполните /connect <адрес собеседника>.")
				continue
			}
			if err := client.Send(ctx, target, line); err != nil {
				fmt.Fprintln(errorOutput, "Ошибка отправки (доставка не подтверждена):", err)
				continue
			}
			fmt.Fprintln(output, "[доставлено]", terminalText(line))
		}
	}
}

func printAddresses(output io.Writer, client *core.Client) {
	for _, address := range client.Addresses() {
		fmt.Fprintln(output, "Адрес:", address)
	}
}

func printHelp(output io.Writer) {
	fmt.Fprintln(output, "/connect <адрес> — подключиться; /to <PeerID> — выбрать собеседника")
	fmt.Fprintln(output, "/addr — свой адрес; /peers — собеседники; /help — помощь; /quit — выход")
	fmt.Fprintln(output, "Введите текст и нажмите Enter, чтобы отправить сообщение.")
}

func terminalText(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return '�'
		}
		return r
	}, text)
}
