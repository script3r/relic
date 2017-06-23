//
// Copyright (c) SAS Institute Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package scdtoken

import (
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/sassoftware/relic/config"
	"github.com/sassoftware/relic/lib/assuan"
	"github.com/sassoftware/relic/lib/passprompt"
	"github.com/sassoftware/relic/signers/sigerrors"
	"github.com/sassoftware/relic/token"
)

var defaultScdSockets = []string{
	"/run/user/$UID/gnupg/S.scdaemon",
	"/var/run/user/$UID/gnupg/S.scdaemon",
	"$HOME/.gnupg/S.scdaemon",
}

type scdToken struct {
	config      *config.Config
	tokenConf   *config.TokenConfig
	sock        *assuan.ScdConn
	serial, pin string
	keyInfos    []*assuan.ScdKey
	mu          sync.Mutex
}

type scdKey struct {
	token     *scdToken
	keyConf   *config.KeyConfig
	key       *assuan.ScdKey
	publicKey crypto.PublicKey
}

func findSock() string {
	uid := fmt.Sprintf("%d", os.Getuid())
	for _, fp := range defaultScdSockets {
		fp = strings.Replace(fp, "$UID", uid, -1)
		fp = os.ExpandEnv(fp)
		_, err := os.Stat(fp)
		if err == nil {
			return fp
		}
	}
	return ""
}

func Open(conf *config.Config, tokenName string, prompt passprompt.PasswordGetter) (tok *scdToken, err error) {
	tconf, err := conf.GetToken(tokenName)
	if err != nil {
		return nil, err
	}
	sockPath := tconf.Provider
	if sockPath == "" {
		sockPath = findSock()
	}
	if sockPath == "" {
		return nil, fmt.Errorf("scdaemon not found; set tokens.%s.provider to the path to the scdaemon socket", tokenName)
	}
	sock, err := assuan.DialScd(sockPath)
	if err != nil {
		return nil, err
	}
	tok = &scdToken{
		config:    conf,
		tokenConf: tconf,
		sock:      sock,
	}
	if err := tok.login(prompt); err != nil {
		sock.Close()
		return nil, err
	}
	return tok, nil
}

func (tok *scdToken) login(prompt passprompt.PasswordGetter) error {
	tconf := tok.tokenConf
	keyInfos, err := tok.sock.Learn()
	if err != nil {
		return err
	}
	tok.keyInfos = keyInfos
	tok.serial = keyInfos[0].Serial
	if tconf.Serial != "" && tconf.Serial != tok.serial {
		return fmt.Errorf("scdaemon token %s has serial %s but configuration specifies %s", tconf.Name(), tok.serial, tconf.Serial)
	}
	loginFunc := func(pin string) (bool, error) {
		if err := tok.sock.CheckPin(pin); err == nil {
			tok.pin = pin
			return true, nil
		} else if _, ok := err.(sigerrors.PinIncorrectError); ok {
			return false, nil
		} else {
			return false, err
		}
	}
	initialPrompt := fmt.Sprintf("PIN for token %s (serial %s): ", tconf.Name(), tok.serial)
	keyringUser := tok.serial
	return token.Login(tconf, prompt, loginFunc, keyringUser, initialPrompt)
}

func (tok *scdToken) Ping() error {
	tok.mu.Lock()
	defer tok.mu.Unlock()
	// TODO
	return nil
}

func (tok *scdToken) Close() error {
	tok.mu.Lock()
	defer tok.mu.Unlock()
	if tok.sock != nil {
		tok.sock.Close()
		tok.sock = nil
	}
	return nil
}

func (tok *scdToken) Config() *config.TokenConfig {
	return tok.tokenConf
}

func (tok *scdToken) GetKey(keyName string) (token.Key, error) {
	tok.mu.Lock()
	defer tok.mu.Unlock()
	keyConf, err := tok.config.GetKey(keyName)
	if err != nil {
		return nil, err
	}
	var key *assuan.ScdKey
	for _, kc := range tok.keyInfos {
		if keyConf.ID != "" && keyConf.ID != kc.KeyId {
			continue
		}
		key = kc
		break
	}
	if key.KeyId == "" {
		return nil, fmt.Errorf("key %s not found in token %s", keyName, tok.tokenConf.Name())
	}
	pubkey, err := key.Public()
	if err != nil {
		return nil, err
	}
	return &scdKey{
		token:     tok,
		keyConf:   keyConf,
		key:       key,
		publicKey: pubkey,
	}, nil
}

func (key *scdKey) Public() crypto.PublicKey {
	return key.publicKey
}

func (key *scdKey) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	key.token.mu.Lock()
	defer key.token.mu.Unlock()
	return key.key.Sign(digest, opts, key.token.pin)
}

func (key *scdKey) Config() *config.KeyConfig {
	return key.keyConf
}

func (key *scdKey) GetID() []byte {
	return []byte(key.token.serial)
}

func (tok *scdToken) Import(keyName string, privKey crypto.PrivateKey) (token.Key, error) {
	return nil, errors.New("function not implemented for tokens of type \"scdaemon\"")
}

func (tok *scdToken) ImportCertificate(cert *x509.Certificate, labelBase string) error {
	return errors.New("function not implemented for tokens of type \"scdaemon\"")
}

func (tok *scdToken) Generate(keyName string, keyType token.KeyType, bits uint) (token.Key, error) {
	// TODO - probably useful
	return nil, errors.New("function not implemented for tokens of type \"scdaemon\"")
}

func (key *scdKey) ImportCertificate(cert *x509.Certificate) error {
	return errors.New("function not implemented for tokens of type \"scdaemon\"")
}
