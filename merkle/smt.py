from angela.common import util
import hashlib

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

	def _hash(self, data: str) -> str:
		# H is an object of type hashlib.hash_name
		H = eval("hashlib.{}({})".format(self.hash_name, data))
		return H.hexdigest()

	def parent(self, node_id: util.bitarray):
		# We create a copy in order to avoid destructively modifying node_id
		p_id = node_id.copy()

		# The parent node_id is always the first n - 1 bits of the input node_id
		p_id.pop()
		return p_id

	def sibling(self, node_id: util.bitarray):
		# We create a copy in order to avoid destructively modifying node_id
		s_id = node_id.copy()

		# Here, we xor the node_id with a bitarray of integer value of 1
		s_id ^= util.bitarray(([0] * (s_id.length() - 1)) + [1]) # FIXME: This list comprehension is probably slow and memory intensive
		return s_id

	"""
	Assume index is valid (for now).

	FIXME: What are the error cases where we return False?
	"""
	def insert(self, index: str, data: str) -> bool:
		# Do the first level hash of data and insert into index-th leaf
		node_id = util.bitarray(index)
		cache[node_id] = self._hash(data)

		# Do a normal update up the tree
		curr_id = node_id.copy()
		while (curr_id.length > 0):
			# Get both the parent and sibling ids
			s_id = self.sibling(curr_id)
			p_id = self.parent(curr_id)

			# Get the digest of the current node and sibling
			curr_digest = self.cache[curr_id]
			s_digest = None
			if s_id in self.cache:
				s_digest = self.cache[s_id]
			else:
				s_digest = self._empty_cache(s_id.length())

			p_digest = self._hash(curr_digest + s_digest)
			self.cache[p_id] = p_digest

			curr_id = p_id

		# Update root
		self.root_digest = self.cache[curr_id]
		return True

	def generate_copath(self, index: str) -> list:
		copath = list()
		curr_id = bitarray(index)

		while (curr_id.length > 0):
			# Get both the parent and sibling ids
			s_id = self.sibling(curr_id)
			copath.append(s_id)
			curr_id = self.parent(curr_id)

		return copath

	def verify_path(self, index, copath: list) -> bool:
		copath.insert(0, index)
		root_digest = reduce(lambda x, y: self._hash(x + y), copath)
		return root_digest == self.root_digest