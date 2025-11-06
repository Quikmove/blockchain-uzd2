# Blockchain Projektas – Centralizuota Blokų Grandinė
**v0.2 versija** | Go 1.24+ | UTXO modelis | Merkle Tree

---

## Turinys
1. [Anotacija](#anotacija)
2. [Funkcionalumas](#funkcionalumas)
3. [Instaliacija ir Naudojimas](#instaliacija-ir-naudojimas)
4. [Esminės Kodo Dalys](#esminės-kodo-dalys)
5. [Architektūra](#architektūra)
6. [AI Pagalbos Naudojimas](#ai-pagalbos-naudojimas)

---

## Anotacija

Šis projektas – **centralizuota blokų grandinės (blockchain) implementacija**, sukurta Go kalba, naudojanti **UTXO (Unspent Transaction Output)** modelį ir **Proof-of-Work (PoW)** konsensuso mechanizmą. Sistema generuoja vartotojus, kuria transakcijas, sudaro blokus ir juos "kasa" (mining) su savo maišos funkcija **ArchasHasher**.

**v0.2 versijoje** pridėta:
- ✅ **Merkle Tree** implementacija su tikru Merkle Root Hash
- ✅ **Transakcijų validacija** (balansų tikrinimas, UTXO egzistavimo patikra, double-spend prevencija)
- ✅ **Thread-safe** operacijos su `sync.RWMutex`
- 🔄 **Kandidatinių blokų kasimas** (planuojama, bet dar neimplementuota)

### Pagrindinės savybės:
- **~50 vartotojų** generavimas su atsitiktiniais balansais (100–1,000,000)
- **~500 transakcijų** generavimas (po 100 transakcijų bloke)
- **Proof-of-Work** – hash'ai turi prasidėti trimis nuliais (`000...`)
- **CLI sąsaja** lokaliam blockchain'o valdymui
- **UTXO tracking** – realiu laiku seka nepanaudotus išvesties (output) balansus

---

## Funkcionalumas

### v0.1 (2025-10-29)
- ✅ Centralizuota blokų grandinė su custom maišos funkcija (**ArchasHasher**)
- ✅ UTXO modelio implementacija
- ✅ Genesis bloko kūrimas su pradiniais fondais
- ✅ Transakcijų sudarymas ir įtraukimas į blokus
- ✅ Proof-of-Work kasimas (`difficulty = 3`)
- ✅ Paprastas vizualus išvedimas į konsolę

### v0.2 (2025-11-05)
- ✅ **Merkle Tree** struktūra transakcijoms (`internal/merkletree/`)
- ✅ **Merkle Root Hash** apskaičiavimas blokų antraštėse
- ✅ **Transakcijų validacija**:
  - Balansų tikrinimas (siuntėjas negali siųsti daugiau nei turi)
  - UTXO egzistavimo patikra
  - Double-spend prevencija bloko viduje
  - Input/output sumų palyginimas
- ✅ **Thread-safe UTXO tracker** su `sync.RWMutex`
- ⏳ **Kandidatinių blokų kasimas** (planuojama)

---

## Instaliacija ir Naudojimas

### Prieš pradedant
Reikalingi įrankiai:
- **Go 1.24+** ([parsisiųsti](https://golang.org/dl/))
- **Make** (macOS/Linux – įdiegta numatyta, Windows – per [Chocolatey](https://chocolatey.org/) arba WSL)

1. Klonuoti projektą
```bash
git clone https://github.com/Quikmove/blockchain-uzd2.git
cd blockchain-uzd2
```
2. Įdiegti priklausomybes
```bash
go mod download
```

3. Sukompiliuoti CLI
```bash
make build
# Arba rankiniu būdu:
# go build -o bin/cli ./cmd/cli
```

4. Paleisti blockchain'ą
```bash
./bin/cli local
```

#### Rezultatas konsolėje:
```
2025/11/06 12:34:56 Version: 1
2025/11/06 12:34:56 Difficulty: 3
2025/11/06 12:34:56 Generating genesis block...
2025/11/06 12:34:58 Found a POW hash successfully with nonce: 123456
2025/11/06 12:35:01 Added new block with nonce: 234567
2025/11/06 12:35:03 Added new block with nonce: 345678
...
```

### 5️⃣ Konfigūracija (.env failas)
Sukurkite `.env` failą projekto šakniniame kataloge:
```env
BLOCK_VERSION=1
BLOCK_DIFFICULTY=3
PORT=8080
```
Nesukūrus `.env`, bus naudojamos numatytosios reikšmės.
### CLI Komandos
| Komanda | Aprašymas |
|---------|-----------|
| `./bin/cli local` | Paleidžia lokalų blockchain'ą su 50 vartotojų ir 500 transakcijų |
| `make clean` | Ištrina sukompiliuotus failus |

---

## Esminės Kodo Dalys

### 🏗️ `Blockchain` struktūra (`internal/blockchain/blockchain.go`)
Pagrindinė blockchain'o klasė, valdanti blokų grandinę.

```go
type Blockchain struct {
    blocks      []Block
    ChainMutex  *sync.RWMutex
    utxoTracker *UTXOTracker
    hasher      Hasher
}
```

**Svarbūs metodai:**
- `AddBlock(b Block)` – prideda bloką po validacijos
- `IsBlockValid(b Block)` – tikrina bloko hash'o galiojimą (PoW)
- `ValidateBlockTransactions(b Block)` – validuoja visas transakcijas
- `GenerateRandomTransactions(...)` – generuoja atsitiktines transakcijas

---

### 🔐 `ArchasHasher` (`internal/blockchain/archas_hasher.go`)
Patobulinta hešavimo funkcija, perkelta iš C++ kodo, modifikuota PoW kasimui. Naudoja:
- Baitų rotaciją (`bits.RotateLeft8`)
- XOR, AND, OR operacijas
- Dinaminį "collapse" mechanizmą 32-baitų hash'ui generuoti

```go
type ArchasHasher struct {
    pc *PeriodicCounter
}

func (h *ArchasHasher) Hash(data []byte) ([]byte, error) {
    // įvairios rotacijos ir bitų operacijos
}
```



---

### `MerkleTree` (`internal/merkletree/merkletree.go`)
Dvejetainė hash medžio struktūra, naudojama transakcijų autentiškumui tikrinti.

```go
type MerkleTree struct {
    Root *Node
}

type Node struct {
    Val   Hash32
    Left  *Node
    Right *Node
}
```

**Kaip veikia:**
1. Kiekviena transakcija hash'uojama → lapai
2. Lapai poruojami ir hash'uojami → viršutiniai mazgai
3. Kartojama, kol lieka vienas `Root` hash

---

### `UTXOTracker` (`internal/blockchain/utxo_tracker.go`)
Seka **nepanaudotus transakcijų išvesties balansus** (UTXO modelis).

```go
type UTXOTracker struct {
    utxoSet   map[Outpoint]UTXO
    UTXOMutex *sync.RWMutex
}
```

**Svarbūs metodai:**
- `ScanBlock(b Block)` – atnaujina UTXO set'ą (pašalina panaudotus inputs, prideda naujus outputs)
- `GetBalance(address Hash32)` – grąžina adreso balansą
- `GetUTXOsForAddress(...)` – grąžina visus adreso UTXO

**Thread-safety:** Naudoja `RWMutex`, kad kelios goroutine'os galėtų skaityti be konfliktų.

---

### `Block` ir `Transaction` struktūros
```go
type Block struct {
    Header Header
    Body   Body
}

type Header struct {
    Version    uint32
    Timestamp  uint32
    PrevHash   Hash32
    MerkleRoot Hash32
    Difficulty uint32
    Nonce      uint32
}

type Transaction struct {
    TxID    Hash32
    Inputs  []TxInput
    Outputs []TxOutput
}
```

---

### ⚙️ Validacijos logika (`ValidateBlockTransactions`)
**Kiekvienos transakcijos tikrinimas:**
1. **Genesis bloke** – tik coinbase transakcijos (be inputs)
2. **Inputs egzistavimas** – tikrina, ar UTXO egzistuoja `utxoTracker`'yje
3. **Double-spend** – užtikrina, kad tas pats UTXO nenaudojamas dukart bloke
4. **Balansų tikrinimas** – `inputSum >= outputSum`
5. **Overflow apsauga** – tikrina aritmetinius perpildymus

```go
if inputSum < outputSum {
    return fmt.Errorf("outputs exceed inputs")
}
```

---

### ⛏️ Proof-of-Work (`FindValidNonce`)
Ieško `nonce` reikšmės, kad bloko hash'as prasidėtų `difficulty` skaičiumi nulių.

```go
func (h Header) FindValidNonce(ctx context.Context, hasher Hasher) (uint32, Hash32, error) {
    for nonce := uint32(0); ; nonce++ {
        h.Nonce = nonce
        hash, _ := h.Hash(hasher)
        if IsHashValid(hash, h.Difficulty) {
            return nonce, hash, nil
        }
    }
}
```

**Kaip veikia `IsHashValid`?**
```go
func IsHashValid(hash Hash32, diff uint32) bool {
    for i := uint32(0); i < diff; i++ {
        if hash[i] != 0 { return false }
    }
    return true
}
```
Jei `difficulty = 3` → hash'as turi prasidėti `000...`.

---

## Architektūra

```
blockchain-uzd2/
├── cmd/
│   ├── cli/cli.go           # CLI entry point
│   └── api/main.go          # HTTP API (planuojama)
├── internal/
│   ├── blockchain/
│   │   ├── blockchain.go     # Blockchain struktūra
│   │   ├── transactions.go   # Transaction logika
│   │   ├── utxo_tracker.go   # UTXO sekimas
│   │   ├── archas_hasher.go  # Custom hash funkcija
│   │   ├── genesis_block.go  # Genesis bloko kūrimas
│   │   └── user.go           # User generavimas
│   ├── merkletree/
│   │   └── merkletree.go     # Merkle Tree implementacija
│   ├── config/
│   │   └── config.go         # Konfigūracija (.env)
│   └── filetolist/
│       └── file_to_list.go   # Failų skaitymas
├── assets/
│   └── name_list.txt         # Vardų sąrašas vartotojams
├── Makefile                  # Build komandos
└── go.mod                    # Go modulių priklausomybės
```

### OOP Principai
- **Enkapsuliacija** – `Blockchain`, `UTXOTracker` vidiniai laukai privatūs
- **Konstruktoriai** – `NewBlockchain()`, `NewUTXOTracker()`, `NewArchasHasher()`
- **Mutex'ai** – `sync.RWMutex` užtikrina thread-safety
- **Interface'ai** – `Hasher` interface leidžia keisti maišos funkciją

---

## AI Pagalbos Naudojimas

Projekte buvo naudojama **AI pagalba** šiems tikslams:

### 🤖 Kodavimas
- **Merkle Tree implementacija** – AI padėjo sukurti rekursyvią medžio struktūrą ir `doubleHashPair` funkciją
- **Transakcijų sudarymo logika** – generuojant atsitiktines transakcijas su UTXO atranka
- **UTXO validacijos optimizavimas** – double-spend patikros ir balansų skaičiavimo logika

### 🐛 Debugging
- **Nonce overflow problemos sprendimas** – AI pasiūlė konteksto (`context.Context`) naudojimą kasimo sustabdymui
- **Mutex deadlock'ų prevencija** – patarė naudoti `RLock()`/`RUnlock()` skaitymo operacijoms
- **Genesis bloko validacijos klaidos** – AI padėjo identifikuoti coinbase transakcijų tikrinimo logiką

### 🧪 Testai
- **Unit testų struktūra** – AI sugeneravo testų šablonus `merkletree_test.go` ir `utxo_tracker_test.go`
- **Edge case'ų identifikavimas** – pvz., ką daryti, kai bloke 0 transakcijų, arba kai UTXO set'as tuščias

### 📝 Dokumentacija
- **README struktūros planavimas** – AI pasiūlė struktūrą su anotacija, naudojimo instrukcijomis ir architektūros aprašu
- **Kodo komentarų gerinimas** – padėjo parašyti aiškesnius docstring'us funkcijoms

### ⚠️ Svarbūs Niuansai
- **~80% kodo parašyta savarankiškai** – pagrindinė logika (blockchain, transakcijos, PoW) sukurta be AI
- **AI naudota kaip "rubber duck"** – daugiausia debugging'ui ir greičiau rasti Go bibliotekų dokumentaciją
- **Kandidatinių blokų kasimas (v0.2)** – dar neimplementuotas, todėl AI pagalba planuojama būsimose versijose

---

## Ateities Planai (v0.3+)

- [ ] **Kandidatinių blokų kasimas** (5 blokai, ribota kasimo trukmė)
- [ ] **HTTP API** su REST endpoint'ais
- [ ] **Blokų eksportas** į JSON
- [ ] **Performance metrikų** rinkimas (avg. mining time, tx/s)
- [ ] **Signature validation** su tikru kripto (ECDSA)

---

## Licencija
Šis projektas sukurtas akademiniams tikslams (VU MIF BGT kursas, 2025).

**Autorius:** Kristupas Arifovas  
**GitHub:** [github.com/Quikmove/blockchain-uzd2](https://github.com/Quikmove/blockchain-uzd2)

