package main

type NewUserReq struct {
	Name string `json:"username"`
}

type User struct {
	Id   int    `json:"id"`
	Name string `json:"username"`
}

func NewUser(name string) *User {
	return &User{Name: name}
}
