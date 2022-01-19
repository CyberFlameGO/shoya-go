package main

type Permission struct {
	BaseModel
	Name      string `json:"name"`
	CreatedBy string
	Params    map[string]interface{}
}
