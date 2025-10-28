package api

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Quikmove/blockchain-uzd2/internal/blockchain"
	"github.com/Quikmove/blockchain-uzd2/internal/config"
)

type Webserver struct {
	b   *blockchain.Blockchain
	cfg *config.Config
}

func (ws *Webserver) handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	chainCopy := ws.b.Blocks()
	bytes, err := json.MarshalIndent(chainCopy, "", "   ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

type Message struct {
	Transactions blockchain.Transactions
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
	if ws.b.Len() == 0 {
		respondWithJSON(w, r, http.StatusInternalServerError, "No genesis block")
		return
	}
	last, err := ws.b.GetLatestBlock()
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, "Failed to get latest block")
		return
	}

	type Result struct {
		block blockchain.Block
		err   error
	}
	ctx := r.Context()
	resCh := make(chan Result, 1)
	body := blockchain.Body{Transactions: m.Transactions}
	go func() {
		block, err := blockchain.GenerateBlock(ctx, last, body, ws.cfg.Version, ws.cfg.Difficulty)
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
		if err := ws.b.AddBlock(newBlock); err == nil {
			respondWithJSON(w, r, http.StatusCreated, newBlock)
			return
		}
		respondWithJSON(w, r, http.StatusConflict, "Block invalid or chain advanced")
	}
}

func (ws *Webserver) makeNewRouter() http.Handler {
	router := http.ServeMux{}
	router.HandleFunc("GET /", ws.handleGetBlockchain)
	router.HandleFunc("POST /", ws.handleWriteBlock)

	return &router
}

func Run(ctx context.Context, b *blockchain.Blockchain, c *config.Config, started chan<- struct{}) error {
	ws := &Webserver{b: b, cfg: c}
	router := ws.makeNewRouter()

	httpAddr := ws.cfg.Port
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
