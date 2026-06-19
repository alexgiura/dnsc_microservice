package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"time"

	"dnsc_microservice/internal/clients/pnrisc"
	"dnsc_microservice/internal/clients/rtir"
	"dnsc_microservice/internal/mappers"
	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/repository"

	"github.com/google/uuid"
)

const rtirImportSource = "rtir-sync"

// ErrRejectNotPending is returned when status change to rejected is not from pending.
var ErrRejectNotPending = errors.New("reject is only allowed from pending status")

// ErrRejectNoTicketID is returned when the domain has no ticket_id on its records to update in RT.
var ErrRejectNoTicketID = errors.New("no ticket ID on this domain: rejecting requires an associated RT ticket")

// DomainService defines the interface for domain business logic
type DomainService interface {
	SaveDomain(ctx context.Context, input models.SaveDomainInput) (*models.Domain, error)
	GetDomainByID(ctx context.Context, id uuid.UUID) (*models.Domain, error)
	GetDomains(ctx context.Context) ([]*models.Domain, error)
	GetPublicBlacklistedDomains(ctx context.Context) ([]*models.PublicDomain, error)
	GetDashboard(ctx context.Context) (*models.DashboardResponse, error)
	ChangeDomainStatus(ctx context.Context, id uuid.UUID, status string, changedBy, notes string) error
	ChangeDomainStatusBulk(ctx context.Context, input models.BulkChangeDomainStatusInput) (*models.BulkChangeDomainStatusResult, error)
	RequestWhitelist(ctx context.Context, domainID uuid.UUID, input models.CreateWhitelistRequestInput) (*models.WhitelistRequest, error)
	AutoWhitelistStaleDomains(ctx context.Context, cutoff time.Time, changedBy, notes string) error
	UpdateDomain(ctx context.Context, id uuid.UUID, input models.UpdateDomainInput) (*models.Domain, error)
	// SyncRTIRDomains searches RTIR tickets, loads each by ID, upserts domains + checkpoint.
	SyncRTIRDomains(ctx context.Context) error
	// SyncPNRISCDomains POSTs domains whose last_updated is newer than pnrisc_last_synced_at (only whitelist/blacklist), then marks sync.
	SyncPNRISCDomains(ctx context.Context) error
	// TryReimportRTIRTicket GETs RTIR ticket by id and runs the same upsert path as the periodic sync.
	TryReimportRTIRTicket(ctx context.Context, ticketID string) error
	// GetRTIRImportErrors returns rows from core.rtir_import_errors (failed ticket syncs).
	GetRTIRImportErrors(ctx context.Context) ([]models.RTIRImportError, error)
}

type domainService struct {
	repo   repository.DomainRepository
	rtir   *rtir.Client
	pnrisc *pnrisc.Client

	rtirTz      string
	rtirOverlap time.Duration
}

// NewDomainService creates a new domain service. rtir may be nil only if RTIR sync is never used; pnrisc may be nil if PNRISC sync is disabled.
func NewDomainService(repo repository.DomainRepository, rtirClient *rtir.Client, pnriscClient *pnrisc.Client, rtirSyncTimezone string, rtirOverlapMinutes int) DomainService {
	overlap := time.Duration(rtirOverlapMinutes) * time.Minute
	if rtirOverlapMinutes <= 0 {
		overlap = 5 * time.Minute
	}
	return &domainService{
		repo:        repo,
		rtir:        rtirClient,
		pnrisc:      pnriscClient,
		rtirTz:      rtirSyncTimezone,
		rtirOverlap: overlap,
	}
}

func domainTypeFromValue(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return models.DomainTypeDomain
	}
	if net.ParseIP(v) != nil {
		return models.DomainTypeIP
	}
	v = strings.TrimSuffix(v, ".")
	if strings.Count(v, ".") >= 2 {
		return models.DomainTypeSubdomain
	}
	return models.DomainTypeDomain
}

// SaveDomain: if no domain with same value+type exists, insert new domain and records; otherwise append input records to existing domain
func (s *domainService) SaveDomain(ctx context.Context, input models.SaveDomainInput) (*models.Domain, error) {
	typ := domainTypeFromValue(input.Value)
	existing, err := s.repo.GetByValueAndType(ctx, input.Value, typ)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		st := strings.TrimSpace(input.Status)
		if st == "" {
			st = models.DomainStatusPending
		}
		if _, err := models.ParseDomainStatus(st); err != nil {
			return nil, err
		}
		domainDescription := ""
		if len(input.Records) > 0 {
			domainDescription = strings.TrimSpace(input.Records[0].Description)
		}
		domain := &models.Domain{
			ID:          uuid.New(),
			Value:       input.Value,
			Type:        typ,
			Status:      st,
			Description: domainDescription,
			Records:     nil,
		}
		for _, r := range input.Records {
			domain.Records = append(domain.Records, models.DomainRecord{
				ID:          uuid.New(),
				DomainID:    domain.ID,
				TicketID:    r.TicketID,
				Description: r.Description,
				Tags:        r.Tags,
				Date:        r.Date,
				Source:      r.Source,
			})
		}
		if err := s.repo.Insert(ctx, domain); err != nil {
			return nil, err
		}
		return domain, nil
	}

	// append new records to existing domain
	if len(input.Records) > 0 {
		newRecs := make([]models.DomainRecord, 0, len(input.Records))
		for _, r := range input.Records {
			newRecs = append(newRecs, models.DomainRecord{
				ID:          uuid.New(),
				DomainID:    existing.ID,
				TicketID:    r.TicketID,
				Description: r.Description,
				Tags:        r.Tags,
				Date:        r.Date,
				Source:      r.Source,
			})
		}
		if err := s.repo.InsertRecords(ctx, existing.ID, newRecs); err != nil {
			return nil, err
		}
		existing.Records = append(existing.Records, newRecs...)
	}
	return existing, nil
}

// GetDomainByID retrieves a domain by ID (with records)
func (s *domainService) GetDomainByID(ctx context.Context, id uuid.UUID) (*models.Domain, error) {
	return s.repo.GetByID(ctx, id)
}

// GetDomains retrieves all domains with their records
func (s *domainService) GetDomains(ctx context.Context) ([]*models.Domain, error) {
	return s.repo.List(ctx)
}

func (s *domainService) GetPublicBlacklistedDomains(ctx context.Context) ([]*models.PublicDomain, error) {
	return s.repo.ListPublicBlacklisted(ctx)
}

// GetDashboard returns aggregated counts, latest domain records, and top tags for the admin dashboard.
func (s *domainService) GetDashboard(ctx context.Context) (*models.DashboardResponse, error) {
	return s.repo.GetDashboard(ctx)
}

// pickTicketIDForRTReject returns the RT ticket id to update when rejecting: newest record with source rtir, else newest with any non-empty ticket_id.
func pickTicketIDForRTReject(d *models.Domain) string {
	if d == nil || len(d.Records) == 0 {
		return ""
	}
	type rec struct {
		ticketID string
		source   string
		date     time.Time
	}
	var list []rec
	for _, r := range d.Records {
		tid := strings.TrimSpace(r.TicketID)
		if tid == "" {
			continue
		}
		list = append(list, rec{tid, r.Source, r.Date})
	}
	if len(list) == 0 {
		return ""
	}
	sort.Slice(list, func(i, j int) bool { return list[i].date.After(list[j].date) })
	for _, x := range list {
		src := strings.TrimSpace(strings.ToLower(x.source))
		if src == "rtir" || src == "rtir-sync" {
			return x.ticketID
		}
	}
	return list[0].ticketID
}

func (s *domainService) ChangeDomainStatus(ctx context.Context, id uuid.UUID, status string, changedBy, notes string) error {
	st, err := models.ParseDomainStatus(status)
	if err != nil {
		return err
	}
	if st == models.DomainStatusRejected {
		d, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if d.Status != models.DomainStatusPending {
			return ErrRejectNotPending
		}
		ticketID := pickTicketIDForRTReject(d)
		if ticketID == "" {
			return ErrRejectNoTicketID
		}
		if err := s.repo.SetDomainStatusWithHistory(ctx, id, st, changedBy, notes); err != nil {
			return err
		}
		allRejected, err := s.repo.AllDomainsRejectedForTicketID(ctx, ticketID)
		if err != nil {
			return err
		}
		if !allRejected {
			return nil
		}
		if s.rtir == nil {
			log.Printf("domain reject: ticket %s has all domains rejected but RTIR is not configured; skipping RT blacklist update", ticketID)
			return nil
		}
		if err := s.rtir.UpdateTicketBlacklistNo(ctx, ticketID); err != nil {
			return fmt.Errorf("RT ticket update failed: %w", err)
		}
		return nil
	}
	return s.repo.SetDomainStatusWithHistory(ctx, id, st, changedBy, notes)
}

func (s *domainService) ChangeDomainStatusBulk(ctx context.Context, input models.BulkChangeDomainStatusInput) (*models.BulkChangeDomainStatusResult, error) {
	if len(input.DomainIDs) == 0 {
		return nil, fmt.Errorf("domainIds is required")
	}
	st, err := models.ParseDomainStatus(strings.TrimSpace(input.Status))
	if err != nil {
		return nil, err
	}
	changeBy := strings.TrimSpace(input.ChangeBy)
	if changeBy == "" {
		return nil, fmt.Errorf("changeBy is required")
	}
	notes := strings.TrimSpace(input.Notes)

	result := &models.BulkChangeDomainStatusResult{
		Succeeded: []uuid.UUID{},
		Failed:    []models.BulkChangeDomainStatusFailure{},
	}
	seen := make(map[uuid.UUID]struct{}, len(input.DomainIDs))

	for _, idStr := range input.DomainIDs {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			result.Failed = append(result.Failed, models.BulkChangeDomainStatusFailure{
				ID:    idStr,
				Error: "invalid id format",
			})
			continue
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			result.Failed = append(result.Failed, models.BulkChangeDomainStatusFailure{
				ID:    idStr,
				Error: "invalid id format",
			})
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}

		if err := s.ChangeDomainStatus(ctx, id, st, changeBy, notes); err != nil {
			result.Failed = append(result.Failed, models.BulkChangeDomainStatusFailure{
				ID:    id.String(),
				Error: err.Error(),
			})
			continue
		}
		result.Succeeded = append(result.Succeeded, id)
	}

	return result, nil
}

func (s *domainService) RequestWhitelist(ctx context.Context, domainID uuid.UUID, input models.CreateWhitelistRequestInput) (*models.WhitelistRequest, error) {
	req := &models.WhitelistRequest{
		ID:        uuid.New(),
		DomainID:  domainID,
		FirstName: strings.TrimSpace(input.FirstName),
		LastName:  strings.TrimSpace(input.LastName),
		Email:     strings.TrimSpace(input.Email),
		Address:   strings.TrimSpace(input.Address),
		Phone:     strings.TrimSpace(input.Phone),
		Reason:    strings.TrimSpace(input.Reason),
	}

	return s.repo.CreateWhitelistRequest(ctx, req)
}

func (s *domainService) AutoWhitelistStaleDomains(ctx context.Context, cutoff time.Time, changedBy, notes string) error {
	ids, err := s.repo.FindAutoWhitelistCandidateDomainIDs(ctx, cutoff)
	if err != nil {
		return err
	}

	for _, id := range ids {
		if err := s.repo.SetDomainStatusWithHistory(ctx, id, models.DomainStatusWhitelist, changedBy, notes); err != nil {
			return err
		}
	}

	return nil
}

func (s *domainService) UpdateDomain(ctx context.Context, id uuid.UUID, input models.UpdateDomainInput) (*models.Domain, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Value != nil {
		v := strings.TrimSpace(*input.Value)
		if v == "" {
			return nil, fmt.Errorf("value cannot be empty")
		}
		current.Value = v
		current.Type = domainTypeFromValue(v)
	}
	if input.Type != nil && input.Value == nil {
		t := strings.TrimSpace(*input.Type)
		if t != models.DomainTypeDomain && t != models.DomainTypeSubdomain && t != models.DomainTypeIP {
			return nil, fmt.Errorf("invalid type: must be %q, %q, or %q", models.DomainTypeDomain, models.DomainTypeSubdomain, models.DomainTypeIP)
		}
		if domainTypeFromValue(current.Value) != t {
			return nil, fmt.Errorf("type does not match value")
		}
		current.Type = t
	}
	if input.Status != nil {
		st, err := models.ParseDomainStatus(strings.TrimSpace(*input.Status))
		if err != nil {
			return nil, err
		}
		current.Status = st
	}
	if input.Description != nil {
		current.Description = strings.TrimSpace(*input.Description)
	}
	if err := s.repo.Update(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *domainService) SyncRTIRDomains(ctx context.Context) error {
	if s.rtir == nil {
		return fmt.Errorf("rtir client is nil")
	}

	syncStartedAt := time.Now().UTC()
	cutoff, loc, err := s.rtirSyncCutoffAndLoc(ctx)
	if err != nil {
		return fmt.Errorf("rtir sync: %w", err)
	}

	refs, err := s.rtir.SearchTickets(ctx, cutoff, loc)
	if err != nil {
		return fmt.Errorf("rtir sync: %w", err)
	}

	var failed []string
	for _, ref := range refs {
		if err := s.syncOneRTIRTicket(ctx, ref.ID, syncStartedAt); err != nil {
			log.Printf("[rtir-sync] ticket %s: %v", ref.ID, err)
			failed = append(failed, ref.ID)
		}
	}

	log.Printf("[rtir-sync] done: tickets=%d failed=%d cutoff=%s",
		len(refs), len(failed), cutoff.Format(time.RFC3339))
	if len(failed) > 0 {
		return fmt.Errorf("rtir sync: failed ticket ids: %v", failed)
	}
	return nil
}

func (s *domainService) rtirSyncCutoffAndLoc(ctx context.Context) (cutoff time.Time, loc *time.Location, err error) {
	lastSync, err := s.repo.GetLastRTIRSync(ctx)
	if err != nil {
		return time.Time{}, nil, err
	}
	loc, err = time.LoadLocation(strings.TrimSpace(s.rtirTz))
	if err != nil || strings.TrimSpace(s.rtirTz) == "" {
		loc = time.UTC
	}
	if lastSync != nil {
		return lastSync.Add(-s.rtirOverlap), loc, nil
	}
	return time.Now().In(loc).Add(-24 * time.Hour), loc, nil
}

func (s *domainService) syncOneRTIRTicket(ctx context.Context, ticketID string, syncStartedAt time.Time) error {
	ticket, err := s.rtir.GetTicketByID(ctx, ticketID)
	if err != nil {
		_ = s.repo.UpsertRTIRImportError(ctx, ticketID, rtirImportSource, fmt.Sprintf("fetch_ticket: %v", err), time.Now().UTC())
		return err
	}
	data, err := mappers.ExtractRTIRTicketData(ticket)
	if err != nil {
		_ = s.repo.UpsertRTIRImportError(ctx, ticketID, rtirImportSource, fmt.Sprintf("extract: %v", err), mappers.RecordTimeFromRTIRTicket(ticket))
		return err
	}
	if len(data.Domains) == 0 {
		_ = s.repo.UpsertRTIRImportError(ctx, ticketID, rtirImportSource, "missing_ioc_domains", data.RecordTime)
		return fmt.Errorf("missing_ioc_domains")
	}
	for _, val := range data.Domains {
		typ := domainTypeFromValue(val)
		rec := models.DomainRTIRRecord{
			TicketID:             ticketID,
			Value:                val,
			Type:                 typ,
			Description:          data.Description,
			Tags:                 data.Tags,
			RecordDate:           data.RecordTime,
			LastSuccessfulSyncAt: syncStartedAt,
		}
		if err := s.repo.UpsertRTIRDomainRecord(ctx, rec); err != nil {
			_ = s.repo.UpsertRTIRImportError(ctx, ticketID, rtirImportSource, fmt.Sprintf("upsert: %v", err), data.RecordTime)
			return err
		}
	}

	if err := s.repo.DeleteRTIRImportError(ctx, ticketID); err != nil {
		log.Printf("[rtir-sync] delete import error row ticket %s: %v", ticketID, err)
	}
	return nil
}

func (s *domainService) TryReimportRTIRTicket(ctx context.Context, ticketID string) error {
	if s.rtir == nil {
		return fmt.Errorf("rtir client is nil")
	}
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return fmt.Errorf("empty ticket id")
	}
	return s.syncOneRTIRTicket(ctx, ticketID, time.Now().UTC())
}

func (s *domainService) GetRTIRImportErrors(ctx context.Context) ([]models.RTIRImportError, error) {
	return s.repo.ListRTIRImportErrors(ctx)
}

func (s *domainService) SyncPNRISCDomains(ctx context.Context) error {
	if s.pnrisc == nil {
		return fmt.Errorf("pnrisc client is nil")
	}
	const batch = 500
	ids, err := s.repo.ListDomainIDsForPNRISCSync(ctx, batch)
	if err != nil {
		return fmt.Errorf("pnrisc sync: %w", err)
	}
	if len(ids) == 0 {
		log.Printf("[pnrisc-sync] nothing to sync")
		return nil
	}
	log.Printf("[pnrisc-sync] syncing %d domain(s)", len(ids))
	var failed []string
	for _, id := range ids {
		payload, err := s.repo.GetPNRISCSyncPayload(ctx, id)
		if err != nil {
			log.Printf("[pnrisc-sync] domain %s: payload: %v", id, err)
			_ = s.repo.MarkPNRISCSyncFailure(ctx, id, err.Error())
			failed = append(failed, id.String())
			continue
		}
		var dateAdded *string
		if payload.DateAdded != nil {
			s := payload.DateAdded.Format("2006-01-02")
			dateAdded = &s
		}
		body := pnrisc.UpsertBody{
			Domain:      payload.Value,
			Type:        payload.Type,
			DateAdded:   dateAdded,
			Blacklisted: payload.Status == models.DomainStatusBlacklist,
			Reason:      payload.Reason,
		}
		remoteID, err := s.pnrisc.UpsertDomain(ctx, body)
		if err != nil {
			log.Printf("[pnrisc-sync] domain %s: upsert: %v", id, err)
			_ = s.repo.MarkPNRISCSyncFailure(ctx, id, err.Error())
			failed = append(failed, id.String())
			continue
		}
		rid := remoteID
		var prid *string
		if rid != "" {
			prid = &rid
		}
		if err := s.repo.MarkPNRISCSyncSuccess(ctx, id, prid, time.Now().UTC()); err != nil {
			log.Printf("[pnrisc-sync] domain %s: mark success: %v", id, err)
			failed = append(failed, id.String())
			continue
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("pnrisc sync: failed domain ids: %v", failed)
	}
	return nil
}
