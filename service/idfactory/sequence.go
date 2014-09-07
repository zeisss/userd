package idfactory

import (
	"fmt"
)

func NewSequenceFactory(format string) *sequenceFactory {
	return &sequenceFactory{format, 0}
}

type sequenceFactory struct {
	Format   string
	Sequence uint
}

func (seq *sequenceFactory) NewUserID() string {
	seq.Sequence++
	if seq.Format == "" {
		return fmt.Sprintf("%d", seq.Sequence)
	} else {
		return fmt.Sprintf(seq.Format, seq.Sequence)
	}
}
func (seq *sequenceFactory) NewResetPasswordToken() string {
	return seq.NewUserID()
}
