package rule

import (
	"fmt"
	"time"

	"github.com/evilsocket/opensnitch/daemon/conman"
	"github.com/evilsocket/opensnitch/daemon/log"
	"github.com/evilsocket/opensnitch/daemon/ui/protocol"
)

// Action of a rule
type Action string

// Actions of rules
const (
	Allow  = Action("allow")
	Deny   = Action("deny")
	Reject = Action("reject")
)

// Duration of a rule
type Duration string

// daemon possible durations
const (
	Once    = Duration("once")
	Restart = Duration("until restart")
	Always  = Duration("always")
)

// Rule represents an action on a connection.
// The fields match the ones saved as json to disk.
// If a .json rule file is modified on disk, it's reloaded automatically.
type Rule struct {
	// Save date fields as string, to avoid issues marshalling Time (#1140).
	Created string `json:"created"`
	Updated string `json:"updated"`

	Name        string   `json:"name"`
	Description string   `json:"description"`
	Action      Action   `json:"action"`
	Duration    Duration `json:"duration"`
	Operator    Operator `json:"operator"`
	Enabled     bool     `json:"enabled"`
	Precedence  bool     `json:"precedence"`
	Nolog       bool     `json:"nolog"`
}

// Create creates a new rule object with the specified parameters.
func Create(name, description string, enabled, precedence, nolog bool, action Action, duration Duration, op *Operator) *Rule {
	return &Rule{
		Created:     time.Now().Format(time.RFC3339),
		Enabled:     enabled,
		Precedence:  precedence,
		Nolog:       nolog,
		Name:        name,
		Description: description,
		Action:      action,
		Duration:    duration,
		Operator:    *op,
	}
}

func (r *Rule) String() string {
	return fmt.Sprintf("%s: if(%s){ %s %s }", r.Name, r.Operator.String(), r.Action, r.Duration)
}

// Match performs on a connection the checks a Rule has, to determine if it
// must be allowed or denied.
func (r *Rule) Match(con *conman.Connection) bool {
	return r.Operator.Match(con)
}

// Deserialize translates back the rule received to a Rule object
func Deserialize(reply *protocol.Rule) (*Rule, error) {
	if reply.Operator == nil {
		log.Warning("Deserialize rule, Operator nil")
		return nil, fmt.Errorf("invalid operator")
	}
	operator, err := NewOperator(
		Type(reply.Operator.Type),
		Sensitive(reply.Operator.Sensitive),
		Operand(reply.Operator.Operand),
		reply.Operator.Data,
		make([]Operator, 0),
	)
	if err != nil {
		log.Warning("Deserialize rule, NewOperator() error: %s", err)
		return nil, err
	}

	return Create(
		reply.Name,
		reply.Description,
		reply.Enabled,
		reply.Precedence,
		reply.Nolog,
		Action(reply.Action),
		Duration(reply.Duration),
		operator,
	), nil
}

// Serialize translates a Rule to the protocol object
func (r *Rule) Serialize() *protocol.Rule {
	if r == nil {
		return nil
	}

	created, err := time.Parse(time.RFC3339, r.Created)
	if err != nil {
		log.Warning("Error parsing rule Created date (it should be in RFC3339 format): %s  (%s)", err, string(r.Name))
		log.Warning("using current time instead: %s", created)
		created = time.Now()
	}

	return &protocol.Rule{
		Created:     created.Unix(),
		Name:        string(r.Name),
		Description: string(r.Description),
		Enabled:     bool(r.Enabled),
		Precedence:  bool(r.Precedence),
		Nolog:       bool(r.Nolog),
		Action:      string(r.Action),
		Duration:    string(r.Duration),
		Operator: &protocol.Operator{
			Type:      string(r.Operator.Type),
			Sensitive: bool(r.Operator.Sensitive),
			Operand:   string(r.Operator.Operand),
			Data:      string(r.Operator.Data),
		},
	}
}
