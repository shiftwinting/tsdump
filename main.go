package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"

	"github.com/urfave/cli"
	"github.com/voidint/tsdump/build"
	"github.com/voidint/tsdump/config"
	"github.com/voidint/tsdump/model"
	"github.com/voidint/tsdump/model/mysql"
	"github.com/voidint/tsdump/view"
	"github.com/voidint/tsdump/view/txt"

	_ "github.com/voidint/tsdump/view/csv"
	_ "github.com/voidint/tsdump/view/json"
	_ "github.com/voidint/tsdump/view/md"
	_ "github.com/voidint/tsdump/view/txt"
)

var (
	username string
)

func init() {
	u, err := user.Current()
	if err == nil {
		username = u.Username
	}
}

var (
	c   config.Config
	out io.Writer = os.Stdout
)

func main() {
	app := cli.NewApp()
	app.Name = "tsdump"
	app.Usage = "Database table structure dump tool."
	app.Version = build.Version("0.1.0")
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "voidnt",
			Email: "voidint@126.com",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "H, host",
			Value:       "127.0.0.1",
			Usage:       "Connect to host.",
			Destination: &c.Host,
		},
		cli.IntFlag{
			Name:        "P, port",
			Value:       3306,
			Usage:       "Port number to use for connection.",
			Destination: &c.Port,
		},
		cli.StringFlag{
			Name:        "u, user",
			Value:       username,
			Usage:       "User for login if not current user.",
			Destination: &c.Username,
		},
		cli.StringFlag{
			Name:        "p, password",
			Usage:       "Password to use when connecting to server.",
			Destination: &c.Password,
		},
		cli.StringFlag{
			Name:        "d, db",
			Usage:       "Database name.",
			Destination: &c.DB,
		},
		cli.StringFlag{
			Name:  "V, viewer",
			Value: txt.Name,
			Usage: fmt.Sprintf(
				"Output viewer. Optional values: %s",
				strings.Join(view.Registered(), "|"),
			),
			Destination: &c.Viewer,
		},
		cli.StringFlag{
			Name:        "o, output",
			Usage:       "Write to a file, instead of STDOUT.",
			Destination: &c.Output,
		},
		cli.BoolFlag{
			Name:        "D, debug",
			Usage:       "Enable debug mode.",
			Destination: &c.Debug,
		},
	}
	app.Action = func(ctx *cli.Context) error {
		repo, err := mysql.NewRepo(&c)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("[tsdump] %s", err.Error()), 1)
		}

		// Get metadata
		var dbs []model.DB
		if c.DB != "" {
			dbs, err = repo.GetDBs(&model.DB{
				Name: c.DB,
			})
		} else {
			dbs, err = repo.GetDBs(nil)
		}
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("[tsdump] %s", err.Error()), 1)
		}

		if len(c.Output) > 0 {
			var f *os.File
			if f, err = os.Create(c.Output); err != nil {
				return cli.NewExitError(fmt.Sprintf("[tsdump] %s", err.Error()), 1)
			}
			defer f.Close()
			out = f
		}

		// Output as target viewer
		v := view.SelectViewer(c.Viewer)
		if v == nil {
			return cli.NewExitError(fmt.Sprintf("[tsdump] unsupported viewer: %q", c.Viewer), 1)
		}
		if err = v.Do(dbs, out); err != nil {
			return cli.NewExitError(fmt.Sprintf("[tsdump] %s", err.Error()), 1)
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("[tsdump] %s", err.Error()))
		os.Exit(1)
	}
}
