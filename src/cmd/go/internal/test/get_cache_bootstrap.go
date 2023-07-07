//go:build cmd_go_bootstrap

package test

func getCache() cacheI {
	return &fsCache{}
}
