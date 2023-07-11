package wg

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tycale/wireguard-manager/app/ui"
)

// CheckConfigFiles checks for '.conf' files in the '${home}/.wg' directory.
// It returns a slice of the filenames (without the '.conf' extension) and any error encountered.
func CheckConfigFiles() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(home, ".wg")

	files, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	var configFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".conf") {
			configName := strings.TrimSuffix(file.Name(), ".conf")
			configFiles = append(configFiles, configName)
		}
	}

	return configFiles, nil
}

func ListenWGChan(app *tview.Application, eventChan chan ui.Event, rows []ui.TableData, statusChan chan ui.StatusMessage) {
	for event := range eventChan {
		if event.Action == "quit" {
			app.Stop()
			return
		}

		home, err := os.UserHomeDir()
		if err != nil {
			statusChan <- ui.StatusMessage{
				Message: fmt.Sprintf("[red:white] Error getting user home directory: %s ![white]", err),
			}
		}

		confPath := fmt.Sprintf("%s/.wg/%s.conf", home, event.Arg)
		statusMessage := fmt.Sprintf("[orange] Trying to %s tunnel %s[white]", event.Action, confPath)
		statusChan <- ui.StatusMessage{Message: statusMessage, Timer: 5}

		cmd := exec.Command("sudo", "wg-quick", event.Action, confPath)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		if err := cmd.Run(); err != nil {
			errorMessage := fmt.Sprintf("[red:white] Error: %s\n%s ![white]", err, out.String())
			statusChan <- ui.StatusMessage{Message: errorMessage}
		}
	}
}

func UpdateStatus(app *tview.Application, row *ui.TableData, table *tview.Table) {
	iface := row.Iface
	if iface == "" {
		app.QueueUpdateDraw(func() {
			disconnectedCell := tview.NewTableCell(ui.StatusNo)
			disconnectedCell.SetTextColor(tcell.ColorOrchid)
			table.SetCell(row.No+1, 0, disconnectedCell)
			allowedIpsCell := tview.NewTableCell(ui.StatusUnknown)
			table.SetCell(row.No+1, 2, allowedIpsCell)
		})
		return
	}
	app.QueueUpdateDraw(func() {

		connectedCell := tview.NewTableCell(ui.StatusConnected)
		connectedCell.SetTextColor(tcell.ColorGreen)
		table.SetCell(row.No+1, 0, connectedCell)

		status, err := getStatus(iface)

		w := row.Infos.BatchWriter()
		defer w.Close()
		w.Clear()

		if err != nil {
			fmt.Fprintf(w, "Error: %v\n", err)
			return
		}

		fmt.Fprintf(w, status)

		allowedsIps, err := getAllowedIps(iface)
		if err != nil {
			fmt.Fprintf(w, "Error: %v\n", err)
			return
		}

		allowedIpsCell := tview.NewTableCell(strings.Join(allowedsIps, ","))
		table.SetCell(row.No+1, 2, allowedIpsCell)
	})
}

func getStatus(iface string) (string, error) {
	cmd := exec.Command("sudo", "wg", "show", iface)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func getAllowedIps(iface string) ([]string, error) {
	cmd := exec.Command("sudo", "wg", "show", iface, "allowed-ips")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")

	var ipRanges []string
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			ipRanges = append(ipRanges, fields[1:]...)
		}
	}

	if len(ipRanges) == 0 {
		return nil, errors.New("No IP ranges found")
	}

	return ipRanges, nil
}

func RefreshInterface(row *ui.TableData) {
	configName := row.Name

	if runtime.GOOS == "linux" {
		row.Iface = configName
		return
	}

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
