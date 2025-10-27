package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/crypto"
)

type Webserver struct {
	b *crypto.Blockchain
}

func (ws *Webserver) handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	ws.b.ChainMutex.RLock()
	chainCopy := append([]crypto.Block(nil), ws.b.Blocks...)
	ws.b.ChainMutex.RUnlock()
	bytes, err := json.MarshalIndent(chainCopy, "", "   ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

type Message struct {
	Transactions crypto.Transactions
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

func (ws *Webserver) handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()
	ws.b.ChainMutex.RLock()
	if len(ws.b.Blocks) == 0 {
		ws.b.ChainMutex.RUnlock()
		respondWithJSON(w, r, http.StatusInternalServerError, "No genesis block")
		return
	}
	last := ws.b.Blocks[len(ws.b.Blocks)-1]
	ws.b.ChainMutex.RUnlock()

	type Result struct {
		block crypto.Block
		err   error
	}
	ctx := r.Context()
	resCh := make(chan Result, 1)

	go func() {
		block, err := crypto.GenerateBlock(last, crypto.Body{})
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
		ws.b.ChainMutex.Lock()
		if ws.b.IsBlockValid(newBlock) {
			ws.b.Blocks = append(ws.b.Blocks, newBlock)
			ws.b.ChainMutex.Unlock()
			respondWithJSON(w, r, http.StatusCreated, newBlock)
			return
		}
		ws.b.ChainMutex.Unlock()
		respondWithJSON(w, r, http.StatusConflict, "Block invalid or chain advanced")
	}
}

func (ws *Webserver) makeNewRouter() http.Handler {
	router := http.ServeMux{}
	router.HandleFunc("GET /", ws.handleGetBlockchain)
	router.HandleFunc("POST /", ws.handleWriteBlock)

	return &router
}

func Run(b *crypto.Blockchain) error {
	ws := &Webserver{b: b}
	router := ws.makeNewRouter()

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
