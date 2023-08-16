// +build darwin

package wg

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/tycale/wireguard-manager/app/ui"
)

func RefreshInterface(row *ui.TableData) {
	configName := row.Name

	// Create a BatchWriter for row.Infos
	w := row.Infos.BatchWriter()
	defer w.Close()

	// Check that WireGuard interfaces exist
	cmd := exec.Command("wg", "show", "interfaces")
	if err := cmd.Run(); err != nil {
		w.Clear()
		fmt.Fprintf(w, "failed to show WireGuard interfaces: %v\n", err)
		row.Iface = ""
		return
	}

	// Check that the .name file exists
	nameFile := fmt.Sprintf("/var/run/wireguard/%s.name", configName)
	if _, err := os.Stat(nameFile); os.IsNotExist(err) {
		w.Clear()
		fmt.Fprintf(w, "name file does not exist: %s\n", nameFile)
		row.Iface = ""
		return
	}

	// Read the real interface name
	interfaceBytes, err := ioutil.ReadFile(nameFile)
	if err != nil {
		w.Clear()
		fmt.Fprintf(w, "failed to read name file: %v\n", err)
		row.Iface = ""
		return
	}

	interfaceName := strings.TrimSpace(string(interfaceBytes))

	// Check that the .sock file exists
	sockFile := fmt.Sprintf("/var/run/wireguard/%s.sock", interfaceName)
	if _, err := os.Stat(sockFile); os.IsNotExist(err) {
		w.Clear()
		fmt.Fprintf(w, "sock file does not exist: %s\n", sockFile)
		row.Iface = ""
		return
	}

	// Check the modification times of the .name and .sock files
	nameStat := syscall.Stat_t{}
	if err := syscall.Stat(nameFile, &nameStat); err != nil {
		w.Clear()
		fmt.Fprintf(w, "failed to stat name file: %v\n", err)
		row.Iface = ""
		return
	}
	sockStat := syscall.Stat_t{}
	if err := syscall.Stat(sockFile, &sockStat); err != nil {
		w.Clear()
		fmt.Fprintf(w, "failed to stat sock file: %v\n", err)
		row.Iface = ""
		return
	}
	nameTime := time.Unix(nameStat.Mtimespec.Sec, nameStat.Mtimespec.Nsec)
	sockTime := time.Unix(sockStat.Mtimespec.Sec, sockStat.Mtimespec.Nsec)
	if diff := sockTime.Sub(nameTime); diff < -2*time.Second || diff > 2*time.Second {
		w.Clear()
		fmt.Fprintf(w, "modification times of name and sock files differ by more than 2 seconds\n")
		row.Iface = ""
	}

	row.Iface = interfaceName
}
