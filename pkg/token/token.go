package token

import (
	"encoding/base64"
	"encoding/json"

	"dev.khulnasoft.com/pkg/ssh"
	"github.com/pkg/errors"
)

type Token struct {
	HostKey        string `json:"hostKey,omitempty"`
	AuthorizedKeys string `json:"authorizedKeys,omitempty"`
}

func GetDevSpaceToken() (string, error) {
	// get host key
	hostKey, err := ssh.GetDevSpaceHostKey()
	if err != nil {
		return "", errors.Wrap(err, "generate host key")
	}

	// get public key
	publicKey, err := ssh.GetDevSpacePublicKey()
	if err != nil {
		return "", errors.Wrap(err, "generate key pair")
	}

	return buildToken(hostKey, publicKey)
}

func buildToken(hostKey string, publicKey string) (string, error) {
	out, err := json.Marshal(&Token{
		HostKey:        hostKey,
		AuthorizedKeys: publicKey,
	})
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(out), nil
}

func ParseToken(token string) (*Token, error) {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}

	t := &Token{}
	err = json.Unmarshal(decoded, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}
