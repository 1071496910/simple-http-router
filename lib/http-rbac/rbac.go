package rbac

import (
	"errors"
	"github.com/1071496910/simple-http-router/lib/dispatcher"
	"path/filepath"
	"sync"
)

type Role interface {
	ActionPolicy(action string, resource string, acceptable bool)
	AllActionPolicy(resource string, acceptable bool)
	CheckPolicy(action string, resource string) bool
}

func NewRole() Role {
	policy := make(map[string]map[string]bool)
	dps := make(map[string]dispatcher.Dispatcher)

	for _, action := range []string{"GET", "POST", "PUT", "DELETE", "HEAD"} {
		policy[action] = make(map[string]bool)
		dps[action] = dispatcher.NewDispatcher()
	}

	return &role{
		policy: policy,
		dps:    dps,
	}

}

type role struct {
	mtx    sync.RWMutex
	policy map[string]map[string]bool
	dps    map[string]dispatcher.Dispatcher
}

func (r *role) AllActionPolicy(resource string, acceptable bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	trimedResource := filepath.Join(resource)
	for _, action := range []string{"GET", "POST", "PUT", "DELETE", "HEAD"} {
		r.policy[action][trimedResource] = acceptable
		r.dps[action].Register(trimedResource)
	}
}

func (r *role) ActionPolicy(action string, resource string, acceptable bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	trimedResource := filepath.Join(resource)

	r.policy[action][trimedResource] = acceptable
	r.dps[action].Register(trimedResource)
}

func (r *role) CheckPolicy(action string, resource string) bool {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	trimedResource := filepath.Join(resource)
	if resources, ok := r.policy[action]; ok {
		if registeredResource, err := r.dps[action].Dispatch(trimedResource); err == nil {
			return resources[registeredResource]
		}
	}
	return false
}

var (
	empty             = struct{}{}
	ErrEmptyUidOrRole = errors.New("uid or role is empty")
)

type RBAC interface {
	Grant(uid string, role string) error
	AddRole(roleName string, role role)
	AddPolicy(roleName string, action string, resource string, acceptable bool)
	CheckPolicy(uid string, action string, resource string) bool
}

type rbac struct {
	userRole map[string]string
	roles    map[string]role
	mtx      sync.RWMutex
}

func (r *rbac) Grant(uid string, role string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if uid == "" || role == "" {
		return ErrEmptyUidOrRole
	}

	r.userRole[uid] = role
	return nil

}

func (r *rbac) AddPolicy(role string, action string, resource string, acceptable bool) {

}

func (r rbac) CheckPolicy(uid string, action string, resource string) bool {
	panic("implement me")
}
