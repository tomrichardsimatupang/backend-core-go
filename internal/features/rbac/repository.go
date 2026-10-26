package rbac

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"backend-core-go/internal/platform/database"
)

type Repository interface {
	// WithDB mengembalikan repository yang memakai db lain (mis. pgx.Tx).
	WithDB(db database.DBTX) Repository

	ListRoles(ctx context.Context) ([]Role, error)
	FindRoleByID(ctx context.Context, id int64) (*Role, error)
	FindRoleByName(ctx context.Context, name string) (*Role, error)
	CreateRole(ctx context.Context, role *Role) error
	UpsertRole(ctx context.Context, name, description string) (*Role, error)
	DeleteRole(ctx context.Context, id int64) error

	ListPermissions(ctx context.Context) ([]Permission, error)
	UpsertPermissions(ctx context.Context, codes []string) error
	DeletePermissionsNotIn(ctx context.Context, codes []string) error
	RolePermissions(ctx context.Context, roleID int64) ([]string, error)
	ReplaceRolePermissions(ctx context.Context, roleID int64, codes []string) error

	UserRoles(ctx context.Context, userID int64) ([]Role, error)
	AssignRole(ctx context.Context, userID, roleID int64) error
	RevokeRole(ctx context.Context, userID, roleID int64) error
	UserPermissions(ctx context.Context, userID int64) ([]string, error)
	FindUserIDByEmail(ctx context.Context, email string) (int64, error)
}

type repository struct {
	db database.DBTX
	qb sq.StatementBuilderType
}

func NewRepository(db database.DBTX) Repository {
	return &repository{db: db, qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}

var roleColumns = []string{"id", "name", "description", "created_at"}

func (r *repository) WithDB(db database.DBTX) Repository {
	return &repository{db: db, qb: r.qb}
}

func (r *repository) ListRoles(ctx context.Context) ([]Role, error) {
	query, args, err := r.qb.Select(roleColumns...).From("roles").OrderBy("name").ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list roles query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	roles, err := pgx.CollectRows(rows, pgx.RowToStructByName[Role])
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	return roles, nil
}

func (r *repository) FindRoleByID(ctx context.Context, id int64) (*Role, error) {
	return r.findRole(ctx, sq.Eq{"id": id})
}

func (r *repository) FindRoleByName(ctx context.Context, name string) (*Role, error) {
	return r.findRole(ctx, sq.Eq{"name": name})
}

func (r *repository) findRole(ctx context.Context, where sq.Eq) (*Role, error) {
	query, args, err := r.qb.Select(roleColumns...).From("roles").Where(where).ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find role query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	role, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Role])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRoleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find role: %w", err)
	}
	return &role, nil
}

func (r *repository) CreateRole(ctx context.Context, role *Role) error {
	query, args, err := r.qb.Insert("roles").
		Columns("name", "description").
		Values(role.Name, role.Description).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert role query: %w", err)
	}
	if err := r.db.QueryRow(ctx, query, args...).Scan(&role.ID, &role.CreatedAt); err != nil {
		if database.IsUniqueViolation(err) {
			return ErrRoleExists
		}
		return fmt.Errorf("insert role: %w", err)
	}
	return nil
}

func (r *repository) UpsertRole(ctx context.Context, name, description string) (*Role, error) {
	query, args, err := r.qb.Insert("roles").
		Columns("name", "description").
		Values(name, description).
		Suffix("ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description RETURNING id, name, description, created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build upsert role query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	role, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Role])
	if err != nil {
		return nil, fmt.Errorf("upsert role %q: %w", name, err)
	}
	return &role, nil
}

func (r *repository) DeleteRole(ctx context.Context, id int64) error {
	query, args, err := r.qb.Delete("roles").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("build delete role query: %w", err)
	}
	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete role %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrRoleNotFound
	}
	return nil
}

func (r *repository) ListPermissions(ctx context.Context) ([]Permission, error) {
	query, args, err := r.qb.Select("id", "code", "description").From("permissions").OrderBy("code").ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list permissions query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	perms, err := pgx.CollectRows(rows, pgx.RowToStructByName[Permission])
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	return perms, nil
}

func (r *repository) UpsertPermissions(ctx context.Context, codes []string) error {
	if len(codes) == 0 {
		return nil
	}
	ib := r.qb.Insert("permissions").Columns("code")
	for _, code := range codes {
		ib = ib.Values(code)
	}
	query, args, err := ib.Suffix("ON CONFLICT (code) DO NOTHING").ToSql()
	if err != nil {
		return fmt.Errorf("build upsert permissions query: %w", err)
	}
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("upsert permissions: %w", err)
	}
	return nil
}

func (r *repository) DeletePermissionsNotIn(ctx context.Context, codes []string) error {
	query, args, err := r.qb.Delete("permissions").Where(sq.NotEq{"code": codes}).ToSql()
	if err != nil {
		return fmt.Errorf("build delete permissions query: %w", err)
	}
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("delete stale permissions: %w", err)
	}
	return nil
}

func (r *repository) RolePermissions(ctx context.Context, roleID int64) ([]string, error) {
	query, args, err := r.qb.Select("p.code").
		From("permissions p").
		Join("role_permissions rp ON rp.permission_id = p.id").
		Where(sq.Eq{"rp.role_id": roleID}).
		OrderBy("p.code").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build role permissions query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	codes, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("role %d permissions: %w", roleID, err)
	}
	return codes, nil
}

// ReplaceRolePermissions menghapus lalu mengisi ulang permission role.
// Panggil di dalam transaksi.
func (r *repository) ReplaceRolePermissions(ctx context.Context, roleID int64, codes []string) error {
	query, args, err := r.qb.Delete("role_permissions").Where(sq.Eq{"role_id": roleID}).ToSql()
	if err != nil {
		return fmt.Errorf("build delete role permissions query: %w", err)
	}
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("delete role %d permissions: %w", roleID, err)
	}
	if len(codes) == 0 {
		return nil
	}

	// Subquery memakai placeholder "?"; builder luar yang mengubahnya ke $n.
	sub := sq.Select().
		Column(sq.Expr("?::bigint", roleID)).
		Column("id").
		From("permissions").
		Where(sq.Eq{"code": codes})
	query, args, err = r.qb.Insert("role_permissions").
		Columns("role_id", "permission_id").
		Select(sub).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert role permissions query: %w", err)
	}
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("insert role %d permissions: %w", roleID, err)
	}
	return nil
}

func (r *repository) UserRoles(ctx context.Context, userID int64) ([]Role, error) {
	query, args, err := r.qb.Select("r.id", "r.name", "r.description", "r.created_at").
		From("roles r").
		Join("user_roles ur ON ur.role_id = r.id").
		Where(sq.Eq{"ur.user_id": userID}).
		OrderBy("r.name").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build user roles query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	roles, err := pgx.CollectRows(rows, pgx.RowToStructByName[Role])
	if err != nil {
		return nil, fmt.Errorf("user %d roles: %w", userID, err)
	}
	return roles, nil
}

func (r *repository) AssignRole(ctx context.Context, userID, roleID int64) error {
	query, args, err := r.qb.Insert("user_roles").
		Columns("user_id", "role_id").
		Values(userID, roleID).
		Suffix("ON CONFLICT DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("build assign role query: %w", err)
	}
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		if database.IsForeignKeyViolation(err) {
			return ErrUserNotFound
		}
		return fmt.Errorf("assign role %d to user %d: %w", roleID, userID, err)
	}
	return nil
}

func (r *repository) RevokeRole(ctx context.Context, userID, roleID int64) error {
	query, args, err := r.qb.Delete("user_roles").
		Where(sq.Eq{"user_id": userID, "role_id": roleID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build revoke role query: %w", err)
	}
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("revoke role %d from user %d: %w", roleID, userID, err)
	}
	return nil
}

func (r *repository) UserPermissions(ctx context.Context, userID int64) ([]string, error) {
	query, args, err := r.qb.Select("p.code").
		Distinct().
		From("permissions p").
		Join("role_permissions rp ON rp.permission_id = p.id").
		Join("user_roles ur ON ur.role_id = rp.role_id").
		Where(sq.Eq{"ur.user_id": userID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build user permissions query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	codes, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("user %d permissions: %w", userID, err)
	}
	return codes, nil
}

func (r *repository) FindUserIDByEmail(ctx context.Context, email string) (int64, error) {
	query, args, err := r.qb.Select("id").From("users").Where(sq.Eq{"email": email}).ToSql()
	if err != nil {
		return 0, fmt.Errorf("build find user query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	id, err := pgx.CollectOneRow(rows, pgx.RowTo[int64])
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrUserNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("find user by email: %w", err)
	}
	return id, nil
}
