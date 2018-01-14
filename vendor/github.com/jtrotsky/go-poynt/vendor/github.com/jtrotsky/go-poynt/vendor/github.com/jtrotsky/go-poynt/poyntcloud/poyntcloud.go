package poyntcloud

import "github.com/satori/go.uuid"

// TODO: Is this the right place for this function?

// GenerateReferenceID returns a UUID.
func GenerateReferenceID() string {
	return uuid.NewV4().String()
}
