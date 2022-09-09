package authorization

import (
	"net"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewAccessControlRules converts a schema.AccessControlConfiguration into an AccessControlRule slice.
func NewAccessControlRules(config schema.AccessControlConfiguration) (rules []*AccessControlRule) {
	networksMap, networksCacheMap := parseSchemaNetworks(config.Networks)

	for i, schemaRule := range config.Rules {
		rules = append(rules, NewAccessControlRule(i+1, schemaRule, networksMap, networksCacheMap))
	}

	return rules
}

// NewAccessControlRule parses a schema ACL and generates an internal ACL.
func NewAccessControlRule(pos int, rule schema.ACLRule, networksMap map[string][]*net.IPNet, networksCacheMap map[string]*net.IPNet) *AccessControlRule {
	return &AccessControlRule{
		Position:  pos,
		Domains:   schemaDomainsToACL(rule.Domains, rule.DomainsRegex),
		Resources: schemaResourcesToACL(rule.Resources),
		Query:     NewAccessControlQuery(rule.Query),
		Methods:   schemaMethodsToACL(rule.Methods),
		Networks:  schemaNetworksToACL(rule.Networks, networksMap, networksCacheMap),
		Subjects:  schemaSubjectsToACL(rule.Subjects),
		Policy:    StringToLevel(rule.Policy),
	}
}

// AccessControlRule controls and represents an ACL internally.
type AccessControlRule struct {
	Position  int
	Domains   []AccessControlDomain
	Resources []AccessControlResource
	Query     []AccessControlQuery
	Methods   []string
	Networks  []*net.IPNet
	Subjects  []AccessControlSubjects
	Policy    Level
}

// IsMatch returns true if all elements of an AccessControlRule match the object and subject.
func (acr *AccessControlRule) IsMatch(subject Subject, object Object) (match bool) {
	if !acr.MatchesDomains(subject, object) {
		return false
	}

	if !acr.MatchesResources(subject, object) {
		return false
	}

	if !acr.MatchesQuery(object) {
		return false
	}

	if !acr.MatchesMethods(object) {
		return false
	}

	if !acr.MatchesNetworks(subject) {
		return false
	}

	if !acr.MatchesSubjects(subject) {
		return false
	}

	return true
}

func (acr *AccessControlRule) MatchesDomains(subject Subject, object Object) (matches bool) {
	// If there are no domains in this rule then the domain condition is a match.
	if len(acr.Domains) == 0 {
		return true
	}

	// Iterate over the domains until we find a match (return true) or until we exit the loop (return false).
	for _, domain := range acr.Domains {
		if domain.IsMatch(subject, object) {
			return true
		}
	}

	return false
}

func (acr *AccessControlRule) MatchesResources(subject Subject, object Object) (matches bool) {
	// If there are no resources in this rule then the resource condition is a match.
	if len(acr.Resources) == 0 {
		return true
	}

	// Iterate over the resources until we find a match (return true) or until we exit the loop (return false).
	for _, resource := range acr.Resources {
		if resource.IsMatch(subject, object) {
			return true
		}
	}

	return false
}

func (acr *AccessControlRule) MatchesQuery(object Object) (match bool) {
	// If there are no query rules in this rule then the query condition is a match.
	if len(acr.Query) == 0 {
		return true
	}

	// Iterate over the queries until we find a match (return true) or until we exit the loop (return false).
	for _, query := range acr.Query {
		if query.IsMatch(object) {
			return true
		}
	}

	return false
}

func (acr *AccessControlRule) MatchesMethods(object Object) (match bool) {
	// If there are no methods in this rule then the method condition is a match.
	if len(acr.Methods) == 0 {
		return true
	}

	return utils.IsStringInSlice(object.Method, acr.Methods)
}

func (acr *AccessControlRule) MatchesNetworks(subject Subject) (match bool) {
	// If there are no networks in this rule then the network condition is a match.
	if len(acr.Networks) == 0 {
		return true
	}

	// Iterate over the networks until we find a match (return true) or until we exit the loop (return false).
	for _, network := range acr.Networks {
		if network.Contains(subject.IP) {
			return true
		}
	}

	return false
}

func (acr *AccessControlRule) MatchesSubjects(subject Subject) (match bool) {
	if subject.IsAnonymous() {
		return true
	}

	return acr.MatchesSubjectExact(subject)
}

func (acr *AccessControlRule) MatchesSubjectExact(subject Subject) (match bool) {
	// If there are no subjects in this rule then the subject condition is a match.
	if len(acr.Subjects) == 0 {
		return true
	} else if subject.IsAnonymous() {
		return false
	}

	// Iterate over the subjects until we find a match (return true) or until we exit the loop (return false).
	for _, subjectRule := range acr.Subjects {
		if subjectRule.IsMatch(subject) {
			return true
		}
	}

	return false
}
