package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/auroralaboratories/pulse"
	"github.com/ghetzel/cli"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/sliceutil"
)

func main() {
	var pa *pulse.Conn

	app := cli.NewApp()
	app.Name = `pulse`
	app.Usage = `A utility for inspecting and controlling a PulseAudio sound server.`
	app.Version = pulse.Version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   `log-level, L`,
			Usage:  `Level of log output verbosity`,
			Value:  `debug`,
			EnvVar: `LOGLEVEL`,
		},
		cli.StringFlag{
			Name:  `format, f`,
			Usage: `The output format of data returned.`,
			Value: `json`,
		},
	}

	app.Before = func(c *cli.Context) error {
		log.SetLevelString(c.String(`log-level`))

		if p, err := pulse.New(`pulse`); err == nil {
			pa = p
		} else {
			log.Fatalf("Cannot connect to PulseAudio: %v", err)
		}

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  `info`,
			Usage: `Show PulseAudio daemon information.`,
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				if info, err := pa.GetServerInfo(); err == nil {
					print(c, info, nil)
				} else {
					log.Fatalf("Cannot get PulseAudio info: %v", err)
				}
			},
		}, {
			Name:  `sinks`,
			Usage: `Inspect and control audio sinks.`,
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				if sinks, err := pa.GetSinks(c.Args()...); err == nil {
					print(c, sinks, nil)
				} else {
					log.Fatalf("PulseAudio: %v", err)
				}
			},
		}, {
			Name:  `sources`,
			Usage: `Inspect and control audio sources.`,
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				if sources, err := pa.GetSources(c.Args()...); err == nil {
					print(c, sources, nil)
				} else {
					log.Fatalf("PulseAudio: %v", err)
				}
			},
		}, {
			Name:  `clients`,
			Usage: `Inspect and control PulseAudio clients.`,
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				if clients, err := pa.GetClients(c.Args()...); err == nil {
					print(c, clients, nil)
				} else {
					log.Fatalf("PulseAudio: %v", err)
				}
			},
		},
	}

	app.Run(os.Args)
}

func print(c *cli.Context, data interface{}, txtfn func()) {
	if data != nil {
		switch c.GlobalString(`format`) {
		case `json`:
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent(``, `  `)
			enc.Encode(data)
		default:
			if txtfn != nil {
				txtfn()
			} else {
				tw := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)

				for _, line := range sliceutil.Compact([]interface{}{data}) {
					fmt.Fprintf(tw, "%v\n", line)
				}

				tw.Flush()
			}
		}
	}
}
