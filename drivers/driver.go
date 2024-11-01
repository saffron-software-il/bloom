package drivers

import "software.saffron/bloom/index"

type Driver interface {
	// GetDocId returns the ID of the documentation that this
	// driver provides
	GetDocId() string

	// GetDocName returns the user-visible name of the language or framework
	// that is provided documentation for by this driver.
	GetDocName() string

	// DownloadSources downloads or scrapes all HTML documentation sources
	// to a temporary directory, and returns paths for all HTML
	// sources that should be scraped.
	DownloadSources() (error, []string)

	// GetSources returns the paths of the downloded HTML sources that should be scraped
	GetSources() []string

	// SourcesExist checks if the sources already exist in the temp directory.
	// Used to determine if we can skip downloads.
	SourcesExist() bool

	// ScrapePage scrapes the data of a single page and returns the scraped
	// data as an IndexBuffer that can be added to the global
	// docset index
	ScrapePage(filePath string, data []byte) (error, *index.IndexBuffer)
}
