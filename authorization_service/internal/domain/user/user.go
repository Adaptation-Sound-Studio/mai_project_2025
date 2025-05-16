package user

type User struct {
	ID        int64  `json:"user_id"`
	Name      string `json:"name"`
	Login     string `json:"login"`
	Pass      string `json:"password"`
	IsDeleted bool   `json:"is_deleted"`
}
