package hub

import (
	"strings"

	"github.com/256dpi/gomqtt/topic"
	"github.com/baetyl/baetyl/baetyl-hub/config"
	"github.com/baetyl/baetyl/baetyl-hub/utils"
)

// all permit actions
const (
	authPublish   = "pub"
	authSubscribe = "sub"
)

type authenticator struct {
	// for client certs
	certs map[string]cert
	// for client account
	accounts map[string]account
}

func newAuthenticator(principals []config.Principal) *authenticator {
	_certs := make(map[string]cert)
	_accounts := make(map[string]account)
	for _, principal := range principals {
		authorizer := newAuthorizer()
		for _, p := range duplicatePubSubPermitRemove(principal.Permissions) {
			for _, topic := range p.Permits {
				authorizer.Add(topic, p.Action)
			}
		}
		if principal.Password == "" {
			_certs[principal.Username] = cert{
				authorizer: authorizer,
			}
		} else {
			_accounts[principal.Username] = account{
				Password:   principal.Password,
				authorizer: authorizer,
			}
		}
	}
	return &authenticator{certs: _certs, accounts: _accounts}
}

func duplicatePubSubPermitRemove(permission []config.Permission) []config.Permission {
	PubPermitList := make(map[string]struct{})
	SubPermitList := make(map[string]struct{})
	for _, _permission := range permission {
		switch _permission.Action {
		case authPublish:
			for _, v := range _permission.Permits {
				PubPermitList[v] = struct{}{}
			}
		case authSubscribe:
			for _, v := range _permission.Permits {
				SubPermitList[v] = struct{}{}
			}
		}
	}
	return []config.Permission{
		{Action: authPublish, Permits: utils.GetKeys(PubPermitList)},
		{Action: authSubscribe, Permits: utils.GetKeys(SubPermitList)},
	}
}

func (a *authenticator) authenticateAccount(username, password string) *authorizer {
	_account, ok := a.accounts[username]
	if ok && len(password) > 0 && strings.Compare(password, _account.Password) == 0 {
		return _account.authorizer
	}
	return nil
}

func (a *authenticator) authenticateCert(serialNumber string) *authorizer {
	_cert, ok := a.certs[serialNumber]
	if ok {
		return _cert.authorizer
	}
	return nil
}

type account struct {
	Password   string
	authorizer *authorizer
}

type cert struct {
	authorizer *authorizer
}

type authorizer struct {
	*topic.Tree
}

func newAuthorizer() *authorizer {
	return &authorizer{Tree: topic.NewTree()}
}

func (p *authorizer) authorize(action, topic string) bool {
	_actions := p.Match(topic)
	for _, _action := range _actions {
		if action == _action.(string) {
			return true
		}
	}
	return false
}
