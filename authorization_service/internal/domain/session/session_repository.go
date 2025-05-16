package session

type Repository interface {
	SetSession(sessionID string, userID int64, ttlSeconds int) error
	GetUserIDBySession(sessionID string) (int64, error)
	DeleteSession(sessionID string) error

	SetUserDeletedFlag(userID int64, isDeleted bool) error
	GetUserDeletedFlag(userID int64) (bool, error)
}
