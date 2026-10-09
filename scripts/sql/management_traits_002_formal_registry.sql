-- Local formal version registry only. Execute separately with explicit approval.
-- No approval seed, old table ALTER, data conversion or PDF pointer changes.
-- MySQL DDL auto-commits; IF NOT EXISTS does not repair drift. Runtime validates.
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS el_mng_formal_version (
  id VARCHAR(64) NOT NULL,
  version_code VARCHAR(64) NOT NULL,
  environment VARCHAR(64) NOT NULL,
  repo_code VARCHAR(64) NOT NULL,
  bundle_id VARCHAR(64) NOT NULL,
  source_snapshot LONGTEXT NOT NULL,
  content_sha CHAR(64) NOT NULL,
  workbook_sha CHAR(64) NOT NULL,
  template_sha VARCHAR(64) NOT NULL,
  binding_sha VARCHAR(64) NOT NULL,
  asset_key VARCHAR(64) NOT NULL,
  identity_sha CHAR(64) NOT NULL,
  state VARCHAR(64) NOT NULL,
  epoch BIGINT NOT NULL,
  created_by BIGINT NOT NULL,
  created_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_mng_formal_code (environment,repo_code,version_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE IF NOT EXISTS el_mng_formal_approval (
  id VARCHAR(64) NOT NULL,
  version_id VARCHAR(64) NOT NULL,
  kind VARCHAR(64) NOT NULL,
  identity_sha CHAR(64) NOT NULL,
  actor_id BIGINT NOT NULL,
  created_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_mng_formal_approval (version_id,kind),
  CONSTRAINT fk_mng_formal_approval FOREIGN KEY (version_id)
    REFERENCES el_mng_formal_version(id) ON UPDATE RESTRICT ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE IF NOT EXISTS el_mng_formal_audit (
  id VARCHAR(64) NOT NULL,
  version_id VARCHAR(64) NOT NULL,
  identity_sha CHAR(64) NOT NULL,
  action VARCHAR(64) NOT NULL,
  actor_id BIGINT NOT NULL,
  epoch BIGINT NOT NULL,
  reason VARCHAR(512) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_mng_formal_audit (version_id,created_at),
  CONSTRAINT fk_mng_formal_audit FOREIGN KEY (version_id)
    REFERENCES el_mng_formal_version(id) ON UPDATE RESTRICT ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;