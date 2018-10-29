import hashlib
from merkle.proof import Proof
from common import util
from functools import reduce

class SparseMerkleTree(object):

	"""docstring for SparseMerkleTree"""
	def __init__(self, hash_name: str) -> None:
		super(SparseMerkleTree, self).__init__()

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
		self.root_digest = self._empty_cache(self.depth)

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
		if node_id.length() == 0:
			return node_id, False

		# We create a copy in order to avoid destructively modifying node_id
		s_id = node_id.copy()

		# # Here, we xor the node_id with a bitarray of integer value of 1
		s_id ^= util.bitarray(([0] * (s_id.length() - 1)) + [1]) # FIXME: This list comprehension is probably slow and memory intensive
		is_left = s_id.copy().pop() == 0
		return s_id, is_left

	def _get_empty_ancestor(self, index: util.bitarray) -> util.bitarray:
		prev_id = curr_id = index.copy()
		while curr_id.length() > 0:
			curr_id.pop()
			prev_id = curr_id
		return prev_id

	"""
	Assume index is valid (for now).

	FIXME: What are the error cases where we return False?
	"""
	def insert(self, index: str, data: str) -> bool:
		# Do the first level hash of data and insert into index-th leaf
		node_id = util.bitarray(index)
		self.cache[node_id] = self._hash(data)

		# Do a normal update up the tree
		curr_id = node_id.copy()
		while (curr_id.length() > 0):
			# Get both the parent and sibling ids
			s_id, is_left = self._sibling(curr_id)
			p_id = self._parent(curr_id)

			# Get the digest of the current node and sibling
			curr_digest = self.cache[curr_id]
			s_digest = None
			if s_id in self.cache:
				s_digest = self.cache[s_id]
			else:
				s_digest = self._empty_cache(s_id.length())

			# Hash the digests of the left and right children
			if is_left:
				p_digest = self._hash(s_digest + curr_digest)
			else:
				p_digest = self._hash(curr_digest + s_digest)
			self.cache[p_id] = p_digest

			# Traverse up the tree by making the current node the parent node
			curr_id = p_id

		# Update root
		self.root_digest = self.cache[curr_id]
		return True

	def generate_proof(self, index: str) -> list:
		copath = list()
		curr_id = util.bitarray(index)
		proof = Proof(index=curr_id)
		proof.proof_type = curr_id in self.cache

		if not proof.proof_type:
			curr_id = self._get_empty_ancestor(curr_id)
		proof.proof_id = curr_id

		# Our stopping condition is length > 0 so we don't add the root to the copath
		while (curr_id.length() > 0):
			# Append the sibling to the copath and advance current node
			s_id, is_left = self._sibling(curr_id)
			s_digest = None

			# Check to see if sibling is cacehd otherwise set to empty value of appropriate length
			if s_id in self.cache:
				s_digest = self.cache[s_id]
			else:
				s_digest = self._empty_cache(len(s_id))

			copath.append((s_id, s_digest))
			curr_id = self._parent(curr_id)

		proof.copath = copath
		return proof

	# Note: copath is destructively modified
	def verify_proof(self, proof: Proof) -> bool:
		proof_id_length = proof.proof_id.length()
		if proof.proof_type == False:
			if proof_id_length > len(proof.index):
				return False
			for i in range(proof_id_length):
				if proof.proof_id[i] != proof.index[i]:
					return False

		root_digest = self.cache.get(proof.proof_id, self._empty_cache(self.depth - proof_id_length))
		for i in range(len(proof.copath)): 
			if proof.copath[i][0][-1] == 0:
				root_digest = self._hash(proof.copath[i][1] + root_digest)
			else:
				root_digest = self._hash(root_digest + proof.copath[i][1])
		return root_digest == self.root_digest
