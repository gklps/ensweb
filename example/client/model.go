package main

type Request struct {
	UserName string `json:"UserName"`
	Password string `json:"Password"`
}

type Response struct {
	Token   string `json:"Token"`
	Message string `json:"Message"`
}
