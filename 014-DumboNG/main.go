package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"learn_DumboNG/014-DumboNG/config"
	"learn_DumboNG/014-DumboNG/core"
	"learn_DumboNG/014-DumboNG/crypto"
	"learn_DumboNG/014-DumboNG/logger"
	"learn_DumboNG/014-DumboNG/node"
	"learn_DumboNG/014-DumboNG/pool"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultRuntimeDir = "014-DumboNG/runtime"
	defaultBasePort   = 9000
	defaultControl    = 10000
)

type cliOptions struct {
	nodes       int
	threshold   int
	id          int
	basePort    int
	controlBase int
	controlPort int
	faults      int
	rate        int
	txSize      int
	batchSize   int
	logLevel    int
	runtimeDir  string
	duration    time.Duration
	payload     string
	keepRunning bool
}

func main() {
	opts := &cliOptions{}
	root := &cobra.Command{
		Use:   "dumbo-ng",
		Short: "A runnable Dumbo-NG/sMVBA teaching implementation",
	}

	root.PersistentFlags().StringVar(&opts.runtimeDir, "runtime", defaultRuntimeDir, "runtime directory for configs, data and logs")

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Generate local keys, threshold keys, committee and parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			return initRuntime(opts)
		},
	}
	initCmd.Flags().IntVar(&opts.nodes, "nodes", 4, "number of nodes")
	initCmd.Flags().IntVar(&opts.threshold, "threshold", 0, "threshold signature threshold; default is 2f+1")
	initCmd.Flags().IntVar(&opts.basePort, "base-port", defaultBasePort, "base consensus port")
	initCmd.Flags().IntVar(&opts.controlBase, "control-base-port", defaultControl, "base control API port")
	initCmd.Flags().IntVar(&opts.faults, "faults", 0, "number of crash/byzantine-simulated nodes")
	initCmd.Flags().IntVar(&opts.rate, "rate", 0, "synthetic tx rate per node; 0 disables synthetic generator")
	initCmd.Flags().IntVar(&opts.txSize, "tx-size", 256, "synthetic tx size")
	initCmd.Flags().IntVar(&opts.batchSize, "batch-size", 8, "batch size")

	keysCmd := &cobra.Command{
		Use:   "keys",
		Short: "Generate ed25519 node keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := configDir(opts.runtimeDir)
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				return err
			}
			config.GenerateKeyFiles(opts.nodes, configDir)
			fmt.Printf("generated %d node key files in %s\n", opts.nodes, configDir)
			return nil
		},
	}
	keysCmd.Flags().IntVar(&opts.nodes, "nodes", 4, "number of node keys")

	tssCmd := &cobra.Command{
		Use:   "threshold-keys",
		Short: "Generate threshold BLS key shares",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := configDir(opts.runtimeDir)
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				return err
			}
			threshold := opts.threshold
			if threshold == 0 {
				threshold = highThreshold(opts.nodes)
			}
			config.GenerateTsKeyFiles(opts.nodes, threshold, configDir)
			fmt.Printf("generated %d threshold key files in %s with threshold %d\n", opts.nodes, configDir, threshold)
			return nil
		},
	}
	tssCmd.Flags().IntVar(&opts.nodes, "nodes", 4, "number of node key shares")
	tssCmd.Flags().IntVar(&opts.threshold, "threshold", 0, "threshold; default is 2f+1")

	nodeCmd := &cobra.Command{
		Use:   "node",
		Short: "Run a single Dumbo-NG node",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNode(opts)
		},
	}
	nodeCmd.Flags().IntVar(&opts.id, "id", 0, "node id")
	nodeCmd.Flags().IntVar(&opts.controlPort, "control-port", 0, "control API port; default runtime control base + id")
	nodeCmd.Flags().IntVar(&opts.controlBase, "control-base-port", defaultControl, "base control API port")
	nodeCmd.Flags().IntVar(&opts.logLevel, "log-level", int(logger.DeployLevel), "log level bitmask")

	localCmd := &cobra.Command{
		Use:   "local",
		Short: "Start a local multi-node Dumbo-NG testbed",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLocal(opts)
		},
	}
	localCmd.Flags().IntVar(&opts.nodes, "nodes", 4, "number of local nodes")
	localCmd.Flags().IntVar(&opts.basePort, "base-port", defaultBasePort, "base consensus port")
	localCmd.Flags().IntVar(&opts.controlBase, "control-base-port", defaultControl, "base control API port")
	localCmd.Flags().IntVar(&opts.faults, "faults", 0, "number of crash/byzantine-simulated nodes")
	localCmd.Flags().IntVar(&opts.rate, "rate", 1000, "synthetic tx rate per node")
	localCmd.Flags().IntVar(&opts.txSize, "tx-size", 256, "synthetic tx size")
	localCmd.Flags().IntVar(&opts.batchSize, "batch-size", 200, "batch size")
	localCmd.Flags().IntVar(&opts.logLevel, "log-level", int(logger.DeployLevel), "log level bitmask")
	localCmd.Flags().DurationVar(&opts.duration, "duration", 20*time.Second, "benchmark duration")
	localCmd.Flags().BoolVar(&opts.keepRunning, "keep-running", false, "keep nodes running instead of stopping after duration")

	consoleCmd := &cobra.Command{
		Use:   "console",
		Short: "Open an interactive CLI against a running node control API",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConsole(opts)
		},
	}
	consoleCmd.Flags().IntVar(&opts.id, "id", 0, "node id")
	consoleCmd.Flags().IntVar(&opts.controlPort, "control-port", 0, "control API port; default control base + id")
	consoleCmd.Flags().IntVar(&opts.controlBase, "control-base-port", defaultControl, "base control API port")

	submitCmd := &cobra.Command{
		Use:   "submit TEXT",
		Short: "Submit one transaction to a running node",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.payload = strings.Join(args, " ")
			resp, err := sendControl(opts, node.ControlRequest{Command: "submit", Payload: opts.payload})
			if err != nil {
				return err
			}
			printJSON(resp)
			return nil
		},
	}
	submitCmd.Flags().IntVar(&opts.id, "id", 0, "node id")
	submitCmd.Flags().IntVar(&opts.controlPort, "control-port", 0, "control API port; default control base + id")
	submitCmd.Flags().IntVar(&opts.controlBase, "control-base-port", defaultControl, "base control API port")

	statusCmd := singleControlCommand("status", opts, node.ControlRequest{Command: "status"})
	commitsCmd := singleControlCommand("commits", opts, node.ControlRequest{Command: "commits"})
	peersCmd := singleControlCommand("peers", opts, node.ControlRequest{Command: "peers"})

	root.AddCommand(initCmd, keysCmd, tssCmd, nodeCmd, localCmd, consoleCmd, submitCmd, statusCmd, commitsCmd, peersCmd)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func singleControlCommand(name string, opts *cliOptions, req node.ControlRequest) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Send " + name + " command to a running node",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := sendControl(opts, req)
			if err != nil {
				return err
			}
			printJSON(resp)
			return nil
		},
	}
	cmd.Flags().IntVar(&opts.id, "id", 0, "node id")
	cmd.Flags().IntVar(&opts.controlPort, "control-port", 0, "control API port; default control base + id")
	cmd.Flags().IntVar(&opts.controlBase, "control-base-port", defaultControl, "base control API port")
	return cmd
}

func initRuntime(opts *cliOptions) error {
	if opts.nodes <= 0 {
		return fmt.Errorf("nodes must be positive")
	}
	if opts.threshold == 0 {
		opts.threshold = highThreshold(opts.nodes)
	}
	if opts.threshold > opts.nodes {
		return fmt.Errorf("threshold cannot be greater than nodes")
	}

	for _, dir := range []string{configDir(opts.runtimeDir), dataDir(opts.runtimeDir), logsDir(opts.runtimeDir)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	config.GenerateKeyFiles(opts.nodes, configDir(opts.runtimeDir))
	config.GenerateTsKeyFiles(opts.nodes, opts.threshold, configDir(opts.runtimeDir))
	if err := writeCommittee(opts); err != nil {
		return err
	}
	if err := writeParameters(opts); err != nil {
		return err
	}

	fmt.Printf("initialized %d-node runtime in %s\n", opts.nodes, opts.runtimeDir)
	fmt.Printf("consensus ports: %d..%d\n", opts.basePort, opts.basePort+opts.nodes-1)
	fmt.Printf("control ports:   %d..%d\n", opts.controlBase, opts.controlBase+opts.nodes-1)
	return nil
}

func runNode(opts *cliOptions) error {
	if opts.controlPort == 0 {
		opts.controlPort = opts.controlBase + opts.id
	}
	if err := os.MkdirAll(nodeDataDir(opts.runtimeDir, opts.id), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(logsDir(opts.runtimeDir), 0o755); err != nil {
		return err
	}
	n, err := node.NewNode(
		filepath.Join(configDir(opts.runtimeDir), fmt.Sprintf(".node-key-%d.json", opts.id)),
		filepath.Join(configDir(opts.runtimeDir), fmt.Sprintf(".node-ts-key-%d.json", opts.id)),
		filepath.Join(configDir(opts.runtimeDir), ".committee.json"),
		filepath.Join(configDir(opts.runtimeDir), ".parameters.json"),
		nodeDataDir(opts.runtimeDir, opts.id),
		logsDir(opts.runtimeDir),
		opts.logLevel,
		opts.id,
	)
	if err != nil {
		return err
	}
	if err := n.StartControl(controlAddr(opts.controlPort)); err != nil {
		return err
	}
	fmt.Printf("node %d running; control=%s\n", opts.id, controlAddr(opts.controlPort))
	n.AnalyzeBlock()
	return nil
}

func runLocal(opts *cliOptions) error {
	if err := initRuntime(opts); err != nil {
		return err
	}

	children := make([]*exec.Cmd, 0, opts.nodes)
	for i := 0; i < opts.nodes; i++ {
		args := []string{
			"node",
			"--runtime", opts.runtimeDir,
			"--id", strconv.Itoa(i),
			"--control-base-port", strconv.Itoa(opts.controlBase),
			"--log-level", strconv.Itoa(opts.logLevel),
		}
		cmd := exec.Command(os.Args[0], args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			stopChildren(children)
			return err
		}
		children = append(children, cmd)
	}

	fmt.Printf("started %d local nodes\n", opts.nodes)
	if opts.keepRunning {
		fmt.Println("keep-running enabled; press Ctrl+C to stop")
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		stopChildren(children)
		return nil
	}

	time.Sleep(opts.duration)
	for i := 0; i < opts.nodes; i++ {
		resp, err := sendControlToPort(opts.controlBase+i, node.ControlRequest{Command: "status"})
		if err == nil {
			fmt.Printf("node %d status: ", i)
			printJSON(resp)
		}
	}
	stopChildren(children)
	return nil
}

func runConsole(opts *cliOptions) error {
	fmt.Println("Dumbo-NG console. Commands: status, commits, peers, submit TEXT, quit")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("dumbo-ng> ")
		if !scanner.Scan() {
			return scanner.Err()
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "quit" || line == "exit" {
			return nil
		}
		fields := strings.Fields(line)
		cmd := fields[0]
		var req node.ControlRequest
		switch cmd {
		case "status", "commits", "peers":
			req = node.ControlRequest{Command: cmd}
		case "submit":
			payload := strings.TrimSpace(strings.TrimPrefix(line, "submit"))
			if payload == "" {
				fmt.Println("usage: submit TEXT")
				continue
			}
			req = node.ControlRequest{Command: "submit", Payload: payload}
		case "help":
			fmt.Println("Commands: status, commits, peers, submit TEXT, quit")
			continue
		default:
			fmt.Println("unknown command; type help")
			continue
		}
		resp, err := sendControl(opts, req)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		printJSON(resp)
	}
}

func writeCommittee(opts *cliOptions) error {
	committee := make(map[string]any)
	for i := 0; i < opts.nodes; i++ {
		keyFile := filepath.Join(configDir(opts.runtimeDir), fmt.Sprintf(".node-key-%d.json", i))
		pub, _, err := config.GenKeysFromFile(keyFile)
		if err != nil {
			return err
		}
		committee[strconv.Itoa(i)] = map[string]any{
			"name":    string(crypto.EncodePublicKey(pub)),
			"node_id": i,
			"addr":    fmt.Sprintf("127.0.0.1:%d", opts.basePort+i),
		}
	}
	return writeJSON(filepath.Join(configDir(opts.runtimeDir), ".committee.json"), committee)
}

func writeParameters(opts *cliOptions) error {
	params := config.Parameters{
		Pool:      pool.DefaultParameters,
		Consensus: core.DefaultParameters,
	}
	params.Pool.Rate = opts.rate
	params.Pool.TxSize = opts.txSize
	params.Pool.BatchSize = opts.batchSize
	params.Consensus.Faults = opts.faults
	return writeJSON(filepath.Join(configDir(opts.runtimeDir), ".parameters.json"), params)
}

func sendControl(opts *cliOptions, req node.ControlRequest) (node.ControlResponse, error) {
	port := opts.controlPort
	if port == 0 {
		port = opts.controlBase + opts.id
	}
	return sendControlToPort(port, req)
}

func sendControlToPort(port int, req node.ControlRequest) (node.ControlResponse, error) {
	conn, err := net.DialTimeout("tcp", controlAddr(port), 2*time.Second)
	if err != nil {
		return node.ControlResponse{}, err
	}
	defer conn.Close()
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return node.ControlResponse{}, err
	}
	var resp node.ControlResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return node.ControlResponse{}, err
	}
	return resp, nil
}

func stopChildren(children []*exec.Cmd) {
	for _, child := range children {
		if child.Process != nil {
			_ = child.Process.Kill()
			_, _ = child.Process.Wait()
		}
	}
}

func writeJSON(path string, value any) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func printJSON(value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fmt.Println(value)
		return
	}
	fmt.Println(string(data))
}

func highThreshold(n int) int {
	return 2*((n-1)/3) + 1
}

func configDir(runtime string) string {
	return filepath.Join(runtime, "config")
}

func dataDir(runtime string) string {
	return filepath.Join(runtime, "data")
}

func logsDir(runtime string) string {
	return filepath.Join(runtime, "logs")
}

func nodeDataDir(runtime string, id int) string {
	return filepath.Join(dataDir(runtime), fmt.Sprintf("node-%d", id))
}

func controlAddr(port int) string {
	return fmt.Sprintf("127.0.0.1:%d", port)
}
