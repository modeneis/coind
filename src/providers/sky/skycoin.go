package sky

import (
	"crypto/rand"
	"strconv"
	"sync"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/skycoin/skycoin/src/api/webrpc"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/coin"
	"github.com/skycoin/skycoin/src/gui"
	"github.com/skycoin/skycoin/src/visor"

	"github.com/modeneis/coind/src/server/model_server"
)

// BlockStore holds fake block data
type BlockStoreSky struct {
	sync.RWMutex
	BestBlockHeight int32
	BlockHashes     map[int64]string
	BlockTX         map[string]string
	HashBlocks      map[string]*visor.ReadableBlocks
	NextHash        string
}

// New creates a new fake SKY, and sets up important connection details.
func New() *Provider {
	rest := &gui.Client{
		Addr: "https://explorer.skycoin.net" + ":" + "443" + "/api/",
	}
	coinFake := &Provider{
		SkyRESTClinet: rest,
	}
	coinFake.Start()
	return coinFake
}

// FakeCoin is the main fields for sky fake coin
type Provider struct {
	DefaultBlockStore *BlockStoreSky
	InitialBlock      visor.ReadableBlocks
	SkyRPCClient      *webrpc.Client
	SkyRESTClinet     *gui.Client
}

func (p *Provider) Start() {
	p.DefaultBlockStore = &BlockStoreSky{
		BlockHashes: make(map[int64]string),
		HashBlocks:  make(map[string]*visor.ReadableBlocks),
		BlockTX:     make(map[string]string),
	}
}

// Name is the name used to retrieve this provider later.
func (p *Provider) Name() string {
	return "skycoin"
}

// Name is SkycoinFake name used to retrieve this report type later.
func (p *Provider) GetType() string {
	return "SKY"
}

func (p *Provider) CreateFakeBlock(deposit model_server.Deposit) (retBlocks interface{}, err error) {
	p.DefaultBlockStore.Lock()
	var blocks *visor.ReadableBlocks
	defer func() {
		//DefaultBlockStore.BestBlockHeight++
		hash := blocks.Blocks[0].Head.BlockHash
		tx := blocks.Blocks[0].Head.BodyHash

		p.DefaultBlockStore.BestBlockHeight = int32(blocks.Blocks[0].Head.BkSeq)
		p.DefaultBlockStore.BlockHashes[int64(p.DefaultBlockStore.BestBlockHeight)] = hash

		p.DefaultBlockStore.BlockTX[hash] = tx
		p.DefaultBlockStore.HashBlocks[hash] = blocks

		// Update NextHash of previous block
		bestHeight := int64(p.DefaultBlockStore.BestBlockHeight)
		prevBlockHash := p.DefaultBlockStore.BlockHashes[bestHeight]

		prevBlock := p.DefaultBlockStore.HashBlocks[prevBlockHash]

		if prevBlock != nil && len(prevBlock.Blocks) > 0 {
			prevBlock.Blocks[0].Head.PreviousBlockHash = hash
			p.DefaultBlockStore.HashBlocks[prevBlockHash] = prevBlock
		}

		// Update Hash of new block
		p.DefaultBlockStore.HashBlocks[blocks.Blocks[0].Head.BlockHash] = blocks

		p.DefaultBlockStore.Unlock()
	}()

	//get metadata
	var meta *visor.BlockchainMetadata
	meta, err = p.SkyRESTClinet.BlockchainMetadata()
	if err != nil {
		return nil, err
	}

	//get latest block
	blocks, err = p.SkyRESTClinet.Blocks(int(meta.Head.BkSeq), int(meta.Head.BkSeq))
	if err != nil {
		return nil, err
	}

	//add a fake block to it
	txFound := visor.ReadableTransactionOutput{}
	for _, block := range blocks.Blocks {

		var actualTX visor.ReadableTransaction

		for _, tx := range block.Body.Transactions {

			txFound = tx.Out[0]
			txFound.Address = deposit.Address
			txFound.Coins = strconv.Itoa(int(deposit.Value))
			txFound.Hours = deposit.Hours

			tx.Out = append(tx.Out, txFound)
			actualTX = tx
			break

		}

		block.Body.Transactions[0] = actualTX

		if txFound.Address != "" {
			break
		}
	}

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
	if seq == 0 {
		seq = int64(p.DefaultBlockStore.BestBlockHeight)
	}
	if hash, ok := p.DefaultBlockStore.BlockHashes[seq]; ok {
		block = &btcjson.GetBestBlockResult{
			Hash:   hash,
			Height: p.DefaultBlockStore.BestBlockHeight,
		}
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
	return int32(p.DefaultBlockStore.BestBlockHeight)
}

//
//func (s *SkycoinFake) CreateFakeBlockRPC(deposit Deposit) (blocks visor.ReadableBlocks) {
//
//	//get blocks
//	param := []uint64{1}
//	blk := visor.ReadableBlocks{}
//	if err := s.SkyRPCClient.Do(&blk, "get_lastblocks", param); err != nil {
//		log.Println(err)
//		return blocks
//	}
//
//	//add a fake block to it
//	txFound := visor.ReadableTransactionOutput{}
//	for _, block := range blk.Blocks {
//
//		for _, tx := range block.Body.Transactions {
//
//			for _, out := range tx.Out {
//				txFound = out
//				txFound.Address = deposit.Address
//				txFound.Coins = string(deposit.Value)
//				txFound.Hours = deposit.Hours
//
//				break
//
//			}
//			if txFound.Address != "" {
//				tx.Out = append(tx.Out, txFound)
//				break
//			}
//		}
//		if txFound.Address != "" {
//			break
//		}
//
//	}
//
//	return blk
//}
//
//func (s *SkycoinFake) CreateFakeBlockFromDeposit(deposit Deposit) (blocks visor.ReadableBlocks) {
//	s.DefaultBlockStore.Lock()
//	defer func() {
//		//DefaultBlockStore.BestBlockHeight++
//		hash := blocks.Blocks[0].Head.BodyHash
//		s.DefaultBlockStore.BestBlockHeight++
//		s.DefaultBlockStore.BlockHashes[int64(s.DefaultBlockStore.BestBlockHeight)] = hash
//		s.DefaultBlockStore.HashBlocks[hash] = &blocks
//
//		s.DefaultBlockStore.Unlock()
//	}()
//
//	coins := strconv.Itoa(int(deposit.Value))
//	hours := deposit.Hours
//	address := deposit.Address
//
//	if address == "" {
//		address = "cBnu9sUvv12dovBmjQKTtfE4rbjMmf3fzW"
//	}
//	if coins == "" {
//		coins = "1"
//	}
//
//	blocks = visor.ReadableBlocks{
//		Blocks: []visor.ReadableBlock{
//			{
//				Head: visor.ReadableBlockHeader{
//					BkSeq:             1,
//					BlockHash:         "662835cc081e037561e1fe05860fdc4b426f6be562565bfaa8ec91be5675064a",
//					PreviousBlockHash: "f680fe1f068a1cd5c3ef9194f91a9bc3cacffbcae4a32359a3c014da4ef7516f",
//					Time:              1,
//					Fee:               20732,
//					Version:           1,
//					BodyHash:          "tx_body_hash",
//				},
//				Body: visor.ReadableBlockBody{
//					Transactions: []visor.ReadableTransaction{
//						{
//							Length:    608,
//							Type:      0,
//							Hash:      "662835cc081e037561e1fe05860fdc4b426f6be562565bfaa8ec91be5675064a",
//							InnerHash: "37f1111bd83d9c995b9e48511bd52de3b0e440dccbf6d2cfd41dee31a10f1aa4",
//							Timestamp: 1,
//							Sigs: []string{
//								"ef0b8e1465557e6f21cb2bfad17136188f0b9bd54bba3db76c3488eb8bc900bc7662e3fe162dd6c236d9e52a7051a2133855081a91f6c1a63e1fce2ae9e3820e00",
//								"800323c8c22a2c078cecdfad35210902f91af6f97f0c63fe324e0a9c2159e9356f2fbbfff589edea5a5c24453ef5fc0cd5929f24bebee28e37057acd6d42f3d700",
//								"ca6a6ef5f5fb67490d88ddeeee5e5d11055246613b03e7ed2ad5cc82d01077d262e2da56560083928f5389580ae29500644719cf0e82a5bf065cecbed857598400",
//								"78ddc117607159c7b4c76fc91deace72425f21f2df5918d44d19a377da68cc610668c335c84e2bb7a8f16cd4f9431e900585fc0a3f1024b722b974fcef59dfd500",
//								"4c484d44072e23e97a437deb03a85e3f6eca0bd8875031efe833e3c700fc17f91491969b9864b56c280ef8a68d18dd728b211ce1d46fe477fe3104d73d55ad6501",
//							},
//							In: []string{
//								"4bd7c68ecf3039c2b2d8c26a5e2983e20cf53b6d62b099e7786546b3c3f600f9",
//								"f9e39908677cae43832e1ead2514e01eaae48c9a3614a97970f381187ee6c4b1",
//								"7e8ac23a2422b4666ff45192fe36b1bd05f1285cf74e077ac92cabf5a7c1100e",
//								"b3606a4f115d4161e1c8206f4fb5ac0e91551c40d0ee6fe40c86040d2faacac0",
//								"305f1983f5b630bba27e2777c229c725b6b57f37a6ddee138d1d82ae56311909",
//							},
//							Out: []visor.ReadableTransactionOutput{
//								{
//									Hash:    "574d7e5afaefe4ee7e0adf6ce1971d979f038adc8ebbd35771b2c19b0bad7e3d",
//									Address: address,
//									Coins:   coins,
//									Hours:   hours,
//								},
//							},
//						},
//					},
//				},
//			},
//		},
//	}
//
//	return blocks
//}

func CreateSkycoinBlock() *coin.Block {
	prev := coin.Block{Head: coin.BlockHeader{Version: 0x02, Time: 100, BkSeq: 98}}
	b := make([]byte, 128)
	rand.Read(b)
	uxHash := cipher.SumSHA256(b)

	tx, _ := createTransaction("", 50e1)
	txns := coin.Transactions{tx}

	// valid block is fine
	fee := uint64(121)
	currentTime := uint64(133)
	block, err := coin.NewBlock(prev, currentTime, uxHash, txns, _makeFeeCalc(fee))
	if err != nil {
		panic(err)
	}
	return block
}

func createTransaction(destAddr string, coins uint64) (coin.Transaction, error) {
	addr, err := cipher.DecodeBase58Address(destAddr)
	if err != nil {
		return coin.Transaction{}, err
	}

	return coin.Transaction{
		Out: []coin.TransactionOutput{
			{
				Address: addr,
				Coins:   coins,
			},
		},
	}, nil
}

func _makeFeeCalc(fee uint64) coin.FeeCalculator {
	return func(t *coin.Transaction) (uint64, error) {
		return fee, nil
	}
}

var SkyBlockString = `{
    "blocks": [
        {
            "header": {
                "version": 0,
                "timestamp": 1477295242,
                "seq": 1,
                "fee": 20732,
                "prev_hash": "f680fe1f068a1cd5c3ef9194f91a9bc3cacffbcae4a32359a3c014da4ef7516f",
                "hash": "662835cc081e037561e1fe05860fdc4b426f6be562565bfaa8ec91be5675064a"
            },
            "body": {
                "txns": [
                    {
                        "length": 608,
                        "type": 0,
                        "txid": "662835cc081e037561e1fe05860fdc4b426f6be562565bfaa8ec91be5675064a",
                        "inner_hash": "37f1111bd83d9c995b9e48511bd52de3b0e440dccbf6d2cfd41dee31a10f1aa4",
                        "sigs": [
                            "ef0b8e1465557e6f21cb2bfad17136188f0b9bd54bba3db76c3488eb8bc900bc7662e3fe162dd6c236d9e52a7051a2133855081a91f6c1a63e1fce2ae9e3820e00",
                            "800323c8c22a2c078cecdfad35210902f91af6f97f0c63fe324e0a9c2159e9356f2fbbfff589edea5a5c24453ef5fc0cd5929f24bebee28e37057acd6d42f3d700",
                            "ca6a6ef5f5fb67490d88ddeeee5e5d11055246613b03e7ed2ad5cc82d01077d262e2da56560083928f5389580ae29500644719cf0e82a5bf065cecbed857598400",
                            "78ddc117607159c7b4c76fc91deace72425f21f2df5918d44d19a377da68cc610668c335c84e2bb7a8f16cd4f9431e900585fc0a3f1024b722b974fcef59dfd500",
                            "4c484d44072e23e97a437deb03a85e3f6eca0bd8875031efe833e3c700fc17f91491969b9864b56c280ef8a68d18dd728b211ce1d46fe477fe3104d73d55ad6501"
                        ],
                        "inputs": [
                            "4bd7c68ecf3039c2b2d8c26a5e2983e20cf53b6d62b099e7786546b3c3f600f9",
                            "f9e39908677cae43832e1ead2514e01eaae48c9a3614a97970f381187ee6c4b1",
                            "7e8ac23a2422b4666ff45192fe36b1bd05f1285cf74e077ac92cabf5a7c1100e",
                            "b3606a4f115d4161e1c8206f4fb5ac0e91551c40d0ee6fe40c86040d2faacac0",
                            "305f1983f5b630bba27e2777c229c725b6b57f37a6ddee138d1d82ae56311909"
                        ],
                        "outputs": [
                            {
                                "uxid": "574d7e5afaefe4ee7e0adf6ce1971d979f038adc8ebbd35771b2c19b0bad7e3d",
                                "dst": "cBnu9sUvv12dovBmjQKTtfE4rbjMmf3fzW",
                                "coins": "1",
                                "hours": 3455
                            },
                            {
                                "uxid": "6d8a9c89177ce5e9d3b4b59fff67c00f0471fdebdfbb368377841b03fc7d688b",
                                "dst": "fyqX5YuwXMUs4GEUE3LjLyhrqvNztFHQ4B",
                                "coins": "5",
                                "hours": 3455
                            }
                        ]
                    }
                ]
            }
        }
    ]
}`
