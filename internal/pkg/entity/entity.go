package entity

// Entity defines the minimum functionality for a Application entitiy.
type Entity interface {
	Primary() int
	Type() Kind
}

type Kind string

const (
	KindUser       Kind = "user"
	KindPermission Kind = "permission"
	KindRole       Kind = "role"
	KindQuote      Kind = "quote"
	KindUnknown    Kind = "unknown"
)

// ResolveKind turns a string representation of a kind into a Kind type.
func ResolveKind(in string) Kind {
	types := map[string]Kind{
		"user":       KindUser,
		"permission": KindPermission,
		"role":       KindRole,
		"quote":      KindQuote,
	}

	k, ok := types[in]
	if !ok {
		return KindUnknown
	}
	return k
}
