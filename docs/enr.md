# 以太坊节点记录

这份规范定义了以太坊节点记录（ENR），这是一种用于点对点连接信息的开放格式。节点记录通常包含节点的网络端点，即节点的IP地址和端口。它还包含有关节点在网络上的作用的信息，以便其他节点决定是否连接到该节点。

以太坊节点记录最初在 [EIP-778] 中提出。

## 记录结构

节点记录的组成部分包括：

- `signature`：记录内容的加密签名
- `seq`：序列号，一个64位无符号整数。节点应在记录更改并重新发布记录时增加该数字。
- 记录的其余部分由任意键/值对组成

记录的签名是根据*身份方案*进行制作和验证的。身份方案还负责在DHT中派生节点的地址。

键/值对必须按键排序并且必须是唯一的，即任何键只能出现一次。键可以是任何字节序列，但首选ASCII文本。下表中的键名具有预定义的含义。

| 键           | 值                                          |
|:------------|:-------------------------------------------|
| `id`        | 身份方案的名称，例如 "v4"                  |
| `secp256k1` | 压缩的 secp256k1 公钥，33字节              |
| `ip`        | IPv4地址，4字节                            |
| `tcp`       | TCP端口，大端整数                          |
| `udp`       | UDP端口，大端整数                          |
| `ip6`       | IPv6地址，16字节                           |
| `tcp6`      | IPv6特定TCP端口，大端整数                  |
| `udp6`      | IPv6特定UDP端口，大端整数                  |

除了`id`之外的所有键都是可选的，包括IP地址和端口。一个没有端点信息的记录只要签名有效就仍然有效。如果没有提供`tcp6`/`udp6`端口，则`tcp`/`udp`端口适用于两种IP地址。声明在`tcp`、`tcp6`或`udp`、`udp6`中使用相同的端口号应避免，但这不会使记录无效。

### RLP编码

节点记录的规范编码是一个RLP列表`[signature, seq, k, v, ...]`。节点记录的最大编码大小为300字节。实现应拒绝超过此大小的记录。

记录的签名和编码方式如下：

    content   = [seq, k, v, ...]
    signature = sign(content)
    record    = [signature, seq, k, v, ...]

### 文本编码

节点记录的文本形式是其RLP表示的base64编码，前缀为`enr:`。实现应使用[URL安全的base64字母表]并省略填充字符。

### “v4”身份方案

这份规范定义了一个单一的身份方案，作为默认使用，直到通过进一步的EIP定义其他方案。"v4"方案与节点发现v4使用的加密系统向后兼容。

- 要使用此方案对记录`content`进行签名，请对`content`应用keccak256哈希函数（由EVM使用），然后对哈希创建签名。结果的64字节签名被编码为`r`和`s`签名值的连接（恢复ID `v`被省略）。

- 要验证记录，请检查签名是否由记录的"secp256k1"键/值对中的公钥制作。

- 要派生节点地址，请取未压缩公钥的keccak256哈希，即`keccak256(x || y)`。注意`x`和`y`必须填充到长度32。

## 基本原理

该格式旨在以两种方式满足未来的需求：

- 添加新的键

/值对：这始终是可能的，不需要实现共识。现有客户端将接受任何键/值对，无论它们是否能解释其内容。
- 添加身份方案：这些需要实现共识，因为网络否则不会接受签名。要引入新的身份方案，请提出一个EIP并实施它。一旦大多数客户端接受，就可以使用该方案。

记录的大小受限，因为记录频繁地被转发，并可能包含在如DNS这样的大小受限的协议中。使用“v4”方案签名的包含IPv4地址的记录大约占120字节，为附加元数据留有足够的空间。

你可能会想，为什么需要这么多预定义的与IP地址和端口相关的键。这种需求是因为住宅和移动网络设置通常将IPv4置于NAT之后，而IPv6流量（如果支持）则直接路由到同一主机。声明两种地址类型确保了节点可以从仅支持IPv4的位置和支持两种协议的位置访问。

## 测试向量

这是一个包含IPv4地址`127.0.0.1`和UDP端口`30303`的示例记录。节点ID为`a448f24c6d18e575453db13171562b71999873db5b286df957af199ec94617f7`。

    enr:-IS4QHCYrYZbAKWCBRlAy5zzaDZXJBGkcnh4MHcBFZntXNFrdvJjX04jRzjzCBOonrkTfj499SZuOh8R33Ls8RRcy5wBgmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxoQPKY0yuDUmstAHYpMa2_oxVtw0RW_QAdpzBQA8yWM0xOIN1ZHCCdl8

该记录使用“v4”身份方案签名，序列号为`1`，使用此私钥：

    b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291

记录的RLP结构为：

    [
      7098ad865b00a582051940cb9cf36836572411a47278783077011599ed5cd16b76f2635f4e234738f30813a89eb9137e3e3df5266e3a1f11df72ecf1145ccb9c,
      01,
      "id",
      "v4",
      "ip",
      7f000001,
      "secp256k1",
      03ca634cae0d49acb401d8a4c6b6fe8c55b70d115bf400769cc1400f3258cd3138,
      "udp",
      765f,
    ]

[EIP-778]: https://eips.ethereum.org/EIPS/eip-778
[URL-safe base64 alphabet]: https://tools.ietf.org/html/rfc4648#section-5

# Ethereum Node Records

This specification defines Ethereum Node Records (ENR), an open format for p2p
connectivity information. A node record usually contains the network endpoints of a node,
i.e. the node's IP addresses and ports. It also holds information about the node's purpose
on the network so others can decide whether to connect to the node.

Ethereum Node Records were originally proposed in [EIP-778].

## Record Structure

The components of a node record are:

- `signature`: cryptographic signature of record contents
- `seq`: The sequence number, a 64-bit unsigned integer. Nodes should increase the number
  whenever the record changes and republish the record.
- The remainder of the record consists of arbitrary key/value pairs

A record's signature is made and validated according to an *identity scheme*. The identity
scheme is also responsible for deriving a node's address in the DHT.

The key/value pairs must be sorted by key and must be unique, i.e. any key may be present
only once. The keys can technically be any byte sequence, but ASCII text is preferred. Key
names in the table below have pre-defined meaning.

| Key         | Value                                      |
|:------------|:-------------------------------------------|
| `id`        | name of identity scheme, e.g. "v4"         |
| `secp256k1` | compressed secp256k1 public key, 33 bytes  |
| `ip`        | IPv4 address, 4 bytes                      |
| `tcp`       | TCP port, big endian integer               |
| `udp`       | UDP port, big endian integer               |
| `ip6`       | IPv6 address, 16 bytes                     |
| `tcp6`      | IPv6-specific TCP port, big endian integer |
| `udp6`      | IPv6-specific UDP port, big endian integer |

All keys except `id` are optional, including IP addresses and ports. A record without
endpoint information is still valid as long as its signature is valid. If no `tcp6` /
`udp6` port is provided, the `tcp` / `udp` port applies to both IP addresses. Declaring
the same port number in both `tcp`, `tcp6` or `udp`, `udp6` should be avoided but doesn't
render the record invalid.

### RLP Encoding

The canonical encoding of a node record is an RLP list of `[signature, seq, k, v, ...]`.
The maximum encoded size of a node record is 300 bytes. Implementations should reject
records larger than this size.

Records are signed and encoded as follows:

    content   = [seq, k, v, ...]
    signature = sign(content)
    record    = [signature, seq, k, v, ...]

### Text Encoding

The textual form of a node record is the base64 encoding of its RLP representation,
prefixed by `enr:`. Implementations should use the [URL-safe base64 alphabet]
and omit padding characters.

### "v4" Identity Scheme

This specification defines a single identity scheme to be used as the default until other
schemes are defined by further EIPs. The "v4" scheme is backwards-compatible with the
cryptosystem used by Node Discovery v4.

- To sign record `content` with this scheme, apply the keccak256 hash function (as used by
  the EVM) to `content`, then create a signature of the hash. The resulting 64-byte
  signature is encoded as the concatenation of the `r` and `s` signature values (the
  recovery ID `v` is omitted).

- To verify a record, check that the signature was made by the public key in the
  "secp256k1" key/value pair of the record.

- To derive a node address, take the keccak256 hash of the uncompressed public key, i.e.
  `keccak256(x || y)`. Note that `x` and `y` must be zero-padded up to length 32.

## Rationale

The format is meant to suit future needs in two ways:

- Adding new key/value pairs: This is always possible and doesn't require implementation
  consensus. Existing clients will accept any key/value pairs regardless of whether they
  can interpret their content.
- Adding identity schemes: these need implementation consensus because the network won't
  accept the signature otherwise. To introduce a new identity scheme, propose an EIP and
  get it implemented. The scheme can be used as soon as most clients accept it.

The size of a record is limited because records are relayed frequently and may be included
in size-constrained protocols such as DNS. A record containing a IPv4 address, when signed
using the "v4" scheme occupies roughly 120 bytes, leaving plenty of room for additional
metadata.

You might wonder about the need for so many pre-defined keys related to IP addresses and
ports. This need arises because residential and mobile network setups often put IPv4
behind NAT while IPv6 traffic—if supported—is directly routed to the same host. Declaring
both address types ensures a node is reachable from IPv4-only locations and those
supporting both protocols.

## Test Vectors

This is an example record containing the IPv4 address `127.0.0.1` and UDP port `30303`.
The node ID is `a448f24c6d18e575453db13171562b71999873db5b286df957af199ec94617f7`.

    enr:-IS4QHCYrYZbAKWCBRlAy5zzaDZXJBGkcnh4MHcBFZntXNFrdvJjX04jRzjzCBOonrkTfj499SZuOh8R33Ls8RRcy5wBgmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxoQPKY0yuDUmstAHYpMa2_oxVtw0RW_QAdpzBQA8yWM0xOIN1ZHCCdl8

The record is signed using the "v4" identity scheme using sequence number `1` and this
private key:

    b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291

The RLP structure of the record is:

    [
      7098ad865b00a582051940cb9cf36836572411a47278783077011599ed5cd16b76f2635f4e234738f30813a89eb9137e3e3df5266e3a1f11df72ecf1145ccb9c,
      01,
      "id",
      "v4",
      "ip",
      7f000001,
      "secp256k1",
      03ca634cae0d49acb401d8a4c6b6fe8c55b70d115bf400769cc1400f3258cd3138,
      "udp",
      765f,
    ]

[EIP-778]: https://eips.ethereum.org/EIPS/eip-778
[URL-safe base64 alphabet]: https://tools.ietf.org/html/rfc4648#section-5