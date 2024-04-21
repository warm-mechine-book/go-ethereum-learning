# Ethereum Wire Protocol (ETH)

这份文档定义了“eth”协议，这是一种在[RLPx]传输上实现的协议，用于在节点之间交换以太坊区块链信息。当前的协议版本是**eth/68**。关于过去协议版本的更改列表，请参阅本文档末尾。

### 基本操作

一旦建立连接，必须发送[Status]消息。接收到对方的Status消息后，以太坊会话激活，可以发送任何其他消息。

在会话中，可以执行三个高级任务：链同步、区块传播和交易交换。这些任务使用不相交的协议消息集，并且客户端通常在所有对等连接上同时进行这些活动。

客户端实现应该对协议消息大小施加限制。底层RLPx传输限制单个消息的大小为16.7 MiB。eth协议的实际限制通常较低，通常为10 MiB。如果接收到的消息大于限制，应断开对等点连接。

除了接收消息的硬限制外，客户端还应对它们发送的请求和响应施加“软”限制。推荐的软限制因消息类型而异。限制请求和响应可以确保并发活动，例如区块同步和交易交换在同一对等连接上顺利进行。

### 链同步

参与eth协议的节点应该了解从创世块到当前最新区块的所有区块的完整链。通过从其他对等点下载获得链。

连接时，双方发送它们的[Status]消息，其中包括它们“最佳”已知区块的总难度(TD)和哈希。

TD较差的客户端随后开始使用[GetBlockHeaders]消息下载区块头。它验证接收到的头部中的工作量证明值，并使用[GetBlockBodies]消息获取区块体。接收到的区块通过以太坊虚拟机执行，重新创建状态树和收据。

请注意，头部下载、区块体下载和区块执行可以同时进行。

### 状态同步（又名“快速同步”）

协议版本eth/63至eth/66还允许同步状态树。从协议版本eth/67开始，以太坊状态树不能再通过eth协议检索，状态下载由辅助[snap协议]提供。

状态同步通常是通过下载区块头链、验证它们的有效性来进行的。区块体的请求如链同步部分所述，但不执行交易，只验证它们的“数据有效性”。客户端选择链头附近的一个区块（“枢轴区块”）并下载该区块的状态。

### 区块传播

**注意：在PoW到PoS的过渡（[The Merge]）之后，区块传播不再由'eth'协议处理。以下文本仅适用于PoW和PoA（clique）网络。区块传播消息（NewBlock, NewBlockHashes...）将在未来版本中被移除。**

新挖掘的区块必须传递给所有节点。这通过区块传播完成，这是一个两步过程。当从对等点收到[NewBlock]公告消息时，客户端首先验证区块的基本头部有效性，检查工作量证明值是否有效。然后，它使用[NewBlock]消息将区块发送给少数连接的对等点（通常是对等点总数的平方根）。

在头部有效性检查之后，客户端通过执行区块中包含的所有交易、计算区块的“后状态”来将区块导入其本地链。区

块的`state-root`哈希必须与计算出的后状态根匹配。一旦区块被完全处理并被认为是有效的，客户端将关于该区块的[NewBlockHashes]消息发送给之前没有通知的所有对等点。这些对等点稍后可能会请求完整区块，如果它们没有通过[NewBlock]从其他人那里接收到的话。

节点不应将区块公告发送回先前宣布同一区块的对等点。这通常通过记住最近转发给每个对等点或从每个对等点接收的大量区块哈希集来实现。

收到区块公告也可能触发链同步，如果该区块不是客户端当前最新区块的直接后继。

### 交易交换

所有节点必须交换待处理交易，以便将它们中继给矿工，矿工将选择它们加入区块链。客户端实现跟踪“交易池”中的待处理交易集。交易池受客户端特定限制，并可包含许多（即数千个）交易。

当建立新的对等连接时，双方的交易池需要同步。最初，双方应发送[NewPooledTransactionHashes]消息，包含本地池中所有交易哈希以开始交换。

收到NewPooledTransactionHashes公告后，客户端过滤接收到的集合，收集其本地池中尚未拥有的交易哈希。然后，它可以使用[GetPooledTransactions]消息请求交易。

当客户端池中出现新交易时，应使用[Transactions]和[NewPooledTransactionHashes]消息将它们传播到网络。Transactions消息传递完整的交易对象，并通常发送给少数随机选择的连接对等点。所有其他对等点接收交易哈希的通知，并可以请求完整的交易对象（如果它们不知道的话）。将完整交易传递给少数对等点通常确保所有节点接收到交易，不需要请求它。

节点不应将交易发送回可以确定已知该交易的对等点（无论是因为之前发送过还是因为最初是从该对等点获知的）。这通常通过记住最近由对等点转发或接收的一组交易哈希来实现。

### 交易编码和有效性

对等点之间交换的交易对象有两种编码。在整个规范中，我们使用标识符`txₙ`引用任一编码的交易。

    tx = {legacy-tx, typed-tx}

未类型化的传统交易给出为一个RLP列表。

    legacy-tx = [
        nonce: P,
        gas-price: P,
        gas-limit: P,
        recipient: {B_0, B_20},
        value: P,
        data: B,
        V: P,
        R: P,
        S: P,
    ]

[EIP-2718]类型化交易被编码为RLP字节数组，其中第一个字节是交易类型（`tx-type`），剩余字节是特定于类型的不透明数据。

    typed-tx = tx-type || tx-data

接收时必须验证交易的有效性。有效性取决于以太坊链状态。此规范关心的特定有效性不是交易是否可以由EVM成功执行，而仅是它是否适合临时存储在本地池中并与其他对等点交换。

根据以下规则验证交易。虽然类型化交易的编码是不透明的，但假设它们的`tx-data`提供了`nonce`、`gas-price`、`gas-limit`的值，并且可以从它们的签名确定交易的发送账户。



- 如果交易是类型化的，则实现必须知道`tx-type`。定义的交易类型可能在它们可以接受包含在区块中之前就被认为是有效的。实现应断开发送未知类型交易的对等点。
- 签名必须根据链支持的签名方案有效。对于类型化交易，签名处理由引入类型的EIP定义。对于传统交易，活跃使用的两种方案是基本的'Homestead'方案和[EIP-155]方案。
- `gas-limit`必须覆盖交易的“固有gas”。
- 从签名派生的交易的发送账户必须有足够的以太币余额来支付交易的成本（`gas-limit * gas-price + value`）。
- 交易的`nonce`必须等于或大于发送账户的当前nonce。
- 在考虑将交易包含在本地池中时，实现可以确定多少“未来”交易的nonce大于当前账户nonce是有效的，以及“nonce间隙”在多大程度上是可以接受的。

实现可以为交易强制执行其他验证规则。例如，通常会拒绝编码的交易大于128 kB的交易。

除非另有说明，实现不得因发送无效交易而断开对等点连接，而应简单地丢弃它们。这是因为对等点可能在略有不同的验证规则下运行。

### 区块编码和有效性

以太坊区块按以下方式编码：

    block = [header, transactions, ommers]
    transactions = [tx₁, tx₂, ...]
    ommers = [header₁, header₂, ...]
    withdrawals = [withdrawal₁, withdrawal₂, ...]
    header = [
        parent-hash: B_32,
        ommers-hash: B_32,
        coinbase: B_20,
        state-root: B_32,
        txs-root: B_32,
        receipts-root: B_32,
        bloom: B_256,
        difficulty: P,
        number: P,
        gas-limit: P,
        gas-used: P,
        time: P,
        extradata: B,
        mix-digest: B_32,
        block-nonce: B_8,
        basefee-per-gas: P,
        withdrawals-root: B_32,
    ]

在某些协议消息中，交易和ommer列表一起作为称为'block body'的单个项目传输。

    block-body = [transactions, ommers, withdrawals]

区块头的有效性取决于它们的使用上下文。对于单个区块头，只能验证工作量证明封印（`mix-digest`, `block-nonce`）的有效性。当头部用于扩展客户端的本地链时，或者在链同步期间顺序处理多个头部时，适用以下规则：

- 头部必须形成一个链，其中区块编号是连续的，每个头部的`parent-hash`与前一个头部的哈希匹配。
- 在扩展本地存储的链时，实现还必须验证`difficulty`、`gas-limit`和`time`的值是否在[Yellow Paper]中给出的协议规则的范围内。
- `gas-used`头字段必须小于或等于`gas-limit`。
- `basefee-per-gas`必须出现在[London hard fork]之后的头部中。对于更早的区块，此规则必须不存在。这一规则由[EIP-1559]添加。
- 对于[The Merge]之后的PoS区块，`ommers-hash`必须是空的keccak256哈希，因为不能存在任何ommer头。
- `withdrawals-root`必须出现在[Shanghai fork]之后的头部中。对于该分叉之前的区块，此字段必须不存在。这一规则由

[EIP-4895]添加。

对于完整区块，我们区分区块的EVM状态转换的有效性，以及（较弱的）区块的“数据有效性”。此规范不处理状态转换规则的定义。我们需要区块的数据有效性是为了立即[block propagation]的目的以及在[state synchronization]期间。

要确定区块的数据有效性，请使用以下规则。实现应断开发送无效区块的对等点连接。

- 区块`header`必须有效。
- 区块中包含的`transactions`必须有效，以便包含在该区块编号的链中。这意味着，除了前面给出的交易验证规则外，还需要验证`tx-type`是否允许在区块编号上，并且验证交易气体必须考虑区块编号。
- 所有交易的`gas-limit`之和不得超过区块的`gas-limit`。
- 必须通过计算和比较交易列表的merkle trie哈希来验证区块的`transactions`是否符合`txs-root`。
- 必须通过计算和比较提款列表的merkle trie哈希来验证区块体的`withdrawals`是否符合`withdrawals-root`。请注意，由于它们是由共识层注入区块的，所以无法对提款进行进一步验证。
- `ommers`列表最多可以包含两个头部。
- `keccak256(ommers)`必须与区块头部的`ommers-hash`匹配。
- `ommers`列表中包含的头部必须是有效的头部。它们的区块编号不得大于包含它们的区块的编号。ommer头部的父哈希必须指向其包含区块的7代或更少祖先的一个，并且它不得已包含在此祖先集中的任何早期区块中。

### 收据编码和有效性

收据是区块的EVM状态转换的输出。与交易一样，收据有两种不同的编码，我们将使用标识符`receiptₙ`引用任一编码的收据。

    receipt = {legacy-receipt, typed-receipt}

未类型化的传统收据按以下方式编码：

    legacy-receipt = [
        post-state-or-status: {B_32, {0, 1}},
        cumulative-gas: P,
        bloom: B_256,
        logs: [log₁, log₂, ...]
    ]
    log = [
        contract-address: B_20,
        topics: [topic₁: B, topic₂: B, ...],
        data: B
    ]

[EIP-2718]类型化收据被编码为RLP字节数组，其中第一个字节给出收据类型（与`tx-type`匹配），剩余字节是特定于类型的不透明数据。

    typed-receipt = tx-type || receipt-data

在以太坊线路协议中，收据总是作为区块中包含的所有收据的完整列表传输。还假设包含收据的区块是有效的并且已知的。当对等点接收到一列区块收据时，必须通过计算并比较列表的merkle trie哈希与区块的`receipts-root`进行验证。由于有效的收据列表由EVM状态转换确定，因此本规范中不需要为收据定义任何进一步的有效性规则。

## 协议消息

在大多数消息中，消息数据列表的第一个元素是`request-id`。对于请求，这是请求方选择的64位整数值。响应方必须在响应消息的`request-id`元素中镜像该值。

### 状态（0x00）

`[version: P, networkid: P, td: P, blockhash: B_32, genesis: B_32, forkid]`

通知对等

点其当前状态。此消息应在连接建立后立即发送，并在任何其他eth协议消息之前发送。

- `version`：当前协议版本
- `networkid`：标识区块链的整数，请参阅下表
- `td`：最佳链的总难度。整数，如区块头中所示。
- `blockhash`：最佳（即最高TD）已知区块的哈希
- `genesis`：创世区块的哈希
- `forkid`：[EIP-2124]叉标识符，编码为`[fork-hash, fork-next]`。

此表列出了常见的网络ID及其对应的网络。还存在未列出的其他ID，即客户端不应要求使用任何特定的网络ID。请注意，网络ID可能与用于交易重放防护的EIP-155链ID相对应，也可能不相对应。

| ID | chain                     |
|----|---------------------------|
| 0  | Olympic (已废弃)         |
| 1  | Frontier (现在是mainnet)    |
| 2  | Morden testnet (已废弃)  |
| 3  | Ropsten testnet (已废弃) |
| 4  | Rinkeby testnet (已废弃) |
| 5  | Goerli testnet            |

有关链ID的社区策划列表，请参阅<https://chainid.network>。

### 新区块哈希（0x01）

`[[blockhash₁: B_32, number₁: P], [blockhash₂: B_32, number₂: P], ...]`

指定一个或多个在网络上出现的新区块。为了尽可能有用，节点应通知对等点所有它们可能不知道的区块。包括发送对等点可能合理地被认为知道的哈希（因为它们以前被通知过，因为该节点已经通过NewBlockHashes宣布了哈希知识）被认为是不良形式，并可能降低发送节点的声誉。包括发送节点后来拒绝通过[GetBlockHeaders]消息兑现的哈希也被认为是不良形式，并可能降低发送节点的声誉。

### 交易（0x02）

`[tx₁, tx₂, ...]`

指定对等方应确保包含在其交易队列中的交易。列表中的项目是以太坊主要规范中描述的格式的交易。Transactions消息必须包含至少一个（新的）交易，不鼓励空的Transactions消息，并可能导致断开连接。

节点在同一会话中不得将相同的交易重新发送给对等点，并且不得将交易中继给接收该交易的对等点。在实践中，这通常通过保持每个对等点的bloom过滤器或已发送或接收的交易哈希集来实现。

### 获取区块头（0x03）

`[request-id: P, [startblock: {P, B_32}, limit: P, skip: P, reverse: {0, 1}]]`

要求对等方返回BlockHeaders消息。响应必须包含若干区块头，当`reverse`为`0`时编号上升，为`1`时下降，`skip`区块间隔，从规范链上的区块`startblock`（由数字或哈希表示）开始，最多`limit`项。

### 区块头（0x04）

`[request-id: P, [header₁, header₂, ...]]`

这是对GetBlockHeaders的响应，包含所请求的头部。头部列表可能为空，如果没有找到任何请求的区块头。单个消息中可以请求的头部数量可能受到实现定义的限制。

BlockHeaders响应的推荐软限制为2 MiB。

### 获取

区块体（0x05）

`[request-id: P, [blockhash₁: B_32, blockhash₂: B_32, ...]]`

此消息通过哈希请求区块体数据。单个消息中可以请求的区块数量可能受到实现定义的限制。

### 区块体（0x06）

`[request-id: P, [block-body₁, block-block-body₂, ...]]`

这是对GetBlockBodies的回应。列表中的项目包含请求的区块的体数据。如果没有请求的区块可用，列表可能为空。

BlockBodies响应的推荐软限制为2 MiB。

### 新区块（0x07）

`[block, td: P]`

指定一个对等方应该知道的完整单个区块。`td`是区块的总难度，即包括此区块在内的所有区块难度之和。

### 新池化交易哈希（0x08）

`[txtypes: B, [txsize₁: P, txsize₂: P, ...], [txhash₁: B_32, txhash₂: B_32, ...]]`

此消息宣布一个或多个在网络中出现且尚未包含在区块中的交易。消息有效负载描述了一列交易，但请注意，它被编码为三个单独的元素。

`txtypes`元素是一个字节数组，包含宣布的[transaction types]。其他两个有效载荷元素指的是宣布的交易的大小和哈希。所有三个有效载荷元素必须包含相同数量的项目。

`txsizeₙ`指的是类型化交易的“共识编码”的长度，即类型化交易的`tx-type || tx-data`的字节大小，以及非类型化传统交易的RLP编码的`legacy-tx`的大小。

此消息的推荐软限制为4096项（~150 KiB）。

为了尽可能有用，节点应通知对等点所有它们可能不知道的交易。然而，节点只应宣布远程对等点可能合理地被认为不知道的交易哈希，但最好返回比在池中有nonce间隙的更多交易。

### 获取池化交易（0x09）

`[request-id: P, [txhash₁: B_32, txhash₂: B_32, ...]]`

此消息通过哈希请求接收方的交易池中的交易。

GetPooledTransactions请求的推荐软限制为256个哈希（8 KiB）。接收方可以对响应（大小或服务时间）施加任意限制，这不应被视为协议违规。

### 池化交易（0x0a）

`[request-id: P, [tx₁, tx₂...]]`

这是对GetPooledTransactions的回应，返回本地池中请求的交易。列表中的项目是以太坊主要规范中描述的格式的交易。

交易必须按请求中的顺序排列，但如果交易不可用，则可以跳过。这样，如果响应大小限制达到，请求者将知道要再次请求哪些哈希（从最后返回的交易开始）以及哪些假设不可用（在最后返回的交易之前的所有间隙）。

允许首先通过NewPooledTransactionHashes宣布一笔交易，但随后拒绝通过PooledTransactions提供它。当交易在公告和请求之间被包含在区块中（并从池中删除）时，可能会出现这种情况。

如果没有任何哈希与池中的交易匹配，对等方可以用空列表响应。

### 获取收据（0x0f）

`[request-id: P, [blockhash₁: B_32, blockhash₂: B_32, ...]]`

要求对等方返回包含给

定区块哈希的收据的Receipts消息。单个消息中可以请求的收据数量可能受到实现定义的限制。

### 收据（0x10）

`[request-id: P, [[receipt₁, receipt₂], ...]]`

这是对GetReceipts的回应，提供请求的区块收据。响应列表中的每个元素对应于GetReceipts请求的一个区块哈希，并且必须包含该区块的完整收据列表。

Receipts响应的推荐软限制为2 MiB。

## 更改日志

### eth/68 ([EIP-5793], 2022年10月)

版本68更改了[NewPooledTransactionHashes]消息，以包括宣布的交易的类型和大小。在此更新之前，消息有效载荷仅是哈希列表：`[txhash₁: B_32, txhash₂: B_32, ...]`。

### eth/67与提款（[EIP-4895], 2022年3月）

PoS验证器提款由[EIP-4895]添加，它更改了区块头的定义以包括`withdrawals-root`，并将`withdrawals`列表包含在区块体中。没有为此更改创建新的线路协议版本，因为它只是对区块有效性规则的向后兼容添加。

### eth/67 ([EIP-4938], 2022年3月)

版本67移除了GetNodeData和NodeData消息。

- GetNodeData (0x0d)
  `[request-id: P, [hash₁: B_32, hash₂: B_32, ...]]`
- NodeData (0x0e)
  `[request-id: P, [value₁: B, value₂: B, ...]]`

### eth/66 ([EIP-2481], 2021年4月)

版本66在消息[GetBlockHeaders], [BlockHeaders], [GetBlockBodies], [BlockBodies], [GetPooledTransactions], [PooledTransactions], GetNodeData, NodeData, [GetReceipts], [Receipts]中添加了`request-id`元素。

### eth/65与类型化交易（[EIP-2976], 2021年4月）

当[EIP-2718]引入类型化交易时，客户端实现者决定在不增加协议版本的情况下接受新的交易和收据格式。这次规范更新还增加了所有共识对象的编码定义，而不是引用Yellow Paper。

### eth/65 ([EIP-2464], 2020年1月)

版本65改进了交易交换，引入了三个附加消息：[NewPooledTransactionHashes], [GetPooledTransactions], 和[PooledTransactions]。

在版本65之前，对等方总是交换完整的交易对象。随着以太坊主网上的活动和交易大小的增加，用于交易交换的网络带宽成为节点运营商的重要负担。通过采用与区块传播类似的两级交易广播系统，更新减少了所需的带宽。

### eth/64 ([EIP-2364], 2019年11月)

版本64将[EIP-2124] ForkID包含在[Status]消息中。这允许对等方在同步区块链之前确定链执行规则的相互兼容性。

### eth/63（2016年）

版本63增加了GetNodeData, NodeData, [GetReceipts] 和 [Receipts] 消息，允许同步交易执行结果。

### eth/62（2015年）

在版本62中，[NewBlockHashes]消息扩展到包括与宣布的哈希一起的区块编号。[Status]中的区块编号被移除。GetBlockHashes (0x03), BlockHashes (0x04), GetBlocks (0x05) 和 Blocks (0x06) 的消息被替换为获取区块头和体的消息。BlockHashesFromNumber (0x08) 消息

被移除。

重新分配/移除消息代码的先前编码如下：

- GetBlockHashes (0x03): `[hash: B_32, max-blocks: P]`
- BlockHashes (0x04): `[hash₁: B_32, hash₂: B_32, ...]`
- GetBlocks (0x05): `[hash₁: B_32, hash₂: B_32, ...]`
- Blocks (0x06): `[[header, transactions, ommers], ...]`
- BlockHashesFromNumber (0x08): `[number: P, max-blocks: P]`

### eth/61（2015年）

版本61添加了BlockHashesFromNumber (0x08)消息，该消息可用于按升序请求区块。它还在[Status]消息中添加了最新区块编号。

### eth/60及以下

60以下的版本号在以太坊PoC开发阶段使用。

- `0x00` 用于 PoC-1
- `0x01` 用于 PoC-2
- `0x07` 用于 PoC-3
- `0x09` 用于 PoC-4
- `0x17` 用于 PoC-5
- `0x1c` 用于 PoC-6

[block propagation]: #block-propagation
[state synchronization]: #state-synchronization-aka-fast-sync
[snap protocol]: ./snap.md
[Status]: #status-0x00
[NewBlockHashes]: #newblockhashes-0x01
[Transactions]: #transactions-0x02
[GetBlockHeaders]: #getblockheaders-0x03
[BlockHeaders]: #blockheaders-0x04
[GetBlockBodies]: #getblockbodies-0x05
[BlockBodies]: #blockbodies-0x06
[NewBlock]: #newblock-0x07
[NewPooledTransactionHashes]: #newpooledtransactionhashes-0x08
[GetPooledTransactions]: #getpooledtransactions-0x09
[PooledTransactions]: #pooledtransactions-0x0a
[GetReceipts]: #getreceipts-0x0f
[Receipts]: #receipts-0x10
[RLPx]: ../rlpx.md
[EIP-155]: https://eips.ethereum.org/EIPS/eip-155
[EIP-1559]: https://eips.ethereum.org/EIPS/eip-1559
[EIP-2124]: https://eips.ethereum.org/EIPS/eip-2124
[EIP-2364]: https://eips.ethereum.org/EIPS/eip-2364
[EIP-2464]: https://eips.ethereum.org/EIPS/eip-2464
[EIP-2481]: https://eips.ethereum.org/EIPS/eip-2481
[EIP-2718]: https://eips.ethereum.org/EIPS/eip-2718
[transaction types]: https://eips.ethereum.org/EIPS/eip-2718
[EIP-2976]: https://eips.ethereum.org/EIPS/eip-2976
[EIP-4895]: https://eips.ethereum.org/EIPS/eip-4895
[EIP-4938]: https://eips.ethereum.org/EIPS/eip-4938
[EIP-5793]: https://eips.ethereum.org/EIPS/eip-5793
[The Merge]: https://eips.ethereum.org/EIPS/eip-3675
[London hard fork]: https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/london.md
[Shanghai fork]: https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/shanghai.md
[Yellow Paper]: https://ethereum.github.io/yellowpaper/paper.pdf