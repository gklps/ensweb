package main

type Request struct {
	UserName string `json:"email"`
	Password string `json:"password"`
}

type Response struct {
	Token   string `json:"Token"`
	Message string `json:"Message"`
}
