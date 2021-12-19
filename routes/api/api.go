package api

const (
	LOGIN        = "/login"
	LOGOUT       = "/logout"
	LOGIN_STATUS = "/login/status"

	MY_USER_SETTING  = "/user/my/setting"
	MY_USER_PASSWORD = "/user/my/password"

	// 用户
	USER_ADD    = "/user/add"
	USER_UPDATE = "/user/update"
	USER_LIST   = "/user/list"
	USER_EXISTS = "/user/exists"
	USER_DETAIL = "/user/detail"
	USER_DELETE = "/user/delete"

	// 角色
	USER_ROLE_PRIV_LIST = "/user/role/privlist"
	USER_ROLE_ADD       = "/user/role/add"
	USER_ROLE_UPDATE    = "/user/role/update"
	USER_ROLE_LIST      = "/user/role/list"
	USER_ROLE_DETAIL    = "/user/role/detail"
	USER_ROLE_DELETE    = "/user/role/delete"
)
