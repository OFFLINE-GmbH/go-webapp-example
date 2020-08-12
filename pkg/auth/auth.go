package auth

import (
	"fmt"

	"go-webapp-example/pkg/db"
	"go-webapp-example/pkg/log"

	casbinsqlx "github.com/OFFLINE-GmbH/casbin-sqlx-adapter"
	"github.com/casbin/casbin"
	casbinlog "github.com/casbin/casbin/log"
)

type Manager struct {
	enforcer *casbin.Enforcer
	logger   log.Logger
}

// createPolicyTable contains the migration to create the
// needed policy table.
const createPolicyTable = `
CREATE TABLE IF NOT EXISTS policies
(
    id     INT(11) AUTO_INCREMENT,
    p_type VARCHAR(32)  NOT NULL DEFAULT '',
    v0     VARCHAR(255) NOT NULL DEFAULT '',
    v1     VARCHAR(255) NOT NULL DEFAULT '',
    v2     VARCHAR(255) NOT NULL DEFAULT '',
    v3     VARCHAR(255) NOT NULL DEFAULT '',
    v4     VARCHAR(255) NOT NULL DEFAULT '',
    v5     VARCHAR(255) NOT NULL DEFAULT '',
    PRIMARY KEY (id)
);
`

// New returns a new instance of an auth manager.
func New(conn *db.Connection, logger log.Logger) (*Manager, error) {
	m := casbin.NewModel()
	// request_definition
	m.AddDef("r", "r", "sub, obj, act")
	// policy_definition
	m.AddDef("p", "p", "sub, obj, act, eft")
	// role_definition
	m.AddDef("g", "g", "_, _")
	// policy_effect
	m.AddDef("e", "e", "some(where (p.eft == allow)) && !some(where (p.eft == deny))")
	// matchers
	m.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")

	_, err := conn.DB.Exec(createPolicyTable)
	if err != nil {
		return nil, err
	}

	a := casbinsqlx.NewAdapterFromOptions(&casbinsqlx.AdapterOptions{
		TableName: "policies",
		Db:        conn.DB,
	})

	e := casbin.NewEnforcer(m, a)
	e.EnableAutoSave(true)
	e.EnableLog(true)

	casbinlog.SetLogger(&casbinLogger{log: logger})

	return &Manager{
		enforcer: e,
		logger:   logger,
	}, nil
}

// Can check if a user has the permission to execute a certain action on a subject.
func (a *Manager) Can(userID int, subject, action string) bool {
	id := userIdentifier(userID)
	if !a.enforcer.Enforce(id, subject, action) {
		a.logger.Debugf("%+v\n", a.enforcer.GetPolicy())
		a.logger.Infof("permission denied to %s for %s.%s\n", id, subject, action)
		return false
	}

	a.logger.Debugf("permission granted to %s for %s.%s\n", id, subject, action)
	return true
}

func (a *Manager) AllowUser(userID int, subject, action string) bool {
	return a.enforcer.AddPermissionForUser(userIdentifier(userID), subject, action, "allow")
}

func (a *Manager) AddRoleForUser(userID, roleID int) bool {
	return a.enforcer.AddRoleForUser(userIdentifier(userID), roleIdentifier(roleID))
}
func (a *Manager) RemoveRoleForUser(userID, roleID int) bool {
	return a.enforcer.DeleteRoleForUser(userIdentifier(userID), roleIdentifier(roleID))
}
func (a *Manager) AddRolePermission(roleID int, subject, action string) bool {
	return a.enforcer.AddPolicy(roleIdentifier(roleID), subject, action, "allow")
}
func (a *Manager) DeleteRole(roleID int) {
	a.enforcer.DeleteRole(roleIdentifier(roleID))
}
func (a *Manager) PermissionsForUser(userID int) [][]string {
	return a.enforcer.GetImplicitPermissionsForUser(userIdentifier(userID))
}
func (a *Manager) PermissionsForRole(roleID int) [][]string {
	return a.enforcer.GetFilteredPolicy(0, roleIdentifier(roleID))
}

func (a *Manager) HasRole(userID, roleID int) bool {
	has, err := a.enforcer.HasRoleForUser(userIdentifier(userID), roleIdentifier(roleID))
	if err != nil {
		a.logger.Errorf("cannot check role %s for user %d", roleID, userID)
		return false
	}
	return has
}

// userIdentifier returns a unique user identifier.
//
// Casbin requires the use of strings as subjects. Therefore
// the User entity cannot be used directly.
func userIdentifier(u int) string {
	return fmt.Sprintf("user-%d", u)
}

// roleIdentifier returns a unique role identifier.
func roleIdentifier(u int) string {
	return fmt.Sprintf("role-%d", u)
}

// casbinLogger implements casbins' logger interface.
type casbinLogger struct {
	log     log.Logger
	enabled bool
}

func (cs *casbinLogger) EnableLog(enable bool) {
	cs.enabled = enable
}

func (cs *casbinLogger) IsEnabled() bool {
	return cs.enabled
}

func (cs *casbinLogger) Print(args ...interface{}) {
	cs.log.Trace(args)
}

func (cs *casbinLogger) Printf(msg string, args ...interface{}) {
	cs.log.Tracef(msg, args...)
}
