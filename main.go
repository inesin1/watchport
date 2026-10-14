package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const defaultTimeout = 5 * time.Second

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("watchport", flag.ContinueOnError)
	flags.SetOutput(stderr)
	timeout := flags.Duration("timeout", defaultTimeout, "maximum wait for each URL")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if *timeout <= 0 {
		fmt.Fprintln(stderr, "watchport: -timeout must be greater than zero")
		return 2
	}

	targets := flags.Args()
	if len(targets) == 0 {
		fmt.Fprintln(stderr, "usage: watchport [-timeout duration] URL [URL ...]")
		flags.PrintDefaults()
		return 2
	}
	for _, target := range targets {
		if err := validateURL(target); err != nil {
			fmt.Fprintf(stderr, "watchport: %v\n", err)
			return 2
		}
	}

	client := &http.Client{Timeout: *timeout}
	results := make([]checkResult, len(targets))
	var checks sync.WaitGroup
	for i, target := range targets {
		checks.Add(1)
		go func(i int, target string) {
			defer checks.Done()
			results[i] = check(target, client)
		}(i, target)
	}
	checks.Wait()

	allOK := true
	for _, result := range results {
		if result.err != nil {
			fmt.Fprintf(stdout, "FAIL %s (%s): %v\n", result.url, result.elapsed.Round(time.Millisecond), result.err)
			allOK = false
			continue
		}

		label := "OK"
		if result.statusCode < http.StatusOK || result.statusCode >= http.StatusMultipleChoices {
			label = "FAIL"
			allOK = false
		}
		fmt.Fprintf(stdout, "%s %s HTTP %d (%s)\n", label, result.url, result.statusCode, result.elapsed.Round(time.Millisecond))
	}
	if !allOK {
		return 1
	}
	return 0
}

func validateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	scheme := strings.ToLower(parsed.Scheme)
	if (scheme != "http" && scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("invalid URL %q: expected an absolute HTTP(S) URL", rawURL)
	}
	return nil
}
