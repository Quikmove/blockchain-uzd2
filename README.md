# Blokų grandinių technologijos – 2 užduotis
**v0.2 versija** | Go 1.24+ | UTXO modelis | Merkle Tree | CLI | Lygiagretus kasimas

## Anotacija

Tai yra **centralizuota blokų grandinės (blockchain) implementacija** Go kalba, realizuojanti **UTXO (Unspent Transaction Output)** modelį ir **Proof-of-Work (PoW)** konsensuso mechanizmą. Sistema generuoja vartotojus, kuria transakcijas, formuoja blokus ir juos "kasa" naudojant modifikuotą maišos funkciją **ArchasHasher** iš 1-os užduoties.

### Pagrindinės savybės
- **50 vartotojų** su atsitiktiniais balansais (100–1,000,000)
- **UTXO modelis** su išsamiu nepanaudotų transakcijų išvesties sekimu
- **Merkle Tree** transakcijų hash'avimui
- **Proof-of-Work** su difficulty = 3 (hash'as prasideda `000...`)
- **Lygiagretus kasimas** su 12 darbuotojų (workers)
- **Thread-safe** operacijos su `sync.RWMutex`
- **CLI sąsaja** su interaktyvia valdymo konsole

### Įgyvendinti reikalavimai
- UTXO modelio realizavimas (+0.5 balo)
- Lygiagretus kasimo procesas v0.2 versijoje (+0.5 balo)
- Merkle Tree su tikru Merkle Root Hash
- Transakcijų validacija (balansų tikrinimas, double-spend prevencija)
- OOP principų taikymas (enkapsuliacija, konstruktoriai, interfaces)

---

## Funkcionalumas

### Versija v0.1 (2025-10-29)

**Pagrindinė realizacija:**
- **Centralizuota blokų grandinė** su modifikuota ArchasHasher maišos funkcija
- **Vartotojų generavimas** – 50 vartotojų su atsitiktiniais balansais
- **Genesis blokas** – pradinis blokas su fondų paskirstymu vartotojams
- **Transakcijų kūrimas** – atsitiktinių transakcijų generavimas tarp vartotojų
- **Proof-of-Work kasimas** – hash'o paieška su `difficulty = 3`
- **Konsolės išvestis** – blokų kasimo proceso vizualizacija

**Realizuotos funkcijos:**
```go
// Blockchain struktūra
type Blockchain struct {
    blocks      []Block
    chainMutex  *sync.RWMutex
    txGenMutex  *sync.Mutex
    utxoTracker *UTXOTracker
    hasher      Hasher
}
```

### Versija v0.2

**Naujos funkcijos:**
- **Merkle Tree implementacija** (`internal/merkletree/merkletree.go`)
  - Rekursyvi medžio konstrukcija
  - Tikras Merkle Root Hash skaičiavimas
  - Transakcijų autentiškumo tikrinimas
  
- **Transakcijų validacija** (`ValidateBlockTransactions`)
  - Balansų tikrinimas per UTXO tracker
  - UTXO egzistavimo patikra
  - Double-spend prevencija bloko viduje
  - Input/output sumų palyginimas
  - Overflow apsauga
  
- **Lygiagretus kasimas** (`MineBlocks`)
  - 12 goroutine'ų kasa blokus lygiagrečiai
  - Context-based atšaukimas
  - Atomic counter'is suskaičiuoja iškastus blokus
  - Thread-safe blockchain prieiga
  
- **Thread-safe UTXO tracker**
  - `sync.RWMutex` užtikrina thread-safety
  - Realaus laiko balansų sekimas
  - UTXO set atnaujinimas kiekvieno bloko metu

**Papildymai:**
- Patobulinta kodo struktūra su geresnėmis OOP praktikomis
- Race condition'ų pašalinimas
- Goroutine leak'ų prevencija
- Išsamus README su pavyzdžiais

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
3. Iškasa 5 blokus po 100 transakcijų kiekviename
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
2025/11/07 15:23:45 Version: 1
2025/11/07 15:23:45 Difficulty: 3
2025/11/07 15:23:45 Generating genesis block...
2025/11/07 15:23:47 Found a POW hash successfully with nonce: 234567
2025/11/07 15:23:47 Added genesis block successfully
```

**Lygiagretus blokų kasimas:**
```
2025/11/07 15:23:50 Worker 3 mined block at index 1 (block #1) with 100 transactions and nonce 156789
2025/11/07 15:23:52 Worker 7 mined block at index 2 (block #2) with 100 transactions and nonce 298456
2025/11/07 15:23:54 Worker 1 mined block at index 3 (block #3) with 100 transactions and nonce 445623
2025/11/07 15:23:56 Worker 5 mined block at index 4 (block #4) with 100 transactions and nonce 589012
2025/11/07 15:23:58 Worker 9 mined block at index 5 (block #5) with 100 transactions and nonce 723456
```

**Balansų peržiūra:**
```
╔═══════════════════════════════════════════════════════════════════════════════════════════╗
║                                     USER BALANCES                                         ║
╠════════════════════════════════╦═══════════════╦══════════════════════════════════════════╣
║            NAME                ║    BALANCE    ║              PUBLIC KEY                  ║
╠════════════════════════════════╬═══════════════╬══════════════════════════════════════════╣
║ Alice                          ║       856234  ║ 3a7f9e2c1b8d4f6e0a5c9b7d3e1f8a2c4b6d    ║
║ Bob                            ║       423156  ║ 8f2e4a6c1d9b5e7a3c8f1b4d6a9e2c5f7b    ║
...
╚════════════════════════════════╩═══════════════╩══════════════════════════════════════════╝
```

### Konfigūracija

Sukurkite `.env` failą projekto šakniniame kataloge:
```env
BLOCK_VERSION=1
BLOCK_DIFFICULTY=3
PORT=8080
```

Parametrai:
- `BLOCK_VERSION` – bloko versijos numeris
- `BLOCK_DIFFICULTY` – kasimo sudėtingumas (kiek nulių hash'o pradžioje)
- `PORT` – HTTP API portas (ateityje)
---

## Sistemos architektūra

### Pagrindinės struktūros

#### 1. Blockchain
```go
type Blockchain struct {
    blocks      []Block          // Blokų grandinė
    chainMutex  *sync.RWMutex    // Thread-safety blokų sąrašui
    txGenMutex  *sync.Mutex      // Thread-safety transakcijų generavimui
    utxoTracker *UTXOTracker     // UTXO sekimo sistema
    hasher      Hasher           // Maišos funkcijos interface
}
```

#### 2. Block
```go
type Block struct {
    header Header   // Bloko antraštė
    body   Body     // Bloko turinys (transakcijos)
}

type Header struct {
    version    uint32    // Bloko versija
    timestamp  uint32    // Sukūrimo laikas
    prevHash   Hash32    // Ankstesnio bloko hash
    merkleRoot Hash32    // Merkle Tree šaknies hash
    difficulty uint32    // Kasimo sudėtingumas
    nonce      uint32    // PoW nonce reikšmė
}
```

#### 3. Transaction (UTXO modelis)
```go
type Transaction struct {
    TxID    Hash32      // Transakcijos hash
    Inputs  []TxInput   // Įvestys (panaudojami UTXO)
    Outputs []TxOutput  // Išvestys (nauji UTXO)
}

type TxInput struct {
    Prev Outpoint  // Nuoroda į panaudojamą UTXO
    Sig  []byte    // Parašas
}

type TxOutput struct {
    To    Hash32    // Gavėjo adresas (public key)
    Value uint32    // Suma
}

type UTXO struct {
    Out   Outpoint  // Transakcijos ID ir output index
    To    Hash32    // Savininko adresas
    Value uint32    // Suma
}
```

#### 4. UTXOTracker
```go
type UTXOTracker struct {
    utxoSet   map[Outpoint]UTXO  // UTXO rinkinys
    UTXOMutex *sync.RWMutex      // Thread-safety
}
```

#### 5. MerkleTree
```go
type MerkleTree struct {
    Root *Node    // Medžio šaknis
}

type Node struct {
    Val   Hash32   // Hash reikšmė
    Left  *Node    // Kairysis vaikas
    Right *Node    // Dešinysis vaikas
}
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
    Id        uint32    // Unikalus ID
    Name      string    // Vardas
    CreatedAt uint32    // Sukūrimo laikas
    PublicKey Hash32    // Viešasis raktas (adresas)
}
```

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
    
    // 5. Pasirašoma ir hash'uojama
    // ...
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
    
    // Rekursyviai jungiami lapai iki lieka vienas Root
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
        // Tikrinama, ar procesas neatšauktas
        if err := ctx.Err(); err != nil {
            return 0, Hash32{}, err
        }
        
        // Hash'uojama su nauju nonce
        h.SetNonce(nonce)
        hash, err := h.Hash(hasher)
        if err != nil {
            return 0, Hash32{}, err
        }
        
        // Tikrinama, ar hash atitinka difficulty
        if IsHashValid(hash, h.GetDifficulty()) {
            return nonce, hash, nil
        }
        
        nonce++
        
        // Overflow apsauga
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
    
    // Skaičiuojamas reikiamų nulių skaičius
    bits := diff * 4  // difficulty * 4 bitai
    fullBytes := bits / 8
    remBits := bits % 8
    
    // Tikrinami pilni baitai
    var zero [32]byte
    if fullBytes > 0 {
        if !bytes.Equal(hash[:fullBytes], zero[:fullBytes]) {
            return false
        }
    }
    
    // Tikrinami likusieji bitai
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
                
                // Generuoti transakcijas
                txs, err := bch.GenerateRandomTransactions(users, low, high, txCount)
                if err != nil {
                    continue
                }
                
                // Generuoti ir kasti bloką
                body := NewBody(txs)
                blk, err := bch.GenerateBlock(ctx, body, version, difficulty)
                if err != nil {
                    continue
                }
                
                // Pridėti bloką (thread-safe)
                err = bch.AddBlock(blk)
                if err != nil {
                    continue
                }
                
                // Atnaujinti skaitiklį (atomic)
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

**Kodėl 12 workers?**
- Modernus CPU turi 8-16 core'ų
- 12 suteikia gerą balansą tarp paralelizmo ir overhead'o
- Kasimas yra CPU-intensive, tad daugiau nei core'ų nėra efektyvu

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

### 8. Thread-Safety mechanizmai

**1. Blockchain mutex:**
```go
// Skaitymas (Read Lock)
func (bch *Blockchain) GetLatestBlock() (Block, error) {
    bch.chainMutex.RLock()
    defer bch.chainMutex.RUnlock()
    // ...
}

// Rašymas (Write Lock)
func (bch *Blockchain) AddBlock(b Block) error {
    bch.chainMutex.Lock()
    defer bch.chainMutex.Unlock()
    // ...
}
```

**2. UTXO Tracker mutex:**
```go
func (t *UTXOTracker) GetBalance(address Hash32) uint32 {
    t.UTXOMutex.RLock()
    defer t.UTXOMutex.RUnlock()
    // ...
}

func (t *UTXOTracker) ScanBlock(b Block, hasher Hasher) {
    t.UTXOMutex.Lock()
    defer t.UTXOMutex.Unlock()
    // ...
}
```

**3. Atomic operacijos:**
```go
totalMined := atomic.Int64{}
// ...
mined := totalMined.Add(1)  // Thread-safe increment
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


**Autorius:** Kristupas Arifovas  
**GitHub:** [github.com/Quikmove/blockchain-uzd2](https://github.com/Quikmove/blockchain-uzd2)

