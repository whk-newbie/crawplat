// Package model 定义 IAM 服务的数据模型。
// 本文件仅包含 User 结构体，无业务逻辑、无持久化标记。
package model

// User 表示平台用户账号，Username 为登录唯一标识。
// Password 存储明文密码（仅内存模式使用），PasswordHash 存储 bcrypt 哈希（数据库模式使用）。
type User struct {
	Username     string
	Password     string
	PasswordHash string
}
