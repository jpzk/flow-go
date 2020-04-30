package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/crypto/nacl/box"
	"google.golang.org/api/option"

	"github.com/dapperlabs/flow-go/model/bootstrap"
)

const (
	FilenameTransitKeyPub      = "%v.transit-key.pub"
	FilenameTransitKeyPriv     = "%v.transit-key.priv"
	FilenameRandomBeaconCipher = bootstrap.FilenameRandomBeaconPriv + ".enc"
)

const fileMode = os.FileMode(0644)

const flowBucket = "flow-genesis-bootstrap"

var (
	filesToUpload = []string{
		FilenameTransitKeyPub,
		bootstrap.FilenameNodeInfoPub,
	}
	filesToDownload = []string{
		bootstrap.FilenameDKGDataPub,
		bootstrap.FilenameNodeInfosPub,
		bootstrap.FilenameGenesisBlock,
	}
)

func main() {

	var bootdir, keydir, wrapId string
	var pull, push bool

	flag.StringVar(&bootdir, "d", "~/bootstrap", "The bootstrap directory containing your node-info files")
	flag.StringVar(&keydir, "t", "", "Token provided by the Flow team to access the transit server")
	flag.BoolVar(&pull, "pull", false, "Fetch keys and metadata from the transit server")
	flag.BoolVar(&push, "push", false, "Upload public keys to the transit server")
	flag.StringVar(&wrapId, "x-server-wrap", "", "(Flow Team Use), wrap response keys")
	flag.Parse()

	// Wrap takes prescedence, so we just do that first
	if wrapId != "" {
		log.Printf("Wrapping response for node %s\n", wrapId)
		err := wrapFile(bootdir, wrapId)
		if err != nil {
			log.Fatalf("Failed to wrap response: %s\n", err)
		}
		return
	}

	var err error
	if pull && push {
		flag.Usage()
		log.Fatal("Only one of -pull or -push may be specified\n")
	}

	if !(pull || push) {
		flag.Usage()
		log.Fatal("One of -pull or -push must be specified")
	}

	if keydir == "" {
		flag.Usage()
		log.Fatal("Access key required")
	}

	nodeId, err := fetchNodeId(bootdir)
	if err != nil {
		log.Fatalf("Could not determine node ID: %s\n", err)
	}

	// timeout remote operations
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if push {
		runPush(ctx, bootdir, keydir, nodeId)
		return
	}

	if pull {
		runPull(ctx, bootdir, keydir, nodeId)
		return
	}
}

// Read the NodeID file to build other filenames from
func fetchNodeId(bootdir string) (string, error) {
	path := filepath.Join(bootdir, bootstrap.FilenameNodeId)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Error reading file %s: %w", path, err)
	}

	return strings.TrimSpace(string(data)), nil
}

// Run the push process
// - create transit keypair
// - upload files to GCS bucket
func runPush(ctx context.Context, bootdir, token, nodeId string) {
	log.Println("Running push")
	err := generateKeys(bootdir, nodeId)
	if err != nil {
		log.Fatalf("Failed to push: %s", err)
	}
	for _, file := range filesToUpload {
		err = bucketUpload(ctx, bootdir, fmt.Sprintf(file, nodeId), token)
		if err != nil {
			log.Fatalf("Failed to push: %s", err)
		}
	}
}

func runPull(ctx context.Context, bootdir, token, nodeId string) {
	filesToDownload = append(filesToDownload, fmt.Sprintf(FilenameRandomBeaconCipher, nodeId))
	log.Println("Running pull")
	var err error
	for _, file := range filesToDownload {
		err = bucketDownload(ctx, bootdir, file, token)
		if err != nil {
			log.Fatalf("Failed to pull: %s", err)
		}
	}

	err = unwrapFile(bootdir, nodeId)
	if err != nil {
		log.Fatalf("Failed to pull: %s", err)
	}
}

// generateKeys creates the transit keypair and writes them to disk for later
func generateKeys(bootdir, nodeId string) error {

	privPath := filepath.Join(bootdir, fmt.Sprintf(FilenameTransitKeyPriv, nodeId))
	pubPath := filepath.Join(bootdir, fmt.Sprintf(FilenameTransitKeyPub, nodeId))

	log.Print("Generating keypair")

	// Generate the keypair
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("Failed to create keys: %w", err)
	}

	// Write private key file
	err = ioutil.WriteFile(privPath, priv[:], fileMode)
	if err != nil {
		return fmt.Errorf("Failed to write pivate key file: %w", err)
	}

	// Write public key file
	err = ioutil.WriteFile(pubPath, pub[:], fileMode)
	if err != nil {
		return fmt.Errorf("Failed to write public key file: %w", err)
	}

	return nil
}

func unwrapFile(bootdir, nodeId string) error {

	log.Print("Decrypting Random Beacon key")

	pubKeyPath := filepath.Join(bootdir, fmt.Sprintf(FilenameTransitKeyPub, nodeId))
	privKeyPath := filepath.Join(bootdir, fmt.Sprintf(FilenameTransitKeyPriv, nodeId))
	ciphertextPath := filepath.Join(bootdir, fmt.Sprintf(FilenameRandomBeaconCipher, nodeId))
	plaintextPath := filepath.Join(bootdir, fmt.Sprintf(bootstrap.FilenameRandomBeaconPriv, nodeId))

	ciphertext, err := ioutil.ReadFile(ciphertextPath)
	if err != nil {
		return fmt.Errorf("Failed to open ciphertext file %s: %w", ciphertextPath, err)
	}

	publicKey, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return fmt.Errorf("Failed to open public keyfile %s: %w", pubKeyPath, err)
	}

	privateKey, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return fmt.Errorf("Failed to open private keyfile %s: %w", privKeyPath, err)
	}

	// NaCl is picky and wants its type to be exactly a [32]byte, but readfile reads a slice
	var pubKeyBytes, privKeyBytes [32]byte
	copy(pubKeyBytes[:], publicKey)
	copy(privKeyBytes[:], privateKey)

	plaintext := make([]byte, 0, len(ciphertext)-box.AnonymousOverhead)
	plaintext, ok := box.OpenAnonymous(plaintext, ciphertext, &pubKeyBytes, &privKeyBytes)
	if !ok {
		return fmt.Errorf("Failed to decrypt ciphertext: unknown error in NaCl")
	}

	err = ioutil.WriteFile(plaintextPath, plaintext, fileMode)
	if err != nil {
		return fmt.Errorf("Failed to write the decrypted file %s: %w", plaintextPath, err)
	}

	return nil
}

func wrapFile(bootdir, nodeId string) error {
	pubKeyPath := filepath.Join(bootdir, fmt.Sprintf(FilenameTransitKeyPub, nodeId))
	plaintextPath := filepath.Join(bootdir, fmt.Sprintf(bootstrap.FilenameRandomBeaconPriv, nodeId))
	ciphertextPath := filepath.Join(bootdir, fmt.Sprintf(FilenameRandomBeaconCipher, nodeId))

	plaintext, err := ioutil.ReadFile(plaintextPath)
	if err != nil {
		return fmt.Errorf("Failed to open plaintext file %s: %w", plaintextPath, err)
	}

	publicKey, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return fmt.Errorf("Faield to open public keyfile %s: %w", pubKeyPath, err)
	}

	var pubKeyBytes [32]byte
	copy(pubKeyBytes[:], publicKey)

	ciphertext := make([]byte, 0, len(plaintext)+box.AnonymousOverhead)

	ciphertext, err = box.SealAnonymous(ciphertext, plaintext, &pubKeyBytes, rand.Reader)
	if err != nil {
		return fmt.Errorf("Could not encrypt file: %w", err)
	}

	err = ioutil.WriteFile(ciphertextPath, ciphertext, fileMode)
	if err != nil {
		return fmt.Errorf("Error writing ciphertext: %w", err)
	}

	return nil
}

func bucketUpload(ctx context.Context, bootdir, filename, token string) error {

	gcsClient, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return fmt.Errorf("Failed to initialize GCS client: %w", err)
	}
	defer gcsClient.Close()

	path := filepath.Join(bootdir, filename)
	log.Printf("Uploading %s\n", path)

	upload := gcsClient.Bucket(flowBucket).
		Object(filepath.Join(token, filename)).
		NewWriter(ctx)
	defer func() {
		err := upload.Close()
		if err != nil {
			log.Fatalf("Failed to close writer stream: %s\n", err)
		}
	}()

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Error opening upload file: %w", err)
	}
	defer file.Close()

	n, err := io.Copy(upload, file)
	if err != nil {
		return fmt.Errorf("Error uploading file: %w", err)
	}

	log.Printf("Uploaded %d bytes\n", n)

	return nil
}

func bucketDownload(ctx context.Context, bootdir, filename, token string) error {

	gcsClient, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return fmt.Errorf("Failed to initialize GCS client: %w", err)
	}
	defer gcsClient.Close()

	path := filepath.Join(bootdir, filename)
	log.Printf("Downloading %s\n", path)

	download, err := gcsClient.Bucket(flowBucket).
		Object(filepath.Join(token, filename)).
		NewReader(ctx)
	if err != nil {
		return fmt.Errorf("Error creating GCS object reader: %w", err)
	}
	defer download.Close()

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Error creating download file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, download)
	if err != nil {
		return fmt.Errorf("Error downloading file: %w", err)
	}

	return nil
}