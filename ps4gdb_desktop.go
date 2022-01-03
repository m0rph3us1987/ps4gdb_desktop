package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type raw_process_entry struct {
	data []byte
	name string
	pid  uint32
}

func main() {
	var command int32
	command_found := false
	fmt.Printf("ps4gdb_desktop v1.0 by m0rph3us1987\n")

	args := os.Args[1:]

	if len(args) < 2 {
		fmt.Printf("usage: ps4gdb_desktop ip:port <command> [param]\n\n")
		fmt.Printf("Commands: \n")
		fmt.Printf("get-pids  \t\t\tRetrieves a list of running processes on ps4.\n")
		fmt.Printf("attach [pid] \t\t\tAttach to the specified pid on ps4.\n")
		fmt.Printf("kill-server \t\t\tKills the gdbstub on ps4.\n\n\n")

		fmt.Printf("Example:\n")
		fmt.Printf("Attach to pid 60:\t\t ps4gdb_desktop 192.168.0.2:8164 attach 60\n\n")
		return
	}

	// Evaluate commands
	if strings.Compare(strings.ToLower(args[1]), "kill-server") == 0 {
		command_found = true
		command = -2
	}
	if strings.Compare(strings.ToLower(args[1]), "get-pids") == 0 {
		command_found = true
		command = -1
	}
	if strings.Compare(strings.ToLower(args[1]), "attach") == 0 {
		command_found = true
		tmpCmd, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Printf("Error: Invalid pid number\n%s\n", err.Error())
			return
		}
		command = int32(tmpCmd)
	}

	if !command_found {
		fmt.Printf("Error: invalid comamnd %s\n", args[1])
		return
	}

	fmt.Printf("Connecting to %s\n", args[0])

	// Connect to ps4
	conn, err := net.Dial("tcp", args[0])
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Executing command...\n")

	// Read process list and length
	// Read len
	pListLenBuffer := make([]byte, 8)
	readBytes, err := conn.Read(pListLenBuffer)
	if readBytes != 8 {
		panic("Wrong length received\n")
	}
	pListLen := binary.LittleEndian.Uint64(pListLenBuffer)

	// Read list
	var pListDataBuffer []byte
	tmp := make([]byte, 256)
	for len(pListDataBuffer) < int(pListLen) {
		n, err := conn.Read(tmp)

		if err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			break
		}
		pListDataBuffer = append(pListDataBuffer, tmp[:n]...)

	}

	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, uint32(command))
	conn.Write(buff)
	conn.Close()

	if command == -1 {
		get_raw_entries_from_data(pListDataBuffer)
	}
}

func get_raw_entries_from_data(raw_data []byte) []raw_process_entry {
	var pos uint32
	var size uint32
	result := make([]raw_process_entry, 0)

	pos = 0

	for int(pos) < len(raw_data) {
		var entry raw_process_entry
		size = binary.LittleEndian.Uint32(raw_data[pos:])

		entry.data = raw_data[pos : pos+size]
		pos += size
		entry.pid = binary.LittleEndian.Uint32(entry.data[0x48:])
		entry.name = string(entry.data[0x1BF : 0x1BF+19])
		fmt.Printf("%03d: %s\n", entry.pid, entry.name)
		result = append(result, entry)
	}

	return result
}
