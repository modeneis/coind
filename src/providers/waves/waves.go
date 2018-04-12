package waves

import (
	"sync"

	"github.com/modeneis/waves-go-client/client"
	"github.com/modeneis/waves-go-client/model"

	"github.com/btcsuite/btcd/btcjson"

	"github.com/modeneis/coind/src/server/model_server"
)

//get lastest block

//inject fake block into it

//return back

// BlockStore holds fake block data
type BlockStoreWaves struct {
	sync.RWMutex
	BestBlockHeight int32
	BlockHashes     map[int64]string
	BlockTX         map[string]string
	HashBlocks      map[string]*model.Blocks
	NextHash        string
}

// WavesFake is the main fields for waves fake coin
type Provider struct {
	DefaultBlockStore *BlockStoreWaves
	InitialBlock      *model.Blocks
	MainNET           string
}

// New creates a new fake SKY, and sets up important connection details.
func New() *Provider {
	coinFake := &Provider{
		MainNET: "",
	}
	coinFake.Start()
	return coinFake
}

// Name is the name used to retrieve this provider later.
func (p *Provider) Name() string {
	return "waves"
}

// Name is SkycoinFake name used to retrieve this report type later.
func (p *Provider) GetType() string {
	return "WAVES"
}

func (p *Provider) Start() {
	p.DefaultBlockStore = &BlockStoreWaves{
		BlockHashes: make(map[int64]string),
		HashBlocks:  make(map[string]*model.Blocks),
		BlockTX:     make(map[string]string),
	}
}

func (p *Provider) CreateFakeBlock(deposit model_server.Deposit) (retBlocks interface{}, err error) {

	var blocks *model.Blocks
	p.DefaultBlockStore.Lock()
	defer func() {
		//DefaultBlockStore.BestBlockHeight++
		hash := blocks.Signature
		tx := blocks.Signature

		p.DefaultBlockStore.BestBlockHeight++
		p.DefaultBlockStore.BlockHashes[int64(p.DefaultBlockStore.BestBlockHeight)] = hash
		p.DefaultBlockStore.HashBlocks[hash] = blocks
		p.DefaultBlockStore.BlockTX[hash] = tx

		// Update NextHash of previous block
		bestHeight := int64(p.DefaultBlockStore.BestBlockHeight)
		prevBlockHash := p.DefaultBlockStore.BlockHashes[bestHeight]

		prevBlock := p.DefaultBlockStore.HashBlocks[prevBlockHash]

		if prevBlock != nil && len(prevBlock.Transactions) > 0 {
			prevBlock.Reference = hash
			p.DefaultBlockStore.HashBlocks[prevBlockHash] = prevBlock
		}

		// Update Hash of new block
		p.DefaultBlockStore.HashBlocks[blocks.Signature] = blocks

		p.DefaultBlockStore.Unlock()
	}()

	blocks, _, err = client.NewBlocksService(p.MainNET).GetBlocksLast()
	if err != nil {
		return nil, err
	}
	//add fake block to existing block

	var fakeTX model.Transactions
	for _, tx := range blocks.Transactions {

		//TODO: properly test and select the best block
		if tx.Recipient != "" {
			fakeTX.ID = tx.ID
			fakeTX.Signature = tx.Signature
			fakeTX.Height = tx.Height

			fakeTX.Recipient = deposit.Address
			fakeTX.Amount = deposit.Value
			break
		}
	}

	//add fake tx to transactions
	blocks.Transactions = append(blocks.Transactions, fakeTX)

	//return blocks with fake tx added

	return blocks, err
}

func (p *Provider) GetBlock(hash string) (block interface{}, err error) {

	if b, ok := p.DefaultBlockStore.HashBlocks[hash]; ok {
		block = b
		return
	} else {
		return nil, &btcjson.RPCError{
			Code:    btcjson.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}

}

func (p *Provider) GetBestBlock(seq int64) (block interface{}, err error) {
	if hash, ok := p.DefaultBlockStore.BlockHashes[int64(p.DefaultBlockStore.BestBlockHeight)]; ok {
		block = &btcjson.GetBestBlockResult{
			Hash:   hash,
			Height: p.DefaultBlockStore.BestBlockHeight,
		}
		return block, nil
	}
	return block, err
}

func (p *Provider) GetGetBlockHash(tx string) (block interface{}, err error) {
	if hash, ok := p.DefaultBlockStore.BlockTX[tx]; ok {
		block = p.DefaultBlockStore.HashBlocks[hash]
	} else {
		err = &btcjson.RPCError{
			Code:    btcjson.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}
	return block, err
}

func (p *Provider) GetBlockCount() (count int32) {
	return count
}
