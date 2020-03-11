package auth

import (
	"strings"

	"github.com/256dpi/gomqtt/topic"
	"github.com/baetyl/baetyl/baetyl-hub/config"
	"github.com/baetyl/baetyl/baetyl-hub/utils"
)

// all permit actions
const (
	Publish   = "pub"
	Subscribe = "sub"
)

// Auth auth
type Auth struct {
	// for client certs
	certs map[string]cert
	// for client account
	accounts map[string]account
}

// NewAuth creates auth
func NewAuth(principals []config.Principal) *Auth {
	_certs := make(map[string]cert)
	_accounts := make(map[string]account)
	for _, principal := range principals {
		authorizer := NewAuthorizer()
		for _, p := range duplicatePubSubPermitRemove(principal.Permissions) {
			for _, topic := range p.Permits {
				authorizer.Add(topic, p.Action)
			}
		}
		if principal.Password == "" {
			_certs[principal.Username] = cert{
				Authorizer: authorizer,
			}
		} else {
			_accounts[principal.Username] = account{
				Password:   principal.Password,
				Authorizer: authorizer,
			}
		}
	}
	return &Auth{certs: _certs, accounts: _accounts}
}

func duplicatePubSubPermitRemove(permission []config.Permission) []config.Permission {
	PubPermitList := make(map[string]struct{})
	SubPermitList := make(map[string]struct{})
	for _, _permission := range permission {
		switch _permission.Action {
		case Publish:
			for _, v := range _permission.Permits {
				PubPermitList[v] = struct{}{}
			}
		case Subscribe:
			for _, v := range _permission.Permits {
				SubPermitList[v] = struct{}{}
			}
		}
	}
	return []config.Permission{
		{Action: Publish, Permits: utils.GetKeys(PubPermitList)},
		{Action: Subscribe, Permits: utils.GetKeys(SubPermitList)},
	}
}

// AuthenticateAccount auth client account, then return authorizer if pass
func (a *Auth) AuthenticateAccount(username, password string) *Authorizer {
	_account, ok := a.accounts[username]
	if ok && len(password) > 0 && strings.Compare(password, _account.Password) == 0 {
		return _account.Authorizer
	}
	return nil
}

// AuthenticateCert auth client cert, then return authorizer if pass
func (a *Auth) AuthenticateCert(serialNumber string) *Authorizer {
	_cert, ok := a.certs[serialNumber]
	if ok {
		return _cert.Authorizer
	}
	return nil
}

type account struct {
	Password   string
	Authorizer *Authorizer
}

type cert struct {
	Authorizer *Authorizer
}

// Authorizer checks topic permission
type Authorizer struct {
	*topic.Tree
}

// NewAuthorizer create a new authorizer
func NewAuthorizer() *Authorizer {
	return &Authorizer{Tree: topic.NewStandardTree()}
}

// Authorize auth action
func (p *Authorizer) Authorize(action, topic string) bool {
	_actions := p.Match(topic)
	for _, _action := range _actions {
		if action == _action.(string) {
			return true
		}
	}
	return false
}
