# devp2p
这个仓库包含了以太坊使用的点对点网络协议的规范。这里的问题跟踪器用于讨论协议的变更。如果你只是有一个问题，也可以开一个 issue。

协议级别的安全问题非常重要！请通过以太坊基金会赏金计划负责任地报告严重问题。

我们有几个低层协议的规范：
- 以太坊节点记录
- DNS节点列表
- 节点发现协议 v4
- 节点发现协议 v5
- RLPx 协议

该仓库还包含许多基于 RLPx 的应用层协议的规范：
- 以太坊线路协议（eth/68）
- 以太坊快照协议（snap/1）
- 轻量以太坊子协议（les/4）
- Parity轻协议（pip/1）
- 以太坊见证协议（wit/0）

## 任务
devp2p 是构成以太坊点对点网络的一套网络协议。这里所说的“以太坊网络”是广义上的，即 devp2p 不特定于某个区块链，而应满足任何与以太坊旗下相关的网络化应用的需求。

我们的目标是实现一个由多种编程环境实现的正交部分的集成系统。该系统提供了发现整个互联网中的其他参与者以及与这些参与者的安全通信的功能。

devp2p 中的网络协议应该易于仅根据规范从零开始实现，并且必须在消费级互联网连接的限制内工作。我们通常采用“先规范后设计”的方法，但任何提出的规范都必须伴随一个工作原型或在合理的时间内可实现。

## 与 libp2p 的关系
libp2p 项目大约在 devp2p 同时启动，旨在成为一系列模块的集合，用于从模块化组件组装点对点网络。关于 devp2p 和 libp2p 之间的关系的问题相当频繁。

这两个项目很难进行比较，因为它们的范围和设计目标不同。devp2p 是一个集成的系统定义，旨在很好地满足以太坊的需求（尽管它也可能适合其他应用），而 libp2p 是一个不特定于任何单一应用的编程库部件集合。

尽管如此，这两个项目在精神上非常相似，devp2p 正在逐渐采用 libp2p 的部分成熟部件。

## 实现
devp2p 是大多数以太坊客户端的一部分。实现包括：

C#：Nethermind https://github.com/NethermindEth/nethermind
C++：Aleth https://github.com/ethereum/aleth
C：Breadwallet https://github.com/breadwallet/breadwallet-core
Elixir：Exthereum https://github.com/exthereum/ex_wire
Go：go-ethereum/geth https://github.com/ethereum/go-ethereum
Java：Tuweni RLPx 库 https://github.com/apache/incubator-tuweni/tree/master/rlpx
Java：Besu https://github.com/hyperledger/besu
JavaScript：EthereumJS https://github.com/ethereumjs/ethereumjs-devp2p
Kotlin：Tuweni Discovery 库 https://github.com/apache/incubator-tuweni/tree/master/devp2p
Nim：Nimbus nim-eth https://github.com/status-im/nim-eth
Python：Trinity https://github.com/ethereum/trinity