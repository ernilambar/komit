package main

import (
	"errors"
	"fmt"
	"os"

	"charm.land/huh/v2"
	"github.com/spf13/pflag"
)

// version is the build version, overridable at release time via
// -ldflags "-X main.version=vX.Y.Z".
var version = "dev"

type options struct {
	dryRun bool
	noPush bool
	all    bool
	model  string
	yes    bool
	count  int
}

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) > 1 && os.Args[1] == "status" {
		return runStatus(os.Args[2:])
	}

	var opts options
	pflag.BoolVar(&opts.dryRun, "dry-run", false, "print subjects without committing")
	pflag.BoolVar(&opts.noPush, "no-push", false, "commit without pushing")
	pflag.BoolVarP(&opts.all, "all", "a", false, "stage modified and deleted tracked files")
	pflag.StringVar(&opts.model, "model", "", "model to use (overrides KOMIT_MODEL)")
	pflag.BoolVarP(&opts.yes, "yes", "y", false, "commit the first suggestion without prompting")
	pflag.IntVarP(&opts.count, "count", "n", 2, "number of subject candidates (1-5)")
	showVersion := pflag.BoolP("version", "V", false, "print version and exit")
	showHelp := pflag.BoolP("help", "h", false, "print help and exit")
	pflag.Parse()

	if *showHelp {
		pflag.CommandLine.SetOutput(os.Stdout)
		pflag.Usage()
		return 0
	}
	if *showVersion {
		fmt.Println(version)
		return 0
	}

	if err := validateFlags(opts, pflag.CommandLine.Changed("count")); err != nil {
		fmt.Fprintln(os.Stderr, "komit:", err)
		return 1
	}
	if opts.yes {
		opts.count = 1
	}
	if !opts.dryRun && !opts.yes && !stdinIsTTY() {
		fmt.Fprintln(os.Stderr, "komit: stdin is not a terminal; use -y or --dry-run")
		return 1
	}

	cfg := loadConfig()
	if opts.model != "" {
		cfg.Model = opts.model
	}
	if err := cfg.validate(); err != nil {
		fmt.Fprintln(os.Stderr, "komit:", err)
		return 1
	}

	if err := ensureRepo(); err != nil {
		fmt.Fprintln(os.Stderr, "komit:", err)
		return 1
	}
	if opts.all {
		if err := stageTracked(); err != nil {
			fmt.Fprintln(os.Stderr, "komit:", err)
			return 1
		}
	}
	if err := checkStaged(); err != nil {
		fmt.Fprintln(os.Stderr, "komit:", err)
		return 1
	}

	input, err := buildModelInput()
	if err != nil {
		fmt.Fprintln(os.Stderr, "komit:", err)
		return 1
	}

	var subject string
	if opts.dryRun || opts.yes {
		subjects, err := suggestSubjects(cfg, input, opts.count)
		if err != nil {
			fmt.Fprintln(os.Stderr, "komit:", err)
			return 1
		}
		if opts.dryRun {
			for i, subject := range subjects {
				fmt.Printf("%d. %s\n", i+1, subject)
			}
			return 0
		}
		subject = subjects[0]
		fmt.Println(subject)
	} else {
		subject, err = chooseSubject(cfg, input, opts.count)
		if errors.Is(err, huh.ErrUserAborted) {
			return 130
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "komit:", err)
			return 1
		}
	}

	if err := commit(subject); err != nil {
		fmt.Fprintln(os.Stderr, "komit:", err)
		return 1
	}
	if opts.noPush {
		return 0
	}
	if err := push(); err != nil {
		fmt.Fprintln(os.Stderr, "komit:", err)
		return 1
	}
	return 0
}

func validateFlags(opts options, countSet bool) error {
	if opts.count < 1 || opts.count > 5 {
		return fmt.Errorf("-n must be between 1 and 5, got %d", opts.count)
	}
	if opts.yes && countSet && opts.count > 1 {
		return fmt.Errorf("-y commits a single subject; remove -n or use -n 1")
	}
	return nil
}
