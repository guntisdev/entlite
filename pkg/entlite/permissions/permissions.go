package permissions

type Permission uint32

const (
	// Layer: Database
	DbRead  Permission = 1 << 0 // 1
	DbWrite Permission = 1 << 1 // 2

	// Layer: API (Proto/Frontend)
	ApiRead  Permission = 1 << 2 // 4
	ApiWrite Permission = 1 << 3 // 8

	// Common Shortcuts (Aliases)
	Default   = DbRead | DbWrite | ApiRead | ApiWrite // 15 (Full CRUD)
	ReadOnly  = DbWrite | DbRead | ApiRead            // 7  (System fields like createdAt)
	WriteOnly = DbRead | DbWrite | ApiWrite           // 11 (Password during signup)
	Internal  = DbRead | DbWrite                      // 3  (Internal metadata, no Proto)
	Virtual   = ApiRead | ApiWrite                    // 12 (Exists only in proto, never stored in DB)
)
