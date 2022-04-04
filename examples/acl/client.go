package acl

/*
* Testing acl Rest API (once deployed)
 */

//
//func main() {
//	var admin acl.RoleJson
//	admin.Id = "admin"
//	admin.Permissions = []string{"delete"}
//
//	sendAddRole(admin)
//
//	user1 := acl.RoleJson{
//		Id:          "user-1",
//		Permissions: []string{"write-1", "write-3"},
//	}
//	sendAddRole(user1)
//
//	user2 := acl.RoleJson{
//		Id:          "user-2",
//		Permissions: []string{"write-2", "write-3"},
//	}
//	sendAddRole(user2)
//
//	var addChildReq = acl.RolePermission{
//		Id:    admin.Id,
//		Child: []string{user1.Id, user2.Id},
//	}
//
//	sendAddChild(addChildReq)
//
//	admin.Permissions = append(admin.Permissions, user1.Permissions...)
//	admin.Permissions = append(admin.Permissions, user2.Permissions...)
//
//	sendCheckPermission(admin)
//
//	user1.Permissions = admin.Permissions
//	sendCheckPermission(user1)
//
//	user2.Permissions = admin.Permissions
//	sendCheckPermission(user2)
//
//}
//
//func sendCheckPermission(user acl.RoleJson) {
//	var r acl.RoleJson
//	for _, permission := range user.Permissions {
//		r.Id = user.Id
//		r.Permissions = []string{permission}
//		marshal, err := json.Marshal(r)
//		if err != nil {
//			return
//		}
//		fmt.Println("Checking permission ", r.Permissions, " for id ", r.Id)
//		send("http://localhost:9999/get", marshal)
//	}
//
//}
//
//func sendAddChild(rolePermission acl.RolePermission) {
//	marshal, err := json.Marshal(rolePermission)
//	if err != nil {
//		return
//	}
//	send("http://localhost:9999/update", marshal)
//}
//
//func sendAddRole(user acl.RoleJson) ([]byte, bool) {
//	marshal, err := json.Marshal(user)
//	if err != nil {
//		return nil, false
//	}
//	fmt.Println("Sending add user role for - ", user.Id)
//	send("http://localhost:9999/add", marshal)
//	return marshal, true
//}
//
//func send(url string, body []byte) {
//	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
//	if err != nil {
//		return
//	}
//
//	client := http.Client{Timeout: 3 * time.Second}
//
//	res, err := client.Do(request)
//	if err != nil {
//		panic(err)
//	}
//	body, err = ioutil.ReadAll(res.Body)
//	_ = res.Body.Close()
//	fmt.Println("response is ", string(body))
//}
