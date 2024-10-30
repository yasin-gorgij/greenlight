package data

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/lib/pq"
)

type Permissions []string

type PermissionModel struct {
	DB *sql.DB
}

func (p Permissions) Include(code string) bool {
	return slices.Contains(p, code)
}

func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
        SELECT permissions.code
        FROM permissions
        INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
        INNER JOIN users ON users_permissions.user_id = users.id
        WHERE users.id = $1`

	rows, err := m.DB.QueryContext(ctx, query, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions
	for rows.Next() {
		var permission string
		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
        INSERT INTO users_permissions
        SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`

	_, err := m.DB.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}
