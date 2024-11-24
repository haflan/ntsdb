package ntsdb

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jmoiron/sqlx"
)

type Note struct {
	ID              int            `db:"id"`
	UUID            string         `db:"uuid"`
	Text            sql.NullString `db:"text"`
	GroupID         int            `db:"group_id"`
	GroupUUID       string         `db:"group_uuid"`
	Type            NoteType       `db:"note_type"`
	Media           sql.NullString `db:"media"`
	Selected        int            `db:"selected"`
	Latitude        float64        `db:"latitude"`
	Longitude       float64        `db:"longitude"`
	CreatedAt       int            `db:"createdAt"`
	UpdatedAt       int            `db:"updatedAt"`
	SiteName        sql.NullString `db:"site_name"`
	LinkTitle       sql.NullString `db:"link_title"`
	LinkImage       sql.NullString `db:"link_image"`
	LinkDescription sql.NullString `db:"link_description"`
}

type NoteType int

const (
	NoteTypeText     NoteType = 1
	NoteTypeImage    NoteType = 2
	NoteTypeAudio    NoteType = 3
	NoteTypeLocation NoteType = 6
)

type NoteGroup struct {
	ID             int            `db:"id"`
	UUID           string         `db:"uuid"`
	Title          string         `db:"title"`
	Image          sql.NullString `db:"image"`
	ImageType      sql.NullInt64  `db:"image_type"`
	Description    sql.NullString `db:"description"`
	Order          sql.NullInt64  `db:"order"`
	Pinned         int            `db:"pinned"`
	LatestNoteType sql.NullInt64  `db:"latest_note_type"`
	LatestNoteText sql.NullString `db:"latest_note_text"`
	CreatedAt      int            `db:"createdAt"`
	UpdatedAt      int            `db:"updatedAt"`
}

type NTSDB struct {
	db    *sqlx.DB
	Close func()
}

func OpenNTSDB(path string) NTSDB {
	cleanup := func() {}
	if strings.HasSuffix(path, ".zip") {
		var err error
		path, err = unzipDB(path)
		if err != nil {
			panic(err)
		}
		cleanup = func() {
			os.Remove(path)
			os.Remove(path + "-wal")
			os.Remove(path + "-shm")
		}
	}
	return NTSDB{
		db:    sqlx.MustConnect("sqlite3", path),
		Close: cleanup,
	}
}

func (nts NTSDB) GetNotes() ([]Note, error) {
	var out []Note
	rows, err := nts.db.Queryx("SELECT * FROM notes ORDER BY createdAt ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var note Note
		err = rows.StructScan(&note)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		out = append(out, note)
	}
	return out, nil
}

func (nts NTSDB) GetNoteGroups() ([]NoteGroup, error) {
	var out []NoteGroup
	rows, err := nts.db.Queryx("SELECT * FROM notegroups")
	if err != nil {
		return nil, fmt.Errorf("failed to query note groups: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var group NoteGroup
		err = rows.StructScan(&group)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		out = append(out, group)
	}
	return out, nil
}

// unzipDB extracts the notetoself.db file from a zip archive
// and returns the path to the extracted file.
func unzipDB(zipPath string) (string, error) {

	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to read zip file: %w", err)
	}

	// Create a reader from the in-memory zip data
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return "", err
	}

	var dbfile *zip.File
	// Iterate over the files in the zip archive
	for _, file := range zipReader.File {
		if file.Name == "notetoself.db" {
			dbfile = file
			break
		}
	}
	if dbfile == nil {
		return "", fmt.Errorf("notetoself.db not found in zip archive")
	}

	// Open the file
	in, err := dbfile.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open notetoself.db: %w", err)
	}
	defer in.Close()

	out, err := os.CreateTemp("", "ntsdb-*.db")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return out.Name(), nil
}
