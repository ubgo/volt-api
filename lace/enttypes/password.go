package enttypes

import (
	"database/sql/driver"
	"lace/crypt"
)

type Password string

// Value store value to database
func (p Password) Value() (driver.Value, error) {
	if len(string(p)) == 0 {
		return nil, nil
	}
	val, err := crypt.Encrypt(string(p))
	return val, err
}

// Scan get value from database
func (p *Password) Scan(src interface{}) error {
	// fmt.Println("sfdsahfkjasdhfkjdsahj", src)
	if src == nil {
		return nil
	}
	var ciphertext string
	switch v := src.(type) {
	case string:
		ciphertext = v
	case []byte:
		ciphertext = string(v)
	}
	// val, err := base64.StdEncoding.DecodeString(ciphertext)
	// if err != nil {
	// 	return err
	// }
	decrypted, err := crypt.Decrypt(ciphertext)
	if err != nil {
		// *p = Password(ciphertext)
		return err
	}
	*p = Password(decrypted)
	return nil
}
