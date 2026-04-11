package db

import (
	"fmt"
	"log/slog"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
)

const normalizeBatchSize = 500

// MigrateNormalizeFeedText reads all rows from the feeds table in batches,
// applies StripHTMLAndNormalizeFeedText to title, source and description,
// and writes the normalized values back. Runs until no rows remain.
// Intended to be run once via the -migrate-normalize-feeds flag.
func MigrateNormalizeFeedText() error {
	var total, updated int
	offset := 0

	slog.Info("[migrate] starting feed text normalization")

	for {
		rows, err := DB.Query(
			`SELECT id, title, source, description FROM feeds ORDER BY id LIMIT ? OFFSET ?`,
			normalizeBatchSize, offset,
		)
		if err != nil {
			return fmt.Errorf("failed to query feeds batch at offset %d: %w", offset, err)
		}

		type row struct {
			id          int
			title       string
			source      string
			description string
		}

		var batch []row
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.id, &r.title, &r.source, &r.description); err != nil {
				_ = rows.Close()
				return fmt.Errorf("failed to scan feed row: %w", err)
			}
			batch = append(batch, r)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("failed to close rows: %w", err)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("rows iteration error: %w", err)
		}

		if len(batch) == 0 {
			break
		}

		total += len(batch)

		tx, err := DB.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		stmt, err := tx.Prepare(`UPDATE feeds SET title = ?, source = ?, description = ? WHERE id = ?`)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to prepare update statement: %w", err)
		}

		for _, r := range batch {
			normTitle := utils.StripHTMLAndNormalizeFeedText(r.title)
			normSource := utils.StripHTMLAndNormalizeFeedText(r.source)
			normDesc := utils.StripHTMLAndNormalizeFeedText(r.description)

			if normTitle == r.title && normSource == r.source && normDesc == r.description {
				continue
			}

			if _, err := stmt.Exec(normTitle, normSource, normDesc, r.id); err != nil {
				_ = stmt.Close()
				_ = tx.Rollback()
				return fmt.Errorf("failed to update feed id %d: %w", r.id, err)
			}
			updated++
		}

		if err := stmt.Close(); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to close statement: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit batch at offset %d: %w", offset, err)
		}

		slog.Info("[migrate] batch done", "offset", offset, "batch_size", len(batch), "updated_so_far", updated)
		offset += normalizeBatchSize
	}

	slog.Info("[migrate] feed text normalization complete", "total_rows", total, "rows_updated", updated)
	return nil
}
