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
		proof = self.T.generate_proof(index)
		self.assertFalse(proof.proof_type)
		self.assertTrue(self.T.verify_proof(proof)) 

	def test_non_membership_large(self):
		indices = list()

		for i in range(100):
			index = random_index()
			is_member = flip_coin()
			if is_member:
				data = "angela{}".format(i)
				self.T.insert(index, data.encode())
			indices.append((index, is_member))

		for index, is_member in indices:
			proof = self.T.generate_proof(index)
			self.assertEqual(proof.proof_type, is_member)
			self.assertTrue(self.T.verify_proof(proof))
		
	def test_membership_small(self):
		index = bitarray('101').to01()
		self.T.insert(index, b"angela")
		proof = self.T.generate_proof(index)
		self.assertTrue(self.T.verify_proof(proof))

	def test_membership(self):
		index = random_index()
		self.T.insert(index, b"angela")
		proof = self.T.generate_proof(index)
		self.assertEqual(len(proof.copath), 256)
		self.assertTrue(self.T.verify_proof(proof))

	def test_membership_large(self):
		indices = [random_index() for i in range(100)]

		for number, index in enumerate(indices):
			data = "angela{}".format(number)
			self.T.insert(index, data.encode())

		proofs = list()
		for index in indices:
			proofs.append(self.T.generate_proof(index))

		for proof in proofs:
			self.assertTrue(self.T.verify_proof(proof))

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

def flip_coin():
	r = random.random()
	if r > 0.5:
		return True
	return False

if __name__ == '__main__':
	unittest.main()
