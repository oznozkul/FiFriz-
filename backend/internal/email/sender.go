package email

import (
	"context"
	"log"
)

type Sender interface {
	SendPasswordResetEmail(ctx context.Context, to, resetLink string) error
}

type ConsoleSender struct {
	logResetLinks bool
}

func NewConsoleSender(logResetLinks bool) *ConsoleSender {
	return &ConsoleSender{logResetLinks: logResetLinks}
}

func (s *ConsoleSender) SendPasswordResetEmail(_ context.Context, to, resetLink string) error {
	if s.logResetLinks {
		log.Printf("[email] password reset email to=%s link=%s", to, resetLink)
	} else {
		log.Printf("[email] password reset email sent to=%s", to)
	}
	return nil
}
