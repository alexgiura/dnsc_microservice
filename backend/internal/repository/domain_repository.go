package repository

import (
	"context"
	"database/sql"
	"dnsc_microservice/internal/models"
	"errors"
	"fmt"
	"strings"
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
	SetDomainStatusWithHistory(ctx context.Context, id uuid.UUID, status string, changedBy, notes string) error
	CreateWhitelistRequest(ctx context.Context, request *models.WhitelistRequest) (*models.WhitelistRequest, error)
	Update(ctx context.Context, domain *models.Domain) error
	GetByValueAndType(ctx context.Context, value, typ string) (*models.Domain, error)
	InsertRecords(ctx context.Context, domainID uuid.UUID, records []models.DomainRecord) error
	// AllDomainsRejectedForTicketID is true when there is at least one domain_record for the ticket
	// and every distinct domain linked to that ticket_id has status rejected.
	AllDomainsRejectedForTicketID(ctx context.Context, ticketID string) (bool, error)
	FindAutoWhitelistCandidateDomainIDs(ctx context.Context, cutoff time.Time) ([]uuid.UUID, error)

	GetLastRTIRSync(ctx context.Context) (*time.Time, error)
	UpsertRTIRDomainRecord(ctx context.Context, rec models.DomainRTIRRecord) error

	ListDomainIDsForPNRISCSync(ctx context.Context, limit int) ([]uuid.UUID, error)
	GetPNRISCSyncPayload(ctx context.Context, domainID uuid.UUID) (*models.PNRISCDomainPayload, error)
	MarkPNRISCSyncSuccess(ctx context.Context, domainID uuid.UUID, remoteID *string, syncedAt time.Time) error
	MarkPNRISCSyncFailure(ctx context.Context, domainID uuid.UUID, errMsg string) error

	UpsertRTIRImportError(ctx context.Context, ticketID, source, errorMessage string, ticketDate time.Time) error
	DeleteRTIRImportError(ctx context.Context, ticketID string) error
	ListRTIRImportErrors(ctx context.Context) ([]models.RTIRImportError, error)

	// ListDistinctTicketIDsFromDomainRecords returns non-empty distinct ticket_id values from core.domain_records.
	ListDistinctTicketIDsFromDomainRecords(ctx context.Context) ([]string, error)
	// UpdateDomainRecordsDateByTicketID sets date for all rows with the given ticket_id; returns pgx RowsAffected().
	UpdateDomainRecordsDateByTicketID(ctx context.Context, ticketID string, date time.Time) (int64, error)

	GetDashboard(ctx context.Context) (*models.DashboardResponse, error)
}

type domainRepository struct {
	db *pgxpool.Pool
}

func domainDescriptionFromNull(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
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
		INSERT INTO core.domains (id, value, type, status, description)
		VALUES ($1, $2, $3, $4, $5)
	`, domain.ID, domain.Value, domain.Type, domain.Status, domain.Description)
	if err != nil {
		return fmt.Errorf("insert domain: %w", err)
	}

	if len(domain.Records) > 0 {
		if err := r.insertRecordsTx(ctx, tx, domain.ID, domain.Records); err != nil {
			return err
		}
	}

	// Initial status history for newly created blacklisted domains.
	if domain.Status == models.DomainStatusBlacklist {
		_, err = tx.Exec(ctx, `
			INSERT INTO core.domain_status (id, domain_id, status, changed_by, notes)
			VALUES ($1, $2, $3, $4, $5)
		`, uuid.New(), domain.ID, models.DomainStatusBlacklist, "system", "first record")
		if err != nil {
			return fmt.Errorf("insert initial domain_status: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// insertRecordsTx inserts domain_records inside an existing transaction
func (r *domainRepository) insertRecordsTx(ctx context.Context, tx pgx.Tx, domainID uuid.UUID, records []models.DomainRecord) error {
	for _, rec := range records {
		var syncAt interface{}
		if rec.LastSuccessfulSyncAt != nil {
			syncAt = *rec.LastSuccessfulSyncAt
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO core.domain_records (id, domain_id, ticket_id, description, tags, date, source, last_successful_sync_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, rec.ID, domainID, rec.TicketID, rec.Description, rec.Tags, rec.Date, rec.Source, syncAt)
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
		SELECT id, value, type, status, description FROM core.domains WHERE id = $1
	`, id)
	var d models.Domain
	var descNS sql.NullString
	if err := row.Scan(&d.ID, &d.Value, &d.Type, &d.Status, &descNS); err != nil {
		return nil, fmt.Errorf("get domain by id: %w", err)
	}
	d.Description = domainDescriptionFromNull(descNS)
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
		SELECT id, domain_id, ticket_id, description, tags, date, source, last_successful_sync_at
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
		var syncAt sql.NullTime
		if err := rows.Scan(&rec.ID, &rec.DomainID, &rec.TicketID, &rec.Description, &rec.Tags, &rec.Date, &rec.Source, &syncAt); err != nil {
			return nil, fmt.Errorf("scan domain_record: %w", err)
		}
		if syncAt.Valid {
			t := syncAt.Time
			rec.LastSuccessfulSyncAt = &t
		}
		list = append(list, rec)
	}
	return list, rows.Err()
}

func (r *domainRepository) AllDomainsRejectedForTicketID(ctx context.Context, ticketID string) (bool, error) {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return false, nil
	}
	var ok bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM core.domain_records WHERE ticket_id = $1
		) AND NOT EXISTS (
			SELECT 1
			FROM core.domain_records dr
			INNER JOIN core.domains d ON d.id = dr.domain_id
			WHERE dr.ticket_id = $1 AND d.status <> $2
		)
	`, ticketID, models.DomainStatusRejected).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("all domains rejected for ticket_id: %w", err)
	}
	return ok, nil
}

func (r *domainRepository) getStatusHistoryByDomainID(ctx context.Context, domainID uuid.UUID) ([]models.DomainStatus, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, domain_id, status, changed_at, changed_by, notes
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
		if err := rows.Scan(&entry.ID, &entry.DomainID, &entry.Status, &entry.ChangedAt, &entry.ChangedBy, &entry.Notes); err != nil {
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
		SELECT id, value, type, status, description FROM core.domains ORDER BY last_updated DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}
	defer rows.Close()
	var domains []*models.Domain
	for rows.Next() {
		var d models.Domain
		var descNS sql.NullString
		if err := rows.Scan(&d.ID, &d.Value, &d.Type, &d.Status, &descNS); err != nil {
			return nil, fmt.Errorf("scan domain: %w", err)
		}
		d.Description = domainDescriptionFromNull(descNS)
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
			MAX(ds.changed_at) FILTER (WHERE ds.status = 'blacklist') AS last_blacklisted_at
		FROM core.domains d
		LEFT JOIN core.domain_status ds ON ds.domain_id = d.id
		WHERE d.status = 'blacklist'
		GROUP BY d.id, d.value, d.type
		HAVING MAX(ds.changed_at) FILTER (WHERE ds.status = 'blacklist') IS NOT NULL
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

// SetDomainStatusWithHistory updates core.domains.status and appends core.domain_status.
func (r *domainRepository) SetDomainStatusWithHistory(ctx context.Context, id uuid.UUID, status string, changedBy, notes string) error {
	if _, err := models.ParseDomainStatus(status); err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE core.domains SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("update domain status: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO core.domain_status (id, domain_id, status, changed_by, notes)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), id, status, changedBy, notes)
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

// Update updates domain fields (value, type, status, description)
func (r *domainRepository) Update(ctx context.Context, domain *models.Domain) error {
	_, err := r.db.Exec(ctx, `
		UPDATE core.domains SET value = $2, type = $3, status = $4, description = $5 WHERE id = $1
	`, domain.ID, domain.Value, domain.Type, domain.Status, domain.Description)
	if err != nil {
		return fmt.Errorf("update domain: %w", err)
	}
	return nil
}

// GetByValueAndType finds a domain by value and type (with records)
func (r *domainRepository) GetByValueAndType(ctx context.Context, value, typ string) (*models.Domain, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, value, type, status, description FROM core.domains WHERE value = $1 AND type = $2
	`, value, typ)
	var d models.Domain
	var descNS sql.NullString
	if err := row.Scan(&d.ID, &d.Value, &d.Type, &d.Status, &descNS); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get by value/type: %w", err)
	}
	d.Description = domainDescriptionFromNull(descNS)
	records, err := r.getRecordsByDomainID(ctx, d.ID)
	if err != nil {
		return nil, err
	}
	d.Records = records
	return &d, nil
}

// FindAutoWhitelistCandidateDomainIDs selects domain IDs where status is blacklist,
// and the maximum domain_records.date is <= cutoff or there are no records.
func (r *domainRepository) FindAutoWhitelistCandidateDomainIDs(ctx context.Context, cutoff time.Time) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `
		SELECT d.id
		FROM core.domains d
		LEFT JOIN core.domain_records r ON r.domain_id = d.id
		WHERE d.status = 'blacklist'
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

func (r *domainRepository) GetLastRTIRSync(ctx context.Context) (*time.Time, error) {
	row := r.db.QueryRow(ctx, `
		SELECT MAX(last_successful_sync_at) FROM core.domain_records WHERE last_successful_sync_at IS NOT NULL
	`)
	var nt sql.NullTime
	if err := row.Scan(&nt); err != nil {
		return nil, fmt.Errorf("get last rtir sync: %w", err)
	}
	if !nt.Valid {
		return nil, nil
	}
	return &nt.Time, nil
}

func (r *domainRepository) UpsertRTIRDomainRecord(ctx context.Context, rec models.DomainRTIRRecord) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var recID, domID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT dr.id, dr.domain_id
		FROM core.domain_records dr
		INNER JOIN core.domains d ON d.id = dr.domain_id
		WHERE dr.ticket_id = $1 AND d.value = $2
	`, rec.TicketID, rec.Value).Scan(&recID, &domID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("lookup domain_record by ticket: %w", err)
	}

	if err == nil {
		_, err = tx.Exec(ctx, `
			UPDATE core.domains SET value = $2, type = $3 WHERE id = $1
		`, domID, rec.Value, rec.Type)
		if err != nil {
			return fmt.Errorf("update domain from rtir: %w", err)
		}
		_, err = tx.Exec(ctx, `
			UPDATE core.domain_records
			SET description = $2, tags = $3, date = $4, source = $5, last_successful_sync_at = $6
			WHERE id = $1
		`, recID, rec.Description, rec.Tags, rec.RecordDate, "rtir", rec.LastSuccessfulSyncAt)
		if err != nil {
			return fmt.Errorf("update domain_record from rtir: %w", err)
		}
		return tx.Commit(ctx)
	}

	var domainID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id FROM core.domains WHERE value = $1 AND type = $2
	`, rec.Value, rec.Type).Scan(&domainID)
	if errors.Is(err, pgx.ErrNoRows) {
		domainID = uuid.New()
		_, err = tx.Exec(ctx, `
			INSERT INTO core.domains (id, value, type, status, description)
			VALUES ($1, $2, $3, $4, $5)
		`, domainID, rec.Value, rec.Type, models.DomainStatusPending, rec.Description)
		if err != nil {
			return fmt.Errorf("insert domain from rtir: %w", err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO core.domain_status (id, domain_id, status, changed_by, notes)
			VALUES ($1, $2, $3, $4, $5)
		`, uuid.New(), domainID, models.DomainStatusPending, "rtir-sync", "first record")
		if err != nil {
			return fmt.Errorf("insert initial domain_status from rtir: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("lookup domain by value: %w", err)
	}

	recordID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO core.domain_records (id, domain_id, ticket_id, description, tags, date, source, last_successful_sync_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, recordID, domainID, rec.TicketID, rec.Description, rec.Tags, rec.RecordDate, "rtir", rec.LastSuccessfulSyncAt)
	if err != nil {
		return fmt.Errorf("insert domain_record from rtir: %w", err)
	}

	return tx.Commit(ctx)
}

const pnriscErrStatusMax = 500

func (r *domainRepository) ListDomainIDsForPNRISCSync(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if limit <= 0 {
		limit = 500
	}
	rows, err := r.db.Query(ctx, `
		SELECT id FROM core.domains
		WHERE (pnrisc_last_synced_at IS NULL OR last_updated > pnrisc_last_synced_at)
		  AND status IN ($2, $3)
		ORDER BY last_updated ASC
		LIMIT $1
	`, limit, models.DomainStatusWhitelist, models.DomainStatusBlacklist)
	if err != nil {
		return nil, fmt.Errorf("list domains for pnrisc sync: %w", err)
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan domain id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *domainRepository) GetPNRISCSyncPayload(ctx context.Context, domainID uuid.UUID) (*models.PNRISCDomainPayload, error) {
	row := r.db.QueryRow(ctx, `
		SELECT value, type, status, description FROM core.domains WHERE id = $1
	`, domainID)
	var p models.PNRISCDomainPayload
	var domainDesc sql.NullString
	p.DomainID = domainID
	if err := row.Scan(&p.Value, &p.Type, &p.Status, &domainDesc); err != nil {
		return nil, fmt.Errorf("get domain for pnrisc: %w", err)
	}

	var dateAdded sql.NullTime
	if err := r.db.QueryRow(ctx, `
		SELECT MAX(changed_at) FROM core.domain_status
		WHERE domain_id = $1 AND status = 'blacklist'
	`, domainID).Scan(&dateAdded); err != nil {
		return nil, fmt.Errorf("pnrisc date_added: %w", err)
	}
	if dateAdded.Valid {
		t := dateAdded.Time
		p.DateAdded = &t
	}

	// PNRISC "reason" = doar descrierea din core.domains (fără fallback la domain_records).
	if domainDesc.Valid {
		p.Reason = strings.TrimSpace(domainDesc.String)
	}

	return &p, nil
}

func (r *domainRepository) MarkPNRISCSyncSuccess(ctx context.Context, domainID uuid.UUID, remoteID *string, syncedAt time.Time) error {
	var rid interface{}
	if remoteID != nil && *remoteID != "" {
		rid = *remoteID
	}
	_, err := r.db.Exec(ctx, `
		UPDATE core.domains SET
			pnrisc_last_synced_at = $2,
			pnrisc_sync_status = $3,
			pnrisc_remote_id = COALESCE($4, pnrisc_remote_id)
		WHERE id = $1
	`, domainID, syncedAt, "ok", rid)
	if err != nil {
		return fmt.Errorf("mark pnrisc sync success: %w", err)
	}
	return nil
}

func (r *domainRepository) MarkPNRISCSyncFailure(ctx context.Context, domainID uuid.UUID, errMsg string) error {
	if len(errMsg) > pnriscErrStatusMax {
		errMsg = errMsg[:pnriscErrStatusMax]
	}
	_, err := r.db.Exec(ctx, `
		UPDATE core.domains SET pnrisc_sync_status = $2 WHERE id = $1
	`, domainID, "error: "+errMsg)
	if err != nil {
		return fmt.Errorf("mark pnrisc sync failure: %w", err)
	}
	return nil
}

const rtirImportErrMsgMax = 8000

func (r *domainRepository) UpsertRTIRImportError(ctx context.Context, ticketID, source, errorMessage string, ticketDate time.Time) error {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return fmt.Errorf("ticket id is empty")
	}
	if len(errorMessage) > rtirImportErrMsgMax {
		errorMessage = errorMessage[:rtirImportErrMsgMax]
	}
	if strings.TrimSpace(source) == "" {
		source = "rtir-sync"
	}
	if ticketDate.IsZero() {
		ticketDate = time.Now().UTC()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO core.rtir_import_errors (id, ticket_id, source, error_message, date, last_sync_try_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (ticket_id) DO UPDATE SET
			source = EXCLUDED.source,
			error_message = EXCLUDED.error_message,
			date = EXCLUDED.date,
			last_sync_try_at = now()
	`, uuid.New(), ticketID, source, errorMessage, ticketDate)
	if err != nil {
		return fmt.Errorf("upsert rtir import error: %w", err)
	}
	return nil
}

func (r *domainRepository) DeleteRTIRImportError(ctx context.Context, ticketID string) error {
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return nil
	}
	_, err := r.db.Exec(ctx, `DELETE FROM core.rtir_import_errors WHERE ticket_id = $1`, ticketID)
	if err != nil {
		return fmt.Errorf("delete rtir import error: %w", err)
	}
	return nil
}

func (r *domainRepository) ListRTIRImportErrors(ctx context.Context) ([]models.RTIRImportError, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, ticket_id, source, error_message, "date", last_sync_try_at
		FROM core.rtir_import_errors
		ORDER BY last_sync_try_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list rtir import errors: %w", err)
	}
	defer rows.Close()
	var list []models.RTIRImportError
	for rows.Next() {
		var e models.RTIRImportError
		if err := rows.Scan(&e.ID, &e.TicketID, &e.Source, &e.ErrorMessage, &e.Date, &e.LastSyncTryAt); err != nil {
			return nil, fmt.Errorf("scan rtir import error: %w", err)
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *domainRepository) GetDashboard(ctx context.Context) (*models.DashboardResponse, error) {
	out := &models.DashboardResponse{
		RecentRecords:      []models.DashboardRecentRecord{},
		TopTags:            []models.DashboardTagCount{},
		BlacklistFollowUps: []models.DashboardBlacklistFollowUp{},
	}

	row := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE status = 'blacklist')::int,
			COUNT(*) FILTER (WHERE status = 'whitelist')::int,
			COUNT(*) FILTER (WHERE status = 'pending')::int,
			COUNT(*) FILTER (WHERE status = 'rejected')::int
		FROM core.domains
	`)
	if err := row.Scan(
		&out.Counts.Total,
		&out.Counts.Blacklist,
		&out.Counts.Whitelist,
		&out.Counts.Pending,
		&out.Counts.Rejected,
	); err != nil {
		return nil, fmt.Errorf("dashboard counts: %w", err)
	}

	recRows, err := r.db.Query(ctx, `
		SELECT dr.id, dr.domain_id, d.value, d.type, d.status, dr.ticket_id, dr.date, dr.source
		FROM core.domain_records dr
		INNER JOIN core.domains d ON d.id = dr.domain_id
		ORDER BY dr.date DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("dashboard recent records: %w", err)
	}
	defer recRows.Close()
	for recRows.Next() {
		var rec models.DashboardRecentRecord
		var srcNS sql.NullString
		if err := recRows.Scan(&rec.ID, &rec.DomainID, &rec.Value, &rec.Type, &rec.Status, &rec.TicketID, &rec.Date, &srcNS); err != nil {
			return nil, fmt.Errorf("scan dashboard record: %w", err)
		}
		if srcNS.Valid {
			rec.Source = strings.TrimSpace(srcNS.String)
		}
		out.RecentRecords = append(out.RecentRecords, rec)
	}
	if err := recRows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard recent records: %w", err)
	}

	followRows, err := r.db.Query(ctx, `
		SELECT d.id, d.value, d.type, d.status, lb.blacklisted_at,
			(
				SELECT COUNT(*)::int
				FROM core.domain_records dr
				WHERE dr.domain_id = d.id AND dr.date > lb.blacklisted_at
			) AS reports_after
		FROM core.domains d
		INNER JOIN LATERAL (
			SELECT MAX(ds.changed_at) AS blacklisted_at
			FROM core.domain_status ds
			WHERE ds.domain_id = d.id AND ds.status = 'blacklist'
		) lb ON lb.blacklisted_at IS NOT NULL
		WHERE d.status = 'blacklist'
			AND EXISTS (
				SELECT 1 FROM core.domain_records dr2
				WHERE dr2.domain_id = d.id AND dr2.date > lb.blacklisted_at
			)
		ORDER BY lb.blacklisted_at DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, fmt.Errorf("dashboard blacklist follow-ups: %w", err)
	}
	defer followRows.Close()
	for followRows.Next() {
		var row models.DashboardBlacklistFollowUp
		if err := followRows.Scan(
			&row.DomainID,
			&row.Value,
			&row.Type,
			&row.Status,
			&row.BlacklistedAt,
			&row.ReportsAfterBlacklist,
		); err != nil {
			return nil, fmt.Errorf("scan blacklist follow-up: %w", err)
		}
		out.BlacklistFollowUps = append(out.BlacklistFollowUps, row)
	}
	if err := followRows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard blacklist follow-ups: %w", err)
	}

	tagRows, err := r.db.Query(ctx, `
		SELECT trim(both from t.tag) AS tag, COUNT(*)::int AS cnt
		FROM core.domain_records dr
		CROSS JOIN LATERAL unnest(dr.tags) AS t(tag)
		WHERE dr.tags IS NOT NULL AND cardinality(dr.tags) > 0
		GROUP BY trim(both from t.tag)
		HAVING trim(both from t.tag) != ''
		ORDER BY cnt DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("dashboard top tags: %w", err)
	}
	defer tagRows.Close()
	for tagRows.Next() {
		var tc models.DashboardTagCount
		if err := tagRows.Scan(&tc.Tag, &tc.Count); err != nil {
			return nil, fmt.Errorf("scan dashboard tag: %w", err)
		}
		out.TopTags = append(out.TopTags, tc)
	}
	if err := tagRows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard top tags: %w", err)
	}

	return out, nil
}

func (r *domainRepository) ListDistinctTicketIDsFromDomainRecords(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT trim(both from ticket_id)
		FROM core.domain_records
		WHERE ticket_id IS NOT NULL AND trim(both from ticket_id) <> ''
		ORDER BY 1
	`)
	if err != nil {
		return nil, fmt.Errorf("list distinct domain_records ticket ids: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan ticket id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *domainRepository) UpdateDomainRecordsDateByTicketID(ctx context.Context, ticketID string, date time.Time) (int64, error) {
	cmd, err := r.db.Exec(ctx, `
		UPDATE core.domain_records SET date = $1 WHERE trim(both from ticket_id) = $2
	`, date.UTC(), strings.TrimSpace(ticketID))
	if err != nil {
		return 0, fmt.Errorf("update domain_records date by ticket: %w", err)
	}
	return cmd.RowsAffected(), nil
}
