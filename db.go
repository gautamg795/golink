package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Link is the structure stored for each go short link.
type Link struct {
	Short    string // the "foo" part of http://go/foo
	Long     string // the target URL or text/template pattern to run
	Created  time.Time
	LastEdit time.Time // when the link was last edited
	Owner    string    // user@domain
	Clicks   int       // number of times this link has been served
}

// DB provides storage for Links.
type DB interface {
	// LoadAll returns all stored Links.
	LoadAll() ([]*Link, error)

	// Load a Link by its short name. It returns fs.ErrNotExist if the link does not exist.
	Load(short string) (*Link, error)

	// Save a Link.
	Save(link *Link) error
}

// FileDB stores Links in JSON files on disk.
type FileDB struct {
	// dir is the directory to store one JSON file per link.
	dir string
}

// NewFileDB returns a new FileDB which will store links in individual JSON
// files in the specified directory.  If mkdir is true, the directory will be
// created if it does not exist.
func NewFileDB(dir string, mkdir bool) (*FileDB, error) {
	if mkdir {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}
	if fi, err := os.Stat(dir); err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", dir)
	}
	return &FileDB{dir: dir}, nil
}

// linkPath returns the path to the file on disk for the specified link. Short
// name is normalized to be case insensitive, remove dashes, and escape some
// characters.
//
// TODO(willnorris): some of this normalization is not unique to FileDB and
// should be moved elsewhere
func (f *FileDB) linkPath(short string) string {
	name := url.PathEscape(strings.ToLower(short))
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, ".", "%2e")
	return filepath.Join(f.dir, name)
}

func (f *FileDB) LoadAll() ([]*Link, error) {
	d, err := os.Open(f.dir)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	names, err := d.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	links := make([]*Link, len(names))
	for i, short := range names {
		link, err := f.Load(short)
		if err != nil {
			return nil, err
		}
		links[i] = link
	}

	return links, nil
}

func (f *FileDB) Load(short string) (*Link, error) {
	data, err := os.ReadFile(f.linkPath(short))
	if err != nil {
		return nil, err
	}
	link := new(Link)
	if err := json.Unmarshal(data, link); err != nil {
		return nil, err
	}
	return link, nil
}

func (f *FileDB) Save(link *Link) error {
	j, err := json.MarshalIndent(link, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(f.linkPath(link.Short), j, 0600); err != nil {
		return err
	}
	return nil
}