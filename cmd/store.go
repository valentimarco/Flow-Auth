package main

import (
	"flowgest/auth/cmd/models"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
)

type InMem struct {
	users    map[string]PasskeyUser
	sessions map[string]webauthn.SessionData

}

func NewInMem() *InMem {
	return &InMem{
		users:    make(map[string]PasskeyUser),
		sessions: make(map[string]webauthn.SessionData),
	}
}

func (i *InMem) GetSession(token string) webauthn.SessionData {
	fmt.Printf("[DEBUG] GetSession: %v", i.sessions[token])
	return i.sessions[token]
}

func (i *InMem) SaveSession(token string, data webauthn.SessionData) {
	fmt.Printf("[DEBUG] SaveSession: %s - %v", token, data)
	i.sessions[token] = data
}

func (i *InMem) DeleteSession(token string) {
	fmt.Printf("[DEBUG] DeleteSession: %v", token)
	delete(i.sessions, token)
}

func (i *InMem) GetUser(userName string) PasskeyUser {
	fmt.Printf("[DEBUG] GetUser: %v", userName)
	if _, ok := i.users[userName]; !ok {
		fmt.Printf("[DEBUG] GetUser: creating new user: %v", userName)
		i.users[userName] = &models.User{
			ID:          []byte(userName),
			DisplayName: userName,
			Name:        userName,
		}
	}

	return i.users[userName]
}

func (i *InMem) SaveUser(user PasskeyUser) {
	fmt.Printf("[DEBUG] SaveUser: %v", user.WebAuthnName())
	fmt.Printf("[DEBUG] SaveUser: %v", user)
	i.users[user.WebAuthnName()] = user
}