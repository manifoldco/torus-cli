// Package db provides in-memory (and eventually on-disk) storage and caching
// of values returned from the registry.
// Sensitive data is stored in the db in its encrypted form.
package db

type DB struct {
	masterKey []byte
}

// MasterKey returns the encrypted bytes of the user's master key.
func (db *DB) MasterKey() []byte {
	return db.masterKey
}

// SetMasterKey stores the user's encrypted master key.
func (db *DB) SetMasterKey(mk []byte) {
	db.masterKey = mk
}

// Clear removes all data stored in the db.
func (db *DB) Clear() {
	db.masterKey = nil
}
