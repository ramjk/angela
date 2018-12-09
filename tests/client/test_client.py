import unittest
import threading
import ray
import time

from common import util
from server.transaction import WriteTransaction
from client.client import Client
from server.server import Server

class TestSparseMerkleTree(unittest.TestCase):

	def test_practice(self):
		indices = [util.random_index() for i in range(128)]
		data = [util.random_string() for i in range(128)]

		print("Inserting")
		for i in range(128):
			Client.insert_leaf(indices[i], data[i])

		print("Getting root")
		root = Client.get_signed_root()

		print("Proving")
		for i in range(128):
			proof = Client.generate_proof(indices[i])
			d = data[i]
			self.assertTrue(Client.verify_proof(proof, d, root))

if __name__ == '__main__':
	unittest.main()
