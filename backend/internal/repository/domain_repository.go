package repository

import (
	"context"
	"dnsc_microservice/internal/models"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// DomainRepository defines the interface for domain and domain_records persistence
type DomainRepository interface {
	Insert(ctx context.Context, domain *models.Domain) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Domain, error)
	List(ctx context.Context) ([]*models.Domain, error)
	ListPublicBlacklisted(ctx context.Context) ([]*models.PublicDomain, error)
	SetWhitelist(ctx context.Context, id uuid.UUID, whitelist bool) error
	SetWhitelistWithStatus(ctx context.Context, id uuid.UUID, whitelist bool, changedBy, notes string) error
	CreateWhitelistRequest(ctx context.Context, request *models.WhitelistRequest) (*models.WhitelistRequest, error)
	Update(ctx context.Context, domain *models.Domain) error
	GetByValueAndType(ctx context.Context, value, typ string) (*models.Domain, error)
	InsertRecords(ctx context.Context, domainID uuid.UUID, records []models.DomainRecord) error
	FindAutoWhitelistCandidateDomainIDs(ctx context.Context, cutoff time.Time) ([]uuid.UUID, error)
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

	// Initial status history for newly created blacklisted domains.
	// This runs inside the same transaction so all inserts roll back together on failure.
	if !domain.Whitelist {
		_, err = tx.Exec(ctx, `
			INSERT INTO core.domain_status (id, domain_id, whitelist, changed_by, notes)
			VALUES ($1, $2, $3, $4, $5)
		`, uuid.New(), domain.ID, false, "system", "first record")
		if err != nil {
			return fmt.Errorf("insert initial domain_status: %w", err)
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

	statusHistory, err := r.getStatusHistoryByDomainID(ctx, id)
	if err != nil {
		return nil, err
	}
	d.StatusHistory = statusHistory

	whitelistRequests, err := r.getWhitelistRequestsByDomainID(ctx, id)
	if err != nil {
		return nil, err
	}
	d.WhitelistRequests = whitelistRequests
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

func (r *domainRepository) getStatusHistoryByDomainID(ctx context.Context, domainID uuid.UUID) ([]models.DomainStatus, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, domain_id, whitelist, changed_at, changed_by, notes
		FROM core.domain_status
		WHERE domain_id = $1
		ORDER BY changed_at DESC
	`, domainID)
	if err != nil {
		return nil, fmt.Errorf("get domain_status: %w", err)
	}
	defer rows.Close()

	var history []models.DomainStatus
	for rows.Next() {
		var entry models.DomainStatus
		if err := rows.Scan(&entry.ID, &entry.DomainID, &entry.Whitelist, &entry.ChangedAt, &entry.ChangedBy, &entry.Notes); err != nil {
			return nil, fmt.Errorf("scan domain_status: %w", err)
		}
		history = append(history, entry)
	}
	return history, rows.Err()
}

func (r *domainRepository) getWhitelistRequestsByDomainID(ctx context.Context, domainID uuid.UUID) ([]models.WhitelistRequest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, domain_id, first_name, last_name, email, address, phone, reason, created_at
		FROM core.whitelist_requests
		WHERE domain_id = $1
		ORDER BY created_at DESC
	`, domainID)
	if err != nil {
		return nil, fmt.Errorf("get whitelist_requests: %w", err)
	}
	defer rows.Close()

	var requests []models.WhitelistRequest
	for rows.Next() {
		var req models.WhitelistRequest
		if err := rows.Scan(&req.ID, &req.DomainID, &req.FirstName, &req.LastName, &req.Email, &req.Address, &req.Phone, &req.Reason, &req.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan whitelist_request: %w", err)
		}
		requests = append(requests, req)
	}
	return requests, rows.Err()
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

		statusHistory, err := r.getStatusHistoryByDomainID(ctx, d.ID)
		if err != nil {
			return nil, err
		}
		d.StatusHistory = statusHistory

		whitelistRequests, err := r.getWhitelistRequestsByDomainID(ctx, d.ID)
		if err != nil {
			return nil, err
		}
		d.WhitelistRequests = whitelistRequests
		domains = append(domains, &d)
	}
	return domains, rows.Err()
}

func (r *domainRepository) ListPublicBlacklisted(ctx context.Context) ([]*models.PublicDomain, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			d.value,
			d.type,
			MAX(ds.changed_at) FILTER (WHERE ds.whitelist = false) AS last_blacklisted_at
		FROM core.domains d
		LEFT JOIN core.domain_status ds ON ds.domain_id = d.id
		WHERE d.whitelist = false
		GROUP BY d.id, d.value, d.type
		HAVING MAX(ds.changed_at) FILTER (WHERE ds.whitelist = false) IS NOT NULL
		ORDER BY last_blacklisted_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list public blacklisted domains: %w", err)
	}
	defer rows.Close()

	var list []*models.PublicDomain
	for rows.Next() {
		var item models.PublicDomain
		if err := rows.Scan(&item.Value, &item.Type, &item.Date); err != nil {
			return nil, fmt.Errorf("scan public blacklisted domain: %w", err)
		}
		list = append(list, &item)
	}
	return list, rows.Err()
}

// SetWhitelist updates the whitelist flag for a domain
func (r *domainRepository) SetWhitelist(ctx context.Context, id uuid.UUID, whitelist bool) error {
	_, err := r.db.Exec(ctx, `UPDATE core.domains SET whitelist = $2 WHERE id = $1`, id, whitelist)
	if err != nil {
		return fmt.Errorf("set whitelist: %w", err)
	}
	return nil
}

// SetWhitelistWithStatus atomically updates core.domains.whitelist and inserts a row in core.domain_status.
func (r *domainRepository) SetWhitelistWithStatus(ctx context.Context, id uuid.UUID, whitelist bool, changedBy, notes string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE core.domains SET whitelist = $2 WHERE id = $1`, id, whitelist)
	if err != nil {
		return fmt.Errorf("update whitelist: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO core.domain_status (id, domain_id, whitelist, changed_by, notes)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), id, whitelist, changedBy, notes)
	if err != nil {
		return fmt.Errorf("insert domain_status: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *domainRepository) CreateWhitelistRequest(ctx context.Context, request *models.WhitelistRequest) (*models.WhitelistRequest, error) {
	if request == nil {
		return nil, fmt.Errorf("whitelist request is nil")
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO core.whitelist_requests (
			id, domain_id, first_name, last_name, email, address, phone, reason
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`,
		request.ID,
		request.DomainID,
		request.FirstName,
		request.LastName,
		request.Email,
		request.Address,
		request.Phone,
		request.Reason,
	)

	if err := row.Scan(&request.CreatedAt); err != nil {
		return nil, fmt.Errorf("create whitelist request: %w", err)
	}

	return request, nil
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

// FindAutoWhitelistCandidateDomainIDs selects domain IDs where:
//   - domain.whitelist = false
//   - and the maximum domain_records.date is <= cutoff
//     OR there are no domain_records at all (MAX(...) IS NULL).
func (r *domainRepository) FindAutoWhitelistCandidateDomainIDs(ctx context.Context, cutoff time.Time) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `
		SELECT d.id
		FROM core.domains d
		LEFT JOIN core.domain_records r ON r.domain_id = d.id
		WHERE d.whitelist = false
		GROUP BY d.id
		HAVING MAX(r.date) IS NULL OR MAX(r.date) <= $1
	`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("find auto whitelist candidate domain ids: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan auto whitelist candidate domain id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
