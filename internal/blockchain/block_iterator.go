package blockchain

type BlockIterator struct {
	data []Block
	i    int
}

func (bc *Blockchain) Iterator() *BlockIterator {
	bc.ChainMutex.RLock()
	snap := append([]Block(nil), bc.blocks...)
	bc.ChainMutex.RUnlock()
	return &BlockIterator{data: snap, i: 0}
}
func (it *BlockIterator) Next() bool {
	it.i++
	return it.i < len(it.data)
}
func (it *BlockIterator) Value() Block {
	return it.data[it.i]
}
