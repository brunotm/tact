package sftp

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/sshmgr/manager"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
	"github.com/brunotm/tact/js"
)

// File represents a remotely open file
// type File struct {
// 	*sftp.File
// 	s *sshmgr.SFTPctx
// }

// Close both file and sftp client ctx
// func (f *File) Close() (err error) {
// 	defer f.s.Close()
// 	return f.File.Close()
// }

// Open opens and returns the given file in RO mode
// func Open(ctx *tact.Context, filePath string) (file *File, err error) {
// 	// Get a sftp ctx for the host from the manager
// 	client, err := sshmgr.Manager.GetSFTPctx(ssh.NewSSHNodeConfig(ctx))
// 	if err != nil {
// 		return nil, fmt.Errorf("sftp: error getting sftp ctx: %s", err.Error())
// 	}

// 	// Open the remote file for reading, defaults to O_RDONLY
// 	f, err := sftpctx.Open(filePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("sftp: opening file %s: %s", filePath, err.Error())
// 	}
// 	return &File{File: f, s: sftpctx}, nil
// }

// Regex opens the given file and parses with the provided parser remembering the last read line
func Regex(ctx *tact.Context, fileName string, rex rexon.DataParser) (events <-chan []byte) {
	outCh := make(chan []byte)
	go regex(ctx, fileName, rex, outCh)
	return outCh
}

func regex(ctx *tact.Context, fileName string, rex rexon.DataParser, outCh chan<- []byte) {
	defer close(outCh)

	// Get the file path for this collector type from the node configuration
	filePath, ok := ctx.Node().LogFiles[fileName]
	if !ok {
		ctx.LogError("sftp: could not find file path for %s", fileName)
		return
	}

	// Get a sftp ctx for the host from the manager
	client, err := manager.SFTPClient(ssh.NewSSHNodeConfig(ctx))
	if err != nil {
		ctx.LogError("sftp: error getting sftp ctx: %s", err)
		return
	}
	defer client.Close()

	// Open the remote file for reading, defaults to O_RDONLY
	file, err := client.Open(filePath)
	if err != nil {
		ctx.LogError("sftp: opening file %s: %s", fileName, err)
		return
	}
	defer file.Close()

	// Read the first 2048 bytes for hashing
	// Used to guarantee we only seek to previous position if dealing with the same file
	hdr := make([]byte, 2048)
	if _, err := file.Read(hdr); err != nil {
		ctx.LogError("sftp: reading file for hashing: %s, error: %s", fileName, err)
		return
	}
	// Create a hash from the first 2048 bytes
	hash := tact.Hash(hdr)

	// Get the current stat of the file
	stat, err := file.Stat()
	if err != nil {
		ctx.LogError("sftp: fetching stat for file: %s, error: %s", fileName, err)
		return
	}

	// Create a document with the file current stat and hash
	currentFstat, _ := js.Set([]byte{}, hash, "hash")
	currentFstat, _ = js.Set(currentFstat, stat.ModTime(), "mtime")
	currentFstat, _ = js.Set(currentFstat, stat.Size(), "size")

	// Get the previous stat document from the KVStore
	// If it exists load the hash and seek position
	var seek int64
	var previousHash string
	previousFstat, err := ctx.Get([]byte(fileName))
	if err == nil && previousFstat != nil {
		seek, _ = js.GetInt(previousFstat, "size")
		previousHash, _ = js.GetUnsafeString(previousFstat, "hash")
	}

	// If the hashes match set the file offset to the last read
	// Else set it to the beginning of the file
	if hash == previousHash {
		if _, err := file.Seek(seek, 0); err != nil {
			ctx.LogError("sftp: seeking offset for file: %s, error: %s", fileName, err)
			return
		}
	} else {
		if _, err := file.Seek(0, 0); err != nil {
			ctx.LogError("sftp: seeking offset for file: %s, error: %s", fileName, err)
			return
		}
	}

	// Use the given rex object to parse the file
	rchan := rex.Parse(ctx.Context(), file)

	// Fetch and send parsed events to upstream ops
	for result := range rchan {
		for e := range result.Errors {
			ctx.LogError("sftp parser error", result.Errors[e])
		}

		if result.Data == nil {
			continue
		}

		if !tact.WrapCtxSend(ctx.Context(), outCh, result.Data) {
			ctx.LogError("sftp timed out sending event to upstream processing")
			return
		}
	}

	// Store the current file stat document
	if err := ctx.Set([]byte(fileName), currentFstat); err != nil {
		ctx.LogError("sftp storing fstat document for file: %s, error: %s", fileName, err)
	}

}
