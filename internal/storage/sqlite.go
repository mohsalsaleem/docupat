package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"docpatch/internal/domain"
	_ "modernc.org/sqlite"
)

type Repository struct {
	db    *sql.DB
	newID func() string
	now   func() string
}

func Open(path string, newID func() string, now func() string) (*Repository, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	repository := &Repository{db: db, newID: newID, now: now}
	if err := repository.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return repository, nil
}

func (r *Repository) Close() error { return r.db.Close() }

func (r *Repository) migrate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS documents (id TEXT PRIMARY KEY, title TEXT NOT NULL, content TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS patches (id TEXT PRIMARY KEY, document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE, base_version INTEGER NOT NULL, start_offset INTEGER NOT NULL, end_offset INTEGER NOT NULL, original_text TEXT NOT NULL, replacement_text TEXT NOT NULL, instruction TEXT NOT NULL, context_json TEXT NOT NULL DEFAULT '[]', assessment_json TEXT NOT NULL DEFAULT '{}', impacts_json TEXT NOT NULL DEFAULT '[]', status TEXT NOT NULL CHECK(status IN ('proposed','applied','rejected')), created_at TEXT NOT NULL, applied_at TEXT);
CREATE TABLE IF NOT EXISTS document_revisions (id TEXT PRIMARY KEY, document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE, version INTEGER NOT NULL, content TEXT NOT NULL, patch_id TEXT REFERENCES patches(id), created_at TEXT NOT NULL, UNIQUE(document_id, version));
CREATE TABLE IF NOT EXISTS document_sections (id TEXT PRIMARY KEY, document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE, document_title TEXT NOT NULL, title TEXT NOT NULL, slug TEXT NOT NULL, level INTEGER NOT NULL, start_offset INTEGER NOT NULL, end_offset INTEGER NOT NULL, content TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS document_links (id INTEGER PRIMARY KEY AUTOINCREMENT, source_section_id TEXT NOT NULL REFERENCES document_sections(id) ON DELETE CASCADE, target_document TEXT NOT NULL, target_heading TEXT NOT NULL, kind TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS section_embeddings (section_id TEXT NOT NULL REFERENCES document_sections(id) ON DELETE CASCADE, model TEXT NOT NULL, content_hash TEXT NOT NULL, vector_json TEXT NOT NULL, PRIMARY KEY(section_id, model));
CREATE INDEX IF NOT EXISTS patches_document_created ON patches(document_id, created_at DESC);
CREATE INDEX IF NOT EXISTS sections_document_position ON document_sections(document_id, start_offset);
CREATE INDEX IF NOT EXISTS sections_slug ON document_sections(slug);
CREATE INDEX IF NOT EXISTS links_source ON document_links(source_section_id);`)
	if err != nil {
		return err
	}
	if err := r.ensureColumn(ctx, "patches", "context_json", "TEXT NOT NULL DEFAULT '[]'"); err != nil {
		return err
	}
	if err := r.ensureColumn(ctx, "patches", "assessment_json", "TEXT NOT NULL DEFAULT '{}'"); err != nil {
		return err
	}
	return r.ensureColumn(ctx, "patches", "impacts_json", "TEXT NOT NULL DEFAULT '[]'")
}

func (r *Repository) ensureColumn(ctx context.Context, table, column, definition string) error {
	rows, err := r.db.QueryContext(ctx, "PRAGMA table_info("+table+")")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, kind string
		var notNull, primaryKey int
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &kind, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, "ALTER TABLE "+table+" ADD COLUMN "+column+" "+definition)
	return err
}

func (r *Repository) Seed(ctx context.Context, title, content string) error {
	var count int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM documents").Scan(&count); err != nil || count > 0 {
		return err
	}
	_, err := r.CreateDocument(ctx, title, content)
	return err
}

func scanDocument(row interface{ Scan(...any) error }) (domain.Document, error) {
	var d domain.Document
	err := row.Scan(&d.ID, &d.Title, &d.Content, &d.Version, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}
func scanPatch(row interface{ Scan(...any) error }) (domain.Patch, error) {
	var p domain.Patch
	var applied sql.NullString
	var contextJSON, assessmentJSON, impactsJSON string
	err := row.Scan(&p.ID, &p.DocumentID, &p.BaseVersion, &p.Start, &p.End, &p.Original, &p.Replacement, &p.Instruction, &contextJSON, &assessmentJSON, &impactsJSON, &p.Status, &p.CreatedAt, &applied)
	if err == nil {
		err = json.Unmarshal([]byte(contextJSON), &p.Context)
	}
	if err == nil {
		err = json.Unmarshal([]byte(assessmentJSON), &p.Assessment)
	}
	if err == nil {
		err = json.Unmarshal([]byte(impactsJSON), &p.Impacts)
	}
	if applied.Valid {
		p.AppliedAt = &applied.String
	}
	return p, err
}

func translate(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func (r *Repository) GetDocument(ctx context.Context, id string) (domain.Document, error) {
	d, err := scanDocument(r.db.QueryRowContext(ctx, "SELECT id,title,content,version,created_at,updated_at FROM documents WHERE id=?", id))
	return d, translate(err)
}
func (r *Repository) ListDocuments(ctx context.Context) ([]domain.Document, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id,title,substr(content,1,600),version,created_at,updated_at FROM documents ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Document{}
	for rows.Next() {
		var d domain.Document
		e := rows.Scan(&d.ID, &d.Title, &d.Excerpt, &d.Version, &d.CreatedAt, &d.UpdatedAt)
		if e != nil {
			return nil, e
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *Repository) CreateDocument(ctx context.Context, title, content string) (domain.Document, error) {
	now := r.now()
	d := domain.Document{ID: r.newID(), Title: title, Content: content, Version: 1, CreatedAt: now, UpdatedAt: now}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return d, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "INSERT INTO documents(id,title,content,version,created_at,updated_at) VALUES(?,?,?,?,?,?)", d.ID, d.Title, d.Content, 1, now, now); err != nil {
		return d, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO document_revisions(id,document_id,version,content,created_at) VALUES(?,?,?,?,?)", r.newID(), d.ID, 1, d.Content, now); err != nil {
		return d, err
	}
	return d, tx.Commit()
}

func (r *Repository) SaveDocument(ctx context.Context, id string, version int, title, content string) (domain.Document, error) {
	now := r.now()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Document{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, "UPDATE documents SET title=?,content=?,version=version+1,updated_at=? WHERE id=? AND version=?", title, content, now, id, version)
	if err != nil {
		return domain.Document{}, err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return domain.Document{}, domain.ErrConflict
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO document_revisions(id,document_id,version,content,created_at) VALUES(?,?,?,?,?)", r.newID(), id, version+1, content, now); err != nil {
		return domain.Document{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Document{}, err
	}
	return r.GetDocument(ctx, id)
}

func (r *Repository) DeleteDocument(ctx context.Context, id string) (domain.Document, error) {
	document, err := r.GetDocument(ctx, id)
	if err != nil {
		return document, err
	}
	result, err := r.db.ExecContext(ctx, "DELETE FROM documents WHERE id=?", id)
	if err != nil {
		return document, err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return document, domain.ErrNotFound
	}
	return document, nil
}

func (r *Repository) GetDocumentRevision(ctx context.Context, id string, version int) (string, error) {
	var content string
	err := r.db.QueryRowContext(ctx, "SELECT content FROM document_revisions WHERE document_id=? AND version=?", id, version).Scan(&content)
	return content, translate(err)
}

const patchSelect = "SELECT id,document_id,base_version,start_offset,end_offset,original_text,replacement_text,instruction,context_json,assessment_json,impacts_json,status,created_at,applied_at FROM patches"

func (r *Repository) CreatePatch(ctx context.Context, p domain.Patch) error {
	contextJSON, err := json.Marshal(p.Context)
	if err != nil {
		return err
	}
	assessmentJSON, err := json.Marshal(p.Assessment)
	if err != nil {
		return err
	}
	impactsJSON, err := json.Marshal(p.Impacts)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, "INSERT INTO patches(id,document_id,base_version,start_offset,end_offset,original_text,replacement_text,instruction,context_json,assessment_json,impacts_json,status,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)", p.ID, p.DocumentID, p.BaseVersion, p.Start, p.End, p.Original, p.Replacement, p.Instruction, string(contextJSON), string(assessmentJSON), string(impactsJSON), p.Status, p.CreatedAt)
	return err
}
func (r *Repository) ListPatches(ctx context.Context, documentID string) ([]domain.Patch, error) {
	rows, err := r.db.QueryContext(ctx, patchSelect+" WHERE document_id=? ORDER BY created_at DESC LIMIT 50", documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Patch{}
	for rows.Next() {
		p, e := scanPatch(rows)
		if e != nil {
			return nil, e
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *Repository) ApplyPatch(ctx context.Context, id string) (domain.Document, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Document{}, err
	}
	defer tx.Rollback()
	p, err := scanPatch(tx.QueryRowContext(ctx, patchSelect+" WHERE id=?", id))
	if err != nil {
		return domain.Document{}, translate(err)
	}
	if p.Status != "proposed" {
		return domain.Document{}, domain.ErrConflict
	}
	var content string
	var version int
	if err = tx.QueryRowContext(ctx, "SELECT content,version FROM documents WHERE id=?", p.DocumentID).Scan(&content, &version); err != nil {
		return domain.Document{}, translate(err)
	}
	if version != p.BaseVersion {
		return domain.Document{}, domain.ErrConflict
	}
	next, err := domain.Apply(content, domain.Selection{Start: p.Start, End: p.End}, p.Original, p.Replacement)
	if err != nil {
		return domain.Document{}, err
	}
	now := r.now()
	result, err := tx.ExecContext(ctx, "UPDATE documents SET content=?,version=version+1,updated_at=? WHERE id=? AND version=?", next, now, p.DocumentID, version)
	if err != nil {
		return domain.Document{}, err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return domain.Document{}, domain.ErrConflict
	}
	if _, err = tx.ExecContext(ctx, "UPDATE patches SET status='applied',applied_at=? WHERE id=?", now, p.ID); err != nil {
		return domain.Document{}, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO document_revisions(id,document_id,version,content,patch_id,created_at) VALUES(?,?,?,?,?,?)", r.newID(), p.DocumentID, version+1, next, p.ID, now); err != nil {
		return domain.Document{}, err
	}
	if err = tx.Commit(); err != nil {
		return domain.Document{}, err
	}
	return r.GetDocument(ctx, p.DocumentID)
}
func (r *Repository) RejectPatch(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "UPDATE patches SET status='rejected' WHERE id=? AND status='proposed'", id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (r *Repository) ReplaceDocumentIndex(ctx context.Context, index domain.DocumentIndex) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "DELETE FROM document_links WHERE source_section_id IN (SELECT id FROM document_sections WHERE document_id=?)", index.DocumentID); err != nil {
		return err
	}
	keep := map[string]bool{}
	for _, section := range index.Sections {
		keep[section.ID] = true
		if _, err = tx.ExecContext(ctx, "INSERT INTO document_sections(id,document_id,document_title,title,slug,level,start_offset,end_offset,content) VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET document_title=excluded.document_title,title=excluded.title,slug=excluded.slug,level=excluded.level,start_offset=excluded.start_offset,end_offset=excluded.end_offset,content=excluded.content", section.ID, section.DocumentID, section.DocumentTitle, section.Title, section.Slug, section.Level, section.Start, section.End, section.Content); err != nil {
			return err
		}
	}
	rows, err := tx.QueryContext(ctx, "SELECT id FROM document_sections WHERE document_id=?", index.DocumentID)
	if err != nil {
		return err
	}
	obsolete := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		if !keep[id] {
			obsolete = append(obsolete, id)
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, id := range obsolete {
		if _, err = tx.ExecContext(ctx, "DELETE FROM document_sections WHERE id=?", id); err != nil {
			return err
		}
	}
	for _, link := range index.Links {
		if _, err = tx.ExecContext(ctx, "INSERT INTO document_links(source_section_id,target_document,target_heading,kind) VALUES(?,?,?,?)", link.SourceSectionID, link.TargetDocument, link.TargetHeading, link.Kind); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) ListIndexedSections(ctx context.Context) ([]domain.IndexedSection, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id,document_id,document_title,title,slug,level,start_offset,end_offset,content FROM document_sections ORDER BY document_id,start_offset")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.IndexedSection{}
	for rows.Next() {
		var item domain.IndexedSection
		if err := rows.Scan(&item.ID, &item.DocumentID, &item.DocumentTitle, &item.Title, &item.Slug, &item.Level, &item.Start, &item.End, &item.Content); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListIndexedLinks(ctx context.Context) ([]domain.IndexedLink, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT source_section_id,target_document,target_heading,kind FROM document_links ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.IndexedLink{}
	for rows.Next() {
		var item domain.IndexedLink
		if err := rows.Scan(&item.SourceSectionID, &item.TargetDocument, &item.TargetHeading, &item.Kind); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListSectionEmbeddings(ctx context.Context, model string) ([]domain.SectionEmbedding, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT section_id,content_hash,model,vector_json FROM section_embeddings WHERE model=?", model)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.SectionEmbedding{}
	for rows.Next() {
		var item domain.SectionEmbedding
		var vectorJSON string
		if err := rows.Scan(&item.SectionID, &item.ContentHash, &item.Model, &vectorJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(vectorJSON), &item.Vector); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpsertSectionEmbeddings(ctx context.Context, items []domain.SectionEmbedding) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, item := range items {
		vectorJSON, marshalErr := json.Marshal(item.Vector)
		if marshalErr != nil {
			return marshalErr
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO section_embeddings(section_id,model,content_hash,vector_json) VALUES(?,?,?,?) ON CONFLICT(section_id,model) DO UPDATE SET content_hash=excluded.content_hash,vector_json=excluded.vector_json", item.SectionID, item.Model, item.ContentHash, string(vectorJSON)); err != nil {
			return err
		}
	}
	return tx.Commit()
}
