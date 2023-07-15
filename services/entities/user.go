package entities

type User struct {
	UserId      string `json:"user_id,omiempty" db:"user_id,omiempty"`
	Username    string `json:"username,omiempty" db:"username,omiempty"`
	PhoneNumber string `json:"phone_number,omiempty" db:"phone_number,omiempty"`
	Membership  string `json:"membership,omiempty" db:"membership,omiempty"`
}
