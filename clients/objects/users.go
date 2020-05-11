package objects

type UsersResponse struct {
	Users []User `json:"users"`
}

type User struct {
	OrgID     string `json:"org_id"`
	AccessKey string `json:"access_key"`
}
