package index

import (
	"database/sql"
	"os"

	_ "github.com/knaka/go-sqlite3-fts5"
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

	_, err = tx.Exec("CREATE VIRTUAL TABLE searchIndex_fts USING fts5(name, type UNINDEXED, path, tokenize='unicode61', prefix='2 3', content=searchIndex, content_rowid=id)")
	if err != nil {
		return nil, err
	}

	triggers := [...]string{
		`CREATE TRIGGER searchIndex_ai AFTER INSERT ON searchIndex
		BEGIN
		  INSERT INTO searchIndex_fts(rowid, name, type, path) VALUES (new.id, new.name, new.type, new.path);
		END;`,
		`
		CREATE TRIGGER searchIndex_au AFTER UPDATE ON searchIndex
		BEGIN
		  UPDATE searchIndex_fts SET name = new.name, type = new.type, path = new.path WHERE rowid = new.id;
		END;
		`,
		`
		CREATE TRIGGER searchIndex_ad AFTER DELETE ON searchIndex
		BEGIN
		  DELETE FROM searchIndex_fts WHERE rowid = old.id;
		END;
		`,
	}
	for _, trigger := range triggers {
		_, err = tx.Exec(trigger)
		if err != nil {
			return nil, err
		}
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
