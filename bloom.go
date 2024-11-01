package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"software.saffron/bloom/drivers"
	"software.saffron/bloom/index"
)

const (
	GenerateCommand = "generate"
)

var Commands = []string{
	GenerateCommand,
}

func scrapePages(driver drivers.Driver, idx *index.Index, paths []string) (error, bool) {
	execs := make(chan struct{}, runtime.NumCPU()*2)
	var wg sync.WaitGroup

	for _, currPath := range paths {
		err := filepath.WalkDir(currPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Check if the file has an .html extension
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".html") {
				execs <- struct{}{}
				go func(driver drivers.Driver) {
					wg.Add(1)
					defer func() {
						wg.Done()
						<-execs
					}()

					fmt.Printf("Scraping page '%s'...\n", path)
					data, err := os.ReadFile(path)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error reading page for scraping '%s': %v", path, err)
						return
					}

					err, buffer := driver.ScrapePage(path, data)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error scraping page '%s': %v", path, err)
						return
					}

					if err = idx.WriteBuffer(buffer); err != nil {
						fmt.Fprintf(os.Stderr, "Error writing to index while scraping page '%s': %v", path, err)
						return
					}
				}(driver)
			}

			return nil
		})

		if err != nil {
			return err, false
		}
	}

	wg.Wait()
	return nil, true
}

func generate(t, url string, skipExists bool, outputPath string) (error, bool) {
	// Figure out which generator to use
	if len(t) == 0 {
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("must specify documentation driver to use.\nSupported drivers are:\n"))

		for _, driver := range drivers.Drivers {
			builder.WriteString(driver)
			builder.WriteString("\n")
		}

		return errors.New(builder.String()), false
	}

	// Output that is not the temp directory might require creating it first
	if outputPath != os.TempDir() {
		err := os.MkdirAll(outputPath, os.ModePerm)
		if err != nil {
			return err, false
		}
	}

	index, err := index.NewIndex("temp/index.db")
	if err != nil {
		return err, false
	}
	defer index.Close()

	var driver drivers.Driver
	switch t {
	case "python":
		driver = drivers.NewPythonDriver("temp")
	default:
		return fmt.Errorf("Unsupported generator '%s'", t), false
	}

	var sources []string
	if driver.SourcesExist() && skipExists {
		fmt.Printf("Warning: Skipping downloading of sources because they exist and skip-exists is ON.\n")
		sources = driver.GetSources()
	} else {
		err, sources = driver.DownloadSources()
		if err != nil {
			return err, false
		}
	}

	if err, ok := scrapePages(driver, index, sources); !ok {
		return err, false
	}

	// TODO: Create metadata
	// Check if some metadata already exists.
	// If it does, use that, just increase the build number.
	// If not we create one from scratch.

	return nil, true
}

func main() {
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	generateType := generateCmd.String("type", "", "The documentation type to generate")
	generateUrl := generateCmd.String("url", "", "The URL to use")
	generateSkipDownload := generateCmd.Bool("skip-exists", false, "Skip downloading sources if files exist")
	generateOutput := generateCmd.String("output", os.TempDir(), "The output directory for all files")

	if len(os.Args) == 1 {
		fmt.Println("Error: too few arguments")
		os.Exit(-1)
	}

	switch os.Args[1] {
	case "generate":
		generateCmd.Parse(os.Args[2:])
		if err, ok := generate(*generateType, *generateUrl, *generateSkipDownload, *generateOutput); !ok {
			fmt.Fprintf(os.Stderr, "Error generating documentation: %v", err)
			os.Exit(-1)
		}
	default:
		fmt.Printf("Unknown command '%s'.\nThe following commands are supported:\n", os.Args[1])
		for _, cmd := range Commands {
			fmt.Printf("%s\n", cmd)
		}
		fmt.Println()
	}
}
