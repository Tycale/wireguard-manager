//go:build linux
// +build linux

package wg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tycale/wireguard-manager/app/ui"
)

func RefreshInterface(row *ui.TableData) {
	configName := row.Name

	// Create a BatchWriter for row.Infos
	w := row.Infos.BatchWriter()
	defer w.Close()

	// Path to the network interface in /sys/class/net/
	ifacePath := filepath.Join("/sys/class/net", configName)

	// Check if the interface directory exists
	if _, err := os.Stat(ifacePath); os.IsNotExist(err) {
		row.Iface = ""
		errMsg := "Interface " + configName + " does not exist."
		fmt.Fprintf(w, "%s\n", errMsg)
		return
	} else if err != nil {
		row.Iface = ""
		errMsg := "Error checking interface " + configName + ": " + err.Error()
		fmt.Fprintf(w, "%s\n", errMsg)
		return
	}

	row.Iface = configName
	return
}
