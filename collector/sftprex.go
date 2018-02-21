package collector

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/sshmgr"
	"github.com/brunotm/tact"
)

// SFTPRex collector base
func SFTPRex(session *tact.Session, fileName string, rex rexon.Parser) <-chan []byte {
	outChan := make(chan []byte)
	go sftpRex(session, fileName, rex, outChan)
	return outChan
}

func sftpRex(session *tact.Session, fileName string, rex rexon.Parser, outChan chan<- []byte) {
	defer close(outChan)

	// Get the file path for this collector type from the node configuration
	filePath, ok := session.Node().LogFiles[fileName]
	if !ok {
		session.LogErr("sftprex: could not find file path for %s", fileName)
		return
	}

	// Get a sftp session for the host from the manager
	sftpSession, err := sshmgr.Manager.GetSFTPSession(NewSSHNodeConfig(session.Node()))
	if err != nil {
		session.LogErr("sftprex: error getting sftp session: %s", err.Error())
		return
	}
	defer sftpSession.Close()

	// Open the remote file for reading, defaults to O_RDONLY
	file, err := sftpSession.Open(filePath)
	if err != nil {
		session.LogErr("sftprex: opening file %s: %s", fileName, err.Error())
		return
	}

	// Read the first 2048 bytes for hashing
	// Used to guarantee we only seek to previous position if dealing with the same file
	hdr := make([]byte, 2048)
	if _, err := file.Read(hdr); err != nil {
		session.LogErr("sftprex: reading file for hashing: %s, error: %s", fileName, err.Error())
		return
	}
	// Create a hash from the first 2048 bytes
	hash := tact.Blake2b(hdr)

	// Get the current stat of the file
	stat, err := file.Stat()
	if err != nil {
		session.LogErr("sftprex: fetching stat for file: %s, error: %s", fileName, err.Error())
		return
	}

	// Create a document with the file current stat and hash
	currentFstat, _ := rexon.JSONSet([]byte{}, hash, "hash")
	currentFstat, _ = rexon.JSONSet(currentFstat, stat.ModTime(), "mtime")
	currentFstat, _ = rexon.JSONSet(currentFstat, stat.Size(), "size")

	// Get the previous stat document from the KVStore
	// If it exists load the hash and seek position
	var seek int64
	var previousHash string
	previousFstat, err := session.Get([]byte(fileName))
	if err == nil && previousFstat != nil {
		seek, _ = rexon.JSONGetInt(previousFstat, "size")
		previousHash, _ = rexon.JSONGetUnsafeString(previousFstat, "hash")
	}

	// If the hashes match set the file offset to the last read
	// Else set it to the beginning of the file
	if hash == previousHash {
		if _, err := file.Seek(seek, 0); err != nil {
			session.LogErr("sftprex: seeking offset for file: %s, error: %s", fileName, err.Error())
			return
		}
	} else {
		if _, err := file.Seek(0, 0); err != nil {
			session.LogErr("sftprex: seeking offset for file: %s, error: %s", fileName, err.Error())
			return
		}
	}

	// Use the given rex object to parse the file
	rchan := rex.Parse(session.Context(), file)

	// Fetch and send parsed events to upstream ops
	for result := range rchan {
		for e := range result.Errors {
			session.LogErr(result.Errors[e].Error())
		}

		if result.Data == nil {
			continue
		}

		if !tact.WrapCtxSend(session.Context(), outChan, result.Data) {
			session.LogErr("sshrex: timed out sending event to upstream processing")
			return
		}
	}

	// Store the current file stat document
	if err := session.Set([]byte(fileName), currentFstat); err != nil {
		session.LogErr("sftprex: storing fstat document for file: %s, error: %s", fileName, err.Error())
	}

}
