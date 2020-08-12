package permission

// Service is used to interact with the entity. It
// allows access to the store by embedding it.
type Service struct {
	*Store
}

// NewService returns a pointer to a new Service.
func NewService(store *Store) *Service {
	return &Service{
		Store: store,
	}
}
