package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Block struct {
	Header Header
	Body   Body
}
type Hash32 [32]byte

func (h Hash32) HexString() string {
	return hex.EncodeToString(h[:])
}
func (h Hash32) MarshalJSON() ([]byte, error) {
	s := "\"" + h.HexString() + "\""
	return []byte(s), nil
}

type Header struct {
	Version    uint32
	Timestamp  uint32
	PrevHash   Hash32
	MerkleRoot Hash32
	Difficulty uint32
	Nonce      uint32
}
type Body struct {
	Transactions Transactions
}

func (h Header) Serialize() []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.LittleEndian, h.Version)
	buf.Write(reverse32(h.PrevHash))
	buf.Write(reverse32(h.MerkleRoot))
	_ = binary.Write(&buf, binary.LittleEndian, h.Timestamp)
	_ = binary.Write(&buf, binary.LittleEndian, h.Difficulty)
	_ = binary.Write(&buf, binary.LittleEndian, h.Nonce)

	return buf.Bytes()
}
func (h Header) Hash() Hash32 {
	b := h.Serialize()
	x := sha256.Sum256(b)
	return sha256.Sum256(x[:])
}
func reverse32(in Hash32) []byte {
	out := make([]byte, 32)
	for i := 0; i < 32; i++ {
		out[i] = in[31-i]
	}
	return out
}
func (b Body) MerkleRootHash() Hash32 {
	return MerkleRootHash(b.Transactions)
}

type Transaction struct {
	TxID     Hash32
	Sender   Hash32
	Receiver Hash32
	Amount   uint32
}

func (t Transaction) Hash() Hash32 {
	var buf bytes.Buffer
	buf.Write(t.TxID[:])
	buf.Write(t.Sender[:])
	buf.Write(t.Receiver[:])
	_ = binary.Write(&buf, binary.LittleEndian, t.Amount)

	return sha256.Sum256(buf.Bytes())
}

type User struct {
	Name      string
	PublicKey Hash32
}
type Transactions []Transaction

var ChainMutex sync.RWMutex
var Blockchain []Block

func HashBytes(bytes []byte) Hash32 {
	h := sha256.New()

	h.Write(bytes)

	return Hash32(h.Sum(nil))
}

func HashString(str string) Hash32 {
	return HashBytes([]byte(str))
}
func MerkleRootHash(t Transactions) Hash32 {
	if len(t) == 0 {
		return Hash32{}
	}

	hashes := make([]Hash32, 0, len(t))
	for _, tx := range t {
		hashes = append(hashes, tx.Hash())
	}

	for len(hashes) > 1 {
		// if odd number of nodes, duplicate last
		if len(hashes)%2 == 1 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}

		var next []Hash32
		for i := 0; i < len(hashes); i += 2 {
			left := hashes[i]
			right := hashes[i+1]

			hash := HashBytes(append(left[:], right[:]...))

			next = append(next, hash)
		}
		hashes = next
	}

	return hashes[0]
}

func calculateHash(block Block) Hash32 {
	// use 6 main header properties: prev block hash, timestamp, version, merkel root hash, nonce, difficulty target
	hash := block.Header.Hash()
	return hash
}
func isHashValid(diff uint32, hash Hash32) bool {
	for i := uint32(0); i < diff && i < 32; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	return true
}
func isBlockValid(newBlock, oldBlock Block) bool {

	if calculateHash(oldBlock) != newBlock.Header.PrevHash {
		return false
	}

	diff := newBlock.Header.Difficulty
	hash := calculateHash(newBlock)

	if isHashValid(diff, hash) {
		return true
	}

	return true
}

func generateBlock(oldBlock Block, body Body) (Block, error) {
	var newBlock Block

	t := time.Now()

	newBlock.Header.Timestamp = uint32(t.UnixNano())
	newBlock.Header.PrevHash = calculateHash(oldBlock)
	newBlock.Body = body

	return newBlock, nil

}

func replaceChain(newBlocks []Block) {
	ChainMutex.Lock()
	defer ChainMutex.Unlock()
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	ChainMutex.RLock()
	chainCopy := append([]Block(nil), Blockchain...)
	ChainMutex.RUnlock()
	bytes, err := json.MarshalIndent(chainCopy, "", "   ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

type Message struct {
	Transactions Transactions
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()
	ChainMutex.RLock()
	if len(Blockchain) == 0 {
		ChainMutex.RUnlock()
		respondWithJSON(w, r, http.StatusInternalServerError, "No genesis block")
		return
	}
	last := Blockchain[len(Blockchain)-1]
	ChainMutex.RUnlock()

	type Result struct {
		block Block
		err   error
	}
	ctx := r.Context()
	resCh := make(chan Result, 1)

	go func() {
		block, err := generateBlock(last, Body{})
		resCh <- Result{block, err}
	}()
	select {
	case <-ctx.Done():
		respondWithJSON(w, r, http.StatusRequestTimeout, "request cancelled")
	case gr := <-resCh:
		newBlock, err := gr.block, gr.err
		if err != nil {
			respondWithJSON(w, r, http.StatusInternalServerError, m)
			return
		}
		ChainMutex.Lock()
		currentLast := Blockchain[len(Blockchain)-1]
		if isBlockValid(newBlock, currentLast) {
			Blockchain = append(Blockchain, newBlock)
			ChainMutex.Unlock()
			respondWithJSON(w, r, http.StatusCreated, newBlock)
			return
		}
		ChainMutex.Unlock()
		respondWithJSON(w, r, http.StatusConflict, "Block invalid or chain advanced")
	}
}

func makeNewRouter() http.Handler {
	router := http.ServeMux{}
	router.HandleFunc("GET /", handleGetBlockchain)
	router.HandleFunc("POST /", handleWriteBlock)

	return &router
}

func run() error {
	router := makeNewRouter()

	httpAddr := os.Getenv("PORT")
	log.Println("Listening on ", httpAddr)

	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

type Users []User

func (u Users) GenerateFunds(low, high uint32) (Transactions, error) {
	txs := Transactions{}

	if len(u) == 0 || high < low {
		return txs, nil
	}

	for _, usr := range u {

		b := make([]byte, 16)
		_, _ = rand.Read(b)
		seed := hex.EncodeToString(b)

		sub := seed
		if len(sub) > 16 {
			sub = sub[:16]
		}

		amount := low
		if v, err := strconv.ParseUint(sub, 16, 32); err == nil {
			amount = uint32(v%uint64(high-low+1) + uint64(low))
		}

		// coinbase transaction creates an unspent output (UTXO) for the user
		tx := Transaction{
			TxID:     HashString("coinbase" + usr.PublicKey.HexString() + strconv.FormatUint(uint64(amount), 10) + strconv.FormatInt(time.Now().UnixNano(), 10)),
			Sender:   HashString("coinbase"),
			Receiver: usr.PublicKey,
			Amount:   amount,
		}

		txs = append(txs, tx)

		// small pause to vary timestamp-derived seeds
		time.Sleep(time.Millisecond)
	}

	return txs, nil
}
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	hashDifficulty := uint32(3)
	func() {
		t := time.Now()
		genesisBlock := Block{Header{1, uint32(t.UnixNano()), Hash32{}, Hash32{}, hashDifficulty, 1}, Body{Transactions: Transactions{}}}
		Blockchain = append(Blockchain, genesisBlock)
	}()
	log.Fatal(run())
}
