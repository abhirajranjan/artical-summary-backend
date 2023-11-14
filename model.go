package main

type baseuser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (m baseuser) Key() string {
	return m.Email
}

type loginModel struct {
	baseuser
}

type registerModel struct {
	baseuser
	Username string `json:"username"`
}

type user struct {
	Id             uint32
	Username       string
	Email          string
	HashedPassword []byte
}

func createUser(m registerModel) user {
	return user{
		Id:             fastrand(),
		Username:       m.Username,
		Email:          m.Email,
		HashedPassword: hash([]byte(m.Password)),
	}
}

type addHistoryModel struct {
	Url string `json:"url"`
}

type delHistoryModel struct {
	Id uint32 `json:"id"`
}

type historyModel struct {
	Id  uint32 `json:"id"`
	Url string `json:"url"`
}

type response struct {
	Error string `json:"error,omitempty"`

	Userid   uint32 `json:"userID,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`

	History []historyModel `json:"history,omitempty"`

	Historyid uint32 `json:"historyID,omitempty"`
}
