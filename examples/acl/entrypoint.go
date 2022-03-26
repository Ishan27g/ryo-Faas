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

// todo for tests only
var leastAccess = Permission{"read-something"}
var midAccess1 = Permission{"write-something-1"}

var midAccess2 = Permission{"write-something-2"}
var midAccess3 = Permission{"write-something-3"}
var mostAccess = Permission{"delete-something"}

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

	admin := NewRole(role.Id, permissions...)
	guest := NewRole("guest", leastAccess)

	admin.AddChild(guest)

	id := store.Get(TableName).Create(role.Id, admin.Marshal())
	if id != "" {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Added new role - " + role.Id))
		return
	}
	w.WriteHeader(http.StatusExpectationFailed)
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
		rl = UnMarshal(doc[0].Data.Value)
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
