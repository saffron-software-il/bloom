package drivers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
	"software.saffron/bloom/index"
)

type PythonDriver struct {
	version  string
	tempPath string
}

func NewPythonDriver(tempFolder string) *PythonDriver {
	return &PythonDriver{
		version:  "3.13",
		tempPath: tempFolder,
	}
}

func symStringToType(str string) (error, index.EntryType) {
	switch str {
	case "annotation":
		return nil, index.Annotation
	case "attribute":
		return nil, index.Attribute
	case "class":
		return nil, index.Class
	case "enum":
		return nil, index.Enum
	case "exception":
		return nil, index.Exception
	case "function":
		return nil, index.Function
	case "keyword":
		return nil, index.Keyword
	case "macro":
		return nil, index.Macro
	case "method":
		return nil, index.Method
	case "module":
		return nil, index.Module
	case "object":
		return nil, index.Object
	case "type":
		return nil, index.Type
	case "member":
		return nil, index.Value
	case "var", "data":
		return nil, index.Variable
	case "struct":
		return nil, index.Struct
	case "envvar":
		return nil, index.Environment
	case "opcode":
		return nil, index.Instruction
	case "monitoring-event":
		return nil, index.Event
	case "option", "cmdoption":
		return nil, index.Option
	case "pdbcommand":
		return nil, index.Command
	}

	return fmt.Errorf("unknown symbol string type: '%s'", str), index.Annotation
}

func downloadFile(url, dest string) error {
	// Create the file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for server errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	// Write the data to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// Extract each file in the ZIP archive
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := extractFile(f, fpath); err != nil {
			return err
		}
	}
	return nil
}

func extractFile(f *zip.File, dest string) error {
	// Open the file inside the ZIP archive
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create the destination file
	outFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Copy the contents of the file to the destination file
	_, err = io.Copy(outFile, rc)
	return err
}

func (p *PythonDriver) GetDocId() string {
	return "python"
}

func (p *PythonDriver) GetDocName() string {
	return fmt.Sprintf("Python v.%s", p.version)
}

func (p *PythonDriver) SourcesExist() bool {
	zipFilePath := filepath.Join(p.tempPath, "python-docs.zip")
	if _, err := os.Stat(zipFilePath); err == nil {
		return true
	}

	return false
}

func (p *PythonDriver) DownloadSources() (error, []string) {
	zipFilePath := filepath.Join(p.tempPath, "python-docs.zip")
	const url = "https://docs.python.org/%s/archives/python-%s-docs-html.zip"
	downloadUrl := fmt.Sprintf(url, p.version, p.version)

	fmt.Println("Downloading file...")
	if err := downloadFile(downloadUrl, zipFilePath); err != nil {
		return err, nil
	}

	fmt.Println("Download completed!")

	// Unzip the downloaded file
	fmt.Println("Extracting ZIP file...")
	if err := unzip(zipFilePath, p.tempPath); err != nil {
		return fmt.Errorf("Failed to extract ZIP file: %v", err), nil
	}
	fmt.Println("Extraction completed!")

	htmlPathBase := filepath.Join(p.tempPath, fmt.Sprintf("python-%s-docs-html", p.version))
	paths := []string{
		filepath.Join(htmlPathBase, "library"),
		filepath.Join(htmlPathBase, "c-api"),
	}

	return nil, paths
}

func (p *PythonDriver) GetSources() []string {
	htmlPathBase := filepath.Join(p.tempPath, fmt.Sprintf("python-%s-docs-html", p.version))
	paths := []string{
		filepath.Join(htmlPathBase, "library"),
		filepath.Join(htmlPathBase, "c-api"),
	}

	return paths
}

func (p *PythonDriver) ScrapePage(filePath string, data []byte) (error, *index.IndexBuffer) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return err, nil
	}

	buffer := index.NewIndexBuffer()
	doc.Find("dl").Each(func(i int, dl *goquery.Selection) {
		dlClass, classExists := dl.Attr("class")
		if !classExists {
			return
		}

		// Some stuff can be skipped
		// TODO: Figure out how to handle "describe"
		if dlClass == "simple" || dlClass == "describe" || dlClass == "field-list simple" {
			return
		}

		parts := strings.Split(dlClass, " ")
		if len(parts) != 2 {
			fmt.Printf("%s: warning: symbol '%s' does not have two parts like expected so skipping it: '%s'\n", filePath, dlClass, dl.Text())
			return
		}

		dt := dl.Find("dt").First()
		symTypeString := parts[1]

		link := dt.Find("a").First().AttrOr("href", "")
		if len(link) == 0 {
			return
		}

		path := fmt.Sprintf("%s%s", filePath, link)

		dt.Find("span.sig-name.descname").Each(func(i int, span *goquery.Selection) {
			sel := dt.Find("em.property > span.pre").First()
			if sel == nil {
				fmt.Printf("%s: warning: symbol '%s' does not have a span.pre describing it: '%s'\n", filePath, dlClass, dl.Text())
				return
			}

			err, symType := symStringToType(symTypeString)
			if err != nil {
				fmt.Printf("%s: warning: unknown symbol type for symbol '%s': %s\n", filePath, symTypeString, dl.Text())
				return
			}

			name, exists := dt.Attr("id")
			if !exists {
				html, err := dl.Html()
				if err != nil {
					fmt.Printf("%s: error: symbol does not contain ID: '%s'\n", filePath, html)
				}

				return
			}

			entry := index.NewIndexEntry(name, path, symType)
			buffer.AddEntry(entry)
		})
	})

	return nil, buffer
}
