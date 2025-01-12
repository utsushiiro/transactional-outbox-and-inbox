package mq

import "github.com/google/uuid"

type Message struct {
	ID      uuid.UUID
	Payload []byte
}

type Messages []*Message
