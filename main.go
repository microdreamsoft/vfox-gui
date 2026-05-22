package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"vfox-gui/internal/vfox"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

type outputWidget struct {
	widget fyne.CanvasObject
	label  *widget.Label
}

func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

func main() {
	myApp := app.New()
	win := myApp.NewWindow("vfox GUI - SDK Version Manager")
	win.Resize(fyne.NewSize(900, 650))
	win.SetIcon(resourceLogoPng)

	mgr := vfox.NewManager()

	tabs := container.NewAppTabs(
		container.NewTabItem("Installed", createInstalledTab(mgr, win)),
		container.NewTabItem("Plugins", createPluginsTab(mgr, win)),
		container.NewTabItem("Install", createInstallTab(mgr, win)),
		container.NewTabItem("Use", createUseTab(mgr, win)),
		container.NewTabItem("Exec", createExecTab(mgr, win)),
		container.NewTabItem("Config", createConfigTab(win)),
	)

	win.SetContent(tabs)
	win.ShowAndRun()
}

func makeOutput(text string) *outputWidget {
	l := widget.NewLabel(text)
	l.Wrapping = fyne.TextWrapWord

	bg := canvas.NewRectangle(theme.InputBackgroundColor())
	bg.CornerRadius = 4

	return &outputWidget{
		widget: container.NewStack(bg, l),
		label:  l,
	}
}

func runAsync(output *widget.Label, fn func() (string, error)) {
	go func() {
		out, err := fn()
		out = stripANSI(out)
		fyne.Do(func() {
			if err != nil {
				output.SetText(fmt.Sprintf("Error: %v", err))
				return
			}
			output.SetText(out)
		})
	}()
}

func showInfo(win fyne.Window, title, message string) {
	dialog.ShowInformation(title, message, win)
}

func showError(win fyne.Window, err error) {
	dialog.ShowError(fmt.Errorf("Error: %v", err), win)
}

func showSuccess(win fyne.Window, msg string) {
	showInfo(win, "Success", msg)
}

func createInstalledTab(mgr *vfox.Manager, win fyne.Window) *fyne.Container {
	output := makeOutput("Click 'Refresh' to list installed SDKs")

	refreshBtn := widget.NewButton("Refresh", func() {
		runAsync(output.label, mgr.List)
	})
	currentBtn := widget.NewButton("Current Versions", func() {
		runAsync(output.label, mgr.Current)
	})
	upgradeBtn := widget.NewButton("Upgrade vfox", func() {
		runAsync(output.label, mgr.Upgrade)
	})

	uninstallEntry := widget.NewEntry()
	uninstallEntry.SetPlaceHolder("sdk-name@version (nodejs@20.0.0)")

	uninstallBtn := widget.NewButton("Uninstall", func() {
		sdk := strings.TrimSpace(uninstallEntry.Text)
		if sdk == "" {
			showInfo(win, "Warning", "Please enter SDK name and version")
			return
		}
		dialog.ShowConfirm("Confirm", fmt.Sprintf("Uninstall %s?", sdk), func(ok bool) {
			if !ok {
				return
			}
			go func() {
				out, err := mgr.Uninstall(sdk)
				out = stripANSI(out)
				if err != nil {
					fyne.Do(func() { showError(win, err) })
					return
				}
				fyne.Do(func() { showSuccess(win, out) })
				fyne.Do(func() { output.label.SetText(out) })
			}()
		}, win)
	})

	topBar := container.NewGridWithColumns(3, refreshBtn, currentBtn, upgradeBtn)
	uninstallRow := container.NewBorder(nil, nil, widget.NewLabel("Uninstall:"), uninstallBtn, uninstallEntry)

	return container.NewBorder(
		container.NewVBox(topBar, uninstallRow),
		nil, nil, nil,
		container.NewScroll(output.widget),
	)
}

func createPluginsTab(mgr *vfox.Manager, win fyne.Window) *fyne.Container {
	output := makeOutput("Click 'Fetch Plugins' to view available plugins")

	fetchBtn := widget.NewButton("Fetch Plugins", func() {
		runAsync(output.label, mgr.Available)
	})

	addEntry := widget.NewEntry()
	addEntry.SetPlaceHolder("Plugin name (nodejs, golang, python...)")

	addBtn := widget.NewButton("Add", func() {
		name := strings.TrimSpace(addEntry.Text)
		if name == "" {
			showInfo(win, "Warning", "Please enter a plugin name")
			return
		}
		runAsync(output.label, func() (string, error) { return mgr.Add(name) })
		addEntry.SetText("")
	})

	updateEntry := widget.NewEntry()
	updateEntry.SetPlaceHolder("Plugin name (blank = update all)")

	updateBtn := widget.NewButton("Update", func() {
		name := strings.TrimSpace(updateEntry.Text)
		if name == "" {
			runAsync(output.label, mgr.UpdateAll)
		} else {
			runAsync(output.label, func() (string, error) { return mgr.Update(name) })
			updateEntry.SetText("")
		}
	})

	removeEntry := widget.NewEntry()
	removeEntry.SetPlaceHolder("Plugin name to remove")

	removeBtn := widget.NewButton("Remove", func() {
		name := strings.TrimSpace(removeEntry.Text)
		if name == "" {
			showInfo(win, "Warning", "Please enter a plugin name")
			return
		}
		dialog.ShowConfirm("Confirm", fmt.Sprintf("Remove plugin '%s'?", name), func(ok bool) {
			if !ok {
				return
			}
			runAsync(output.label, func() (string, error) { return mgr.Remove(name) })
			removeEntry.SetText("")
		}, win)
	})

	topBar := container.NewBorder(nil, nil, nil, nil, fetchBtn)
	addRow := container.NewBorder(nil, nil, widget.NewLabel("Add:"), addBtn, addEntry)
	updateRow := container.NewBorder(nil, nil, widget.NewLabel("Update:"), updateBtn, updateEntry)
	removeRow := container.NewBorder(nil, nil, widget.NewLabel("Remove:"), removeBtn, removeEntry)

	return container.NewBorder(
		container.NewVBox(topBar, addRow, updateRow, removeRow),
		nil, nil, nil,
		container.NewScroll(output.widget),
	)
}

func createInstallTab(mgr *vfox.Manager, win fyne.Window) *fyne.Container {
	output := makeOutput("Search results will appear here")

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("SDK name (nodejs, golang, python...)")
	searchEntryScroll := container.NewScroll(searchEntry)
	searchEntryScroll.SetMinSize(fyne.NewSize(500, 0))

	searchBtn := widget.NewButton("Search", func() {
		name := strings.TrimSpace(searchEntry.Text)
		if name == "" {
			showInfo(win, "Warning", "Please enter an SDK name")
			return
		}
		runAsync(output.label, func() (string, error) { return mgr.Search(name) })
	})

	installEntry := widget.NewEntry()
	installEntry.SetPlaceHolder("sdk-name@version (nodejs@21.5.0)")

	installBtn := widget.NewButton("Install", func() {
		sdk := strings.TrimSpace(installEntry.Text)
		if sdk == "" {
			showInfo(win, "Warning", "Please enter SDK name and version")
			return
		}
		go func() {
			out, err := mgr.Install(sdk)
			out = stripANSI(out)
			if err != nil {
				fyne.Do(func() { showError(win, err) })
				return
			}
			fyne.Do(func() { showSuccess(win, out) })
			fyne.Do(func() { installEntry.SetText("") })
		}()
	})

	infoEntry := widget.NewEntry()
	infoEntry.SetPlaceHolder("sdk-name[@version]")

	infoBtn := widget.NewButton("Info", func() {
		sdk := strings.TrimSpace(infoEntry.Text)
		if sdk == "" {
			showInfo(win, "Warning", "Please enter SDK name")
			return
		}
		runAsync(output.label, func() (string, error) { return mgr.Info(sdk) })
	})

	searchRow := container.NewHBox(widget.NewLabel("Search:"), searchEntry, searchBtn)
	installRow := container.NewBorder(nil, nil, widget.NewLabel("Install:"), installBtn, installEntry)
	infoRow := container.NewBorder(nil, nil, widget.NewLabel("Info:"), infoBtn, infoEntry)

	return container.NewBorder(
		container.NewVBox(searchRow, installRow, infoRow),
		nil, nil, nil,
		container.NewScroll(output.widget),
	)
}

func createUseTab(mgr *vfox.Manager, win fyne.Window) *fyne.Container {
	output := makeOutput("Use this tab to switch SDK versions")

	sdkEntry := widget.NewEntry()
	sdkEntry.SetPlaceHolder("sdk-name@version (nodejs@20.0.0)")

	scopeSelect := widget.NewSelect([]string{"global", "project", "session"}, func(s string) {})
	scopeSelect.SetSelected("global")

	useBtn := widget.NewButton("Use", func() {
		sdk := strings.TrimSpace(sdkEntry.Text)
		if sdk == "" {
			showInfo(win, "Warning", "Please enter SDK name and version")
			return
		}
		runAsync(output.label, func() (string, error) { return mgr.Use(sdk, scopeSelect.Selected) })
	})

	unuseEntry := widget.NewEntry()
	unuseEntry.SetPlaceHolder("SDK name to unuse")

	unuseBtn := widget.NewButton("Unuse", func() {
		sdk := strings.TrimSpace(unuseEntry.Text)
		if sdk == "" {
			showInfo(win, "Warning", "Please enter SDK name")
			return
		}
		runAsync(output.label, func() (string, error) { return mgr.Unuse(sdk, scopeSelect.Selected) })
	})

	currentEntry := widget.NewEntry()
	currentEntry.SetPlaceHolder("SDK name (optional)")

	currentBtn := widget.NewButton("Current", func() {
		sdk := strings.TrimSpace(currentEntry.Text)
		if sdk == "" {
			runAsync(output.label, mgr.Current)
		} else {
			runAsync(output.label, func() (string, error) { return mgr.CurrentSDK(sdk) })
		}
	})

	scopeRow := container.NewHBox(widget.NewLabel("Scope:"), scopeSelect, layout.NewSpacer())
	useRow := container.NewBorder(nil, nil, widget.NewLabel("Use:"), useBtn, sdkEntry)
	unuseRow := container.NewBorder(nil, nil, widget.NewLabel("Unuse:"), unuseBtn, unuseEntry)
	currentRow := container.NewBorder(nil, nil, widget.NewLabel("Current:"), currentBtn, currentEntry)

	return container.NewBorder(
		container.NewVBox(scopeRow, useRow, unuseRow, currentRow),
		nil, nil, nil,
		container.NewScroll(output.widget),
	)
}

func createExecTab(mgr *vfox.Manager, win fyne.Window) *fyne.Container {
	output := makeOutput("Execute commands in vfox managed environment")

	sdkEntry := widget.NewEntry()
	sdkEntry.SetPlaceHolder("sdk-name[@version] (nodejs@20.0.0)")

	cmdEntry := widget.NewEntry()
	cmdEntry.SetPlaceHolder("Command (node -v)")

	execBtn := widget.NewButton("Execute", func() {
		sdk := strings.TrimSpace(sdkEntry.Text)
		cmd := strings.TrimSpace(cmdEntry.Text)
		if sdk == "" || cmd == "" {
			showInfo(win, "Warning", "Please fill in both SDK and command fields")
			return
		}
		runAsync(output.label, func() (string, error) { return mgr.Exec(sdk, cmd) })
	})

	shellBtn := widget.NewButton("Run in cmd.exe", func() {
		sdk := strings.TrimSpace(sdkEntry.Text)
		cmd := strings.TrimSpace(cmdEntry.Text)
		if sdk == "" || cmd == "" {
			showInfo(win, "Warning", "Please fill in both SDK and command fields")
			return
		}
		fullCmd := fmt.Sprintf("vfox exec %s -- %s", sdk, cmd)
		c := exec.Command("cmd", "/C", "start", "cmd", "/K", fullCmd)
		_ = c.Start()
	})

	sdkRow := container.NewBorder(nil, nil, widget.NewLabel("SDK:"), nil, sdkEntry)
	cmdRow := container.NewBorder(nil, nil, widget.NewLabel("Command:"), nil, cmdEntry)
	btnRow := container.NewHBox(execBtn, shellBtn, layout.NewSpacer())

	return container.NewBorder(
		container.NewVBox(sdkRow, cmdRow, btnRow),
		nil, nil, nil,
		container.NewScroll(output.widget),
	)
}

func createConfigTab(win fyne.Window) *fyne.Container {
	output := makeOutput("Config output will appear here")

	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("Config key (cache.availableHookDuration)")

	getBtn := widget.NewButton("Get", func() {
		key := strings.TrimSpace(keyEntry.Text)
		if key == "" {
			showInfo(win, "Warning", "Please enter a config key")
			return
		}
		go func() {
			cmd := exec.Command("vfox", "config", key)
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			out, err := cmd.Output()
			result := stripANSI(strings.TrimSpace(string(out)))
			fyne.Do(func() {
				if err != nil {
					output.label.SetText(fmt.Sprintf("Error: %v", err))
					return
				}
				output.label.SetText(result)
			})
		}()
	})

	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("Config value")

	setBtn := widget.NewButton("Set", func() {
		key := strings.TrimSpace(keyEntry.Text)
		value := strings.TrimSpace(valueEntry.Text)
		if key == "" {
			showInfo(win, "Warning", "Please enter a config key")
			return
		}
		go func() {
			cmd := exec.Command("vfox", "config", key, value)
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			out, err := cmd.Output()
			result := stripANSI(strings.TrimSpace(string(out)))
			fyne.Do(func() {
				if err != nil {
					showError(win, err)
					return
				}
				showSuccess(win, result)
			})
		}()
	})

	keyRow := container.NewBorder(nil, nil, widget.NewLabel("Key:"),
		container.NewHBox(getBtn, setBtn), keyEntry)
	valueRow := container.NewBorder(nil, nil, widget.NewLabel("Value:"), nil, valueEntry)

	return container.NewBorder(
		container.NewVBox(keyRow, valueRow),
		nil, nil, nil,
		container.NewScroll(output.widget),
	)
}
