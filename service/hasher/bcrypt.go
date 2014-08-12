package hasher

import (
	"code.google.com/p/go.crypto/bcrypt"
)

const BcryptDefaultCost = bcrypt.DefaultCost

func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost {
		cost = BcryptDefaultCost
	}
	if cost > bcrypt.MaxCost {
		panic("Cost is too high.")
	}
	return &BcryptHasher{cost}
}

type BcryptHasher struct {
	Cost int
}

func (bh *BcryptHasher) Hash(password string) string {
	byteHash, err := bcrypt.GenerateFromPassword([]byte(password), bh.Cost)
	if err != nil {
		panic(err)
	}
	return string(byteHash)
}

func (bh *BcryptHasher) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (bh *BcryptHasher) NeedsRehash(hash string) bool {
	hashCost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		panic(err)
	}

	return (hashCost < bh.Cost)
}
