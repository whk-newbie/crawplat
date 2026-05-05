// Package model 定义镜像仓库认证引用模型。
package model

// RegistryAuthRef 表示私有镜像仓库的认证引用。
type RegistryAuthRef struct {
	Ref    string `json:"ref"`
	Server string `json:"server"`
}
