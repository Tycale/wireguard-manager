package wg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/tycale/wireguard-manager/app/ui"
)

// getHomeDir retrieves the home directory path based on the user's environment.
// It checks if the program is run with sudo and returns the appropriate home directory.
func getHomeDir() (string, error) {
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser != "" {
		// Get home from user executing sudo
		u, err := user.Lookup(sudoUser)
		if err != nil {
			return "", err
		}
		return u.HomeDir, nil
	}
	return os.UserHomeDir()
}

// CheckConfigFiles checks for '.conf' files in the '${home}/.wg' directory.
// It returns a slice of the filenames (without the '.conf' extension) and any error encountered.
func CheckConfigFiles() ([]string, error) {
	var home string
	var err error

	home, err = getHomeDir()
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
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	for event := range eventChan {
		if event.Action == "quit" {
			app.Stop()
			return
		}

		home, err := getHomeDir()
		if err != nil {
			statusChan <- ui.StatusMessage{
				Message: fmt.Sprintf("[red:white] Error getting user home directory: %s ![white]", err),
			}
		}

		confPath := fmt.Sprintf("%s/.wg/%s.conf", home, event.Arg)
		statusMessage := fmt.Sprintf("[orange] Trying to %s tunnel %s[white]", event.Action, confPath)
		statusChan <- ui.StatusMessage{Message: statusMessage, Timer: 5}

		cmd := exec.Command("wg-quick", event.Action, confPath)
		cmd.Stdout = &stderr
		cmd.Stderr = &stdout

		if err := cmd.Run(); err != nil {
			errorMessage := fmt.Sprintf("[red:white] Error: %s Stderr: %s ![white]", err, stderr.String(), stdout.String())
			statusChan <- ui.StatusMessage{Message: errorMessage}
		}

		app.QueueUpdateDraw(func() {
			w := rows[event.No].Infos.BatchWriter()
			defer w.Close()
			fmt.Fprintf(w, "Stderr: %s\n", stdout.String())
			fmt.Fprintf(w, "Stdout: %s\n", stderr.String())
		})
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
	cmd := exec.Command("wg", "show", iface)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func getAllowedIps(iface string) ([]string, error) {
	cmd := exec.Command("wg", "show", iface, "allowed-ips")
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
