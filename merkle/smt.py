import hashlib
from merkle.proof import Proof
from common import util
from functools import reduce

class SparseMerkleTree(object):

	"""docstring for SparseMerkleTree"""
	def __init__(self, hash_name: str) -> None:
		super(SparseMerkleTree, self).__init__()

		self.root_digest = None

		self.hash_name = hash_name

		# We need to initialize the hash function to determine the digest_size
		H = eval("hashlib.{}()".format(hash_name))
		self.depth = 8 * H.digest_size

		""" 
		The Key space can be partitioned into the set of keys in the cache
		and the set of keys that are empty and in the empty_cache.
		"""
		self.cache = {}
		self.empty_cache = [self._hash(b'0')]

	"""
	https://www.links.org/files/RevocationTransparency.pdf
	"""
	def _empty_cache(self, n: int):
		if len(self.empty_cache) <= n:
			t = self._empty_cache(n - 1)
			t = self._hash(t + t)
			assert len(self.empty_cache) == n
			self.empty_cache.append(t)
		return self.empty_cache[n]	

	def _hash(self, data) -> str:
		# H is an object of type hashlib.hash_name

		H = eval("hashlib.{}()".format(self.hash_name))
		H.update(util.to_bytes(data))
		return H.hexdigest()

	def _parent(self, node_id: util.bitarray):
		# We create a copy in order to avoid destructively modifying node_id
		p_id = node_id.copy()

		# The parent node_id is always the first n - 1 bits of the input node_id
		p_id.pop()
		return p_id

	def _sibling(self, node_id: util.bitarray):
		# We create a copy in order to avoid destructively modifying node_id
		s_id = node_id.copy()

		# Here, we xor the node_id with a bitarray of integer value of 1
		s_id ^= util.bitarray(([0] * (s_id.length() - 1)) + [1]) # FIXME: This list comprehension is probably slow and memory intensive
		return s_id

	def _get_empty_ancestor(index: util.bitarray) -> util.bitarray:
		

	"""
	Assume index is valid (for now).

	FIXME: What are the error cases where we return False?
	"""
	def insert(self, index: str, data: str) -> bool:
		# Do the first level hash of data and insert into index-th leaf
		node_id = util.bitarray(index)
		print("insert: node_id = {}".format(node_id))
		self.cache[node_id] = self._hash(data)

		# Do a normal update up the tree
		curr_id = node_id.copy()
		while (curr_id.length() > 0):
			# Get both the parent and sibling ids
			s_id = self._sibling(curr_id)
			p_id = self._parent(curr_id)

			# Get the digest of the current node and sibling
			curr_digest = self.cache[curr_id]
			s_digest = None
			if s_id in self.cache:
				s_digest = self.cache[s_id]
			else:
				s_digest = self._empty_cache(s_id.length())

			# Hash the digests of the left and right children
			p_digest = self._hash(curr_digest + s_digest)
			self.cache[p_id] = p_digest

			# Traverse up the tree by making the current node the parent node
			curr_id = p_id

		# Update root
		self.root_digest = self.cache[curr_id]
		return True

	def generate_copath(self, index: str) -> list:
		membership_proof = index in self.cache
		copath = list()
		curr_id = util.bitarray(index)

		if not membership_proof:
			curr_id = self._get_empty_ancestor(index)

		# Our stopping condition is length > 0 so we don't add the root to the copath
		while (curr_id.length() > 0):
			# Append the sibling to the copath and advance current node
			s_id = self._sibling(curr_id)
			s_digest = None

			# Check to see if sibling is cacehd otherwise set to empty value of appropriate length
			if s_id in self.cache:
				s_digest = self.cache[s_id]
			else:
				s_digest = self._empty_cache(len(s_id))

			copath.append((s_id, s_digest))
			curr_id = self._parent(curr_id)

		return Proof(membership_proof, index, copath)

	def verify_path(self, proof: Proof) -> bool:
		# Convert index into a bitarray and check if its in the cache
		node_id = util.bitarray(proof.index)
		# print("verify_path: node_id = {}".format(node_id))
		node_digest = self.cache[node_id]

		# Reducing the copath models the hash invariant of the root
		root_digest = node_digest
		for i in range(len(proof.copath)): 
			root_digest = self._hash(root_digest + proof.copath[i][1])

		return root_digest == self.root_digest
