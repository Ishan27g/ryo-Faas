package acl

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Ishan27g/ryo-Faas/store"
)

const TableName = "user-roles"

type RoleJson struct {
	Id          string   `json:"Id"`
	Permissions []string `json:"Permissions"`
}
type RolePermission struct {
	Id    string   `json:"Id"`
	Child []string `json:"Child"`
}

// todo for tests only
var guestAccess = Permission{"read-resources"}

func AddRole(w http.ResponseWriter, r *http.Request) {
	var role RoleJson
	err := json.NewDecoder(r.Body).Decode(&role)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	var permissions []Permission
	for _, permission := range role.Permissions {
		permissions = append(permissions, Permission{Name: permission})
	}

	user := NewRole(role.Id, permissions...)
	guest := NewRole("guest", guestAccess)

	user.AddChild(guest)

	id := store.Get(TableName).Create(role.Id, user.Marshal())
	if id != "" {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Added new role - " + role.Id))
		return
	}
	w.WriteHeader(http.StatusExpectationFailed)
}
func AddChildPermission(w http.ResponseWriter, r *http.Request) {
	var role RolePermission
	err := json.NewDecoder(r.Body).Decode(&role)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	// get parent
	var parent Role
	if doc := store.Get(TableName).Get(role.Id); len(doc) == 1 {
		parent = UnMarshal(doc[0].Data)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Parent id not found " + role.Id))
	}
	// get children, add their permissions to parent
	childrenRoles := store.Get(TableName).Get(role.Child...)
	if len(childrenRoles) != len(role.Child) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Could not find all child ids"))
		return
	}
	for _, id := range childrenRoles {
		rl := UnMarshal(id.Data)
		parent.AddPermissions(rl.getPermissions()...)
	}
	// update database

	if !store.Get(TableName).Update(role.Id, parent.Marshal()) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Could not update parent permission in db"))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Added new permissions for " + role.Id))
}
func CheckPermission(w http.ResponseWriter, r *http.Request) {
	var role RoleJson
	err := json.NewDecoder(r.Body).Decode(&role)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	var permission = Permission{Name: role.Permissions[0]}
	var rl Role
	if doc := store.Get(TableName).Get(role.Id); len(doc) == 1 {
		rl = UnMarshal(doc[0].Data)
		fmt.Println("Checking ", role.Id, " for ", permission)
		if rl.CanAccess(permission) {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("Accepted "))
			fmt.Println(rl)
			return
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not allowed "))
			return
		}
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Not found " + role.Id))
}
