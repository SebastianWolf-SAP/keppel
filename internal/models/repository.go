// SPDX-FileCopyrightText: 2024 SAP SE
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"time"
)

// Repository contains a record from the `repos` table.
type Repository struct {
	ID                      int64       `db:"id"`
	AccountName             AccountName `db:"account_name"`
	Name                    string      `db:"name"`
	NextBlobMountSweepAt    *time.Time  `db:"next_blob_mount_sweep_at"` // see tasks.BlobMountSweepJob
	NextManifestSyncAt      *time.Time  `db:"next_manifest_sync_at"`    // see tasks.ManifestSyncJob (only set for replica accounts)
	NextGarbageCollectionAt *time.Time  `db:"next_gc_at"`               // see tasks.GarbageCollectManifestsJob
}

// FullName prepends the account name to the repository name.
func (r Repository) FullName() string {
	return string(r.AccountName) + `/` + r.Name
}
