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
	Standard  = DbRead | DbWrite | ApiRead | ApiWrite // 15 (Full CRUD)
	ReadOnly  = DbRead | ApiRead                      // 5  (System fields like createdAt)
	WriteOnly = DbWrite | ApiWrite                    // 10 (Password during signup)
	Internal  = DbRead | DbWrite                      // 3  (Internal metadata, no Proto)
	Sensitive = DbRead | DbWrite | ApiWrite           // 11 (Can set it, but never see it back)
	Virtual   = ApiRead | ApiWrite                    // 12 (Exists only in proto, never stored in DB)
)
