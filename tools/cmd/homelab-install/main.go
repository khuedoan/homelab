package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"go.universe.tf/netboot/out/ipxe"
	"go.universe.tf/netboot/pixiecore"
)

// TODO do not harcode map of MAC → flake target
var hostMap = map[string]string{
	"bc:24:11:d0:28:34": ".#metal2",
	"bc:24:11:0d:2f:20": ".#metal1",
}

var (
	installed = make(map[string]bool)
	inFlight  = make(map[string]bool)
	mu        sync.Mutex

	server     *http.Server
	pxeServer  *pixiecore.Server
	shutdownCh = make(chan struct{})
)

func getSubsystemStyle(subsystem string) lipgloss.Style {
	switch subsystem {
	case "DHCP":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	case "TFTP":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	case "HTTP":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	case "CALLBACK":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	case "INSTALL":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	case "PXE":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	}
}

var (
	kernelPath = flag.String("kernel", "", "Path to kernel bzImage")
	initrdPath = flag.String("initrd", "", "Path to initrd")
	initPath   = flag.String("init", "", "Path to init in the system")
	address    = flag.String("address", "0.0.0.0", "Address to listen on (default: 0.0.0.0 for all interfaces)")
	debug      = flag.Bool("debug", false, "Enable debug mode")
)

type callback struct {
	MAC string
	IP  string
}

type bootHandler struct {
	kernelPath string
	initrdPath string
	initPath   string
}

func (b bootHandler) BootSpec(m pixiecore.Machine) (*pixiecore.Spec, error) {
	mac := m.MAC.String()
	if _, ok := hostMap[mac]; ok {
		return &pixiecore.Spec{
			Kernel:  pixiecore.ID("kernel"),
			Initrd:  []pixiecore.ID{"initrd"},
			Cmdline: fmt.Sprintf("init=%s loglevel=4", b.initPath),
		}, nil
	}
	return nil, fmt.Errorf("unknown MAC address: %s", mac)
}

func (b bootHandler) ReadBootFile(id pixiecore.ID) (io.ReadCloser, int64, error) {
	var path string
	switch string(id) {
	case "kernel":
		path = b.kernelPath
	case "initrd":
		path = b.initrdPath
	default:
		return nil, -1, fmt.Errorf("unknown file ID: %s", id)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, -1, err
	}

	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, -1, err
	}

	return f, stat.Size(), nil
}

func (b bootHandler) WriteBootFile(id pixiecore.ID, body io.Reader) error {
	return fmt.Errorf("WriteBootFile not supported")
}

func main() {
	flag.Parse()

	if *kernelPath == "" || *initrdPath == "" || *initPath == "" {
		log.Fatal("Usage: homelab-install -kernel <path> -initrd <path> -init <path>")
	}

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	ipxeMap := make(map[pixiecore.Firmware][]byte)

	efi64Data, err := ipxe.Asset("third_party/ipxe/src/bin-x86_64-efi/ipxe.efi")
	if err != nil {
		log.Fatal("Failed to load embedded EFI64 iPXE binary", "error", err)
	}
	ipxeMap[pixiecore.FirmwareEFI64] = efi64Data
	ipxeMap[pixiecore.FirmwareEFIBC] = efi64Data

	pxeServer = &pixiecore.Server{
		Booter: bootHandler{
			kernelPath: *kernelPath,
			initrdPath: *initrdPath,
			initPath:   *initPath,
		},
		Address:    *address,
		DHCPNoBind: true,
		HTTPPort:   8080,
		Ipxe:       ipxeMap,
		Debug: func(subsystem, msg string) {
			coloredSubsystem := getSubsystemStyle(subsystem).Render(fmt.Sprintf("[%s]", subsystem))
			log.Debugf("%s %s", coloredSubsystem, msg)
		},
		Log: func(subsystem, msg string) {
			coloredSubsystem := getSubsystemStyle(subsystem).Render(fmt.Sprintf("[%s]", subsystem))
			log.Infof("%s %s", coloredSubsystem, msg)
		},
	}

	go func() {
		if err := pxeServer.Serve(); err != nil {
			log.Fatal("PXE server error", "error", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/report", reportHandler)
	server = &http.Server{Addr: ":5000", Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server error", "error", err)
		}
	}()

	// These ports can technically be set for testing, but the
	// protocols burned in firmware on the client side hardcode these,
	// so if you change them in production, nothing will work.
	log.Info(
		"Running on ports",
		"DHCP", 67,
		"TFTP", 69,
		"PXE", 4001,
		"HTTP", 8080,
		"CALLBACK", 5000,
	)

	log.Info("Servers to install", "hosts", hostMap)

	// Handle SIGINT for cleanup
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigCh:
		stopServer()
	case <-shutdownCh:
		stopServer()
	}
}

func reportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	mac := r.FormValue("mac")
	ip := r.FormValue("ip")

	if mac == "" || ip == "" {
		http.Error(w, "missing mac or ip", http.StatusBadRequest)
		return
	}

	coloredSubsystem := getSubsystemStyle("CALLBACK").Render("[CALLBACK]")
	log.Infof("%s Got callback mac=%s ip=%s", coloredSubsystem, mac, ip)

	mu.Lock()
	defer mu.Unlock()

	if flake, ok := hostMap[mac]; ok {
		if !installed[mac] && !inFlight[mac] {
			coloredSubsystem := getSubsystemStyle("INSTALL").Render("[INSTALL]")
			log.Infof("%s Starting installation flake=%s ip=%s", coloredSubsystem, flake, ip)
			inFlight[mac] = true

			go monitorInstall(mac, ip, flake)
		} else {
			log.Warn("Host already installed or in progress", "mac", mac)
		}
	} else {
		log.Warn("Unknown MAC, ignoring", "mac", mac)
	}

	w.WriteHeader(http.StatusOK)
}

func monitorInstall(mac, ip, flake string) {
	coloredSubsystem := getSubsystemStyle("INSTALL").Render("[INSTALL]")
	log.Infof("%s Running installation flake=%s ip=%s", coloredSubsystem, flake, ip)

	cmd := exec.Command("nixos-anywhere",
		"--env-password",
		"--no-substitute-on-destination",
		"--flake", flake,
		fmt.Sprintf("root@%s", ip),
	)

	cmd.Env = append(os.Environ(), "SSHPASS=nixos-installer")

	if *debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		coloredSubsystem := getSubsystemStyle("INSTALL").Render("[INSTALL]")
		log.Errorf("%s Error installing host mac=%s error=%v", coloredSubsystem, mac, err)
		mu.Lock()
		delete(inFlight, mac)
		mu.Unlock()
		return
	}

	mu.Lock()
	defer mu.Unlock()
	delete(inFlight, mac)
	installed[mac] = true

	coloredSubsystem = getSubsystemStyle("INSTALL").Render("[INSTALL]")
	log.Infof("%s Successfully installed host mac=%s", coloredSubsystem, mac)
	log.Infof("%s Installation status: installed=%v", coloredSubsystem, installed)

	pending := make(map[string]struct{})
	for k := range hostMap {
		if !installed[k] {
			pending[k] = struct{}{}
		}
	}
	log.Infof("%s Pending servers: %v", coloredSubsystem, pending)

	if len(pending) == 0 {
		log.Infof("%s All servers installed, shutting down", coloredSubsystem)
		close(shutdownCh)
	}
}

func stopServer() {
	log.Info("Stopping servers...")
	if server != nil {
		server.Close()
		log.Info("HTTP server shutdown complete")
	}
	if pxeServer != nil {
		pxeServer.Shutdown()
		log.Info("PXE server stopped")
	}
	os.Exit(0)
}
