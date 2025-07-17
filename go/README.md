# File Share System

A simple peer-to-peer file sharing system built in Go that allows you to send files between devices over a local network.

## Features

- ✅ Send files over LAN or same Wi-Fi network
- ✅ No internet connection required
- ✅ Simple CLI interface
- ✅ Supports any file type
- ✅ Shows file metadata (name, size)
- ✅ Progress feedback during transfer

## Usage

### Basic Commands

#### Send a file:
```bash
go run main.go send <ip> <port_no>
```
You'll be prompted to enter the file path (just the filename if it's in the current directory, or full path).

#### Receive a file:
```bash
go run main.go receive <port_no>
```

### Example Usage

#### Step 1: Start the receiver (on destination device)
```bash
go run main.go receive 9000
```
This will:
- Start listening on port 9000
- Show your IP address for others to connect to
- Wait for incoming file transfers

#### Step 2: Send the file (from source device)
```bash
go run main.go send 192.168.1.5 9000
```
This will:
- Prompt you to enter the file path (e.g., "document.pdf" or "C:\Users\user\file.txt")
- Connect to the device at IP `192.168.1.5` on port 9000
- Transfer the file content

### Testing on Same Machine

You can test the system on the same machine using `localhost`:

**Terminal 1 (Receiver):**
```bash
go run main.go receive 9000
```

**Terminal 2 (Sender):**
```bash
go run main.go send localhost 9000
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
