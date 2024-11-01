package index

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type Index struct {
	DB   *sql.DB
	Path string
}

func NewIndex(path string) (*Index, error) {
	_, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("CREATE TABLE searchIndex(id INTEGER PRIMARY KEY, name TEXT, type TEXT, path TEXT)")
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("CREATE UNIQUE INDEX anchor ON searchIndex(name, type, path)")
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &Index{
		DB:   db,
		Path: path,
	}, nil
}

func (idx *Index) Close() error {
	return idx.DB.Close()
}

func (idx *Index) WriteBuffer(b *IndexBuffer) error {
	if err, ok := b.Commit(idx.DB); !ok {
		return err
	}

	return nil
}
