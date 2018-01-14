package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/porty/rts-game/pak"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list files in PAK file",
			Action:  listFiles,
		},
		{
			Name:    "extract",
			Aliases: []string{"x"},
			Usage:   "extract files in PAK file",
			Action:  extractFiles,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func listFiles(c *cli.Context) error {
	if c.NArg() != 1 {
		return errors.New("failed to specify PAK filename")
	}

	f, err := os.Open(c.Args()[0])
	if err != nil {
		return fmt.Errorf("failed to open %q: %s", os.Args[1], err)
	}
	defer f.Close()

	r := pak.NewReader(f)
	records, err := r.ListFiles()
	if err != nil {
		return err
	}
	for _, rec := range records {
		log.Printf("%08x: %s", rec.Offset, rec.Filename)
	}
	return nil
}

func extractFiles(c *cli.Context) error {
	if c.NArg() != 1 {
		return errors.New("failed to specify PAK filename")
	}

	f, err := os.Open(c.Args()[0])
	if err != nil {
		return fmt.Errorf("failed to open %q: %s", os.Args[1], err)
	}
	defer f.Close()

	r := pak.NewReader(f)
	return r.ExtractFiles(".")
}
