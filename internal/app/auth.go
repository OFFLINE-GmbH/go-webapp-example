package app

// setupAuth makes sure all permissions are up to date.
func (k *Kernel) setupAuth() {
	// Make sure the admin user is always a member of the admin role
	k.Auth.AddRoleForUser(1, 1)
}
