package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tomo-kay/tene/internal/domain"
)

// DeviceRepo implements handler.DeviceStore with PostgreSQL.
type DeviceRepo struct {
	pool *pgxpool.Pool
}

// NewDeviceRepo creates a PostgreSQL device repository.
func NewDeviceRepo(pool *pgxpool.Pool) *DeviceRepo {
	return &DeviceRepo{pool: pool}
}

// RegisterDevice inserts a new device record.
// The device's ID, CreatedAt, and LastSeenAt are populated from the database.
func (r *DeviceRepo) RegisterDevice(d *domain.Device) error {
	query := `
		INSERT INTO devices (user_id, device_name, x25519_public_key)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, last_seen_at`

	return r.pool.QueryRow(context.Background(), query,
		d.UserID, d.DeviceName, d.X25519PublicKey,
	).Scan(&d.ID, &d.CreatedAt, &d.LastSeenAt)
}

// ListDevices returns all devices for a user.
func (r *DeviceRepo) ListDevices(userID string) ([]domain.Device, error) {
	query := `
		SELECT id, user_id, device_name, x25519_public_key,
		       last_seen_at, created_at
		FROM devices WHERE user_id = $1
		ORDER BY last_seen_at DESC`

	rows, err := r.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("device: list: %w", err)
	}
	defer rows.Close()

	var devices []domain.Device
	for rows.Next() {
		var d domain.Device
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.DeviceName, &d.X25519PublicKey,
			&d.LastSeenAt, &d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("device: list scan: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// DeleteDevice removes a device by ID and owner.
func (r *DeviceRepo) DeleteDevice(id, userID string) error {
	ct, err := r.pool.Exec(context.Background(),
		`DELETE FROM devices WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("device: delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UpdateLastSeen updates the last_seen_at timestamp for a device.
func (r *DeviceRepo) UpdateLastSeen(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE devices SET last_seen_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("device: update last seen: %w", err)
	}
	return nil
}
