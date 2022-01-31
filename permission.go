package main

type Permission struct {
	BaseModel
	UserID    string
	Name      string `json:"name"`
	CreatedBy string
	// TODO: Implement Gorm-compatible parameters.
}
