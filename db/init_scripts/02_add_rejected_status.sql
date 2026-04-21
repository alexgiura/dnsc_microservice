-- Existing databases: widen status CHECK to allow 'rejected'.
-- Safe to re-run: DROP IF EXISTS then ADD.

ALTER TABLE core.domains DROP CONSTRAINT IF EXISTS domains_status_check;
ALTER TABLE core.domains ADD CONSTRAINT domains_status_check
  CHECK (status IN ('whitelist', 'blacklist', 'pending', 'rejected'));

ALTER TABLE core.domain_status DROP CONSTRAINT IF EXISTS domain_status_status_check;
ALTER TABLE core.domain_status ADD CONSTRAINT domain_status_status_check
  CHECK (status IN ('whitelist', 'blacklist', 'pending', 'rejected'));
