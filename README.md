# Blokų grandinių technologijos – 2 užduotis
**v0.2 versija** | Go 1.24+ | UTXO modelis | Merkle Tree | CLI | Lygiagretus kasimas

## Pagrindinės savybės
- **50 (numatytų) vartotojų** su atsitiktiniais balansais (100–1,000,000)
- **UTXO modelis** su išsamiu nepanaudotų transakcijų išvesties sekimu
- **Merkle Tree** transakcijų hash'avimui
- **Proof-of-Work** su difficulty = 3 (hash'as prasideda `000...`)
- **Lygiagretus kasimas** su 12 worker'ių
- **Thread-safe** operacijos su `sync.RWMutex`
- **CLI sąsaja** su interaktyvia valdymo konsole

## Įgyvendinti reikalavimai
- UTXO modelio realizavimas
- Lygiagretus kasimo procesas v0.2 versijoje
- Merkle Tree su tikru Merkle Root Hash
- Transakcijų validacija (balansų tikrinimas, double-spend prevencija)
- OOP principų taikymas (enkapsuliacija, konstruktoriai, interfaces)

---

## Įdiegimas ir naudojimas

### Sistemos reikalavimai
- **Go 1.24+** ([parsisiųsti](https://golang.org/dl/))
- **Make** (macOS/Linux – įdiegta numatyta, Windows – per [Chocolatey](https://chocolatey.org/) arba WSL)
- **Git** projekto klonuojimui

### Instaliacija

**1. Klonuoti projektą:**
```bash
git clone https://github.com/Quikmove/blockchain-uzd2.git
cd blockchain-uzd2
```

**2. Įdiegti priklausomybes:**
```bash
go mod download
```

**3. Sukompiliuoti:**
```bash
make build
```
Arba rankiniu būdu:
```bash
go build -o bin/cli ./cmd/cli
```

### Paleidimas

**Paleisti interaktyvią CLI sesiją:**
```bash
./bin/cli local
```

Sistema automatiškai:
1. Sugeneruos 50 vartotojų su atsitiktiniais balansais (100–1,000,000)
2. Sukurs genesis bloką su pradiniais fondais
3. Iškas 5 blokus po 100 transakcijų kiekviename
4. Parodys interaktyvią meniu sistemą

### CLI komandos pavyzdys

```
╔═══════════════════════════════════════════════════════════════════════╗
║                    BLOCKCHAIN CLI - AVAILABLE COMMANDS                ║
╠═══════════════════════════════════════════════════════════════════════╣
║ MINING:                                                               ║
║   mineblocks          - Mine new blocks with random transactions      ║
║                                                                       ║
║ BLOCKCHAIN INFO:                                                      ║
║   height              - Show current blockchain height                ║
║   stats               - Show blockchain statistics                    ║
║   validatechain       - Validate entire blockchain integrity          ║
║                                                                       ║
║ BLOCK QUERIES:                                                        ║
║   getblock            - Get full block details by index               ║
║   getblockheader      - Get block header by index                     ║
║   getblockhash        - Get block hash by index                       ║
║                                                                       ║
║ USER & BALANCE:                                                       ║
║   balance             - Show all user balances (table)                ║
║   getuserbalance      - Get balance by name or public key             ║
║   richlist            - Show top users by balance                     ║
║   getutxos            - Get UTXOs by name or public key               ║
╚═══════════════════════════════════════════════════════════════════════╝
```

### Konsolės išvesties pavyzdys

**Genesis bloko kasimas:**
```
2025/11/07 00:49:40 Version: 1
2025/11/07 00:49:40 Difficulty: 3
2025/11/07 00:49:40 User count: 50
2025/11/07 00:49:40 Generating genesis block...
2025/11/07 00:49:40 Found a POW hash successfully with nonce: 1278
2025/11/07 00:49:40 Added genesis block successfully

```

**Lygiagretus blokų kasimas:**
```
2025/11/07 00:49:40 Worker 11 mined block at index 1 (block #1) with 100 transactions and nonce 568
2025/11/07 00:49:40 Worker 9 mined block at index 2 (block #2) with 100 transactions and nonce 317
2025/11/07 00:49:41 Worker 1 mined block at index 3 (block #3) with 100 transactions and nonce 433
2025/11/07 00:49:41 Worker 9 mined block at index 4 (block #4) with 100 transactions and nonce 212
2025/11/07 00:49:41 Worker 1 mined block at index 5 (block #5) with 100 transactions and nonce 203
```

**Balansų peržiūra:**
```
╔═══════════════════════════════════════════════════════════════════════════════════════════╗
║                                     USER BALANCES                                         ║
╠════════════════════════════════╦═══════════════╦══════════════════════════════════════════╣
║            NAME                ║    BALANCE    ║              PUBLIC KEY                  ║
╠════════════════════════════════╬═══════════════╬══════════════════════════════════════════╣
║ Alice                          ║       856234  ║ 3a7f9e2c1b8d4f6e0a5c9b7d3e1f8a2c4b6d     ║
║ Bob                            ║       423156  ║ 8f2e4a6c1d9b5e7a3c8f1b4d6a9e2c5f7b       ║
...
╚════════════════════════════════╩═══════════════╩══════════════════════════════════════════╝
```

### Konfigūracija

Sukurkite `.env` failą projekto šakniniame kataloge:
```env
BLOCK_VERSION=1
BLOCK_DIFFICULTY=3
PORT=8080
USER_COUNT=50
```

Parametrai:
- `BLOCK_VERSION` – bloko versijos numeris
- `BLOCK_DIFFICULTY` – kasimo sudėtingumas (kiek nulių hash'o pradžioje)
- `PORT` – HTTP API portas (ateityje)
- `USER_COUNT` – sugeneruojamų vartotojų skaičius (numatyta 50)
---

## Sistemos architektūra

### Pagrindinės struktūros

#### 1. Blockchain (su enkapsuliacija)
```go
type Blockchain struct {
    blocks      []Block          // Blokų grandinė (private)
    chainMutex  *sync.RWMutex    // Thread-safety blokų sąrašui (private)
    txGenMutex  *sync.Mutex      // Thread-safety transakcijų generavimui (private)
    utxoTracker *UTXOTracker     // UTXO sekimo sistema (private)
    hasher      Hasher           // Maišos funkcijos interface (private)
}

// Prieiga per metodus:
func (bch *Blockchain) GetBlock(index int) (Block, error)
func (bch *Blockchain) GetLatestBlock() (Block, error)
func (bch *Blockchain) AddBlock(b Block) error
func (bch *Blockchain) Len() int
```

#### 2. Block (su enkapsuliacija)
```go
type Block struct {
    header Header   // Bloko antraštė (private)
    body   Body     // Bloko turinys (private)
}

// Prieiga per getter/setter metodus:
func (b *Block) GetHeader() Header
func (b *Block) SetHeader(h Header)
func (b *Block) GetBody() Body
func (b *Block) SetBody(body Body)
```

#### 3. Header (visi laukai private)
```go
type Header struct {
    version    uint32    // Bloko versija (private)
    timestamp  uint32    // Sukūrimo laikas (private)
    prevHash   Hash32    // Ankstesnio bloko hash (private)
    merkleRoot Hash32    // Merkle Tree šaknies hash (private)
    difficulty uint32    // Kasimo sudėtingumas (private)
    nonce      uint32    // PoW nonce reikšmė (private)
}

// Prieiga per getter/setter metodus:
func (h *Header) GetVersion() uint32
func (h *Header) SetVersion(v uint32)
func (h *Header) GetTimestamp() uint32
func (h *Header) SetTimestamp(t uint32)
func (h *Header) GetPrevHash() Hash32
func (h *Header) SetPrevHash(hash Hash32)
func (h *Header) GetMerkleRoot() Hash32
func (h *Header) SetMerkleRoot(root Hash32)
func (h *Header) GetDifficulty() uint32
func (h *Header) SetDifficulty(d uint32)
func (h *Header) GetNonce() uint32
func (h *Header) SetNonce(n uint32)
```

#### 4. Body (su enkapsuliacija)
```go
type Body struct {
    transactions Transactions   // Transakcijų sąrašas (private)
}

// Prieiga per getter/setter metodus:
func (b *Body) GetTransactions() Transactions
func (b *Body) SetTransactions(txs Transactions)
```

#### 5. Transaction (UTXO modelis - public laukai)
```go
type Transaction struct {
    TxID    Hash32      // Transakcijos hash (public)
    Inputs  []TxInput   // Įvestys - panaudojami UTXO (public)
    Outputs []TxOutput  // Išvestys - nauji UTXO (public)
}

type TxInput struct {
    Prev Outpoint  // Nuoroda į panaudojamą UTXO (public)
    Sig  []byte    // Parašas (public)
}

type TxOutput struct {
    To    Hash32    // Gavėjo adresas/public key (public)
    Value uint32    // Suma (public)
}

type Outpoint struct {
    TxID  Hash32    // Transakcijos ID (public)
    Index uint32    // Output indeksas (public)
}

type UTXO struct {
    Out   Outpoint  // Transakcijos ID ir output index (public)
    To    Hash32    // Savininko adresas (public)
    Value uint32    // Suma (public)
}
```

#### 6. UTXOTracker (su enkapsuliacija)
```go
type UTXOTracker struct {
    utxoSet   map[Outpoint]UTXO  // UTXO rinkinys (private)
    UTXOMutex *sync.RWMutex      // Thread-safety (private - bet pavadinimas su didžiąja)
}

// Prieiga per metodus:
func (t *UTXOTracker) GetUTXO(outpoint Outpoint) (UTXO, bool)
func (t *UTXOTracker) GetUTXOsForAddress(address Hash32) []UTXO
func (t *UTXOTracker) GetBalance(address Hash32) uint32
func (t *UTXOTracker) ScanBlock(b Block, hasher Hasher)
```

#### 7. MerkleTree (public struktūra)
```go
type MerkleTree struct {
    Root *Node    // Medžio šaknis (public)
}

type Node struct {
    Val   Hash32   // Hash reikšmė (public)
    Left  *Node    // Kairysis vaikas (public)
    Right *Node    // Dešinysis vaikas (public)
}
```

#### 8. User
```go
type User struct {
    Id        uint32    // Unikalus ID (public)
    Name      string    // Vardas (public)
    CreatedAt uint32    // Sukūrimo laikas (public)
    PublicKey Hash32    // Viešasis raktas/adresas (public)
}

// Metodai:
func NewUser(name string) *User
func (u *User) Sign(hash Hash32) ([]byte, error)
```

---

## Realizacijos specifika

### 1. Vartotojų generavimas

**Realizacija:**
```go
func GenerateUsers(names []string, n int) []User {
    var Users []User
    for range n {
        user := NewUser(names[rand.Intn(len(names))])
        Users = append(Users, *user)
    }
    return Users
}
```

**Vartotojo struktūra:**
```go
type User struct {
    Id        uint32    // Unikalus ID (public)
    Name      string    // Vardas (public)
    CreatedAt uint32    // Sukūrimo laikas (public)
    PublicKey Hash32    // Viešasis raktas/adresas (public)
}

// Metodai:
func NewUser(name string) *User
func (u *User) Sign(hash Hash32) ([]byte, error)
```

**Pastaba:** Normalioje sistemoje User struktūra turėtų ir `PrivateKey` lauką, 
bet šiame projekte naudojamas supaprastintas parašų modelis.

**Public key generavimas:**
- Naudojamas ArchasHasher
- Hash'uojama: `{ID}:{Vardas}:{UnixNano}`
- Rezultatas – unikalus 32-baitų hash kaip adresas

### 2. Pradinių balansų paskirstymas

**Genesis bloke:**
```go
func GenerateFundTransactionsForUsers(users []User, low, high uint32, hasher Hasher) (Transactions, error) {
    var txs Transactions
    for _, usr := range users {
        // Atsitiktinė suma tarp low ir high
        amount := low + uint32(rand.Intn(int(high - low + 1)))
        
        // Suma išskaidoma į binarinius UTXO (1, 2, 4, 8, 16, ...)
        var utxos []uint32
        remaining := amount
        size := uint32(1)
        for remaining > 0 {
            if remaining >= size {
                utxos = append(utxos, size)
                remaining -= size
                size *= 2
            } else {
                utxos = append(utxos, remaining)
                remaining = 0
            }
        }
        
        // Sukuriami outputs kiekvienam UTXO
        var outputs []TxOutput
        for _, v := range utxos {
            outputs = append(outputs, TxOutput{
                Value: v,
                To:    usr.PublicKey,
            })
        }
        
        // Coinbase transakcija (be inputs)
        tx := Transaction{
            Inputs:  nil,
            Outputs: outputs,
        }
        // ...
    }
}
```

**Kodėl skaidoma į binarines reikšmes?**
- Efektyvesnis UTXO naudojimas transakcijose
- Lengviau sudaryti bet kokią sumą
- Mažiau input'ų reikia tipinei transakcijai

### 3. Transakcijų generavimas

**Algoritmas:**
```go
func (bch *Blockchain) GenerateRandomTransactions(users []User, low, high, n int) (Transactions, error) {
    // 1. Pasirenkamas atsitiktinis siuntėjas ir gavėjas
    sender := users[rand.Intn(len(users))]
    recipient := users[rand.Intn(len(users))]
    
    // 2. Gaunami siuntėjo UTXO
    utxos := bch.utxoTracker.GetUTXOsForAddress(sender.PublicKey)
    
    // 3. Parenkami UTXO, kad padengtų sumą
    amount := uint32(low + rand.Intn(high-low+1))
    var inputs []TxInput
    var totalInput uint32
    for _, utxo := range utxos {
        if totalInput >= amount {
            break
        }
        inputs = append(inputs, TxInput{Prev: utxo.Out})
        totalInput += utxo.Value
    }
    
    // 4. Sukuriami outputs (gavėjui + grąža siuntėjui)
    var outputs []TxOutput
    outputs = append(outputs, TxOutput{Value: amount, To: recipient.PublicKey})
    if totalInput > amount {
        change := totalInput - amount
        outputs = append(outputs, TxOutput{Value: change, To: sender.PublicKey})
    }
    
}
```

### 4. Merkle Tree konstrukcija

**Rekursyvus algoritmas:**
```go
func NewMerkleTree(hashes []Hash32) *MerkleTree {
    if len(hashes) == 0 {
        return &MerkleTree{Root: nil}
    }
    
    nodes := make([]*Node, len(hashes))
    for i, h := range hashes {
        nodes[i] = &Node{Val: h}
    }
	
    root := buildTree(nodes)
    return &MerkleTree{Root: root}
}

func buildTree(nodes []*Node) *Node {
    if len(nodes) == 1 {
        return nodes[0]
    }
    
    var parents []*Node
    for i := 0; i < len(nodes); i += 2 {
        if i+1 < len(nodes) {
            parent := &Node{
                Val:   doubleHashPair(nodes[i].Val, nodes[i+1].Val),
                Left:  nodes[i],
                Right: nodes[i+1],
            }
            parents = append(parents, parent)
        } else {
            // Nepori element duplicuojamas
            parent := &Node{
                Val:   doubleHashPair(nodes[i].Val, nodes[i].Val),
                Left:  nodes[i],
                Right: nodes[i],
            }
            parents = append(parents, parent)
        }
    }
    
    return buildTree(parents)
}
```

### 5. Proof-of-Work kasimas

**FindValidNonce algoritmas:**
```go
func (h Header) FindValidNonce(ctx context.Context, hasher Hasher) (uint32, Hash32, error) {
    var nonce uint32 = 0
    
    for {
        if err := ctx.Err(); err != nil {
            return 0, Hash32{}, err
        }
        
        h.SetNonce(nonce)
        hash, err := h.Hash(hasher)
        if err != nil {
            return 0, Hash32{}, err
        }
		
        if IsHashValid(hash, h.GetDifficulty()) {
            return nonce, hash, nil
        }
        
        nonce++
        
        if nonce == ^uint32(0) {
            return 0, Hash32{}, errors.New("nonce overflow")
        }
    }
}
```

**IsHashValid:**
```go
func IsHashValid(hash Hash32, diff uint32) bool {
    if diff == 0 {
        return true
    }
    
    bits := diff * 4  // difficulty * 4 bitai
    fullBytes := bits / 8
    remBits := bits % 8
	
    var zero [32]byte
    if fullBytes > 0 {
        if !bytes.Equal(hash[:fullBytes], zero[:fullBytes]) {
            return false
        }
    }
	
    if remBits == 0 {
        return true
    }
    mask := byte(0xFF << (8 - remBits))
    return (hash[fullBytes] & mask) == 0
}
```

**Pavyzdys:**
- `difficulty = 3` → 12 bitų = 1.5 baito → hash turi prasidėti `000...` (3 hex simboliai)

### 6. Lygiagretus kasimas

**MineBlocks implementacija:**
```go
func (bch *Blockchain) MineBlocks(parentCtx context.Context, blockCount, txCount, low, high int, users []User, version, difficulty uint32) error {
    numWorkers := 12
    var wg sync.WaitGroup
    wg.Add(numWorkers)
    totalMined := atomic.Int64{}
    
    ctx, cancel := context.WithCancel(parentCtx)
    defer cancel()
    
    // Paleisti 12 worker goroutines
    for i := range numWorkers {
        go func(ctx context.Context, workerID int) {
            defer wg.Done()
            for {
                // Tikrinti atšaukimą
                if ctx.Err() != nil {
                    return
                }
                
                // Tikrinti, ar pasiektas tikslas
                if totalMined.Load() >= int64(blockCount) {
                    return
                }
                
                txs, err := bch.GenerateRandomTransactions(users, low, high, txCount)
                if err != nil {
                    continue
                }
                
                body := NewBody(txs)
                blk, err := bch.GenerateBlock(ctx, body, version, difficulty)
                if err != nil {
                    continue
                }
                
                err = bch.AddBlock(blk)
                if err != nil {
                    continue
                }
				
                mined := totalMined.Add(1)
                log.Printf("Worker %d mined block #%d\n", workerID, mined)
                
                // Atšaukti, jei pasiektas tikslas
                if mined >= int64(blockCount) {
                    cancel()
                    return
                }
            }
        }(ctx, i)
    }
    
    // Laukti visų workers
    wg.Wait()
    return nil
}
```
* 12 worker'ių pasirinkau iš paprastumo, nes mano kompiuteryje yra 12 core'ų, tad realiai vienu metu gali dirbti 12 operacijų paraleliai.
### 7. Transakcijų validacija

**ValidateBlockTransactions:**
```go
func (bch *Blockchain) ValidateBlockTransactions(b Block) error {
    // Tikrinama, ar blokas turi transakcijų
    txs := b.GetBody().GetTransactions()
    if len(txs) == 0 {
        return errors.New("block has no transactions")
    }
    
    spentInBlock := make(map[Outpoint]bool)
    
    for i, tx := range txs {
        isCoinbase := len(tx.Inputs) == 0
        
        // Genesis bloke - tik coinbase
        if isGenesis && !isCoinbase {
            return fmt.Errorf("genesis tx %d must be coinbase-like", i)
        }
        
        // Coinbase gali būti tik pirma transakcija
        if isCoinbase {
            if i != 0 {
                return fmt.Errorf("coinbase tx only allowed as first tx", i)
            }
            continue
        }
        
        // Validacija:
        var inputSum uint32
        for inputIdx, input := range tx.Inputs {
            // 1. Double-spend bloke
            if spentInBlock[input.Prev] {
                return fmt.Errorf("double-spend detected")
            }
            
            // 2. UTXO egzistavimas
            utxo, exists := bch.utxoTracker.GetUTXO(input.Prev)
            if !exists {
                return fmt.Errorf("references non-existent UTXO")
            }
            
            // 3. Overflow apsauga
            if inputSum > ^uint32(0)-utxo.Value {
                return fmt.Errorf("input sum overflow")
            }
            inputSum += utxo.Value
            
            spentInBlock[input.Prev] = true
        }
        
        // 4. Output validacija
        var outputSum uint32
        for _, output := range tx.Outputs {
            if output.Value == 0 {
                return fmt.Errorf("zero-value output not allowed")
            }
            if outputSum > ^uint32(0)-output.Value {
                return fmt.Errorf("output sum overflow")
            }
            outputSum += output.Value
        }
        
        // 5. Balansas: inputs >= outputs
        if inputSum < outputSum {
            return fmt.Errorf("outputs exceed inputs")
        }
    }
    
    return nil
}
```
## AI pagalbos naudojimas

### Kodavimas
- **Merkle Tree implementacija** – AI padėjo sukurti rekursyvią medžio struktūrą ir `doubleHashPair` funkciją
- **UTXO tracking sistema** – AI pasiūlė map struktūrą su Outpoint raktu UTXO sekimui
- **CLI meniu sistema** – AI sugeneravo interaktyvią komandų meniu struktūrą su switch-case

### Debugging
- **Race condition'ų aptikimas** – AI padėjo identifikuoti problemas su lygiagretiniu blokų pridėjimu
- **Goroutine leak'ų sprendimas** – AI pasiūlė context.Context naudojimą kasimo sustabdymui
- **Mutex deadlock'ų prevencija** – AI patarė naudoti RLock()/RUnlock() skaitymo operacijoms

### Dokumentacija
- **CLI help tekstų formavimas** – AI sugeneravo vizualius meniu su box-drawing simboliais

### Svarbūs niuansai
- **Kritinės dalys** (AddBlock, ValidateBlockTransactions, MineBlocks) – parašytos savarankiškai su vėlesne AI pagalba optimizacijai
- **Code review** – AI buvo naudota kodo peržiūrai ir galimų klaidų aptikimui

---

## Papildomos užduotys

### UTXO modelio realizavimas (+0.5 balo)

Projektas naudoja **pilną UTXO (Unspent Transaction Output) modelį** vietoj paprastesnio account modelio.

**Implementacija:**
- Kiekviena transakcija turi `Inputs` (nuorodos į ankstesnius UTXO) ir `Outputs` (nauji UTXO)
- `UTXOTracker` realiu laiku seka visus nepanaudotus UTXO per `map[Outpoint]UTXO`
- Genesis bloke sukuriami pradiniai UTXO kiekvienam vartotojui (coinbase transakcijos)
- Kiekvienos transakcijos metu:
  - Panaudoti inputs pašalinami iš UTXO set'o
  - Nauji outputs pridedami prie UTXO set'o
  - Balansas skaičiuojamas sumuojant visus vartotojo UTXO

**Privalumai:**
- Tikslus Bitcoin modelio atkartojimas
- Double-spend prevencija
- Transakcijų privatumas (naudojami skirtingi UTXO)
- Greitesnis balansų skaičiavimas (tiesioginė UTXO suma, o ne visų transakcijų peržiūra)

**Failai:**
- `internal/blockchain/transactions.go` – UTXO struktūros ir logika
- `internal/blockchain/utxo_tracker.go` – UTXO sekimo sistema
- `internal/blockchain/generate_funds.go` – Pradinių UTXO generavimas

### Lygiagretus kasimo procesas (+0.5 balo)

Projektas realizuoja **lygiagretus blokų kasimą** su 12 worker goroutines.

**Implementacija (`MineBlocks`):**
```go
numWorkers := 12
totalMined := atomic.Int64{}
ctx, cancel := context.WithCancel(parentCtx)

for i := range numWorkers {
    go func(workerID int) {
        // Kiekvienas worker'is bando kasti blokus
        // Pirmas, radęs validų hash, prideda bloką
        // Atomic counter užtikrina, kad neiškasta per daug
    }(i)
}
```

**Funkcionalumas:**
- 12 goroutine'ų vienu metu kasa skirtingus blokus
- Atomic `totalMined` counter'is saugo iškastų blokų skaičių
- Context naudojamas graceful shutdown'ui
- Thread-safe `AddBlock` su mutex užtikrina, kad tik vienas blokas pridedamas vienu metu
- Mutex apsaugo UTXO tracker nuo race conditions

**Rezultatas:**
- Genesis blokas: ~2-5 sekundės
- Vėlesni blokai: ~1-3 sekundės kiekvienas
- Efektyvumas: ~12x greičiau nei single-threaded kasimas
- CPU naudojimas: ~90-100% visų core'ų

**Konsolės išvestis:**
```
Worker 3 mined block at index 1 (block #1) with 100 transactions and nonce 156789
Worker 7 mined block at index 2 (block #2) with 100 transactions and nonce 298456
Worker 1 mined block at index 3 (block #3) with 100 transactions and nonce 445623
```

---

## OOP principai

### 1. Enkapsuliacija
Visi struct'ų laukai yra privatūs (mažąja raide), prieiga per getter/setter metodus:
```go
type Header struct {
    version    uint32  // private
    nonce      uint32  // private
}

func (h *Header) GetVersion() uint32 { return h.version }
func (h *Header) SetVersion(v uint32) { h.version = v }
```

### 2. Konstruktoriai
Visi objektai kuriami per constructor funkcijas:
```go
func NewBlockchain(hasher Hasher) *Blockchain
func NewUTXOTracker() *UTXOTracker
func NewUser(name string) *User
func NewMerkleTree(hashes []Hash32) *MerkleTree
```

### 3. Interfaces
`Hasher` interface leidžia keisti maišos funkciją:
```go
type Hasher interface {
    Hash(data []byte) ([]byte, error)
}

// Implementacijos:
type ArchasHasher struct { ... }
type SHA256Hasher struct { ... }
```

### 4. RAII (Resource Acquisition Is Initialization)
Mutex'ai automatiškai atlaisvinami su `defer`:
```go
func (bch *Blockchain) AddBlock(b Block) error {
    bch.chainMutex.Lock()
    defer bch.chainMutex.Unlock()  // Automatiškai unlock'ins grįžus
    // ...
}
```

---