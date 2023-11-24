package model

import (
	"requestHandler/pkg/conf"

	"crypto/aes"
	"encoding/hex"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Person struct {
	gorm.Model
	NationalID string
	Email      string
	Lastname   string
	IP         string
	Image1     string
	Image2     string
	State      string
}

func OpenDB(cfg conf.Config) (*gorm.DB, error) {
	dsn := cfg.Mysql.User + ":" + cfg.Mysql.Passwd + "@tcp(" + cfg.Mysql.Addr + ")/" + cfg.Mysql.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Person{})
	return db, nil
}

func (p *Person) BeforeCreate(tx *gorm.DB) (err error) {
	cfg := conf.Load()
	en, err := EncryptUpTo32(cfg.EncKey, p.Lastname)
	p.Lastname = en
	return err

}
func (p *Person) AfterCreate(tx *gorm.DB) (err error) {
	cfg := conf.Load()
	ciphertext, _ := hex.DecodeString(p.Lastname)

	c, err := aes.NewCipher([]byte(cfg.EncKey))
	if err != nil {
		return err
	}

	pt := make([]byte, len(ciphertext))
	c.Decrypt(pt, ciphertext)
	p.Lastname = string(pt[:])
	return nil
}

func EncryptUpTo32(key string, text string) (string, error) {
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// allocate space for ciphered data
	text = text + "                                           "
	text = text[:32]
	out := make([]byte, len(text))

	// encrypt
	c.Encrypt(out, []byte(text))
	// return hex string
	encrypted := hex.EncodeToString(out)[:]
	return encrypted, nil
}
