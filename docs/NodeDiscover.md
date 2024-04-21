# 节点发现协议Node Discovery Protocol
这份规范定义了节点发现协议版本4，这是一个类似Kademlia的分布式哈希表（DHT），用于存储有关以太坊节点的信息。选择Kademlia结构是因为它是组织分布式节点索引的有效方式，并且产生了低直径的拓扑结构。

当前协议版本为4。在本文档的末尾，您可以找到以前协议版本的更改列表。

## 节点身份
每个节点都有一个加密身份，即secp256k1椭圆曲线上的一个密钥。节点的公钥作为其标识符或“节点ID”。

两个节点密钥之间的“距离”是对公钥的哈希进行按位异或运算的结果，以数字形式取得。

distance(n₁, n₂) = keccak256(n₁) XOR keccak256(n₂)

## 节点记录
参与发现协议的节点应该维护一个节点记录（ENR），包含最新的信息。所有记录必须使用“v4”身份方案。其他节点可以随时通过发送ENRRequest数据包来请求本地记录。

要解析任何节点公钥的当前记录，请使用FindNode数据包执行Kademlia查找。当找到节点时，向其发送ENRRequest并从响应中返回记录。

## Kademlia表
节点发现协议中的节点会保留其邻居节点的信息。邻居节点存储在由“k-bucket”组成的路由表中。对于0 ≤ i < 256中的每个i，每个节点保持一个k-bucket，其中包含与自身距离在2^i ≤ distance < 2^(i+1)范围内的邻居。

节点发现协议使用k = 16，即每个k-bucket包含多达16个节点条目。这些条目按最后一次看到的时间排序——最近最少看到的节点在头部，最近最频繁看到的在尾部。

每当遇到新节点N₁时，可以将其插入相应的桶中。如果桶中的条目少于k个，N₁可以简单地添加为最后一个条目。如果桶已经包含k个条目，则需要通过发送Ping数据包重新验证桶中最少看到的节点N₂。如果没有收到N₂的回复，则认为它已死亡，将其移除，并将N₁添加到桶的尾部。

## 端点证明
为了防止流量放大攻击，实现必须验证查询的发送者是否参与发现协议。如果发送者在过去12小时内发送了有效的Pong响应，并且ping哈希匹配，则认为其已验证。

## 递归查找
“查找”是定位与节点ID最接近的k个节点。

查找发起者首先选择它已知的与目标最接近的α个节点。然后，发起者向这些节点发送并发的FindNode数据包。α是一个系统范围内的并发参数，比如3。在递归步骤中，发起者重新发送FindNode到它从之前的查询中了解到的节点。在发起者已经听说的与目标最接近的k个节点中，它选择尚未查询过的α个节点，然后重新发送FindNode给它们。未能快速响应的节点将被排除在外，直到它们做出响应。

如果一轮FindNode查询未能返回比已见的最近节点更近的节点，发起者重新发送find node给所有尚未查询过的k个最近节点。当发起者已经查询并得到了它见过的k个最近节点的响应时，查找终止。

## 有线协议
节点发现消息作

为UDP数据报发送。任何数据包的最大大小为1280字节。

packet = packet-header || packet-data
每个数据包都以头部开始：

packet-header = hash || signature || packet-type
hash = keccak256(signature || packet-type || packet-data)
signature = sign(packet-type || packet-data)
哈希的存在是为了使数据包格式在同一个UDP端口上运行多个协议时可识别。它没有其他用途。

每个数据包都由节点的身份密钥签名。签名被编码为长度为65字节的字节数组，为签名值r, s和“恢复id”v的连接。

packet-type是一个定义消息类型的单字节。有效的数据包类型如下所示。头部之后的数据是特定于数据包类型的，并编码为RLP列表。实现应该忽略数据包数据列表中的任何额外元素以及列表之后的任何额外数据。

### Ping数据包（0x01）
packet-data = [version, from, to, expiration, enr-seq ...]
version = 4
from = [sender-ip, sender-udp-port, sender-tcp-port]
to = [recipient-ip, recipient-udp-port, 0]
expiration字段是一个绝对UNIX时间戳。包含过去时间戳的数据包可能不会被处理。

enr-seq字段是发送者的当前ENR序列号。此字段是可选的。

当接收到ping数据包时，接收者应该回复一个Pong数据包。它还可以考虑将发送者添加到本地表中。实现应该忽略版本不匹配的情况。

如果在过去的12小时内没有与发送者的通信，除了pong外还应发送ping以接收端点证明。

### Pong数据包（0x02）
packet-data = [to, ping-hash, expiration, enr-seq, ...]
Pong是对ping的回复。

ping-hash应等于相应ping数据包的哈希。实现应忽略不包含最新ping数据包哈希的未经请求的pong数据包。

enr-seq字段是发送者的当前ENR序列号。此字段是可选的。

### FindNode数据包（0x03）
packet-data = [target, expiration, ...]
FindNode数据包请求有关接近目标的节点的信息。目标是一个64字节的secp256k1公钥。接收FindNode时，接收者应回复包含其本地表中找到的与目标最接近的16个节点的Neighbors数据包。

为了防御流量放大攻击，只有在通过端点证明程序验证了FindNode的发送者后，才应发送Neighbors回复。

### Neighbors数据包（0x04）
packet-data = [nodes, expiration, ...]
nodes = [[ip, udp-port, tcp-port, node-id], ...]
Neighbors是对FindNode的回复。

### ENRRequest数据包（0x05）
packet-data = [expiration]
接收到此类型的数据包时，节点应回复一个ENRResponse数据包，包含其节点记录的当前版本。

为了防御放大攻击，ENRRequest的发送者应该最近回复了一个ping数据包（就像对FindNode一样）。expiration字段，一个UNIX时间戳，应像处理所有其他现有数据包一样处理，即如果它引用过去的时间，则不应发送回复。

### ENRResponse数据包（0x06）
packet-data = [request-hash, ENR]
这个数据包是对ENRRequest的回应。

request-hash是被回复的整个ENRRequest数据包的哈希。
ENR是节点记录。
接收数据包的节点应验证节点记录是否由签署回应数据包的公钥签名。

## 更改日志
### 当前版本中的已知问题
所有数据包中存在的expiration字段旨在防止数据包重放。由于它是一个绝对时间戳，节点的时钟必须准确才能正确验证它。自2016年协议启动以来，我们收到了无数关于用户时钟错误导致的连接问题的报告。

端点证明不精确

，因为FindNode的发送者永远无法确定接收者是否已看到足够新的pong。Geth按以下方式处理它：如果在过去的12小时内没有与接收者的通信，发起程序通过发送ping。等待对方的ping，回复它然后发送FindNode。

## EIP-868（2019年10月）
EIP-868增加了ENRRequest和ENRResponse数据包。它还修改了Ping和Pong以包括本地ENR序列号。

## EIP-8（2017年12月）
EIP-8要求实现忽略Ping版本中的不匹配以及packet-data中的任何附加列表元素。

# Node Discovery Protocol

This specification defines the Node Discovery protocol version 4, a Kademlia-like DHT that
stores information about Ethereum nodes. The Kademlia structure was chosen because it is
an efficient way to organize a distributed index of nodes and yields a topology of low
diameter.

The current protocol version is **4**. You can find a list of changes in past protocol
versions at the end of this document.

## Node Identities

Every node has a cryptographic identity, a key on the secp256k1 elliptic curve. The public
key of the node serves as its identifier or 'node ID'.

The 'distance' between two node keys is the bitwise exclusive or on the hashes of the
public keys, taken as the number.

    distance(n₁, n₂) = keccak256(n₁) XOR keccak256(n₂)

## Node Records

Participants in the Discovery Protocol are expected to maintain a [node record] \(ENR\)
containing up-to-date information. All records must use the "v4" identity scheme. Other
nodes may request the local record at any time by sending an [ENRRequest] packet.

To resolve the current record of any node public key, perform a Kademlia lookup using
[FindNode] packets. When the node is found, send ENRRequest to it and return the record
from the response.

## Kademlia Table

Nodes in the Discovery Protocol keep information about other nodes in their neighborhood.
Neighbor nodes are stored in a routing table consisting of 'k-buckets'. For each `i` in
`0 ≤ i < 256`, every node keeps a k-bucket of neighbors with distance
`2^i ≤ distance < 2^(i+1)` from itself.

The Node Discovery Protocol uses `k = 16`, i.e. every k-bucket contains up to 16 node
entries. The entries are sorted by time last seen — least-recently seen node at the head,
most-recently seen at the tail.

Whenever a new node N₁ is encountered, it can be inserted into the corresponding bucket.
If the bucket contains less than `k` entries N₁ can simply be added as the last entry. If
the bucket already contains `k` entries, the least recently seen node in the bucket, N₂,
needs to be revalidated by sending a [Ping] packet. If no reply is received from N₂ it is
considered dead, removed and N₁ added to the tail of the bucket.

## Endpoint Proof

To prevent traffic amplification attacks, implementations must verify that the sender of a
query participates in the discovery protocol. The sender of a packet is considered
verified if it has sent a valid [Pong] response with matching ping hash within the last 12
hours.

## Recursive Lookup

A 'lookup' locates the `k` closest nodes to a node ID.

The lookup initiator starts by picking `α` closest nodes to the target it knows of. The
initiator then sends concurrent [FindNode] packets to those nodes. `α` is a system-wide
concurrency parameter, such as 3. In the recursive step, the initiator resends FindNode to
nodes it has learned about from previous queries. Of the `k` nodes the initiator has heard
of closest to the target, it picks `α` that it has not yet queried and resends [FindNode]
to them. Nodes that fail to respond quickly are removed from consideration until and
unless they do respond.

If a round of FindNode queries fails to return a node any closer than the closest already
seen, the initiator resends the find node to all of the `k` closest nodes it has not
already queried. The lookup terminates when the initiator has queried and gotten responses
from the `k` closest nodes it has seen.

## Wire Protocol

Node discovery messages are sent as UDP datagrams. The maximum size of any packet is 1280
bytes.

    packet = packet-header || packet-data

Every packet starts with a header:

    packet-header = hash || signature || packet-type
    hash = keccak256(signature || packet-type || packet-data)
    signature = sign(packet-type || packet-data)

The `hash` exists to make the packet format recognizable when running multiple protocols
on the same UDP port. It serves no other purpose.

Every packet is signed by the node's identity key. The `signature` is encoded as a byte
array of length 65 as the concatenation of the signature values `r`, `s` and the 'recovery
id' `v`.

The `packet-type` is a single byte defining the type of message. Valid packet types are
listed below. Data after the header is specific to the packet type and is encoded as an
RLP list. Implementations should ignore any additional elements in the `packet-data` list
as well as any extra data after the list.

### Ping Packet (0x01)

    packet-data = [version, from, to, expiration, enr-seq ...]
    version = 4
    from = [sender-ip, sender-udp-port, sender-tcp-port]
    to = [recipient-ip, recipient-udp-port, 0]

The `expiration` field is an absolute UNIX time stamp. Packets containing a time stamp
that lies in the past are expired may not be processed.

The `enr-seq` field is the current ENR sequence number of the sender. This field is
optional.

When a ping packet is received, the recipient should reply with a [Pong] packet. It may
also consider the sender for addition into the local table. Implementations should ignore
any mismatches in version.

If no communication with the sender has occurred within the last 12h, a ping should be
sent in addition to pong in order to receive an endpoint proof.

### Pong Packet (0x02)

    packet-data = [to, ping-hash, expiration, enr-seq, ...]

Pong is the reply to ping.

`ping-hash` should be equal to `hash` of the corresponding ping packet. Implementations
should ignore unsolicited pong packets that do not contain the hash of the most recent
ping packet.

The `enr-seq` field is the current ENR sequence number of the sender. This field is
optional.

### FindNode Packet (0x03)

    packet-data = [target, expiration, ...]

A FindNode packet requests information about nodes close to `target`. The `target` is a
64-byte secp256k1 public key. When FindNode is received, the recipient should reply with
[Neighbors] packets containing the closest 16 nodes to target found in its local table.

To guard against traffic amplification attacks, Neighbors replies should only be sent if
the sender of FindNode has been verified by the endpoint proof procedure.

### Neighbors Packet (0x04)

    packet-data = [nodes, expiration, ...]
    nodes = [[ip, udp-port, tcp-port, node-id], ...]

Neighbors is the reply to [FindNode].

### ENRRequest Packet (0x05)

    packet-data = [expiration]

When a packet of this type is received, the node should reply with an ENRResponse packet
containing the current version of its [node record].

To guard against amplification attacks, the sender of ENRRequest should have replied to a
ping packet recently (just like for FindNode). The `expiration` field, a UNIX timestamp,
should be handled as for all other existing packets i.e. no reply should be sent if it
refers to a time in the past.

### ENRResponse Packet (0x06)

    packet-data = [request-hash, ENR]

This packet is the response to ENRRequest.

- `request-hash` is the hash of the entire ENRRequest packet being replied to.
- `ENR` is the node record.

The recipient of the packet should verify that the node record is signed by the public key
which signed the response packet.

# Change Log

## Known Issues in the Current Version

The `expiration` field present in all packets is supposed to prevent packet replay. Since
it is an absolute time stamp, the node's clock must be accurate to verify it correctly.
Since the protocol's launch in 2016 we have received countless reports about connectivity
issues related to the user's clock being wrong.

The endpoint proof is imprecise because the sender of FindNode can never be sure whether
the recipient has seen a recent enough pong. Geth handles it as follows: If no
communication with the recipient has occurred within the last 12h, initiate the procedure
by sending a ping. Wait for a ping from the other side, reply to it and then send
FindNode.

## EIP-868 (October 2019)

[EIP-868] adds the [ENRRequest] and [ENRResponse] packets. It also modifies [Ping] and
[Pong] to include the local ENR sequence number.

## EIP-8 (December 2017)

[EIP-8] mandated that implementations ignore mismatches in Ping version and any additional
list elements in `packet-data`.

[Ping]: #ping-packet-0x01
[Pong]: #pong-packet-0x02
[FindNode]: #findnode-packet-0x03
[Neighbors]: #neighbors-packet-0x04
[ENRRequest]: #enrrequest-packet-0x05
[ENRResponse]: #enrresponse-packet-0x06
[EIP-8]: https://eips.ethereum.org/EIPS/eip-8
[EIP-868]: https://eips.ethereum.org/EIPS/eip-868
[node record]: ./enr.md