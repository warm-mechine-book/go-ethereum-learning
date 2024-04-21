# RLPx
这份规范定义了RLPx传输协议，这是一个基于TCP的传输协议，用于以太坊节点之间的通信。该协议传输加密消息，这些消息属于在连接建立期间协商的一个或多个“能力”。RLPx以[RLP]序列化格式命名。这个名称不是首字母缩略词，没有特定含义。

当前协议版本为**5**。在本文档的末尾，您可以找到过去版本的更改列表。

## 标记

`X || Y`\
    表示X和Y的连接。\
`X ^ Y`\
    是X和Y的逐字节XOR。\
`X[:N]`\
    表示X的N字节前缀。\
`[X, Y, Z, ...]`\
    表示作为RLP列表的递归编码。\
`keccak256(MESSAGE)`\
    是以太坊使用的Keccak256哈希函数。\
`ecies.encrypt(PUBKEY, MESSAGE, AUTHDATA)`\
    是RLPx使用的非对称认证加密函数。\
    AUTHDATA是认证数据，不是结果密文的一部分，\
    但在生成消息标签之前写入HMAC-256。\
`ecdh.agree(PRIVKEY, PUBKEY)`\
    是私钥和公钥之间的椭圆曲线Diffie-Hellman密钥协议。

## ECIES加密

ECIES（椭圆曲线集成加密方案）是RLPx握手中使用的一种非对称加密方法。RLPx使用的密码系统为：

- 使用生成元`G`的椭圆曲线secp256k1。
- `KDF(k, len)`：NIST SP 800-56的串联密钥派生函数。
- `MAC(k, m)`：使用SHA-256哈希函数的HMAC。
- `AES(k, iv, m)`：CTR模式下的AES-128加密函数。

Alice想发送一个加密消息，只有Bob的静态私钥 <code>k<sub>B</sub></code> 能解密。Alice知道Bob的静态公钥 <code>K<sub>B</sub></code>。

为了加密消息`m`，Alice生成一个随机数`r`及相应的椭圆曲线公钥`R = r * G`，并计算共享密钥 <code>S = P<sub>x</sub></code>，其中 <code>(P<sub>x</sub>, P<sub>y</sub>) = r * K<sub>B</sub></code>。她派生加密和认证的密钥材料为
<code>k<sub>E</sub> || k<sub>M</sub> = KDF(S, 32)</code>，以及一个随机初始化向量`iv`。Alice发送加密消息`R || iv || c || d`给Bob，其中 <code>c = AES(k<sub>E</sub>, iv, m)</code> 和
<code>d = MAC(sha256(k<sub>M</sub>), iv || c)</code>。

为了解密消息`R || iv || c || d`，Bob派生共享密钥 <code>S = P<sub>x</sub></code>，其中 <code>(P<sub>x</sub>, P<sub>y</sub>) = k<sub>B</sub> * R</code> 以及加密和认证密钥 <code>k<sub>E</sub> || k<sub>M</sub> = KDF(S, 32)</code>。Bob通过检查 <code>d == MAC(sha256(k<sub>M</sub>), iv || c)</code> 验证消息的真实性，然后获得明文 <code>m = AES(k<sub>E</sub>, iv, c)</code>。

## 节点身份

所有加密操作都基于secp256k1椭圆曲线。每个节点都应维护一个静态secp256k1私钥，该私钥在会

话间保存和恢复。建议只能手动重置私钥，例如通过删除文件或数据库条目。

## 初始握手

通过创建TCP连接并同意进一步加密和认证通信的临时密钥材料来建立RLPx连接。创建这些会话密钥的过程称为“握手”，在“发起者”（打开TCP连接的节点）和“接收者”（接受它的节点）之间进行。

1. 发起者连接到接收者并发送其`auth`消息
2. 接收者接受、解密并验证`auth`（检查签名恢复 == `keccak256(ephemeral-pubk)`）
3. 接收者生成`auth-ack`消息从`remote-ephemeral-pubk`和`nonce`
4. 接收者派生密钥并发送包含[Hello]消息的第一个加密帧
5. 发起者接收`auth-ack`并派生密钥
6. 发起者发送包含发起者[Hello]消息的第一个加密帧
7. 接收者接收并认证第一个加密帧
8. 发起者接收并认证第一个加密帧
9. 如果双方第一个加密帧的MAC有效，则加密握手完成

如果第一个帧数据包的认证失败，任何一方都可能断开连接。

握手消息：

    auth = auth-size || enc-auth-body
    auth-size = enc-auth-body的大小，编码为大端16位整数
    auth-vsn = 4
    auth-body = [sig, initiator-pubk, initiator-nonce, auth-vsn, ...]
    enc-auth-body = ecies.encrypt(recipient-pubk, auth-body || auth-padding, auth-size)
    auth-padding = 任意数据

    ack = ack-size || enc-ack-body
    ack-size = enc-ack-body的大小，编码为大端16位整数
    ack-vsn = 4
    ack-body = [recipient-ephemeral-pubk, recipient-nonce, ack-vsn, ...]
    enc-ack-body = ecies.encrypt(initiator-pubk, ack-body || ack-padding, ack-size)
    ack-padding = 任意数据

实现必须忽略`auth-vsn`和`ack-vsn`的任何不匹配。实现还必须忽略`auth-body`和`ack-body`中的任何额外列表元素。

握手消息交换后生成的密钥：

    static-shared-secret = ecdh.agree(privkey, remote-pubk)
    ephemeral-key = ecdh.agree(ephemeral-privkey, remote-ephemeral-pubk)
    shared-secret = keccak256(ephemeral-key || keccak256(nonce || initiator-nonce))
    aes-secret = keccak256(ephemeral-key || shared-secret)
    mac-secret = keccak256(ephemeral-key || aes-secret)

## 分帧

初始握手后的所有消息都进行分帧。一个帧携带属于一个能力的单个加密消息。

分帧的目的是在单个连接上多路复用多个能力。其次，由于分帧消息为消息认证码提供了合理的划分点，支持加密和认证的流变得简单。帧通过握手期间生成的密钥材料进行加密和认证。

帧头提供有关消息大小和消息来源能力的信息。使用填充以防止缓冲区饥饿，使帧组件与密码的块大小字节对齐。

    frame = header-ciphertext || header-mac || frame-ciphertext || frame-mac
    header-ciphertext = aes(aes-secret, header)
    header = frame-size || header-data || header-padding
    header-data = [capability-id, context-id]
    capability-id = 整数，始终为零
    context-id = 整数，始终为零
    header-padding = 将头部填充到16字节边界
    frame-ciphertext = aes(aes-secret, frame-data || frame-padding)
    frame-padding = 将

帧数据填充到16字节边界

有关`frame-data`和`frame-size`的定义，请参见[Capability Messaging]部分。

### MAC

RLPx中的消息认证使用两个keccak256状态，每个通信方向一个。`egress-mac`和`ingress-mac` keccak状态不断更新，以发送（egress）或接收（ingress）的字节的密文。在初始握手后，MAC状态如下初始化：

发起者：

    egress-mac = keccak256.init((mac-secret ^ recipient-nonce) || auth)
    ingress-mac = keccak256.init((mac-secret ^ initiator-nonce) || ack)

接收者：

    egress-mac = keccak256.init((mac-secret ^ initiator-nonce) || ack)
    ingress-mac = keccak256.init((mac-secret ^ recipient-nonce) || auth)

发送帧时，通过使用要发送的数据更新`egress-mac`状态来计算相应的MAC值。更新是通过将头部与其对应的MAC的加密输出异或来执行的。这样做是为了确保对明文MAC和密文执行统一操作。所有MAC都以明文发送。

    header-mac-seed = aes(mac-secret, keccak256.digest(egress-mac)[:16]) ^ header-ciphertext
    egress-mac = keccak256.update(egress-mac, header-mac-seed)
    header-mac = keccak256.digest(egress-mac)[:16]

计算`frame-mac`：

    egress-mac = keccak256.update(egress-mac, frame-ciphertext)
    frame-mac-seed = aes(mac-secret, keccak256.digest(egress-mac)[:16]) ^ keccak256.digest(egress-mac)[:16]
    egress-mac = keccak256.update(egress-mac, frame-mac-seed)
    frame-mac = keccak256.digest(egress-mac)[:16]

验证入站帧上的MAC是通过以与`egress-mac`相同的方式更新`ingress-mac`状态并与入站帧中的`header-mac`和`frame-mac`值进行比较来完成的。这应该在解密`header-ciphertext`和`frame-ciphertext`之前完成。

# 能力消息

初始握手后的所有消息都与“能力”相关联。任何数量的能力都可以在单个RLPx连接上同时使用。

能力由一个短的ASCII名称（最多八个字符）和版本号标识。连接双方支持的能力在属于“p2p”能力的[Hello]消息中交换，这是所有连接上必须可用的。

## 消息编码

初始[Hello]消息的编码如下：

    frame-data = msg-id || msg-data
    frame-size = frame-data的长度，编码为24位大端整数

其中`msg-id`是一个RLP编码的整数，用于识别消息，`msg-data`是一个包含消息数据的RLP列表。

Hello之后的所有消息都使用Snappy算法压缩。

    frame-data = msg-id || snappyCompress(msg-data)
    frame-size = frame-data的长度，编码为24位大端整数

请注意，压缩消息的`frame-size`是指`msg-data`的压缩大小。由于压缩消息在解压缩后可能会膨胀到非常大的大小，实现应在解码消息前检查数据的未压缩大小。这是可能的，因为[snappy格式]包含长度头。携带未压缩数据大于16 MiB的消息应通过关闭连接被拒绝。

## 基于消息ID的多路复用

虽然帧层支持`capability-id`，但当前版本的RLPx并未使用该字段在不同能力之间进行多路复用。相反，多路复用完全依赖于消息ID。

每个能力都被分配尽可能多的消息ID空间。所有这些能力都必须静态指定它们需要多少消息ID。在连接和接收[Hello]消息后，双方都具有

关于它们共享的能力（包括版本）的等效信息，并能够就消息ID空间的组成形成共识。

假定消息ID从ID 0x10开始是紧凑的（0x00-0x0f保留给“p2p”能力），并按字母顺序分配给每个共享的（版本相同，名称相同）能力。能力名称区分大小写。未共享的能力被忽略。如果共享了同一个（名称相同）能力的多个版本，则数值最高的获胜，其他的被忽略。

## “p2p”能力

“p2p”能力存在于所有连接上。在初始握手后，连接的双方必须发送[Hello]或[Disconnect]消息。接收到[Hello]消息后，会话激活，可以发送任何其他消息。实现必须忽略任何协议版本差异以实现向前兼容。与较低版本的对等方通信时，实现应尝试模仿该版本。

在协议协商后的任何时间，都可以发送[Disconnect]消息。

### Hello（0x00）

`[protocolVersion: P, clientId: B, capabilities, listenPort: P, nodeKey: B_64, ...]`

连接上发送的第一个数据包，双方都发送一次。在收到Hello之前，不得发送其他消息。实现必须忽略Hello中的任何额外列表元素，因为它们可能被未来版本使用。

- `protocolVersion`是“p2p”能力的版本，**5**。
- `clientId`指定客户端软件身份，为人类可读的字符串（例如
  "Ethereum(++)/1.0.0"）。
- `capabilities`是支持的能力及其版本的列表：
  `[[cap1, capVersion1], [cap2, capVersion2], ...]`。
- `listenPort`（遗留）指定客户端正在监听的端口（在当前连接的接口上）。如果为0，则表示客户端没有监听。应忽略此字段。
- `nodeId`是与节点的私钥对应的secp256k1公钥。

### Disconnect（0x01）

`[reason: P]`

通知对等方即将断开连接；如果收到，对等方应立即断开连接。发送时，行为良好的主机会给对等方一个斗争的机会（读：等待2秒）在断开连接之前断开连接自己。

`reason`是一个可选的整数，指定断开连接的多个原因之一：

| 原因   | 含义                                                      |
|--------|:-------------------------------------------------------------|
| `0x00` | 请求断开连接                                             |
| `0x01` | TCP子系统错误                                             |
| `0x02` | 协议违规，例如格式错误的消息，错误的RLP，...               |
| `0x03` | 无用的对等点                                               |
| `0x04` | 对等点过多                                                 |
| `0x05` | 已连接                                                    |
| `0x06` | 不兼容的P2P协议版本                                        |
| `0x07` | 收到空节点身份 - 这自动无效                                |
| `0x08` | 客户端退出                                                |
| `0x09` | 握手中出现意外身份                                         |
| `0x0a` | 身份与此节点相同（即连接到自己）                           |
| `0x0b` | Ping超时                                                  |
| `0x10` | 特定于子协议的其他原因                                      |

### Ping（0x02）

`[]`

请求对等方立即回复[Pong]。

### Pong（0x03）

`[]`

回复对等方的[Ping]数据包。

# 更改日志

### 当前版本的已知问题

- 帧加密/MAC方案被认为是“破解的”，因为`aes-secret

`和`mac-secret`用于读写两用。RLPx连接的两侧从相同的密钥、nonce和IV生成两个CTR流。如果攻击者知道一个明文，他们可以解密重用密钥流的未知明文。
- 审阅者的普遍反馈是，使用keccak256状态作为MAC累加器和在MAC算法中使用AES是一种不常见且过于复杂的消息认证方式，但可以认为是安全的。
- 帧编码提供了`capability-id`和`context-id`字段用于多路复用目的，但这些字段未被使用。

### 版本5（EIP-706，2017年9月）

[EIP-706]增加了Snappy消息压缩。

### 版本4（EIP-8，2015年12月）

[EIP-8]改变了初始握手中`auth-body`和`ack-body`的编码为RLP，增加了握手的版本号，并规定实现应忽略握手消息和[Hello]中的额外列表元素。

# 参考文献

- Elaine Barker, Don Johnson, and Miles Smid. NIST Special Publication 800-56A Section 5.8.1,
  Concatenation Key Derivation Function. 2017.\
  URL <https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-56ar.pdf>

- Victor Shoup. A proposal for an ISO standard for public key encryption, Version 2.1. 2001.\
  URL <http://www.shoup.net/papers/iso-2_1.pdf>

- Mike Belshe and Roberto Peon. SPDY Protocol - Draft 3. 2014.\
  URL <http://www.chromium.org/spdy/spdy-protocol/spdy-protocol-draft3>

- Snappy compressed format description. 2011.\
  URL <https://github.com/google/snappy/blob/master/format_description.txt>

版权所有 &copy; 2014 Alex Leverington.
<a rel="license" href="http://creativecommons.org/licenses/by-nc-sa/4.0/">
本作品采用
Creative Commons Attribution-NonCommercial-ShareAlike
4.0 International License授权</a>。

[Hello]: #hello-0x00
[Disconnect]: #disconnect-0x01
[Ping]: #ping-0x02
[Pong]: #pong-0x03
[Capability Messaging]: #capability-messaging
[EIP-8]: https://eips.ethereum.org/EIPS/eip-8
[EIP-706]: https://eips.ethereum.org/EIPS/eip-706
[RLP]: https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp
[snappy format]: https://github.com/google/snappy/blob/master/format_description.txt


# The RLPx Transport Protocol

This specification defines the RLPx transport protocol, a TCP-based transport protocol
used for communication among Ethereum nodes. The protocol carries encrypted messages
belonging to one or more 'capabilities' which are negotiated during connection
establishment. RLPx is named after the [RLP] serialization format. The name is not an
acronym and has no particular meaning.

The current protocol version is **5**. You can find a list of changes in past versions at
the end of this document.

## Notation

`X || Y`\
    denotes concatenation of X and Y.\
`X ^ Y`\
    is byte-wise XOR of X and Y.\
`X[:N]`\
    denotes an N-byte prefix of X.\
`[X, Y, Z, ...]`\
    denotes recursive encoding as an RLP list.\
`keccak256(MESSAGE)`\
    is the Keccak256 hash function as used by Ethereum.\
`ecies.encrypt(PUBKEY, MESSAGE, AUTHDATA)`\
    is the asymmetric authenticated encryption function as used by RLPx.\
    AUTHDATA is authenticated data which is not part of the resulting ciphertext,\
    but written to HMAC-256 before generating the message tag.\
`ecdh.agree(PRIVKEY, PUBKEY)`\
    is elliptic curve Diffie-Hellman key agreement between PRIVKEY and PUBKEY.

## ECIES Encryption

ECIES (Elliptic Curve Integrated Encryption Scheme) is an asymmetric encryption method
used in the RLPx handshake. The cryptosystem used by RLPx is

- The elliptic curve secp256k1 with generator `G`.
- `KDF(k, len)`: the NIST SP 800-56 Concatenation Key Derivation Function
- `MAC(k, m)`: HMAC using the SHA-256 hash function.
- `AES(k, iv, m)`: the AES-128 encryption function in CTR mode.

Alice wants to send an encrypted message that can be decrypted by Bobs static private key
<code>k<sub>B</sub></code>. Alice knows about Bobs static public key
<code>K<sub>B</sub></code>.

To encrypt the message `m`, Alice generates a random number `r` and corresponding elliptic
curve public key `R = r * G` and computes the shared secret <code>S = P<sub>x</sub></code>
where <code>(P<sub>x</sub>, P<sub>y</sub>) = r * K<sub>B</sub></code>. She derives key
material for encryption and authentication as
<code>k<sub>E</sub> || k<sub>M</sub> = KDF(S, 32)</code> as well as a random
initialization vector `iv`. Alice sends the encrypted message `R || iv || c || d` where
<code>c = AES(k<sub>E</sub>, iv , m)</code> and
<code>d = MAC(sha256(k<sub>M</sub>), iv || c)</code> to Bob.

For Bob to decrypt the message `R || iv || c || d`, he derives the shared secret
<code>S = P<sub>x</sub></code> where
<code>(P<sub>x</sub>, P<sub>y</sub>) = k<sub>B</sub> * R</code> as well as the encryption and
authentication keys <code>k<sub>E</sub> || k<sub>M</sub> = KDF(S, 32)</code>. Bob verifies
the authenticity of the message by checking whether
<code>d == MAC(sha256(k<sub>M</sub>), iv || c)</code> then obtains the plaintext as
<code>m = AES(k<sub>E</sub>, iv || c)</code>.

## Node Identity

All cryptographic operations are based on the secp256k1 elliptic curve. Each node is
expected to maintain a static secp256k1 private key which is saved and restored between
sessions. It is recommended that the private key can only be reset manually, for example,
by deleting a file or database entry.

## Initial Handshake

An RLPx connection is established by creating a TCP connection and agreeing on ephemeral
key material for further encrypted and authenticated communication. The process of
creating those session keys is the 'handshake' and is carried out between the 'initiator'
(the node which opened the TCP connection) and the 'recipient' (the node which accepted it).

1. initiator connects to recipient and sends its `auth` message
2. recipient accepts, decrypts and verifies `auth` (checks that recovery of signature ==
   `keccak256(ephemeral-pubk)`)
3. recipient generates `auth-ack` message from `remote-ephemeral-pubk` and `nonce`
4. recipient derives secrets and sends the first encrypted frame containing the [Hello] message
5. initiator receives `auth-ack` and derives secrets
6. initiator sends its first encrypted frame containing initiator [Hello] message
7. recipient receives and authenticates first encrypted frame
8. initiator receives and authenticates first encrypted frame
9. cryptographic handshake is complete if MAC of first encrypted frame is valid on both sides

Either side may disconnect if authentication of the first framed packet fails.

Handshake messages:

    auth = auth-size || enc-auth-body
    auth-size = size of enc-auth-body, encoded as a big-endian 16-bit integer
    auth-vsn = 4
    auth-body = [sig, initiator-pubk, initiator-nonce, auth-vsn, ...]
    enc-auth-body = ecies.encrypt(recipient-pubk, auth-body || auth-padding, auth-size)
    auth-padding = arbitrary data

    ack = ack-size || enc-ack-body
    ack-size = size of enc-ack-body, encoded as a big-endian 16-bit integer
    ack-vsn = 4
    ack-body = [recipient-ephemeral-pubk, recipient-nonce, ack-vsn, ...]
    enc-ack-body = ecies.encrypt(initiator-pubk, ack-body || ack-padding, ack-size)
    ack-padding = arbitrary data

Implementations must ignore any mismatches in `auth-vsn` and `ack-vsn`. Implementations
must also ignore any additional list elements in `auth-body` and `ack-body`.

Secrets generated following the exchange of handshake messages:

    static-shared-secret = ecdh.agree(privkey, remote-pubk)
    ephemeral-key = ecdh.agree(ephemeral-privkey, remote-ephemeral-pubk)
    shared-secret = keccak256(ephemeral-key || keccak256(nonce || initiator-nonce))
    aes-secret = keccak256(ephemeral-key || shared-secret)
    mac-secret = keccak256(ephemeral-key || aes-secret)

## Framing

All messages following the initial handshake are framed. A frame carries a single
encrypted message belonging to a capability.

The purpose of framing is multiplexing multiple capabilities over a single connection.
Secondarily, as framed messages yield reasonable demarcation points for message
authentication codes, supporting an encrypted and authenticated stream becomes
straight-forward. Frames are encrypted and authenticated via key material generated during
the handshake.

The frame header provides information about the size of the message and the message's
source capability. Padding is used to prevent buffer starvation, such that frame
components are byte-aligned to block size of cipher.

    frame = header-ciphertext || header-mac || frame-ciphertext || frame-mac
    header-ciphertext = aes(aes-secret, header)
    header = frame-size || header-data || header-padding
    header-data = [capability-id, context-id]
    capability-id = integer, always zero
    context-id = integer, always zero
    header-padding = zero-fill header to 16-byte boundary
    frame-ciphertext = aes(aes-secret, frame-data || frame-padding)
    frame-padding = zero-fill frame-data to 16-byte boundary

See the [Capability Messaging] section for definitions of `frame-data` and `frame-size.`

### MAC

Message authentication in RLPx uses two keccak256 states, one for each direction of
communication. The `egress-mac` and `ingress-mac` keccak states are continuously updated
with the ciphertext of bytes sent (egress) or received (ingress). Following the initial
handshake, the MAC states are initialized as follows:

Initiator:

    egress-mac = keccak256.init((mac-secret ^ recipient-nonce) || auth)
    ingress-mac = keccak256.init((mac-secret ^ initiator-nonce) || ack)

Recipient:

    egress-mac = keccak256.init((mac-secret ^ initiator-nonce) || ack)
    ingress-mac = keccak256.init((mac-secret ^ recipient-nonce) || auth)

When a frame is sent, the corresponding MAC values are computed by updating the
`egress-mac` state with the data to be sent. The update is performed by XORing the header
with the encrypted output of its corresponding MAC. This is done to ensure uniform
operations are performed for both plaintext MAC and ciphertext. All MACs are sent
cleartext.

    header-mac-seed = aes(mac-secret, keccak256.digest(egress-mac)[:16]) ^ header-ciphertext
    egress-mac = keccak256.update(egress-mac, header-mac-seed)
    header-mac = keccak256.digest(egress-mac)[:16]

Computing `frame-mac`:

    egress-mac = keccak256.update(egress-mac, frame-ciphertext)
    frame-mac-seed = aes(mac-secret, keccak256.digest(egress-mac)[:16]) ^ keccak256.digest(egress-mac)[:16]
    egress-mac = keccak256.update(egress-mac, frame-mac-seed)
    frame-mac = keccak256.digest(egress-mac)[:16]

Verifying the MAC on ingress frames is done by updating the `ingress-mac` state in the
same way as `egress-mac` and comparing to the values of `header-mac` and `frame-mac` in
the ingress frame. This should be done before decrypting `header-ciphertext` and
`frame-ciphertext`.

# Capability Messaging

All messages following the initial handshake are associated with a 'capability'. Any
number of capabilities can be used concurrently on a single RLPx connection.

A capability is identified by a short ASCII name (max eight characters) and version number. The capabilities
supported on either side of the connection are exchanged in the [Hello] message belonging
to the 'p2p' capability which is required to be available on all connections.

## Message Encoding

The initial [Hello] message is encoded as follows:

    frame-data = msg-id || msg-data
    frame-size = length of frame-data, encoded as a 24bit big-endian integer

where `msg-id` is an RLP-encoded integer identifying the message and `msg-data` is an RLP
list containing the message data.

All messages following Hello are compressed using the Snappy algorithm.

    frame-data = msg-id || snappyCompress(msg-data)
    frame-size = length of frame-data encoded as a 24bit big-endian integer

Note that the `frame-size` of compressed messages refers to the compressed size of
`msg-data`. Since compressed messages may inflate to a very large size after
decompression, implementations should check for the uncompressed size of the data before
decoding the message. This is possible because the [snappy format] contains a length
header. Messages carrying uncompressed data larger than 16 MiB should be rejected by
closing the connection.

## Message ID-based Multiplexing

While the framing layer supports a `capability-id`, the current version of RLPx doesn't
use that field for multiplexing between different capabilities. Instead, multiplexing
relies purely on the message ID.

Each capability is given as much of the message-ID space as it needs. All such
capabilities must statically specify how many message IDs they require. On connection and
reception of the [Hello] message, both peers have equivalent information about what
capabilities they share (including versions) and are able to form consensus over the
composition of message ID space.

Message IDs are assumed to be compact from ID 0x10 onwards (0x00-0x0f is reserved for the
"p2p" capability) and given to each shared (equal-version, equal-name) capability in
alphabetic order. Capability names are case-sensitive. Capabilities which are not shared
are ignored. If multiple versions are shared of the same (equal name) capability, the
numerically highest wins, others are ignored.

## "p2p" Capability

The "p2p" capability is present on all connections. After the initial handshake, both
sides of the connection must send either [Hello] or a [Disconnect] message. Upon receiving
the [Hello] message a session is active and any other message may be sent. Implementations
must ignore any difference in protocol version for forward-compatibility reasons. When
communicating with a peer of lower version, implementations should try to mimic that
version.

At any time after protocol negotiation, a [Disconnect] message may be sent.

### Hello (0x00)

`[protocolVersion: P, clientId: B, capabilities, listenPort: P, nodeKey: B_64, ...]`

First packet sent over the connection, and sent once by both sides. No other messages may
be sent until a Hello is received. Implementations must ignore any additional list elements
in Hello because they may be used by a future version.

- `protocolVersion` the version of the "p2p" capability, **5**.
- `clientId` Specifies the client software identity, as a human-readable string (e.g.
  "Ethereum(++)/1.0.0").
- `capabilities` is the list of supported capabilities and their versions:
  `[[cap1, capVersion1], [cap2, capVersion2], ...]`.
- `listenPort` (legacy) specifies the port that the client is listening on (on the
  interface that the present connection traverses). If 0 it indicates the client is
  not listening. This field should be ignored.
- `nodeId` is the secp256k1 public key corresponding to the node's private key.

### Disconnect (0x01)

`[reason: P]`

Inform the peer that a disconnection is imminent; if received, a peer should disconnect
immediately. When sending, well-behaved hosts give their peers a fighting chance (read:
wait 2 seconds) to disconnect to before disconnecting themselves.

`reason` is an optional integer specifying one of a number of reasons for disconnect:

| Reason | Meaning                                                      |
|--------|:-------------------------------------------------------------|
| `0x00` | Disconnect requested                                         |
| `0x01` | TCP sub-system error                                         |
| `0x02` | Breach of protocol, e.g. a malformed message, bad RLP, ...   |
| `0x03` | Useless peer                                                 |
| `0x04` | Too many peers                                               |
| `0x05` | Already connected                                            |
| `0x06` | Incompatible P2P protocol version                            |
| `0x07` | Null node identity received - this is automatically invalid  |
| `0x08` | Client quitting                                              |
| `0x09` | Unexpected identity in handshake                             |
| `0x0a` | Identity is the same as this node (i.e. connected to itself) |
| `0x0b` | Ping timeout                                                 |
| `0x10` | Some other reason specific to a subprotocol                  |

### Ping (0x02)

`[]`

Requests an immediate reply of [Pong] from the peer.

### Pong (0x03)

`[]`

Reply to the peer's [Ping] packet.

# Change Log

### Known Issues in the current version

- The frame encryption/MAC scheme is considered 'broken' because `aes-secret` and
  `mac-secret` are reused for both reading and writing. The two sides of a RLPx connection
  generate two CTR streams from the same key, nonce and IV. If an attacker knows one
  plaintext, they can decrypt unknown plaintexts of the reused keystream.
- General feedback from reviewers has been that the use of a keccak256 state as a MAC
  accumulator and the use of AES in the MAC algorithm is an uncommon and overly complex
  way to perform message authentication but can be considered safe.
- The frame encoding provides `capability-id` and `context-id` fields for multiplexing
  purposes, but these fields are unused.

### Version 5 (EIP-706, September 2017)

[EIP-706] added Snappy message compression.

### Version 4 (EIP-8, December 2015)

[EIP-8] changed the encoding of `auth-body` and `ack-body` in the initial handshake to
RLP, added a version number to the handshake and mandated that implementations should
ignore additional list elements in handshake messages and [Hello].

# References

- Elaine Barker, Don Johnson, and Miles Smid. NIST Special Publication 800-56A Section 5.8.1,
  Concatenation Key Derivation Function. 2017.\
  URL <https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-56ar.pdf>

- Victor Shoup. A proposal for an ISO standard for public key encryption, Version 2.1. 2001.\
  URL <http://www.shoup.net/papers/iso-2_1.pdf>

- Mike Belshe and Roberto Peon. SPDY Protocol - Draft 3. 2014.\
  URL <http://www.chromium.org/spdy/spdy-protocol/spdy-protocol-draft3>

- Snappy compressed format description. 2011.\
  URL <https://github.com/google/snappy/blob/master/format_description.txt>

Copyright &copy; 2014 Alex Leverington.
<a rel="license" href="http://creativecommons.org/licenses/by-nc-sa/4.0/">
This work is licensed under a
Creative Commons Attribution-NonCommercial-ShareAlike
4.0 International License</a>.

[Hello]: #hello-0x00
[Disconnect]: #disconnect-0x01
[Ping]: #ping-0x02
[Pong]: #pong-0x03
[Capability Messaging]: #capability-messaging
[EIP-8]: https://eips.ethereum.org/EIPS/eip-8
[EIP-706]: https://eips.ethereum.org/EIPS/eip-706
[RLP]: https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp
[snappy format]: https://github.com/google/snappy/blob/master/format_description.txt