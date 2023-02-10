package memory

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

const SECRET_KEY = "2F6418794E02F3A1FCF692A3FFBE09F4A33859E71175AB0EEA4565B0B268777D"
const EXPIRY_TIME = 10 * time.Second // 10 sec expiry just for check
const HASH_COST = bcrypt.DefaultCost
