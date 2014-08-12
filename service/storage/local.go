package storage

import (
	"../../user"

	"errors"
	"sync"
)

var (
	InvalidUserObject = errors.New("Invalid user object")
	UserNotFound      = errors.New("No user found.")

	LoginNameAlreadyTaken = errors.New("The given loginName is already taken.")
	EmailAlreadyTaken     = errors.New("The given email address is already taken.")
)

func NewLocalStorage() *localStorage {
	return &localStorage{
		// Lock
		Lock: &sync.Mutex{},

		// Table
		Users: make(map[string]user.User),

		// Index
		LoginNames: NewIndex(),
		Emails:     NewIndex(),
	}
}

type localStorage struct {
	Lock *sync.Mutex

	// Map{userID => user}
	Users map[string]user.User

	// Map{loginName => userID}
	LoginNames *Index

	// Map{email => userID}
	Emails *Index
}

func (s *localStorage) Save(user user.User) error {
	if user.ID == "" || user.Email == "" || user.LoginName == "" {
		return InvalidUserObject
	}

	s.Lock.Lock()
	defer s.Lock.Unlock()

	// Unique Index Validation
	if s.checkTakenByOtherUser(s.Emails, user.Email, user.ID) {
		return EmailAlreadyTaken
	}

	if s.checkTakenByOtherUser(s.LoginNames, user.LoginName, user.ID) {
		return LoginNameAlreadyTaken
	}

	// Write
	oldUser, err := s.noLockLookup(user.ID)
	if err != nil && err != UserNotFound {
		return err
	}

	if oldUser.Email != user.Email {
		s.Emails.Remove(oldUser.Email)
	}

	if oldUser.LoginName != user.LoginName {
		s.LoginNames.Remove(oldUser.LoginName)
	}

	s.Emails.Put(user.Email, user.ID)
	s.LoginNames.Put(user.LoginName, user.ID)
	s.Users[user.ID] = user

	return nil
}

func (s *localStorage) checkTakenByOtherUser(index *Index, key, userID string) bool {
	otherUserID, taken := index.Lookup(key)
	if taken && otherUserID != userID {
		return true
	}
	return false
}

func (s *localStorage) Get(userID string) (user.User, error) {
	if userID == "" {
		panic("Invalid parameter: userID is empty.")
	}

	s.Lock.Lock()
	defer s.Lock.Unlock()

	return s.noLockLookup(userID)
}

func (s *localStorage) FindByLoginName(loginName string) (user.User, error) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	userID, ok := s.LoginNames.Lookup(loginName)
	if !ok {
		return user.User{}, UserNotFound
	}
	return s.noLockLookup(userID)
}

func (s *localStorage) noLockLookup(userID string) (user.User, error) {
	user, ok := s.Users[userID]
	if !ok {
		return user, UserNotFound
	} else {
		return user, nil
	}
}

// -------------------------------------------------

func NewIndex() *Index {
	return &Index{make(map[string]string)}
}

// Index is a helper struct
type Index struct {
	Data map[string]string
}

func (i *Index) Put(key, value string) {
	i.Data[key] = value
}

func (i *Index) Remove(key string) {
	delete(i.Data, key)
}

func (i *Index) Lookup(key string) (string, bool) {
	value, ok := i.Data[key]
	return value, ok
}
