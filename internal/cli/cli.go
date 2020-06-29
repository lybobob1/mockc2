package cli

import (
	"io"
	"strings"

	"github.com/chzyer/readline"
	"megaman.genesis.local/sknight/mockc2/internal/log"
	"megaman.genesis.local/sknight/mockc2/pkg/protocol"
)

type shellMenu int

const (
	main shellMenu = iota
	agent
)

// A Shell represents an interactive interface to the mock C2 code.
type Shell struct {
	rl             *readline.Instance
	mainCompleter  *readline.PrefixCompleter
	agentCompleter *readline.PrefixCompleter
	menu           shellMenu
}

func (s *Shell) initCompleters() {
	s.mainCompleter = readline.NewPrefixCompleter(
		readline.PcItem("debug",
			readline.PcItem("on"),
			readline.PcItem("off"),
		),
		readline.PcItem("exit"),
		readline.PcItem("help"),
		readline.PcItem("interact"),
		readline.PcItem("listener",
			readline.PcItem("start",
				readline.PcItemDynamic(protocolNames()),
			),
			readline.PcItem("stop"),
		),
		readline.PcItem("list"),
		readline.PcItem("version"),
	)

	s.agentCompleter = readline.NewPrefixCompleter(
		readline.PcItem("exec"),
		readline.PcItem("exit"),
		readline.PcItem("help"),
		readline.PcItem("download"),
		readline.PcItem("main"),
		readline.PcItem("upload"),
	)
}

func protocolNames() func(string) []string {
	return func(line string) []string {
		return protocol.Names()
	}
}

func (s *Shell) completer() *readline.PrefixCompleter {
	switch s.menu {
	default:
		fallthrough
	case main:
		return s.mainCompleter
	case agent:
		return s.agentCompleter
	}
}

func (s *Shell) initReadline() {
	s.initCompleters()

	l, err := readline.NewEx(&readline.Config{
		Prompt:              "mockc2> ",
		HistoryFile:         "/tmp/mockc2.tmp",
		HistorySearchFold:   true,
		AutoComplete:        s.completer(),
		FuncFilterInputRune: filterInput,
	})

	if err != nil {
		panic(err)
	}

	s.rl = l
	s.setMenu(main)
}

func (s *Shell) prompt() string {
	switch s.menu {
	default:
		fallthrough
	case main:
		return "mockc2> "
	case agent:
		return "agent[1]> "
	}
}

func (s *Shell) setMenu(menu shellMenu) {
	s.menu = menu
	s.rl.Config.AutoComplete = s.completer()
	s.rl.SetPrompt(s.prompt())
}

// Run starts the interactive shell running.
func (s *Shell) Run() {
	s.initReadline()
	defer s.rl.Close()

	for {
		line, err := s.rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		cmd := strings.Fields(line)

		if len(cmd) > 0 {
			switch s.menu {
			case main:
				s.mainMenuHandler(cmd)
			case agent:
				s.agentMenuHandler(cmd)
			}
		}
	}
}

func (s *Shell) mainMenuHandler(cmd []string) {
	switch cmd[0] {
	case "debug":
		debugCommand(cmd)
	case "exit", "quit":
		exitCommand(cmd)
	case "help", "?":
		mainMenuCommand(cmd)
	case "interact":
		s.setMenu(agent)
	case "list":
		listCommand(cmd)
	case "listener":
		listenerCommand(cmd)
	case "version":
		versionCommand(cmd)
	default:
		log.Warn("Invalid command")
	}
}

func (s *Shell) agentMenuHandler(cmd []string) {
	switch cmd[0] {
	case "exit", "quit":
		exitCommand(cmd)
	case "help", "?":
		agentMenuCommand(cmd)
	case "main":
		s.setMenu(main)
	default:
		log.Warn("Invalid command")
	}
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}