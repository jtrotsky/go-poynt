package poyntcloud

import "github.com/google/uuid"

// GenerateReferenceID returns a UUID.
func GenerateReferenceID() string {
	return uuid.New().String()
}
