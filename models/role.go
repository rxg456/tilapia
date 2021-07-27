package models

// 角色
type Role struct {
	Model
	Name string
	Desc string
}

//角色权限
type RolePermissionRel struct {
	Model
	Rid int
	Pid int
}
