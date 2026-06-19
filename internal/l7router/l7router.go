package l7router

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
)

// Rule is one L7 routing rule: if method+path match, send to cluster.
// Rules are checked in Priority order (lowest first). First match wins.
//
// L7 routing is what makes an LB "smart" — it reads the HTTP request before
// deciding where to send it. L4 routing only sees IP:port; L7 sees the
// full request. The cost: you must parse HTTP before forwarding, which adds
// latency and CPU. L4 is always faster; use L7 only when you need it.
type Rule struct {
	ID         string `json:"id"`
	Method     string `json:"method"`      // "GET", "POST", "*" = any method
	PathPrefix string `json:"path_prefix"` // "/api/", "/" = catch-all fallback
	Cluster    string `json:"cluster"`     // which cluster to route to
	Priority   int    `json:"priority"`    // lower = checked first
}

type L7Router struct {
	mu    sync.RWMutex
	rules []*Rule
}

func New() *L7Router { return &L7Router{} }

func (r *L7Router) AddRule(rule *Rule) {
	if rule.ID == "" {
		rule.ID = fmt.Sprintf("r-%06x", rand.Intn(0xffffff))
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules = append(r.rules, rule)
	sort.Slice(r.rules, func(i, j int) bool {
		return r.rules[i].Priority < r.rules[j].Priority
	})
}

func (r *L7Router) DeleteRule(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, rule := range r.rules {
		if rule.ID == id {
			r.rules = append(r.rules[:i], r.rules[i+1:]...)
			return true
		}
	}
	return false
}

func (r *L7Router) ListRules() []*Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Rule, len(r.rules))
	copy(out, r.rules)
	return out
}

// Match returns the first rule that matches method+path, or nil (no match).
func (r *L7Router) Match(method, path string) *Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rule := range r.rules {
		if matchesMethod(rule.Method, method) && strings.HasPrefix(path, rule.PathPrefix) {
			return rule
		}
	}
	return nil
}

func matchesMethod(ruleMethod, reqMethod string) bool {
	return ruleMethod == "*" || strings.EqualFold(ruleMethod, reqMethod)
}
