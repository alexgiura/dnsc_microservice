package repository

import (
	"context"
	"dnsc_microservice/internal/models"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// DomainRepository defines the interface for domain and domain_records persistence
type DomainRepository interface {
	Insert(ctx context.Context, domain *models.Domain) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Domain, error)
	List(ctx context.Context) ([]*models.Domain, error)
	SetWhitelist(ctx context.Context, id uuid.UUID, whitelist bool) error
	Update(ctx context.Context, domain *models.Domain) error
	GetByValueAndType(ctx context.Context, value, typ string) (*models.Domain, error)
	InsertRecords(ctx context.Context, domainID uuid.UUID, records []models.DomainRecord) error
}

type domainRepository struct {
	db *pgxpool.Pool
}

// NewDomainRepository creates a new domain repository
func NewDomainRepository(db *pgxpool.Pool) DomainRepository {
	return &domainRepository{db: db}
}

// Insert persists a domain and its records (in one transaction)
func (r *domainRepository) Insert(ctx context.Context, domain *models.Domain) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO core.domains (id, value, type, whitelist)
		VALUES ($1, $2, $3, $4)
	`, domain.ID, domain.Value, domain.Type, domain.Whitelist)
	if err != nil {
		return fmt.Errorf("insert domain: %w", err)
	}

	if len(domain.Records) > 0 {
		if err := r.insertRecordsTx(ctx, tx, domain.ID, domain.Records); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// insertRecordsTx inserts domain_records inside an existing transaction
func (r *domainRepository) insertRecordsTx(ctx context.Context, tx pgx.Tx, domainID uuid.UUID, records []models.DomainRecord) error {
	for _, rec := range records {
		_, err := tx.Exec(ctx, `
			INSERT INTO core.domain_records (id, domain_id, ticket_id, description, tags, date, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, rec.ID, domainID, rec.TicketID, rec.Description, rec.Tags, rec.Date, rec.Source)
		if err != nil {
			return fmt.Errorf("insert domain_record: %w", err)
		}
	}
	return nil
}

// InsertRecords appends records to an existing domain
func (r *domainRepository) InsertRecords(ctx context.Context, domainID uuid.UUID, records []models.DomainRecord) error {
	if len(records) == 0 {
		return nil
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := r.insertRecordsTx(ctx, tx, domainID, records); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// GetByID retrieves a domain by ID and loads its records
func (r *domainRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Domain, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, value, type, whitelist FROM core.domains WHERE id = $1
	`, id)
	var d models.Domain
	if err := row.Scan(&d.ID, &d.Value, &d.Type, &d.Whitelist); err != nil {
		return nil, fmt.Errorf("get domain by id: %w", err)
	}
	records, err := r.getRecordsByDomainID(ctx, id)
	if err != nil {
		return nil, err
	}
	d.Records = records
	return &d, nil
}

func (r *domainRepository) getRecordsByDomainID(ctx context.Context, domainID uuid.UUID) ([]models.DomainRecord, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, domain_id, ticket_id, description, tags, date, source
		FROM core.domain_records
		WHERE domain_id = $1
		ORDER BY date DESC
	`, domainID)
	if err != nil {
		return nil, fmt.Errorf("get domain_records: %w", err)
	}
	defer rows.Close()
	var list []models.DomainRecord
	for rows.Next() {
		var rec models.DomainRecord
		if err := rows.Scan(&rec.ID, &rec.DomainID, &rec.TicketID, &rec.Description, &rec.Tags, &rec.Date, &rec.Source); err != nil {
			return nil, fmt.Errorf("scan domain_record: %w", err)
		}
		list = append(list, rec)
	}
	return list, rows.Err()
}

// List retrieves all domains with their records
func (r *domainRepository) List(ctx context.Context) ([]*models.Domain, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, value, type, whitelist FROM core.domains ORDER BY value
	`)
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}
	defer rows.Close()
	var domains []*models.Domain
	for rows.Next() {
		var d models.Domain
		if err := rows.Scan(&d.ID, &d.Value, &d.Type, &d.Whitelist); err != nil {
			return nil, fmt.Errorf("scan domain: %w", err)
		}
		records, err := r.getRecordsByDomainID(ctx, d.ID)
		if err != nil {
			return nil, err
		}
		d.Records = records
		domains = append(domains, &d)
	}
	return domains, rows.Err()
}

// SetWhitelist updates the whitelist flag for a domain
func (r *domainRepository) SetWhitelist(ctx context.Context, id uuid.UUID, whitelist bool) error {
	_, err := r.db.Exec(ctx, `UPDATE core.domains SET whitelist = $2 WHERE id = $1`, id, whitelist)
	if err != nil {
		return fmt.Errorf("set whitelist: %w", err)
	}
	return nil
}

// Update updates domain fields (value, type, whitelist only)
func (r *domainRepository) Update(ctx context.Context, domain *models.Domain) error {
	_, err := r.db.Exec(ctx, `
		UPDATE core.domains SET value = $2, type = $3, whitelist = $4 WHERE id = $1
	`, domain.ID, domain.Value, domain.Type, domain.Whitelist)
	if err != nil {
		return fmt.Errorf("update domain: %w", err)
	}
	return nil
}

// GetByValueAndType finds a domain by value and type (with records)
func (r *domainRepository) GetByValueAndType(ctx context.Context, value, typ string) (*models.Domain, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, value, type, whitelist FROM core.domains WHERE value = $1 AND type = $2
	`, value, typ)
	var d models.Domain
	if err := row.Scan(&d.ID, &d.Value, &d.Type, &d.Whitelist); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get by value/type: %w", err)
	}
	records, err := r.getRecordsByDomainID(ctx, d.ID)
	if err != nil {
		return nil, err
	}
	d.Records = records
	return &d, nil
}
