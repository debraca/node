package auth

import (
	log "github.com/cihub/seelog"
	"github.com/mysterium/node/openvpn"
	"github.com/mysterium/node/openvpn/management"
	"regexp"
)

// CredentialsProvider returns client's current auth primitives (i.e. customer identity signature / node's sessionId)
type CredentialsProvider func() (username string, password string, err error)

type middleware struct {
	fetchCredentials CredentialsProvider
	connection       management.Connection
	lastUsername     string
	lastPassword     string
	state            openvpn.State
}

// NewMiddleware creates client user_auth challenge authentication middleware
func NewMiddleware(credentials CredentialsProvider) *middleware {
	return &middleware{
		fetchCredentials: credentials,
		connection:       nil,
	}
}

func (m *middleware) Start(connection management.Connection) error {
	m.connection = connection
	log.Info("starting client user-pass provider middleware")
	return nil
}

func (m *middleware) Stop(connection management.Connection) error {
	return nil
}

func (m *middleware) ConsumeLine(line string) (consumed bool, err error) {
	rule, err := regexp.Compile("^>PASSWORD:Need 'Auth' username/password$")
	if err != nil {
		return false, err
	}

	match := rule.FindStringSubmatch(line)
	if len(match) == 0 {
		return false, nil
	}
	username, password, err := m.fetchCredentials()
	log.Info("authenticating user ", username, " with pass: ", password)

	_, err = m.connection.SingleLineCommand("password 'Auth' %s", password)
	if err != nil {
		return true, err
	}

	_, err = m.connection.SingleLineCommand("username 'Auth' %s", username)
	if err != nil {
		return true, err
	}
	return true, nil
}
