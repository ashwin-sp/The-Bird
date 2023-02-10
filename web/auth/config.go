package auth

import (
	"github.com/os3224/final-project-b0c9bd62-as14091-sp6370/web/auth/storage"
)

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

const EXPIRY_TIME = storage.EXPIRY_TIME
