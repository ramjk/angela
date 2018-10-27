import hashlib
import unittest
from angela.merkle.smt import SparseMerkleTree

class TestSparseMerkleTree(unittest.TestCase):
	def test_constructor(self):
		T = SparseMerkleTree("sha256")
		self.assertEquals(T.hash_name, "sha256")
		self.assertEquals(T.depth, 256)
		self.assertEquals(len(T.empty_cache), 1)

		SHA256 = hashlib.sha256(b'0')
		self.assertEquals(T.empty_cache[0], SHA256.hexdigest())

if __name__ == '__main__':
	unittest.main()