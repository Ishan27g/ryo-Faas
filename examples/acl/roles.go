package acl

type Role interface {
	AddChild(children ...Role)
	AddPermissions(permission ...Permission)
	RemovePermissions(permission ...Permission)

	CanAccess(permission Permission) bool

	getPermissions() []Permission
	getChildren() []Role

	Marshal() map[string]interface{}
}
type role struct {
	Id          string                 `json:"Id"`
	Permissions map[string]*Permission `json:"Permissions"`
	Children    []Role                 `json:"Children"`
}

func UnMarshal(m map[string]interface{}) Role {
	var r = new(role)
	if ms := m["role"]; ms != nil {
		rl := ms.(map[string]interface{})
		r.Id = rl["Id"].(string)
		if rl["Permissions"] != nil {
			r.Permissions = make(map[string]*Permission)
			p := rl["Permissions"].(map[string]interface{})
			for name, _ := range p {
				r.Permissions[name] = &Permission{Name: name}
			}
		}
	}
	return r
}
func (r *role) Marshal() map[string]interface{} {
	var m = make(map[string]interface{})
	rm := r
	m["role"] = rm
	return m
}

func (r *role) RemovePermissions(permissions ...Permission) {
	for _, p := range permissions {
		delete(r.Permissions, p.Name)
	}
}

func (r *role) CanAccess(permission Permission) bool {
	return r.Permissions[permission.Name] != nil
}

func (r *role) getPermissions() []Permission {
	var p []Permission
	for _, permission := range r.Permissions {
		p = append(p, *permission)
	}
	return p
}
func (r *role) getChildren() []Role {
	return r.Children
}

func (r *role) AddPermissions(permissions ...Permission) {
	for _, p := range permissions {
		r.Permissions[p.Name] = &Permission{Name: p.Name}
	}
}

func (r *role) AddChild(children ...Role) {
	for _, child := range children {
		r.addChildPermissions(child)
	}
	r.Children = append(r.Children, children...)
	return
}

func (r *role) addChildPermissions(child Role) {
	r.AddPermissions(child.getPermissions()...)
	for _, c := range child.getChildren() {
		r.addChildPermissions(c)
	}
}

type Permission struct {
	Name string `json:"Name"`
}

func NewRole(name string, permissions ...Permission) Role {
	r := role{
		Id:          name,
		Permissions: make(map[string]*Permission),
		Children:    nil,
	}
	for _, perm := range permissions {
		r.Permissions[perm.Name] = &Permission{Name: perm.Name}
	}
	return &r
}
