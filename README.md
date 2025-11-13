# Blokų grandinių technologijos – 2 užduotis
**v0.2 versija** | Go 1.24+ | UTXO modelis | Merkle Tree | CLI | Lygiagretus kasimas | Decentralized Mining | Kriptografiniai parašai

## Pagrindinės savybės
- **100 (numatytų) vartotojų** su atsitiktiniais balansais (100–1,000,000), konfigūruojamas per USER_COUNT
- **UTXO modelis** su išsamiu nepanaudotų transakcijų išvesties sekimu
- **Merkle Tree** transakcijų hash'avimui
- **Proof-of-Work** su difficulty = 3 (hash'as prasideda `000...`)
- **Lygiagretus kasimas** su kompiuterio core'ų kiekio worker'ių (runtime.NumCPU())
- **Decentralized mining simulation** su keliais kandidatais ir laiko limitais
- **Kriptografiniai parašai** - tikri secp256k1 parašai su verifikacija
- **HASH160 adresai** - PublicAddress generavimas naudojant SHA256 + RIPEMD160
- **Timestamp validacija** - blokų laiko žymų tikrinimas
- **Transaction ID validacija** - transakcijų hash'ų tikrinimas
- **Coinbase validacija** - maksimalių atlygių limitai
- **User registry sistema** - vartotojų registravimas ir valdymas
- **Thread-safe** operacijos su `sync.RWMutex`
- **CLI sąsaja** su interaktyvia valdymo konsole

## Įgyvendinti reikalavimai
- UTXO modelio realizavimas
- Lygiagretus kasimo procesas v0.2 versijoje
- Merkle Tree su tikru Merkle Root Hash
- Transakcijų validacija (balansų tikrinimas, double-spend prevencija)
- OOP principų taikymas (enkapsuliacija, konstruktoriai, interfaces)

---

## Įgyvendinta funkcionalumas

### Versija v0.1 reikalavimai

#### 1. Bloko struktūra
- ✅ **Antraštė (Header)**:
  - Ankstesnio bloko maišos reikšmė (Prev Block Hash)
  - Laiko žyma (Timestamp)
  - Duomenų struktūros versija (Version)
  - Visų transakcijų maišos reikšmė (Merkle Root Hash)
  - Atsitiktinis skaičius – Nonce, naudojamas Proof-of-Work procese
  - Sudėtingumo lygis (Difficulty Target)

- ✅ **Turinys (Body)**:
  - Transakcijų sąrašas

#### 2. Vartotojų generavimas (~1000 vartotojų)
- ✅ Vardas
- ✅ Public_key (viešasis raktas)
- ✅ Balansas (atsitiktinis nuo 100 iki 1 000 000)

#### 3. Transakcijų generavimas (~10 000 įrašų)
- ✅ Transaction_id (kitų laukų hash)
- ✅ Sender (siuntėjo raktas)
- ✅ Receiver (gavėjo raktas)
- ✅ Amount (siunčiama suma)

#### 4. Naujo bloko formavimas
- ✅ Atsitiktinai pasirenkama 100 transakcijų iš sąrašo
- ✅ Transakcijos paruošiamos įtraukimui į naują bloką

#### 5. Bloko kasimas (Proof-of-Work)
- ✅ Hash'uojant 6 pagrindinius bloko antraštės (block header) elementus randamas bloko hash
- ✅ Hash'as prasideda bent trimis nuliais (000...)
- ✅ Naudojama savo maišos funkcija (modifikuota, kad tiktų šiam tikslui)
- ✅ Nonce iteracija iki galiojamo hash'o

#### 6. Bloko patvirtinimas ir įtraukimas
- ✅ Į bloką įtrauktos transakcijos pašalinamos iš sąrašo
- ✅ Vartotojų balansai atnaujinami
- ✅ Naujas blokas pridedamas prie grandinės

#### 7. Procesų kartojimas
- ✅ Procesas kartojamas (3–5 žingsniai), kol neliks neįtrauktų transakcijų

#### 8. Transakcijų ir blokų kūrimo proceso matomumas
- ✅ Transakcijų ir blokų kūrimo procesas matomas (išvedamas į konsolę)
- ✅ Išvedimo kokybė ir vizualumas

### Versija v0.2 reikalavimai

#### 1. Merkle Tree realizavimas
- ✅ Įgyvendintas Merkle Tree ir tikras Merkle Root Hash
- ✅ Vietoje paprasto visų transakcijų ID maišo naudojamas tikras Merkle Tree

#### 2. Transakcijų verifikacija
- ✅ **Balanso tikrinimas**: siuntėjas negali siųsti daugiau, nei turi
- ✅ **Transakcijos ID tikrinimas**: maišos reikšmės teisingumas

#### 3. Patobulintas kasimo procesas
- ✅ Sugeneruojami 5 kandidatiniai blokai (~100 transakcijų kiekviename)
- ✅ Bandymai „kasti" ribotą laiką (pvz., 5s) arba iki riboto bandymų skaičiaus
- ✅ Jei nė vienas neiškastas – padidinamas laikas/bandymai ir procesas kartojamas
- ✅ Decentralizuoto kasimo simuliacija: keli kandidatai kasa paraleliai, pirmas sėkmingai iškastas blokas pridedamas


### Papildomi įgyvendinti funkcionalumai

#### 1. UTXO modelio realizavimas
- ✅ UTXO (Unspent Transaction Output) modelis vietoje sąskaitos modelio
- ✅ Išsamus nepanaudotų transakcijų išvesties sekimas
- ✅ UTXO paieška pagal adresą
- ✅ Balanso skaičiavimas iš UTXO sąrašo

#### 2. Lygiagretus kasimo procesas v0.2 versijoje
- ✅ Lygiagretus kasimas su keliais worker'iais
- ✅ Keli kandidatai kasa paraleliai

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
1. Sugeneruos 100 vartotojų su atsitiktiniais balansais (100–1,000,000)
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
║   simulatedecentralizedmining - Simulate decentralized mining         ║
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
║   getblocktransactions- Get block transactions by index               ║
║   getallheaders       - Get all block headers                         ║
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
2025/11/07 00:49:40 User count: 100
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
USER_COUNT=100
```

Parametrai:
- `BLOCK_VERSION` – bloko versijos numeris
- `BLOCK_DIFFICULTY` – kasimo sudėtingumas (kiek nulių hash'o pradžioje)
- `PORT` – HTTP API portas (ateityje)
- `USER_COUNT` – sugeneruojamų vartotojų skaičius (numatyta 100)

**Pastaba:** Worker skaičius kasimo metu yra dinamiškas ir nustatomas pagal kompiuterio CPU core'ų skaičių (runtime.NumCPU())
---

## Sistemos architektūra

### Pagrindinės struktūros

- **Blockchain** – blokų grandinė su UTXO tracker, user registry, thread-safe operacijomis
- **Block** – blokas su Header (version, timestamp, prevHash, merkleRoot, difficulty, nonce) ir Body (transactions)
- **Transaction** – UTXO modelis su Inputs (Outpoint + signature) ir Outputs (PublicAddress + Value)
- **User** – vartotojas su PublicKey (33B), PublicAddress (20B HASH160), PrivateKey (32B)
- **UTXOTracker** – nepanaudotų transakcijų išvesties sekimo sistema
- **MerkleTree** – transakcijų hash'avimo medis
- **TransactionSigner** – secp256k1 parašų generavimo ir verifikacijos interface
- **KeyGenerator** – raktų porų generavimo interface

---

## Realizacijos specifika

### 1. Transakcijų generavimas

```
FUNKCIJA GenerateRandomTransactions(users, low, high, n):
    sender = atsitiktinis_vartotojas(users)
    recipient = atsitiktinis_vartotojas(users)
    
    utxos = gauti_UTXO(sender.address)
    amount = atsitiktinis_skaicius(low, high)
    
    inputs = []
    totalInput = 0
    KIEKVIENAM utxo IN utxos:
        JEI totalInput >= amount:
            BREAK
        inputs.prideti(utxo.outpoint)
        totalInput += utxo.value
    
    outputs = [TxOutput(amount, recipient.address)]
    JEI totalInput > amount:
        change = totalInput - amount
        outputs.prideti(TxOutput(change, sender.address))
    
    tx = Transaction(inputs, outputs)
    tx.txID = hash(tx.serialize_be_signatures())
    
    KIEKVIENAM input IN tx.inputs:
        hashToSign = SignatureHash(tx, utxo.value, utxo.address)
        input.sig = sign(hashToSign, sender.privateKey)
    
    RETURN tx
```

### 2. Merkle Tree konstrukcija

```
FUNKCIJA BuildMerkleTree(transaction_hashes):
    JEI transaction_hashes.tuscias():
        RETURN nil
    
    nodes = []
    KIEKVIENAM hash IN transaction_hashes:
        nodes.prideti(Node(hash))
    
    RETURN buildTree(nodes)

FUNKCIJA buildTree(nodes):
    JEI nodes.length == 1:
        RETURN nodes[0]
    
    parents = []
    I = 0
    KOL I < nodes.length:
        JEI I+1 < nodes.length:
            parent = Node(
                hash = doubleHash(nodes[I].val, nodes[I+1].val),
                left = nodes[I],
                right = nodes[I+1]
            )
        KITAIP:
            // Nepori element duplicuojamas
            parent = Node(
                hash = doubleHash(nodes[I].val, nodes[I].val),
                left = nodes[I],
                right = nodes[I]
            )
        parents.prideti(parent)
        I += 2
    
    RETURN buildTree(parents)
```

### 3. Proof-of-Work kasimas

```
FUNKCIJA FindValidNonce(header, difficulty):
    nonce = 0
    
    KOL TRUE:
        header.nonce = nonce
        hash = hash(header.serialize())
        
        JEI IsHashValid(hash, difficulty):
            RETURN nonce, hash
        
        nonce++
        
        JEI nonce == MAX_UINT32:
            RETURN ERROR("nonce overflow")

FUNKCIJA IsHashValid(hash, difficulty):
    JEI difficulty == 0:
        RETURN TRUE
    
    bits = difficulty * 4
    fullBytes = bits / 8
    remBits = bits % 8
    
    // Tikrinti pilnus baitus
    KIEKVIENAM I IN [0..fullBytes):
        JEI hash[I] != 0:
            RETURN FALSE
    
    // Tikrinti likusius bitus
    JEI remBits > 0:
        mask = 0xFF << (8 - remBits)
        JEI (hash[fullBytes] & mask) != 0:
            RETURN FALSE
    
    RETURN TRUE
```

### 4. Lygiagretus kasimas

```
FUNKCIJA MineBlocks(blockCount, txCount, users):
    numWorkers = runtime.NumCPU()
    
    KIEKVIENAM round IN [1..blockCount]:
        blockChan = naujas_kanalas()
        wg = naujas_WaitGroup()
        
        // Paleisti worker'ius
        KIEKVIENAM workerID IN [0..numWorkers):
            wg.prideti(1)
            GOROUTINE:
                KOL TRUE:
                    txs = GenerateRandomTransactions(users, low, high, txCount)
                    body = Body(txs)
                    block = GenerateBlock(body, version, difficulty)
                    
                    JEI AddBlock(block) == SUCCESS:
                        blockChan <- block
                        RETURN
        
        // Laukti pirmo sėkmingai iškasto bloko
        blk = <-blockChan
        log("Round {round}: mined block with {blk.txCount} transactions")
        
        wg.laukti()
```

### 5. Decentralized Mining

```
FUNKCIJA MineBlocksDecentralized(config, users):
    KIEKVIENAM round IN [1..config.blockCount]:
        timeLimit = config.initialTimeLimit
        miningSuccess = FALSE
        
        KOL NOT miningSuccess:
            // Generuoti kandidatų blokus
            candidates = []
            KIEKVIENAM I IN [1..config.candidateCount]:
                txs = GenerateRandomTransactions(users, config.low, config.high, config.txCount)
                candidates.prideti(Body(txs))
            
            blockChan = naujas_kanalas()
            timeoutCtx = context.WithTimeout(timeLimit)
            
            // Kiekvienas kandidatas kasa paraleliai
            KIEKVIENAM candidate IN candidates:
                GOROUTINE:
                    KOL timeoutCtx.NOT_DONE():
                        block = bandyti_iskasti(candidate)
                        JEI block != NULL:
                            blockChan <- block
                            RETURN
            
            SELECT:
                CASE blk = <-blockChan:
                    AddBlock(blk)
                    miningSuccess = TRUE
                CASE <-timeoutCtx.DONE():
                    timeLimit = timeLimit * config.timeMultiplier
                    log("Timeout, retrying with timeLimit={timeLimit}")
```

### 6. Validacija

```
FUNKCIJA ValidateBlockTransactions(block, users):
    spentInBlock = {}
    
    KIEKVIENAM tx IN block.transactions:
        // 1. Transaction ID validacija
        expectedTxID = hash(tx.serialize_be_signatures())
        JEI tx.txID != expectedTxID:
            RETURN ERROR("TxID mismatch")
        
        // 2. Coinbase validacija
        JEI tx.inputs.tuscias():
            JEI tx.index != 0:
                RETURN ERROR("Coinbase only as first tx")
            KIEKVIENAM output IN tx.outputs:
                JEI output.value == 0 ARBA output.value > MAX_COINBASE:
                    RETURN ERROR("Invalid coinbase output")
            CONTINUE
        
        // 3. Input validacija
        inputSum = 0
        KIEKVIENAM input IN tx.inputs:
            JEI spentInBlock[input.prev]:
                RETURN ERROR("Double-spend")
            
            utxo = GetUTXO(input.prev)
            JEI utxo == NULL:
                RETURN ERROR("UTXO not found")
            
            inputSum += utxo.value
            
            // Parašo verifikacija
            hashToVerify = SignatureHash(tx, utxo.value, utxo.address)
            JEI NOT verify_signature(hashToVerify, input.sig, utxo.publicKey):
                RETURN ERROR("Invalid signature")
            
            spentInBlock[input.prev] = TRUE
        
        // 4. Output validacija
        outputSum = suma(tx.outputs)
        JEI inputSum < outputSum:
            RETURN ERROR("Outputs exceed inputs")
    
    RETURN SUCCESS
```

---

## Kriptografija

### HASH160 adresų generavimas

```
FUNKCIJA GenerateAddress(publicKey):
    // 1. SHA256 hash'as iš public key
    sha256Hash = SHA256(publicKey)
    
    // 2. RIPEMD160 hash'as iš SHA256 rezultato
    address = RIPEMD160(sha256Hash)
    
    RETURN address  // 20 baitų
```

### Parašų sistema

- **secp256k1 parašai** – tikri kriptografiniai parašai su verifikacija kiekvienam transakcijos input'ui
- **KeyGenerator** – generuoja secp256k1 raktų poras (PrivateKey 32B, PublicKey 33B) iš mnemonic

---

## Decentralized Mining

Simuliacija su keliais kandidatų blokais, kurie kasa paraleliai su laiko limitu. Pirmas sėkmingai iškastas blokas pridedamas į grandinę. Jei per laiko limitą niekas nebuvo iškastas, laiko limitas padidinamas (TimeMultiplier: 2.0) ir procesas kartojamas.


---

## Validacijos

### Bloko validacija

```
FUNKCIJA ValidateBlock(block):
    // 1. Merkle root validacija
    computedRoot = MerkleRootHash(block.body.transactions)
    JEI block.header.merkleRoot != computedRoot:
        RETURN ERROR("Merkle root mismatch")
    
    // 2. Timestamp validacija
    currentTime = dabartinis_laikas()
    JEI block.header.timestamp > currentTime + 7200:
        RETURN ERROR("Timestamp too far in future")
    JEI block.header.timestamp < currentTime - 7200:
        RETURN ERROR("Timestamp too far in past")
    
    // 3. Hash validacija (non-genesis)
    JEI NOT isGenesis:
        hash = hash(block.header)
        JEI NOT IsHashValid(hash, block.header.difficulty):
            RETURN ERROR("Hash doesn't meet difficulty")
    
    RETURN SUCCESS
```

### Transakcijų validacija

- **Timestamp** – blokas negali būti daugiau nei ±7200s nuo dabartinio laiko
- **Transaction ID** – tikrinamas, kad TxID atitinka transakcijos duomenis (be parašų)
- **Coinbase** – tik pirmoji transakcija, maksimalus atlygis 1,000,000
- **Signature verification** – kiekvienas input turi galiojantį secp256k1 parašą


---

## Techninės detalės

- **Domain package** (`internal/domain/`) – pagrindiniai tipai (Block, Transaction, User, UTXO) ir struktūruoti klaidų tipai
- **Crypto package** (`internal/crypto/`) – Hasher, TransactionSigner, KeyGenerator interfaces su implementacijomis
- **Config sistema** – `.env` failas su BLOCK_VERSION, BLOCK_DIFFICULTY, USER_COUNT, PORT parametrais

---