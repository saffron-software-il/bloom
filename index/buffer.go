package index

import (
	"context"
	"database/sql"
)

type IndexBuffer struct {
	Entries []IndexEntry
}

func NewIndexBuffer() *IndexBuffer {
	return &IndexBuffer{
		Entries: nil,
	}
}

func (b *IndexBuffer) AddEntry(e IndexEntry) {
	b.Entries = append(b.Entries, e)
}

func (b *IndexBuffer) Clear() {
	b.Entries = nil
}

func (b *IndexBuffer) Len() int {
	return len(b.Entries)
}

func (b *IndexBuffer) Commit(db *sql.DB) (error, bool) {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err, false
	}
	defer tx.Rollback()

	stmt, err := db.Prepare("INSERT OR IGNORE INTO searchIndex(name, type, path) VALUES (?, ?, ?)")
	if err != nil {
		return err, false
	}
	defer stmt.Close()

	for _, entry := range b.Entries {
		stmt.Exec(entry.Name, entry.Type, entry.Path)
	}

	if err := tx.Commit(); err != nil {
		return err, false
	}

	return err, true
}
