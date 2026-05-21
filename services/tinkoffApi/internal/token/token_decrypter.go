package token

import "github.com/gladinov/cryptotoken"

type Decrypter struct {
	tokenCrypter *cryptotoken.TokenCrypter
}

func NewDecrypter(key string) *Decrypter {
	return &Decrypter{
		tokenCrypter: cryptotoken.NewTokenCrypter(key),
	}
}

func (d *Decrypter) DecryptTokenFromBase64(tokenInBase64 string) (string, error) {
	encryptedToken, err := cryptotoken.GetEncryptedTokenFromBase64(tokenInBase64)
	if err != nil {
		return "", err
	}

	return cryptotoken.DecryptToken(&encryptedToken, d.tokenCrypter.KeyInBase64)
}
