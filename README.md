# bittorrent

# Instruction to run the program

# go run ./main.go "path to .torrent file" "file save name" 

# or 

# 1 go build ./main.go 
# 2 ./main "path to .torrent file" "file save name" 


# General description of all project folders


# bitField
The code snippet defines a Go package named "bitfield" that includes two functions operating on a slice of bytes. The slice of bytes is a bitfield that represents the availability of pieces of data in a larger data set, often used in peer-to-peer file sharing.

The first function, "HasPiece", takes an integer index as input and returns a boolean value indicating whether the corresponding bit in the bitfield is set or not. It does this by calculating the byte index and bit offset of the given index and checking if the corresponding bit in the byte is set.

The second function, "SetPiece", takes an integer index as input and sets the corresponding bit in the bitfield. It does this by calculating the byte index and bit offset of the given index and setting the corresponding bit in the byte.

Overall, these functions provide a convenient way to manipulate and check the availability of pieces of data in a larger data set represented as a bitfield.


# client
The code defines a Client struct that is a wrapper around a net.Conn and is used to communicate with peers implementing the BitTorrent protocol. The Client struct has methods for completing the handshake with a peer, receiving a bitfield message, sending various types of messages such as request, interested, not interested, unchoke, piece, have, and keep-alive messages. The New() function creates a new Client by dialing a connection to a peer and completing the handshake. The Close() method closes the connection. The Read() method reads and consumes a message from the connection.


# handshake
This is a Go language package for handling the handshake message used in the BitTorrent protocol. The package defines a struct HandShake with fields for the protocol string (Pstr), information hash (InfoHash), and peer ID (PeerID). The package provides functions for creating a new handshake message, serializing a handshake into a byte slice, and reading a handshake from a reader. The Read function reads a handshake message from an input stream and returns a pointer to a HandShake struct containing the message data.


# message
This is a Go package called "message" that contains functions and types to parse messages from the peer in the BitTorrent client implementation.

The package includes:

A messageID type that is an unsigned 8-bit integer (uint8).
Several constants representing different types of BitTorrent messages, each with a unique ID.
A Message type that stores the ID and payload of a message.
Functions to format and parse specific types of messages, including FormatPiece, FormatRequest, ParsePiece, ParseHave, and ParseRequest.


# peer2peer
This code contains the main logic for a BitTorrent client using peer-to-peer protocol.

The Torrent struct holds the data required to download a torrent such as Peers, PeerID, InfoHash, PieceHashes, PieceLength, Length, and Name.

The startDownloadWorker function starts a worker that downloads pieces from a peer and puts them on the results queue when done downloading them (or when an error occurs).

The readMessage function reads a message from the peer and updates the pieceProgress struct accordingly if the message is a piece message.

The attemptDownloadPiece function attempts to download a piece from a peer and returns the piece data or an error if it fails.

The Download function creates a work queue and a results queue, creates a worker for each peer, creates a pieceWork for each piece, puts it on the work queue, and waits for all pieces to be downloaded.

This implementation uses channels to communicate between workers. The workQueue channel is used to distribute work to the workers, and the results channel is used to collect the results.

This implementation also uses a pieceProgress struct to keep track of the progress of downloading a piece, including the number of bytes downloaded, the number of bytes requested, and the number of outstanding requests.

# peer
This is a Go package named "peers" which defines a Peer struct, an Unmarshal function, and a String method for the Peer struct.

The Peer struct has two fields: an IP field of type net.IP, and a Port field of type uint16.

The Unmarshal function takes a byte slice (peersBin) and returns a slice of Peer structs and an error if one occurred. It first checks if the length of peersBin is a multiple of the peerSize constant (6 bytes, 4 for IP, 2 for port), and returns an error if it is not. Then it creates a slice of Peer structs with a length of numPeers (calculated from the length of peersBin and peerSize), and fills each Peer struct's IP and Port fields by parsing the byte slice.

The String method for Peer struct returns a string representation of the Peer struct in the format of "IP:Port", where IP and Port are the respective fields of the Peer struct, using the JoinHostPort function from the net package to join them together.

# torrent
This is a Go language package that provides types and functions for connecting to peers, downloading and saving files using the BitTorrent protocol. The package includes a TorrentFile struct, which represents metadata of a .torrent file, and methods to parse the .torrent file and connect to peers.

The TorrentFile struct has the following fields: Announce (string), InfoHash ([20]byte), PieceHashes ([][20]byte), PieceLength (int), Length (int), and Name (string).

The bencodeInfo and bencodeTorrent types are used to parse the metadata in the .torrent file.

The ConnectToPeers function connects to peers concurrently and returns a slice of pointers to client.Client structs.

The DownloadToFile function downloads the file described by the torrent file and saves it to the specified path.

The Open function parses a .torrent file and returns a TorrentFile struct.

Overall, the package provides functionality to connect to peers, download files, and parse .torrent files, which are necessary components for BitTorrent clients.


# tracker
This code defines functions for handling tracker requests and responses in a BitTorrent client written in Go. The package torrent is imported along with other necessary Go libraries.

The bencodeTrackerResp type is defined to hold the decoded tracker response data, which includes an interval and a list of peer addresses.

The buildTrackerURL function takes a peer ID and port number, uses the url library to construct a URL with query parameters required by the BitTorrent protocol, and returns the URL as a string.

The requestPeers function uses the buildTrackerURL function to construct a tracker request URL, makes an HTTP GET request to the tracker, and decodes the response using the bencode library. It then returns a slice of peers.Peer structs containing the peer addresses in the tracker response.

Note that bencode and peers are custom packages used in this codebase and are not part of the standard Go library.


# main
This is the main entry point of a BitTorrent client program. The program takes two arguments: the path to the .torrent file and the path to the file to be downloaded. It opens the torrent file, connects to peers and downloads the file, and then starts seeding the file to the peers that are connected to it. The program waits for the user to press enter to exit.

The main() function first reads the input arguments and opens the torrent file using the torrent.Open() function. It then gets the torrent metadata using the GetTorrent() method of the TorrentFile type. It then connects to the peers using the ConnectToPeers() function and downloads the file to the specified output path using the DownloadToFile() method.

The function also starts seeding the file using the SeedFile() function from the seeder package in a separate goroutine. The program waits for the seeding to finish using a sync.WaitGroup and then prints a message indicating that the main function has completed. The program also sends keep-alive messages to the peers using a separate goroutine to maintain the connection.

If any errors occur during the process, the program logs the error using the log.Fatal() function and exits.