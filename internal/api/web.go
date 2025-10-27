package api

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/config"
	"github.com/Quikmove/blockchain-uzd2/internal/crypto"
)

type Webserver struct {
	b      *crypto.Blockchain
	config *config.Config
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
	body := crypto.Body{Transactions: m.Transactions}
	go func() {
		block, err := crypto.GenerateBlock(ctx, last, body, ws.config.Version, ws.config.Difficulty)
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

func Run(ctx context.Context, b *crypto.Blockchain, c *config.Config, started chan<- struct{}) error {
	ws := &Webserver{b: b, config: c}
	router := ws.makeNewRouter()

	httpAddr := ws.config.Port
	log.Printf("Listening on :%s", httpAddr)

	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	ln, err := net.Listen("tcp", ":"+httpAddr)
	if err != nil {
		return err
	}
	if started != nil {
		close(started)
	}

	errChan := make(chan error, 1)
	go func() {
		err := s.Serve(ln)
		errChan <- err
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.Println("Shutting down web server...")
		if err := s.Shutdown(shutdownCtx); err != nil {
			return err
		}
		if err := <-errChan; err != nil && err != http.ErrServerClosed {
			return err
		}
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
	}
	return nil
}
