# DNS 节点列表

点对点节点软件通常包含硬编码的引导节点列表。更新这些列表需要软件更新，软件维护者需要付出努力确保列表是最新的。因此，提供的列表通常很小，给软件很少的选择进入网络的初始入口点。

这份规范描述了一种通过 DNS 检索的经过身份验证、可更新节点列表的方案。要使用这样的列表，客户端只需要知道 DNS 名称和签署列表的公钥信息。

基于 DNS 的发现最初在 [EIP-1459] 中提出。

## 节点列表

“节点列表”是一个包含任意长度的['节点记录'（ENRs）](./enr.md)的列表。列表可能通过使用链接引用其他列表。整个列表使用 secp256k1 私钥签名。为了验证列表，客户端必须知道相应的公钥。

## URL 方案

要引用 DNS 节点列表，客户端使用带有 'enrtree' 方案的 URL。URL 包含列表所在的 DNS 名称以及签名列表的公钥。公钥包含在 URL 的用户名部分，并采用压缩的 32 字节二进制公钥的 base32 编码。

示例：

    enrtree://AM5FCQLWIZX2QFPNJAP7VUERCCRNGRHWZG3YYHIUV7BVDQ5FDPRT2@nodes.example.org

此 URL 指向位于 DNS 名称 'nodes.example.org' 的节点列表，并由公钥

    0x049f88229042fef9200246f49f94d9b77c4e954721442714e85850cb6d9e5daf2d880ea0e53cb3ac1a75f9923c2726a4f941f7d326781baa6380754a360de5c2b6

签名。

## DNS 记录结构

列表中的节点被编码为通过 DNS 协议分发的默克尔树。默克尔树的条目包含在 DNS TXT 记录中。树的根是一个 TXT 记录，内容如下：

    enrtree-root:v1 e=<enr-root> l=<link-root> seq=<sequence-number> sig=<signature>

其中：

- `enr-root` 和 `link-root` 指向包含节点和链接子树的根哈希。
- `sequence-number` 是树的更新序列号，一个十进制整数。
- `signature` 是对记录内容的 keccak256 哈希进行的 65 字节 secp256k1 EC 签名，排除了 `sig=` 部分，编码为 URL 安全的 base64。

在子域上的进一步 TXT 记录将哈希映射到三种条目类型之一。任何条目的子域名称是其文本内容的（缩写的）keccak256 哈希的 base32 编码。

- `enrtree-branch:<h₁>,<h₂>,...,<hₙ>` 是一个中间树条目，包含子树条目的哈希。
- `enrtree://<key>@<fqdn>` 是一个叶子，指向位于另一个完全合格域名的不同列表。注意这种格式与 URL 编码匹配。此类型的条目只能出现在 `link-root` 指向的子树中。
- `enr:<node-record>` 是一个包含节点记录的叶子。节点记录编码为 URL 安全的 base64 字符串。注意，这种类型的条目与规范的 ENR 文本编码匹配。它只能出现在 `enr-root` 子树中。

树没有定义特定的排序或结构。每当树更新时，其序列号应该增加。任何 TXT 记录的内容应足够小，以适应 UDP DNS 数据包强加的 512 字节限制。这限

制了可以放入 `enrtree-branch` 条目的哈希数量。

区域文件格式示例：

    ; name                        ttl     class type  content
    @                             60      IN    TXT   enrtree-root:v1 e=JWXYDBPXYWG6FX3GMDIBFA6CJ4 l=C7HRFPF3BLGF3YR4DY5KX3SMBE seq=1 sig=o908WmNp7LibOfPsr4btQwatZJ5URBr2ZAuxvK4UWHlsB9sUOTJQaGAlLPVAhM__XJesCHxLISo94z5Z2a463gA
    C7HRFPF3BLGF3YR4DY5KX3SMBE    86900   IN    TXT   enrtree://AM5FCQLWIZX2QFPNJAP7VUERCCRNGRHWZG3YYHIUV7BVDQ5FDPRT2@morenodes.example.org
    JWXYDBPXYWG6FX3GMDIBFA6CJ4    86900   IN    TXT   enrtree-branch:2XS2367YHAXJFGLZHVAWLQD4ZY,H4FHT4B454P6UXFD7JCYQ5PWDY,MHTDO6TMUBRIA2XWG5LUDACK24
    2XS2367YHAXJFGLZHVAWLQD4ZY    86900   IN    TXT   enr:-HW4QOFzoVLaFJnNhbgMoDXPnOvcdVuj7pDpqRvh6BRDO68aVi5ZcjB3vzQRZH2IcLBGHzo8uUN3snqmgTiE56CH3AMBgmlkgnY0iXNlY3AyNTZrMaECC2_24YYkYHEgdzxlSNKQEnHhuNAbNlMlWJxrJxbAFvA
    H4FHT4B454P6UXFD7JCYQ5PWDY    86900   IN    TXT   enr:-HW4QAggRauloj2SDLtIHN1XBkvhFZ1vtf1raYQp9TBW2RD5EEawDzbtSmlXUfnaHcvwOizhVYLtr7e6vw7NAf6mTuoCgmlkgnY0iXNlY3AyNTZrMaECjrXI8TLNXU0f8cthpAMxEshUyQlK-AM0PW2wfrnacNI
    MHTDO6TMUBRIA2XWG5LUDACK24    86900   IN    TXT   enr:-HW4QLAYqmrwllBEnzWWs7I5Ev2IAs7x_dZlbYdRdMUx5EyKHDXp7AV5CkuPGUPdvbv1_Ms1CPfhcGCvSElSosZmyoqAgmlkgnY0iXNlY3AyNTZrMaECriawHKWdDRk2xeZkrOXBQ0dfMFLHY4eENZwdufn1S1o

## 客户端协议

要在给定的 DNS 名称处找到节点，比如说 "mynodes.org"：

1. 解析该名称的 TXT 记录并检查是否包含有效的 "enrtree-root=v1" 条目。假设条目中包含的 `enr-root` 哈希是 "CFZUWDU7JNQR4VTCZVOJZ5ROV4"。
2. 验证根签名是否与已知公钥匹配，并检查序列号是否大于或等于该名称先前看到的任何数字。
3. 解析哈希子域的 TXT 记录，例如 "CFZUWDU7JNQR4VTCZVOJZ5ROV4.mynodes.org"，并验证内容是否与哈希匹配。
4. 下一步取决于找到的条目类型：
   - 对于 `enrtree-branch`：解析哈希列表并继续解析它们（步骤 3）。
   - 对于 `enr`：解码，验证节点记录并将其导入本地节点

存储。

在遍历过程中，客户端必须跟踪已解析的哈希和域，以避免进入无限循环。客户端最好随机遍历树。

客户端实现应避免在正常操作期间一次性下载整个树。最好在需要时通过 DNS 请求条目，即当客户端正在寻找对等方时。

## 基本原理

使用 DNS 是因为它是一个低延迟协议，几乎可以保证可用。

作为默克尔树，任何节点列表都可以通过对根的单一签名进行验证。哈希子域保护列表的完整性。在最坏的情况下，中间解析器可以阻止访问列表或不允许更新它，但不能破坏其内容。序列号防止用旧版本替换根。

在客户端同步更新可以增量进行，这对于大型列表很重要。树的单个条目足够小，可以适应单个 UDP 数据包，确保与只能使用基本 UDP DNS 的环境兼容。树格式也适合使用缓存解析器：只有树的根需要短 TTL。中间条目和叶子可以缓存数天。

### 为什么存在链接子树？

列表之间的链接启用了联盟和信任网络功能。大列表的操作员可以将维护委托给其他列表提供商。如果两个节点列表互相链接，用户可以使用任何一个列表并从两者中获取节点。

链接子树与包含 ENRs 的树分开，这样做是为了使客户端实现能够独立同步这些树。一个想要尽可能多获取节点的客户端将首先同步链接树，然后将所有链接的名称添加到同步视野中。

## 参考文献

1. 用于表示二进制数据的 base64 和 base32 编码在 [RFC 4648] 中定义。对于 base64 和 base32 数据，不使用填充。

[EIP-1459]: https://eips.ethereum.org/EIPS/eip-1459
[RFC 4648]: https://tools.ietf.org/html/rfc4648

# DNS Node Lists

Peer-to-peer node software often contains hard-coded bootstrap node lists. Updating those
lists requires a software update and, effort is required from the software maintainers to
ensure the list is up-to-date. As a result, the provided lists are usually small, giving
the software little choice of initial entry point into the network.

This specification describes a scheme for authenticated, updateable node lists retrievable
via DNS. In order to use such a list, the client only requires information about the DNS
name and the public key that signs the list.

DNS-based discovery was initially proposed in [EIP-1459].

## Node Lists

A 'node list' is a list of ['node records' (ENRs)](./enr.md) of arbitrary length. Lists
may refer to other lists using links. The entire list is signed using a secp256k1 private
key. The corresponding public key must be known to the client in order to verify the list.

## URL Scheme

To refer to a DNS node list, clients use a URL with 'enrtree' scheme. The URL contains the
DNS name on which the list can be found as well as the public key that signed the list.
The public key is contained in the username part of the URL and is the base32 encoding of
the compressed 32-byte binary public key.

Example:

    enrtree://AM5FCQLWIZX2QFPNJAP7VUERCCRNGRHWZG3YYHIUV7BVDQ5FDPRT2@nodes.example.org

This URL refers to a node list at the DNS name 'nodes.example.org' and is signed by the
public key

    0x049f88229042fef9200246f49f94d9b77c4e954721442714e85850cb6d9e5daf2d880ea0e53cb3ac1a75f9923c2726a4f941f7d326781baa6380754a360de5c2b6

## DNS Record Structure

The nodes in a list are encoded as a merkle tree for distribution via the DNS protocol.
Entries of the merkle tree are contained in DNS TXT records. The root of the tree is a TXT
record with the following content:

    enrtree-root:v1 e=<enr-root> l=<link-root> seq=<sequence-number> sig=<signature>

where

- `enr-root` and `link-root` refer to the root hashes of subtrees containing nodes and
  links subtrees.
- `sequence-number` is the tree's update sequence number, a decimal integer.
- `signature` is a 65-byte secp256k1 EC signature over the keccak256 hash of the record
  content, excluding the `sig=` part, encoded as URL-safe base64.

Further TXT records on subdomains map hashes to one of three entry types. The subdomain
name of any entry is the base32 encoding of the (abbreviated) keccak256 hash of its text
content.

- `enrtree-branch:<h₁>,<h₂>,...,<hₙ>` is an intermediate tree entry containing hashes of
  subtree entries.
- `enrtree://<key>@<fqdn>` is a leaf pointing to a different list located at another fully
  qualified domain name. Note that this format matches the URL encoding. This type of
  entry may only appear in the subtree pointed to by `link-root`.
- `enr:<node-record>` is a leaf containing a node record. The node record is encoded as a
  URL-safe base64 string. Note that this type of entry matches the canonical ENR text
  encoding. It may only appear in the `enr-root` subtree.

No particular ordering or structure is defined for the tree. Whenever the tree is updated,
its sequence number should increase. The content of any TXT record should be small enough
to fit into the 512 byte limit imposed on UDP DNS packets. This limits the number of
hashes that can be placed into an `enrtree-branch` entry.

Example in zone file format:

    ; name                        ttl     class type  content
    @                             60      IN    TXT   enrtree-root:v1 e=JWXYDBPXYWG6FX3GMDIBFA6CJ4 l=C7HRFPF3BLGF3YR4DY5KX3SMBE seq=1 sig=o908WmNp7LibOfPsr4btQwatZJ5URBr2ZAuxvK4UWHlsB9sUOTJQaGAlLPVAhM__XJesCHxLISo94z5Z2a463gA
    C7HRFPF3BLGF3YR4DY5KX3SMBE    86900   IN    TXT   enrtree://AM5FCQLWIZX2QFPNJAP7VUERCCRNGRHWZG3YYHIUV7BVDQ5FDPRT2@morenodes.example.org
    JWXYDBPXYWG6FX3GMDIBFA6CJ4    86900   IN    TXT   enrtree-branch:2XS2367YHAXJFGLZHVAWLQD4ZY,H4FHT4B454P6UXFD7JCYQ5PWDY,MHTDO6TMUBRIA2XWG5LUDACK24
    2XS2367YHAXJFGLZHVAWLQD4ZY    86900   IN    TXT   enr:-HW4QOFzoVLaFJnNhbgMoDXPnOvcdVuj7pDpqRvh6BRDO68aVi5ZcjB3vzQRZH2IcLBGHzo8uUN3snqmgTiE56CH3AMBgmlkgnY0iXNlY3AyNTZrMaECC2_24YYkYHEgdzxlSNKQEnHhuNAbNlMlWJxrJxbAFvA
    H4FHT4B454P6UXFD7JCYQ5PWDY    86900   IN    TXT   enr:-HW4QAggRauloj2SDLtIHN1XBkvhFZ1vtf1raYQp9TBW2RD5EEawDzbtSmlXUfnaHcvwOizhVYLtr7e6vw7NAf6mTuoCgmlkgnY0iXNlY3AyNTZrMaECjrXI8TLNXU0f8cthpAMxEshUyQlK-AM0PW2wfrnacNI
    MHTDO6TMUBRIA2XWG5LUDACK24    86900   IN    TXT   enr:-HW4QLAYqmrwllBEnzWWs7I5Ev2IAs7x_dZlbYdRdMUx5EyKHDXp7AV5CkuPGUPdvbv1_Ms1CPfhcGCvSElSosZmyoqAgmlkgnY0iXNlY3AyNTZrMaECriawHKWdDRk2xeZkrOXBQ0dfMFLHY4eENZwdufn1S1o

## Client Protocol

To find nodes at a given DNS name, say "mynodes.org":

1. Resolve the TXT record of the name and check whether it contains a valid
   "enrtree-root=v1" entry. Let's say the `enr-root` hash contained in the entry is
   "CFZUWDU7JNQR4VTCZVOJZ5ROV4".
2. Verify the signature on the root against the known public key and check whether the
   sequence number is larger than or equal to any previous number seen for that name.
3. Resolve the TXT record of the hash subdomain, e.g.
   "CFZUWDU7JNQR4VTCZVOJZ5ROV4.mynodes.org" and verify whether the content matches the
   hash.
4. The next step depends on the entry type found:
   - for `enrtree-branch`: parse the list of hashes and continue resolving them (step 3).
   - for `enr`: decode, verify the node record and import it to local node storage.

During traversal, the client must track hashes and domains which are already resolved to
avoid going into an infinite loop. It's in the client's best interest to traverse the tree
in random order.

Client implementations should avoid downloading the entire tree at once during normal
operation. It's much better to request entries via DNS when-needed, i.e. at the time when
the client is looking for peers.

## Rationale

DNS is used because it is a low-latency protocol that is pretty much guaranteed to be
available.

Being a merkle tree, any node list can be authenticated by a single signature on the root.
Hash subdomains protect the integrity of the list. At worst intermediate resolvers can
block access to the list or disallow updates to it, but cannot corrupt its content. The
sequence number prevents replacing the root with an older version.

Synchronizing updates on the client side can be done incrementally, which matters for
large lists. Individual entries of the tree are small enough to fit into a single UDP
packet, ensuring compatibility with environments where only basic UDP DNS can be used. The
tree format also works well with caching resolvers: only the root of the tree needs a
short TTL. Intermediate entries and leaves can be cached for days.

### Why does the link subtree exist?

Links between lists enable federation and web-of-trust functionality. The operator of a
large list can delegate maintenance to other list providers. If two node lists link to
each other, users can use either list and get nodes from both.

The link subtree is separate from the tree containing ENRs. This is done to enable client
implementations to sync these trees independently. A client wanting to get as many nodes
as possible will sync the link tree first and add all linked names to the sync horizon.

## References

1. The base64 and base32 encodings used to represent binary data are defined in [RFC
   4648]. No padding is used for base64 and base32 data.

[EIP-1459]: https://eips.ethereum.org/EIPS/eip-1459
[RFC 4648]: https://tools.ietf.org/html/rfc4648