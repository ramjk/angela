import hashlib
import unittest
import random
from merkle.smt import SparseMerkleTree
from common.util import bitarray

class TestSparseMerkleTree(unittest.TestCase):
	def setUp(self):
		self.T = SparseMerkleTree("sha256")

	def test_constructor(self):
		self.assertEqual(self.T.hash_name, "sha256")
		self.assertEqual(self.T.depth, 256)
		self.assertEqual(len(self.T.empty_cache), 1)

		SHA256 = hashlib.sha256(b'0')
		self.assertEqual(self.T.empty_cache[0], SHA256.hexdigest())

	"""
		Recall the CONIKS construction of a Merkle Tree and non-membership proofs on said Tree.
		Given an index j that is not mapped to any value, we are meant to provide an authentication
		path for an index i that has a value mapped to it where i meant to be the longest prefix
		match of j.
	"""
	def test_non_membership(self):
		index = random_index()
		copath = self.T.generate_copath(index)
		self.assertNotEqual(len(copath), 255)
		self.assertTrue(self.T.verify_path(index, copath)) 

	def test_membership(self):
		index = random_index()
		self.T.insert(index, b"angela")
		copath = self.T.generate_copath(index)
		self.assertEqual(len(copath), 255)
		self.assertTrue(self.T.verify_path(index, copath))

def random_index(digest_size: int = 256) -> str:
	bitarr = list()
	for i in range(digest_size):
		r = random.random()
		if r > 0.5:
			bitarr.append(1)
		else:
			bitarr.append(0)
	bitstring = bitarray(bitarr).to01()
	return bitstring


if __name__ == '__main__':
	unittest.main()