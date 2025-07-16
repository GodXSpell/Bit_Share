# File Share System

A simple peer-to-peer file sharing system built in Go that allows you to send files between devices over a local network.

## Features

- ✅ Send files over LAN or same Wi-Fi network
- ✅ No internet connection required
- ✅ Simple CLI interface
- ✅ Supports any file type
- ✅ Shows file metadata (name, size)
- ✅ Progress feedback during transfer

## Project Structure

```
fileshare/
├── cmd/
│   └── main.go              # Entry point: handles CLI commands
├── internal/
│   ├── transfer/
│   │   └── transfer.go      # Logic for sending and receiving files
│   └── utils/
│       └── utils.go         # Utility functions
├── go.mod                   # Go module declaration
├── test.txt                 # Sample file to send
└── README.md               # This file
```

## Installation

1. Make sure you have Go installed (version 1.21+)
2. Clone or download this project
3. Navigate to the project directory
4. Build the application:
   ```bash
   go build -o fileshare.exe .
   ```

## Usage

### Basic Commands

#### Send a file:
```bash
go run main.go send <file_path> <receiver_ip> <port>
```

#### Receive a file:
```bash
go run main.go receive <save_path> <listen_port>
```

### Example Usage

#### Step 1: Start the receiver (on destination device)
```bash
go run main.go receive received_test.txt 9000
```
This will:
- Start listening on port 9000
- Wait for incoming file transfers
- Save the received file as `received_test.txt`

#### Step 2: Send the file (from source device)
```bash
go run main.go send test.txt 192.168.1.5 9000
```
This will:
- Send `test.txt` to the device at IP `192.168.1.5`
- Connect to port 9000
- Transfer the file content

### Finding Your IP Address

To find your device's IP address:

**Windows:**
```cmd
ipconfig
```

**macOS/Linux:**
```bash
ifconfig
```

Look for your local network IP (usually starts with 192.168.x.x or 10.x.x.x)

## Testing on Same Machine

You can test the system on the same machine using `localhost`:

**Terminal 1 (Receiver):**
```bash
go run main.go receive received_test.txt 9000
```

**Terminal 2 (Sender):**
```bash
go run main.go send test.txt localhost 9000
```

## Network Requirements

- Both devices must be on the same network (LAN/Wi-Fi)
- The receiver's port must be accessible (not blocked by firewall)
- No internet connection required

## File Transfer Process

### Sender Side:
1. Reads file from disk
2. Connects to receiver's IP:PORT via TCP
3. Sends filename and file size first
4. Streams file content over the connection
5. Closes connection when complete

### Receiver Side:
1. Starts TCP listener on specified port
2. Accepts incoming connection
3. Reads filename and file size metadata
4. Creates output file
5. Receives and writes file content
6. Confirms successful transfer

## Error Handling

The system handles common errors:
- File not found
- Network connection issues
- Invalid IP addresses or ports
- Permission errors
- Incomplete transfers

## Future Enhancements

- [ ] Large file chunking (for files >100MB)
- [ ] Connection retry mechanism
- [ ] File encryption
- [ ] Multiple file transfers
- [ ] Progress bars
- [ ] Peer discovery over LAN
- [ ] Checksum verification

## Troubleshooting

### Common Issues:

1. **"Connection refused"**: Make sure the receiver is running and listening on the correct port
2. **"File not found"**: Check the file path and ensure the file exists
3. **"Permission denied"**: Ensure you have read permissions for the source file and write permissions for the destination
4. **Firewall blocking**: Check if your firewall is blocking the port

### Testing Network Connectivity:

```bash
# Test if port is reachable
telnet <receiver_ip> <port>
```

## License

This project is open source and available under the MIT License.
