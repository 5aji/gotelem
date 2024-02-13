package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/kschamplin/gotelem/internal/db"
	"github.com/kschamplin/gotelem/internal/logparsers"
	"github.com/kschamplin/gotelem/skylab"
	"github.com/urfave/cli/v2"
)

var parsersString string

func init() {
	subCmds = append(subCmds, clientCmd)
	parsersString = func() string {
		// create a string like "'telem', 'candump', 'anotherparser'"
		keys := make([]string, len(logparsers.ParsersMap))
		i := 0
		for k := range logparsers.ParsersMap {
			keys[i] = k
			i++
		}
		s := strings.Join(keys, "', '")
		return "'" + s + "'"
	}()
}

var importCmd = &cli.Command{
	Name:      "import",
	Aliases:   []string{"i"},
	Usage:     "import a log file into a database",
	ArgsUsage: "[log file]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "format",
			Aliases: []string{"f"},
			Usage:   "the format of the log file. One of " + parsersString,
			Value:   "telem",
		},
		&cli.PathFlag{
			Name:    "database",
			Aliases: []string{"d", "db"},
			Usage:   "the path of the database",
			Value:   "./gotelem.db",
		},
		&cli.UintFlag{
			Name:  "batch-size",
			Usage: "the maximum size of each SQL transaction",
			Value: 800,
		},
	},
	Action: importAction,
}

// importAction peforms a file import to the database. It can use any of the parsers provided
// by logparsers. Adding new parsers there will work.
func importAction(ctx *cli.Context) error {
	path := ctx.Args().Get(0)
	if path == "" {
		fmt.Println("missing log file!")
		cli.ShowAppHelpAndExit(ctx, -1)
	}
	fstream, err := os.Open(path)
	if err != nil {
		return err
	}
	fReader := bufio.NewReader(fstream)

	pfun, ok := logparsers.ParsersMap[ctx.String("format")]
	if !ok {
		fmt.Println("invalid format provided: must be one of " + parsersString)
		cli.ShowAppHelpAndExit(ctx, -1)
	}

	dbPath := ctx.Path("database")
	db, err := db.OpenTelemDb(dbPath)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	// we should batch data, avoiding individual transactions to the database.
	bSize := ctx.Uint("batch-size")
	eventsBatch := make([]skylab.BusEvent, bSize)

	batchIdx := 0

	// stats for imports
	var n_pkt atomic.Int64

	delegateInsert := func(events []skylab.BusEvent) {
		n, err := db.AddEventsCtx(ctx.Context, events...)
		if err != nil {
			fmt.Printf("%v", err)
		}
		n_pkt.Add(n)
	}

	var wg sync.WaitGroup
	var linenum int64 = 0
	n_unknown := 0
	n_error := 0
	for {
		line, err := fReader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break // end of file, go to the flush sequence
			}
			return err
		}
		f, err := pfun(line)
		var idErr *skylab.UnknownIdError
		if errors.As(err, &idErr) {
			fmt.Printf("unknown id %v\n", idErr.Error())
			n_unknown++
			continue
		} else if err != nil {
			fmt.Printf("got an error processing line %d: %v\n", linenum, err)
			n_error++
			continue
		}
		eventsBatch[batchIdx] = f
		linenum++
		batchIdx++
		if batchIdx >= int(bSize) {
			// flush it!!!!
			wg.Add(1)
			e := make([]skylab.BusEvent, bSize)
			copy(e, eventsBatch)
			go func(e []skylab.BusEvent) {
				delegateInsert(e)
				wg.Done()
			}(e)
			batchIdx = 0 // reset the batch
		}

	}
	// check if we have remaining packets and flush them
	if batchIdx > 0 {
		wg.Add(1)

		go func() {
			// since we don't do any modification
			// we can avoid the copy
			delegateInsert(eventsBatch[:batchIdx])
			wg.Done()
		}()
	}
	// wait for any goroutines.
	wg.Wait()
	fmt.Printf("import status: %d successful, %d unknown, %d errors\n", n_pkt.Load(), n_unknown, n_error)

	return nil
}

var clientCmd = &cli.Command{
	Name:        "client",
	Aliases:     []string{"c"},
	Subcommands: []*cli.Command{importCmd},
	Usage:       "Client utilities and tools",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "gui",
			Aliases: []string{"g"},
			Usage:   "start a local TUI",
		},
	},
	Description: `
Connects to a gotelem server or relay. Also acts as a helper command line tool.
	`,
	Action: client,
}

func client(ctx *cli.Context) error {
	return nil
}
