//go:build linux
// +build linux

package wg

import (
	"github.com/tycale/wireguard-manager/app/ui"
)

func RefreshInterface(row *ui.TableData) {
	configName := row.Name

	row.Iface = configName
	return
}
