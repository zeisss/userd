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
	return bh.serialize(byteHash)
}

func (bh *BcryptHasher) Verify(password, hash string) bool {
	byteHash := bh.unserialize(hash)

	err := bcrypt.CompareHashAndPassword(byteHash, []byte(password))
	return err == nil
}

func (bh *BcryptHasher) NeedsRehash(hash string) bool {
	hashCost, err := bcrypt.Cost(bh.unserialize(hash))
	if err != nil {
		panic(err)
	}

	return (hashCost < bh.Cost)
}

func (bh *BcryptHasher) serialize(byteHash []byte) string {
	return string(byteHash)
}

func (bh *BcryptHasher) unserialize(hash string) []byte {
	return []byte(hash)
}
