package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"go.universe.tf/netboot/out/ipxe"
	"go.universe.tf/netboot/pixiecore"
)

// Hardcoded map of MAC → flake target
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

var (
	kernelPath = flag.String("kernel", "", "Path to kernel bzImage")
	initrdPath = flag.String("initrd", "", "Path to initrd")
	initPath   = flag.String("init", "", "Path to init in the system")
	address    = flag.String("address", "0.0.0.0", "Address to listen on (default: 0.0.0.0 for all interfaces)")
)

type phoneHome struct {
	MAC string
	IP  string
}

// Pixiecore boot handler
type bootHandler struct {
	kernelPath string
	initrdPath string
	initPath   string
}

func (b bootHandler) BootSpec(m pixiecore.Machine) (*pixiecore.Spec, error) {
	mac := m.MAC.String()
	if _, ok := hostMap[mac]; ok {
		// Serve the NixOS installer kernel/initrd
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

	fmt.Println("Servers to install:", hostMap)

	// Load pixiecore's embedded iPXE binaries
	ipxeMap := make(map[pixiecore.Firmware][]byte)

	// Load BIOS/PXE binary
	biosData, err := ipxe.Asset("third_party/ipxe/src/bin/undionly.kpxe")
	if err != nil {
		log.Fatalf("Failed to load embedded BIOS iPXE binary: %v", err)
	}
	ipxeMap[pixiecore.FirmwareX86PC] = biosData

	// Load EFI64 binary
	efi64Data, err := ipxe.Asset("third_party/ipxe/src/bin-x86_64-efi/ipxe.efi")
	if err != nil {
		log.Fatalf("Failed to load embedded EFI64 iPXE binary: %v", err)
	}
	ipxeMap[pixiecore.FirmwareEFI64] = efi64Data
	ipxeMap[pixiecore.FirmwareEFIBC] = efi64Data

	// Load EFI32 binary
	efi32Data, err := ipxe.Asset("third_party/ipxe/src/bin-i386-efi/ipxe.efi")
	if err != nil {
		log.Fatalf("Failed to load embedded EFI32 iPXE binary: %v", err)
	}
	ipxeMap[pixiecore.FirmwareEFI32] = efi32Data

	// Start Pixiecore PXE server
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
			log.Printf("[DEBUG] %s: %s", subsystem, msg)
		},
		Log: func(subsystem, msg string) {
			log.Printf("[INFO] %s: %s", subsystem, msg)
		},
	}

	go func() {
		if err := pxeServer.Serve(); err != nil {
			log.Fatalf("PXE server error: %v", err)
		}
	}()

	// Start HTTP phone-home server
	mux := http.NewServeMux()
	mux.HandleFunc("/report", reportHandler)
	server = &http.Server{Addr: ":5000", Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	fmt.Println("PXE server listening on port 8080")
	fmt.Println("Phone-home server listening on port 5000")

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

	fmt.Printf("Got phone-home from %s at %s\n", mac, ip)

	mu.Lock()
	defer mu.Unlock()

	if flake, ok := hostMap[mac]; ok {
		if !installed[mac] && !inFlight[mac] {
			fmt.Printf("→ Starting nixos-anywhere for %s on %s\n", flake, ip)
			inFlight[mac] = true

			go monitorInstall(mac, ip, flake)
		} else {
			fmt.Printf("%s already installed or in progress, ignoring\n", mac)
		}
	} else {
		fmt.Println("Unknown MAC, ignoring")
	}

	w.WriteHeader(http.StatusOK)
}

func monitorInstall(mac, ip, flake string) {
	fmt.Printf("Running nixos-anywhere install of %s on %s\n", flake, ip)

	// Run nixos-anywhere with password authentication
	cmd := exec.Command("nixos-anywhere",
		"--env-password",
		"--no-substitute-on-destination",
		"--flake", flake,
		fmt.Sprintf("root@%s", ip),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Set SSHPASS environment variable with the hardcoded installer password
	cmd.Env = append(os.Environ(), "SSHPASS=nixos-installer")

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error installing %s: %v\n", mac, err)
		mu.Lock()
		delete(inFlight, mac)
		mu.Unlock()
		return
	}

	mu.Lock()
	defer mu.Unlock()
	delete(inFlight, mac)
	installed[mac] = true

	fmt.Printf("Successfully installed %s\n", mac)
	fmt.Printf("Installed servers: %+v\n", installed)

	pending := make(map[string]struct{})
	for k := range hostMap {
		if !installed[k] {
			pending[k] = struct{}{}
		}
	}
	fmt.Printf("Pending servers: %+v\n", pending)

	if len(pending) == 0 {
		fmt.Println("All servers installed, shutting down.")
		close(shutdownCh)
	}
}

func stopServer() {
	fmt.Println("Stopping servers...")
	if server != nil {
		server.Close()
		fmt.Println("HTTP server shutdown complete.")
	}
	if pxeServer != nil {
		pxeServer.Shutdown()
		fmt.Println("PXE server stopped.")
	}
	os.Exit(0)
}
