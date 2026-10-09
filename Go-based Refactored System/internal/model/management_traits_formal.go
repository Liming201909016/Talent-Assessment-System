package model

import "time"

type ManagementTraitsFormalVersion struct {
	ID             string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	VersionCode    string    `gorm:"column:version_code;type:varchar(64);uniqueIndex:uk_mng_formal_code,priority:3" json:"versionCode"`
	Environment    string    `gorm:"column:environment;type:varchar(64);uniqueIndex:uk_mng_formal_code,priority:1" json:"environment"`
	RepoCode       string    `gorm:"column:repo_code;type:varchar(64);uniqueIndex:uk_mng_formal_code,priority:2" json:"repoCode"`
	BundleID       string    `gorm:"column:bundle_id;type:varchar(64)" json:"bundleId"`
	SourceSnapshot string    `gorm:"column:source_snapshot;type:longtext" json:"-"`
	ContentSHA     string    `gorm:"column:content_sha;type:char(64)" json:"contentSha"`
	WorkbookSHA    string    `gorm:"column:workbook_sha;type:char(64)" json:"workbookSha"`
	TemplateSHA    string    `gorm:"column:template_sha;type:varchar(64)" json:"templateSha"`
	BindingSHA     string    `gorm:"column:binding_sha;type:varchar(64)" json:"bindingSha"`
	AssetKey       string    `gorm:"column:asset_key;type:varchar(64)" json:"assetKey"`
	IdentitySHA    string    `gorm:"column:identity_sha;type:char(64)" json:"identitySha"`
	State          string    `gorm:"column:state;type:varchar(64)" json:"state"`
	Epoch          int64     `gorm:"column:epoch;type:bigint" json:"epoch"`
	CreatedBy      int64     `gorm:"column:created_by;type:bigint" json:"createdBy"`
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime(6)" json:"createdAt"`
}

func (ManagementTraitsFormalVersion) TableName() string { return "el_mng_formal_version" }

type ManagementTraitsFormalApproval struct {
	ID          string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	VersionID   string    `gorm:"column:version_id;type:varchar(64);uniqueIndex:uk_mng_formal_approval,priority:1" json:"versionId"`
	Kind        string    `gorm:"column:kind;type:varchar(64);uniqueIndex:uk_mng_formal_approval,priority:2" json:"kind"`
	IdentitySHA string    `gorm:"column:identity_sha;type:char(64)" json:"identitySha"`
	ActorID     int64     `gorm:"column:actor_id;type:bigint" json:"actorId"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime(6)" json:"createdAt"`
}

func (ManagementTraitsFormalApproval) TableName() string { return "el_mng_formal_approval" }

type ManagementTraitsFormalAudit struct {
	ID          string    `gorm:"column:id;type:varchar(64);primaryKey" json:"id"`
	VersionID   string    `gorm:"column:version_id;type:varchar(64)" json:"versionId"`
	IdentitySHA string    `gorm:"column:identity_sha;type:char(64)" json:"identitySha"`
	Action      string    `gorm:"column:action;type:varchar(64)" json:"action"`
	ActorID     int64     `gorm:"column:actor_id;type:bigint" json:"actorId"`
	Epoch       int64     `gorm:"column:epoch;type:bigint" json:"epoch"`
	Reason      string    `gorm:"column:reason;type:varchar(512)" json:"reason"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime(6)" json:"createdAt"`
}

func (ManagementTraitsFormalAudit) TableName() string { return "el_mng_formal_audit" }
